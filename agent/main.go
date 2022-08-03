package main

import (
	"embed"
	"flag"
	"fmt"
	"github.com/thomasglachant/spooter/core"
	"os"
	"time"
)

var (
	logPrefix = "agent"

	// config
	config *AgentConfig

	// services
	alerter *Alerter
	watcher *Watcher
	emitter *Emitter
)

// Filesystem which contains templates
//go:embed templates/**/*
var templateFs embed.FS

// Filesystem which contains assets
//go:embed assets/**/*
var assetFs embed.FS

// auto-generated during build
var version = "?"
var date = time.Now().Format("2006-01-02")

func main() {
	fmt.Printf("# spooter-agent v%s (%s)\n", version, date)

	var configFile string
	var checkConfig bool
	flag.StringVar(&configFile, "config", "", "Path to config file")
	flag.BoolVar(&checkConfig, "check-config", false, "Check config file is valid")
	flag.Parse()

	//
	// check config special case
	if checkConfig {
		config = &AgentConfig{}
		if err := core.ReadConfig(configFile, config); err != nil {
			fmt.Printf("Invalid config file : %s\n", err)
			os.Exit(1)
		}
		fmt.Printf("Config file is valid\n")
		os.Exit(0)
	}

	//
	// start agent
	core.Logger.Infof(logPrefix, "load config from %s", configFile)
	config = &AgentConfig{}
	if err := core.ReadConfig(configFile, config); err != nil {
		core.Logger.Criticalf(logPrefix, "unable to load agent config : %s", err)
	}

	// setup embedded fs to core
	core.TemplateFs = templateFs
	core.AssetFs = assetFs
	core.Logger.DebugEnabled = config.Debug

	// start processes
	alerter = &Alerter{}
	watcher = &Watcher{}
	processes := []core.ProcessInterface{
		&core.ProcessStruct{RunningProcess: watcher},
		&core.ProcessStruct{RunningProcess: alerter},
	}

	if config.Emitter.Enabled {
		emitter = &Emitter{}
		processes = append(processes, &core.ProcessStruct{RunningProcess: emitter})
	}

	core.RunProcesses(processes)
}
