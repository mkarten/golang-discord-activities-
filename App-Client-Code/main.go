package main

import (
	"App-Client-Code/constants"
	"fmt"
	"log"
	"strings"
	"syscall/js"
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

	tempSdk.ready = false
	tempSdk.clientId = clientId
	if config == (Configuration{}) {
		tempSdk.configuration = GetDefaultConfiguration()
	} else {
		tempSdk.configuration = config
	}
	// add the message event listener

	tempSdk.sourceOrigin = "*"

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

	tempSdk.source, tempSdk.sourceOrigin = getRPCServerSource()

	return tempSdk
}

func getRPCServerSource() (js.Value, string) {
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
}
