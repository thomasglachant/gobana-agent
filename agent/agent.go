package agent

import "spooter-agent/core"

var (
	logPrefix = "agent"

	AppConfig = &AgentConfig{}

	// services
	alerter *Alerter
	watcher *Watcher
	emitter *Emitter
)

var AppVersion = "?"

func StartAgent(configFile string) {
	core.Logger.Infof(logPrefix, "load AppConfig from %s", configFile)
	if err := core.ReadConfig(configFile, AppConfig); err != nil {
		core.Logger.Criticalf(logPrefix, "unable to load agent AppConfig : %s", err)
	}

	// start processes
	alerter = &Alerter{}
	watcher = &Watcher{}
	processes := []core.ProcessInterface{
		&core.ProcessStruct{RunningProcess: watcher},
		&core.ProcessStruct{RunningProcess: alerter},
	}

	if AppConfig.Emitter.Enabled {
		emitter = &Emitter{}
		processes = append(processes, &core.ProcessStruct{RunningProcess: emitter})
	}

	core.RunProcesses(processes)
}
