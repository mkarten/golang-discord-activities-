package constants

const (
	DISPATCH                                     = "DISPATCH"
	AUTHORIZE                                    = "AUTHORIZE"
	AUTHENTICATE                                 = "AUTHENTICATE"
	GET_GUILDS                                   = "GET_GUILDS"
	GET_GUILD                                    = "GET_GUILD"
	GET_CHANNEL                                  = "GET_CHANNEL"
	GET_CHANNELS                                 = "GET_CHANNELS"
	SELECT_VOICE_CHANNEL                         = "SELECT_VOICE_CHANNEL"
	SELECT_TEXT_CHANNEL                          = "SELECT_TEXT_CHANNEL"
	SUBSCRIBE                                    = "SUBSCRIBE"
	UNSUBSCRIBE                                  = "UNSUBSCRIBE"
	CAPTURE_SHORTCUT                             = "CAPTURE_SHORTCUT"
	SET_CERTIFIED_DEVICES                        = "SET_CERTIFIED_DEVICES"
	SET_ACTIVITY                                 = "SET_ACTIVITY"
	GET_SKUS                                     = "GET_SKUS"
	GET_ENTITLEMENTS                             = "GET_ENTITLEMENTS"
	GET_SKUS_EMBEDDED                            = "GET_SKUS_EMBEDDED"
	GET_ENTITLEMENTS_EMBEDDED                    = "GET_ENTITLEMENTS_EMBEDDED"
	START_PURCHASE                               = "START_PURCHASE"
	SET_CONFIG                                   = "SET_CONFIG"
	SEND_ANALYTICS_EVENT                         = "SEND_ANALYTICS_EVENT"
	USER_SETTINGS_GET_LOCALE                     = "USER_SETTINGS_GET_LOCALE"
	OPEN_EXTERNAL_LINK                           = "OPEN_EXTERNAL_LINK"
	ENCOURAGE_HW_ACCELERATION                    = "ENCOURAGE_HW_ACCELERATION"
	CAPTURE_LOG                                  = "CAPTURE_LOG"
	SET_ORIENTATION_LOCK_STATE                   = "SET_ORIENTATION_LOCK_STATE"
	OPEN_INVITE_DIALOG                           = "OPEN_INVITE_DIALOG"
	GET_PLATFORM_BEHAVIORS                       = "GET_PLATFORM_BEHAVIORS"
	GET_CHANNEL_PERMISSIONS                      = "GET_CHANNEL_PERMISSIONS"
	OPEN_SHARE_MOMENT_DIALOG                     = "OPEN_SHARE_MOMENT_DIALOG"
	INITIATE_IMAGE_UPLOAD                        = "INITIATE_IMAGE_UPLOAD"
	GET_ACTIVITY_INSTANCE_CONNECTED_PARTICIPANTS = "GET_ACTIVITY_INSTANCE_CONNECTED_PARTICIPANTS"
)

var Commands = []string{
	AUTHORIZE,
	AUTHENTICATE,
	GET_GUILDS,
	GET_GUILD,
	GET_CHANNEL,
	GET_CHANNELS,
	SELECT_VOICE_CHANNEL,
	SELECT_TEXT_CHANNEL,
	SUBSCRIBE,
	UNSUBSCRIBE,
	CAPTURE_SHORTCUT,
	SET_CERTIFIED_DEVICES,
	SET_ACTIVITY,
	GET_SKUS,
	GET_ENTITLEMENTS,
	GET_SKUS_EMBEDDED,
	GET_ENTITLEMENTS_EMBEDDED,
	START_PURCHASE,
	SET_CONFIG,
	SEND_ANALYTICS_EVENT,
	USER_SETTINGS_GET_LOCALE,
	OPEN_EXTERNAL_LINK,
	ENCOURAGE_HW_ACCELERATION,
	CAPTURE_LOG,
	SET_ORIENTATION_LOCK_STATE,
	OPEN_INVITE_DIALOG,
	GET_PLATFORM_BEHAVIORS,
	GET_CHANNEL_PERMISSIONS,
	OPEN_SHARE_MOMENT_DIALOG,
	INITIATE_IMAGE_UPLOAD,
	GET_ACTIVITY_INSTANCE_CONNECTED_PARTICIPANTS,
}
