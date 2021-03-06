package signalwire

// SDK consts
const (
	WSTimeOut              = 5
	JSONRPCVer             = "2.0"
	WssHost                = "relay.signalwire.com"
	SDKVersion             = "1.0.1"
	BladeVersionMajor      = 2
	BladeVersionMinor      = 3
	BladeRevision          = 0
	MaxSimCalls            = 100
	MaxPlay                = 10
	TaskingEndpoint        = "https://relay.signalwire.com/api/relay/rest/tasks"
	UserAgent              = "Go SDK"
	BladeConnectionRetries = -1 // how many times to retry to connect before giving up . -1 keeps trying forever.
	/* internal */
	BroadcastEventTimeout   = 10 /* seconds */
	DefaultRingTimeout      = 30 /* seconds */
	SimActionsOfTheSameKind = 10 /* Buffered Channels - how many simultaneous actions of the same kind we can start */
	EventQueue              = 10 /* Buffered Channels - max unprocessed events per action */
	DefaultActionTimeout    = 30
	HTTPClientTimeout       = 60 /* seconds */
	MaxCallDuration         = 3600 * 4
)
