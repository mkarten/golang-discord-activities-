package main

import (
	"App-Client-Code/constants"
	"App-Client-Code/eventEmitter"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"log"
	"strings"
	"syscall/js"
	"time"
)

const (
	// Opcodes
	HANDSHAKE = iota
	FRAME
	CLOSE
	HELLO
)

var (
	ALLOWED_ORIGINS = []string{
		js.Global().Get("location").Get("origin").String(),
		"https://discord.com",
		"https://discordapp.com",
		"https://ptb.discord.com",
		"https://ptb.discordapp.com",
		"https://canary.discord.com",
		"https://canary.discordapp.com",
		"https://staging.discord.co",
		"http://localhost:3333",
		"https://pax.discord.com",
		"null",
	}
)

type Configuration struct {
	disableConsoleLogOverride bool
}

func GetDefaultConfiguration() Configuration {
	return Configuration{disableConsoleLogOverride: false}
}

type DiscordSdk struct {
	channelId       string
	clientId        string
	frameId         string
	guildId         string
	instanceId      string
	platform        string
	ready           bool
	configuration   Configuration
	source          js.Value
	sourceOrigin    string
	EventBus        *eventEmitter.EventEmitter
	PendingCommands map[string]func(js.Value, error)
}

func (sdk *DiscordSdk) String() string {
	return fmt.Sprintf("DiscordSdk{channelId: %s, clientId: %s, frameId: %s, guildId: %s, instanceId: %s, platform: %s, ready: %t, configuration: %v, source: %v, sourceOrigin: %s}",
		sdk.channelId, sdk.clientId, sdk.frameId, sdk.guildId, sdk.instanceId, sdk.platform, sdk.ready, sdk.configuration, sdk.source, sdk.sourceOrigin)

}

func defaultErrorHandler(msg js.Value, err error) {
	if err != nil {
		// call console.error
		js.Global().Get("console").Call("error", msg.Get("message").String())
	}
}

func NewDiscordSdk(clientId string, config Configuration) *DiscordSdk {
	tempSdk := new(DiscordSdk)

	// setup the eventBus
	tempSdk.EventBus = eventEmitter.NewEventEmitter()

	// set source and sourceOrigin to default values
	tempSdk.source = js.Null()
	tempSdk.sourceOrigin = ""
	tempSdk.PendingCommands = make(map[string]func(js.Value, error))

	// setup the pendingCommandMap

	tempSdk.ready = false
	tempSdk.clientId = clientId
	if config == (Configuration{}) {
		tempSdk.configuration = GetDefaultConfiguration()
	} else {
		tempSdk.configuration = config
	}
	// add the message event listener
	if !js.Global().Get("window").IsUndefined() {
		addEventListener(js.Global().Get("window"), "message", tempSdk.onMessage)
	} else {
		// return to an error state
		tempSdk.frameId = ""
		tempSdk.instanceId = ""
		tempSdk.platform = constants.DESKTOP
		tempSdk.guildId = ""
		tempSdk.channelId = ""
		return tempSdk
	}
	// get all the iframe parameters
	tempSdk.frameId = GetParam("frame_id")
	if tempSdk.frameId == "" {
		log.Fatal("Missing frame_id")
	}

	tempSdk.instanceId = GetParam("instance_id")
	if tempSdk.instanceId == "" {
		log.Fatal("Missing instance_id")
	}

	tempSdk.platform = GetParam("platform")
	if tempSdk.platform == "" {
		log.Fatal("Missing platform")
	}

	if tempSdk.platform != constants.DESKTOP && tempSdk.platform != constants.MOBILE {
		log.Fatal("Invalid platform", tempSdk.platform)
	}

	tempSdk.channelId = GetParam("channel_id")
	tempSdk.guildId = GetParam("guild_id")

	tempSdk.source, tempSdk.sourceOrigin = tempSdk.getRPCServerSource()
	tempSdk.addOnReadyListener()
	tempSdk.handshake()
	return tempSdk
}

// GetRPCServerSource returns the source and the sourceOrigin of the RPC server
func (sdk *DiscordSdk) getRPCServerSource() (js.Value, string) {
	var a js.Value
	parent := js.Global().Get("window").Get("parent")
	opener := parent.Get("opener")
	if !opener.IsNull() && !opener.IsUndefined() {
		a = opener
	} else {
		a = parent
	}
	referrer := js.Global().Get("document").Get("referrer")
	if referrer.Truthy() {
		return a, referrer.String()
	}
	return a, "*"
}

// GetTransfer returns the transfer object from the payload
func (sdk *DiscordSdk) getTransfer(payload js.Value) js.Value {
	switch payload.Get("cmd").String() {
	case constants.SUBSCRIBE, constants.UNSUBSCRIBE:
		return js.Undefined()
	default:
		transfer := payload.Get("transfer")
		if !transfer.IsNull() && !transfer.IsUndefined() {
			return transfer
		}
		return js.Undefined()
	}
}

// SendCommand sends a command to the Discord RPC server when the command is answered the callback function is called with the response and an non-nil error if there was an error
func (sdk *DiscordSdk) SendCommand(payload map[string]interface{}, callback func(js.Value, error)) {
	if sdk.source.IsNull() {
		log.Fatal("Attempting to send message before initialization")
	}
	nonce := uuid.New().String()
	payload["nonce"] = nonce
	sdk.source.Call("postMessage", []interface{}{FRAME, payload}, sdk.sourceOrigin, js.Undefined())
	sdk.PendingCommands[nonce] = callback
}

func (sdk *DiscordSdk) addOnReadyListener() {
	sdk.EventBus.Once(constants.READY, func(args ...interface{}) {
		sdk.ready = true
	}, nil)
}

// Ready is a blocking call that waits for the SDK to be ready and then executes the provided function
// it is prefreable to call this function with a goroutine to prevent blocking the main thread
func (sdk *DiscordSdk) Ready(fn func(...interface{}), context interface{}) {
	for !sdk.ready {
		time.Sleep(100 * time.Millisecond)
	}
	fn(context)
}

func (sdk *DiscordSdk) handshake() {
	a := sdk.source
	if a.IsNull() || a.IsUndefined() {
		log.Fatal("Missing source")
	}
	encoding := "json"
	a.Call("postMessage", []interface{}{
		HANDSHAKE,
		map[string]interface{}{
			"v":         1,
			"encoding":  encoding,
			"client_id": sdk.clientId,
			"frame_id":  sdk.frameId,
		}}, sdk.sourceOrigin,
	)
}

func (sdk *DiscordSdk) onMessage(event js.Value) {
	origin := event.Get("origin").String()
	if !contains(ALLOWED_ORIGINS, origin) {
		return
	}
	ev := event.Get("data")
	// check if the event is an array
	if ev.Type() != js.TypeObject {
		return
	}
	// check any of the opcode or data is undefined or null
	if ev.Index(0).IsUndefined() || ev.Index(0).IsNull() || ev.Index(1).IsUndefined() || ev.Index(1).IsNull() {
		return
	}
	opcode, data := ev.Index(0).Int(), ev.Index(1)
	switch opcode {
	case HELLO:
		// backwards compatibility for older applications
		break
	case CLOSE:
		go sdk.handleClose(data)
		break
	case HANDSHAKE:
		go sdk.handleHandshake()
		break
	case FRAME:
		go sdk.handleFrame(data)
		break
	default:
		log.Fatal("Invalid opcode", opcode)
	}
}

func (sdk *DiscordSdk) handleClose(payload js.Value) {
	fmt.Println("Handle Close")
}

func (sdk *DiscordSdk) handleHandshake() {}

func (sdk *DiscordSdk) Subscribe(event string, fn func(...interface{}), context interface{}, args ...interface{}) {
	listenerCount := sdk.EventBus.ListenerCount(event)
	sdk.EventBus.On(event, fn, context)
	// if this is the first suscriber to the event, send the subscribe command to the RPC server
	if listenerCount == 0 && event != constants.READY && contains(constants.Events, event) {
		arguments := make([]interface{}, len(args))
		payload := map[string]interface{}{
			"cmd":  constants.SUBSCRIBE,
			"args": arguments,
			"evt":  event,
		}
		sdk.SendCommand(payload, func(msg js.Value, err error) { defaultErrorHandler(msg, err) })
	}
}

type SDKError struct {
	Code    int
	Message string
	Name    string
}

func (sdk *DiscordSdk) handleFrame(payload js.Value) {
	fmt.Println("Handle Frame")
	incomingPayload := parseIncomingPayload(payload)
	// check if the payload cmd is a dispatch
	if incomingPayload.Cmd == constants.DISPATCH {
		// dispatch the event and the data in the event bus
		fmt.Println("Emitting", incomingPayload.Evt)
		sdk.EventBus.Emit(incomingPayload.Evt, incomingPayload.Data)
	} else {
		if incomingPayload.Evt == constants.ERROR {
			// in response to a command
			if incomingPayload.Nonce != "" {
				// check if the nonce is in the pendingCommands map
				if _, ok := sdk.PendingCommands[incomingPayload.Nonce]; ok {
					// call the callback function with the error
					sdk.PendingCommands[incomingPayload.Nonce](incomingPayload.Data, errors.New("error"))
					// remove the command from the pendingCommands map
					delete(sdk.PendingCommands, incomingPayload.Nonce)
				}
				// general error
				sdk.EventBus.Emit("error", SDKError{
					Code:    incomingPayload.Data.Get("code").Int(),
					Message: incomingPayload.Data.Get("message").String(),
					Name:    "Discord SDK Error",
				})
			}
		}
	}
	if incomingPayload.Nonce == "" {
		log.Fatal("Missing nonce")
	}
	if _, ok := sdk.PendingCommands[incomingPayload.Nonce]; ok {
		// call the callback function with the response
		sdk.PendingCommands[incomingPayload.Nonce](incomingPayload.Data, nil)
		// remove the command from the pendingCommands map
		delete(sdk.PendingCommands, incomingPayload.Nonce)
	}
}

type IncomingPayload struct {
	Evt   string   `json:"evt"`
	Nonce string   `json:"nonce"`
	Data  js.Value `json:"data"`
	Cmd   string   `json:"cmd"`
}

func parseIncomingPayload(payload js.Value) IncomingPayload {
	if !payload.Get("evt").IsNull() {
		// check if the event is an error
		if payload.Get("evt").String() == constants.ERROR {
			return parseErrorPayload(payload)
		}
		// if the event is not an error, parse the payload as an event Payload
		return parseEventPayload(payload)
	} else {
		// if the event is null, parse the payload as a Response Payload
		return parseResponsePayload(payload)
	}
}

func parseErrorPayload(payload js.Value) IncomingPayload {
	nonce := payload.Get("nonce").String()
	data := payload.Get("data")
	cmd := payload.Get("cmd").String()
	return IncomingPayload{
		Evt:   constants.ERROR,
		Nonce: nonce,
		Data:  data,
		Cmd:   cmd,
	}

}

func parseEventPayload(payload js.Value) IncomingPayload {
	evt := payload.Get("evt").String()
	if !contains(constants.Events, evt) {
		log.Fatal("Invalid event", evt)
	}
	nonce := payload.Get("nonce").String()
	data := payload.Get("data")
	cmd := payload.Get("cmd").String()
	return IncomingPayload{
		Evt:   evt,
		Nonce: nonce,
		Data:  data,
		Cmd:   cmd,
	}
}

func parseResponsePayload(payload js.Value) IncomingPayload {
	nonce := payload.Get("nonce").String()
	data := payload.Get("data")
	cmd := payload.Get("cmd").String()
	// check if it is a known command
	if !contains(constants.Commands, cmd) {
		log.Fatal("Unknown command", cmd)
	}
	return IncomingPayload{
		Evt:   "",
		Nonce: nonce,
		Data:  data,
		Cmd:   cmd,
	}

}

func contains[c comparable](arr []c, str c) bool {
	for _, a := range arr {
		if a == str {
			return true
		}
	}
	return false
}

func addEventListener(location js.Value, event string, callback func(js.Value)) {
	location.Call("addEventListener", event, js.FuncOf(func(this js.Value, p []js.Value) interface{} {
		callback(p[0])
		return nil
	}))
}

func GetParams() map[string]string {
	params := make(map[string]string)
	windowLocation := js.Global().Get("location").Get("search").String()
	if len(windowLocation) > 0 {
		windowLocation = windowLocation[1:]
		pairs := strings.Split(windowLocation, "&")
		for _, pair := range pairs {
			kv := strings.Split(pair, "=")
			if len(kv) == 2 {
				params[kv[0]] = kv[1]
			}
		}
	}
	return params
}

func GetParam(key string) string {
	return GetParams()[key]
}

func main() {
	sdk := NewDiscordSdk("1219594199188377651", Configuration{})
	fmt.Println("Discord SDK", sdk)

	// display a discord sdk information on the web page
	//js.Global().Get("document").Call("write", fmt.Sprintf("<h1>%s</h1>", sdk.String()))

	// wait for the SDK to be ready
	sdk.Ready(func(args ...interface{}) {
		fmt.Println("SDK is ready")
	}, nil)

	// subscribe to the VOICE_STATE_UPDATE event
	sdk.Subscribe(constants.VOICE_STATE_UPDATE, func(args ...interface{}) {
		fmt.Println("Voice State Update", args)
	}, nil)

	//prevent the program from exiting
	select {}
}
