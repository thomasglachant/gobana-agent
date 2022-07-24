package main

import (
	"embed"
	"flag"

	"github.com/thomasglachant/spooter/core"
)

var (
	logPrefix = "agent"

	alerter *Alerter
	watcher *Watcher
)

// Filesystem which contains templates
//go:embed templates/**/*
var templateFs embed.FS

// Filesystem which contains assets
//go:embed assets/**/*
var assetFs embed.FS

func main() {
	var configFile string
	flag.StringVar(&configFile, "config", "", "Path to config file")

	// parse command line
	flag.Parse()

	// load config
	core.Logger.Infof(logPrefix, "load config from %s", configFile)
	agentConfig = &AgentConfig{}
	if err := core.ReadConfig(configFile, agentConfig); err != nil {
		core.Logger.Criticalf(logPrefix, "unable to load config : %s", err)
	}

	// setup embedded fs to core
	core.TemplateFs = templateFs
	core.AssetFs = assetFs
	core.Logger.DebugEnabled = agentConfig.Debug

	// start processes
	alerter = &Alerter{}
	watcher = &Watcher{}
	core.RunProcesses([]core.ProcessInterface{
		&core.ProcessStruct{RunningProcess: watcher},
		&core.ProcessStruct{RunningProcess: alerter},
	})
}