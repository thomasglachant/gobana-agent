package main

import (
	"embed"
	"flag"
	"fmt"
	"gobana-agent/agent"
	"gobana-agent/core"
	"os"
)

const AppName = "gobana-agent"

var Version = "?"
var Date = "-"

// Filesystem which contains templates
//
//go:embed templates/**/*
var templateFs embed.FS

func main() {
	var configFile string
	var checkConfig bool
	var showVersion bool
	flag.StringVar(&configFile, "config", "config.yaml", "Path to config file")
	flag.BoolVar(&checkConfig, "check-config", false, "Check config file is valid")
	flag.BoolVar(&showVersion, "version", false, "Show Version number")
	flag.Parse()

	// Version
	if showVersion {
		fmt.Printf("%s %s (%s)\n", AppName, Version, Date)
		os.Exit(0)
	}

	core.Logger.Infof("app", "Start %s (%s)", AppName, Version)
	defer core.Logger.Infof("app", "Exit %s", AppName)

	//
	// check config special case
	if checkConfig {
		agent.CheckConfig(configFile)
	}

	// setup embedded fs to core
	core.TemplateFs = templateFs
	// setup vars
	core.Logger.DebugEnabled = agent.AppConfig.Debug
	agent.AppVersion = Version

	// parse config
	core.Logger.Infof("config", "load config from %s", configFile)
	if err := core.ReadConfig(configFile, agent.AppConfig); err != nil {
		core.Logger.Criticalf("config", "unable to load agent config : %s", err)
	}

	// start processes
	agent.Alerter = &agent.AlerterProcess{}
	agent.Watcher = &agent.WatcherProcess{}
	processes := []core.ProcessInterface{
		&core.ProcessStruct{RunningProcess: agent.Watcher},
		&core.ProcessStruct{RunningProcess: agent.Alerter},
	}

	if agent.AppConfig.Emitter.Enabled {
		agent.Emitter = &agent.EmitterProcess{}
		processes = append(processes, &core.ProcessStruct{RunningProcess: agent.Emitter})
	}

	core.RunProcesses(processes)
}
