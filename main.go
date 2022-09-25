package main

import (
	"embed"
	"flag"
	"fmt"
	"os"
	"time"

	"gobana-agent/agent"
	"gobana-agent/core"
)

const AppName = "gobana-agent"

// Filesystem which contains templates
//
//go:embed templates/**/*
var templateFs embed.FS

// auto-generated during build
var version = "?"
var date = time.Now().Format("2006-01-02")

func main() {
	var configFile string
	var checkConfig bool
	var showVersion bool
	flag.StringVar(&configFile, "config", "config.yaml", "Path to config file")
	flag.BoolVar(&checkConfig, "check-config", false, "Check config file is valid")
	flag.BoolVar(&showVersion, "version", false, "Show version number")
	flag.Parse()

	// version
	if showVersion {
		fmt.Printf("%s v%s (%s)\n", AppName, version, date)
		os.Exit(0)
	}

	core.Logger.Infof("app", "Start %s (v%s)", AppName, version)
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
	agent.AppVersion = version

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
