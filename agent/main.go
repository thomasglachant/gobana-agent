package main

import (
	"embed"
	"flag"
	"fmt"
	"os"

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
	var checkConfig bool
	flag.StringVar(&configFile, "config", "", "Path to config file")
	flag.BoolVar(&checkConfig, "check-config", false, "Check config file is valid")
	flag.Parse()

	//
	// check config special case
	if checkConfig {
		if err := loadConfig(configFile); err != nil {
			fmt.Printf("Invalid config file : %s\n", err)
			os.Exit(1)
		}
		fmt.Printf("Config file is valid\n")
		os.Exit(0)
	}

	//
	// start agent
	core.Logger.Infof(logPrefix, "load config from %s", configFile)
	if err := loadConfig(configFile); err != nil {
		core.Logger.Criticalf(logPrefix, "unable to load agent config : %s", err)
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
