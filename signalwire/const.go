package signalwire

// SDK consts
const (
	WSTimeOut         = 5
	JSONRPCVer        = "2.0"
	WssHost           = "relay.signalwire.com"
	SDKVersion        = "1.0"
	BladeVersionMajor = 2
	BladeVersionMinor = 3
	BladeRevision     = 0
	MaxSimCalls       = 100
	/* internal */
	BroadcastEventTimeout   = 10 /*seconds*/
	DefaultRingTimeout      = 30 /*seconds*/
	SimActionsOfTheSameKind = 10 /* Buffered Channels - how many simultaneous actions of the same kind we can start */
	EventQueue              = 10 /* Buffered Channels - max unprocessed events per action */
	DefaultActionTimeout    = 30
)
