package main

import (
	"embed"
	"flag"
	"fmt"
	"os"

	"gobana-agent/agent"
	"gobana-agent/core"
)

const AppName = "Gobana Agent"

var (
	version = "?"
	commit  = "-"
)

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
		fmt.Printf("%s version %s, build %s\n", AppName, version, commit)
		os.Exit(0)
	}

	core.Logger.Infof("app", "Start %s version %s", AppName, version)
	defer core.Logger.Infof("app", "Exit %s", AppName)

	//
	// check config special case
	if checkConfig {
		agent.CheckConfig(configFile)
	}

	// parse config
	core.Logger.Infof("config", "load config from %s", configFile)
	if err := core.ReadConfig(configFile, agent.AppConfig); err != nil {
		core.Logger.Criticalf("config", "unable to load agent config : %s", err)
	}

	// setup embedded fs to core
	core.TemplateFs = templateFs
	// setup vars
	core.Logger.DebugEnabled = agent.AppConfig.Debug
	agent.AppVersion = version

	// start processes
	agent.Alerter = &agent.AlerterProcess{}
	agent.Watcher = &agent.WatcherProcess{}
	processes := []core.ProcessInterface{
		&core.ProcessStruct{RunningProcess: agent.Watcher},
		&core.ProcessStruct{RunningProcess: agent.Alerter},
	}

	core.RunProcesses(processes)
}
