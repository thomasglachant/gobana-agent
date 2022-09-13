package agent

var (
	AppConfig = &AgentConfig{}

	// services
	Alerter *AlerterProcess
	Watcher *WatcherProcess
	Emitter *EmitterProcess
)

var AppVersion = "?"
