package agent

var (
	AppConfig = &AgentConfig{}

	// services
	Alerter *AlerterProcess
	Watcher *WatcherProcess
)

var AppVersion = "?"
