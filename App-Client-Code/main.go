package main

import (
	"App-Client-Code/constants"
	"fmt"
	"log"
	"strings"
	"syscall/js"
)

const (
	HANDSHAKE = 0
	FRAME     = 1
	CLOSE     = 2
	HELLO     = 3
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
	channelId     string
	clientId      string
	frameId       string
	guildId       string
	instanceId    string
	platform      string
	ready         bool
	configuration Configuration
	source        js.Value
	sourceOrigin  string
}

func NewDiscordSdk(clientId string, config Configuration) DiscordSdk {
	tempSdk := DiscordSdk{}

	// setup the eventBus

	// set source and sourceOrigin to default values
	tempSdk.source = js.Null()
	tempSdk.sourceOrigin = ""

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

func (sdk *DiscordSdk) addOnReadyListener() {
	fmt.Println("Add on ready listener")
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
	fmt.Println("Handshake sent")
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
	opcode, data := ev.Index(0).Int(), ev.Index(1)
	switch opcode {
	case HELLO:
		// backwards compatibility for older applications
		break
	case CLOSE:
		sdk.handleClose(data)
		break
	case HANDSHAKE:
		sdk.handleHandshake()
		break
	case FRAME:
		sdk.handleFrame(data)
		break
	default:
		log.Fatal("Invalid opcode", opcode)
	}
}

func (sdk *DiscordSdk) handleClose(payload js.Value) {
	fmt.Println("Handle Close")
}

func (sdk *DiscordSdk) handleHandshake() {}

func (sdk *DiscordSdk) handleFrame(payload js.Value) {
	fmt.Println("Handle Frame")
}

func contains(arr []string, str string) bool {
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
	//prevent the program from exiting
	select {}
}
