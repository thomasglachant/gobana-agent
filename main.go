package main

import (
	"embed"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"spooter/agent"

	yaml "gopkg.in/yaml.v3"

	"spooter/core"
)

const logPrefix = "main"

// auto-generated during build
var version string
var date string

// Filesystem which contains templates
//go:embed templates/**/*
var templateFs embed.FS

// Filesystem which contains assets
//go:embed assets/**/*
var assetFs embed.FS

//nolint:funlen
func main() {
	// setup embedded fs to core
	core.TemplateFs = templateFs
	core.AssetFs = assetFs

	fmt.Printf("# spooter v%s (%s)\n", version, date)

	var agentFlag bool
	flag.BoolVar(&agentFlag, "agent", false, "Enable/disable agent mode")

	var configFile string
	flag.StringVar(&configFile, "config", "", "Path to config file")

	// parse command line
	flag.Parse()

	// resolve default config file
	if configFile == "" && agentFlag {
		configFile = "/etc/spooter/agent.yml"
	}

	/*
	   Gen config
	*/
	//
	// load config from file
	//
	core.Logger.Infof(logPrefix, "load config from %s", configFile)
	var readErr error
	data, readErr := os.ReadFile(configFile)
	if readErr != nil {
		core.Logger.Criticalf(logPrefix, "unable to open config file : %s", readErr)
	}
	if err := yaml.Unmarshal(data, &core.AppConfig); err != nil {
		core.Logger.Criticalf(logPrefix, "unable to decode config : %s", err)
	}

	core.Logger.Infof(logPrefix, "Validate config file")
	validationErr := core.CheckConfig(core.AppConfig)
	if validationErr != nil {
		core.Logger.Criticalf(logPrefix, "Invalid config file : %s", validationErr)
	}

	/*
	   Start App
	*/
	if !agentFlag {
		core.Logger.Criticalf(logPrefix, "You must add \"-agent\" argument")
	}

	// enable services
	processesToRun := []core.ProcessInterface{}
	if agentFlag {
		processes, err := agent.GetProcesses()
		if err != nil {
			core.Logger.Criticalf(logPrefix, "unable to get processes : %s", err)
		}
		for _, p := range processes {
			processesToRun = append(processesToRun, &core.ProcessStruct{RunningProcess: p})
		}
	}

	nbCtrlC := 0
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt)
	signal.Notify(sigChan, syscall.SIGTERM)

	// Catch ctrl+c signal
	go func(processesToRun []core.ProcessInterface) {
		for range sigChan {
			nbCtrlC++

			if nbCtrlC >= 2 {
				core.Logger.Infof(logPrefix, "Force exit !")
				os.Exit(0)
			}
			core.Logger.Infof(logPrefix, "Receive stop signal : kill process properly ... press Ctrl+c again to force kill")

			// Shutdown all running processes
			for _, process := range processesToRun {
				process.Stop()
			}
		}
	}(processesToRun)

	// Start services
	c := make(chan bool)
	for _, process := range processesToRun {
		go func(process core.ProcessInterface) {
			defer func() {
				if r := recover(); r != nil {
					core.Logger.Criticalf(process.GetName(), "Panic occurred : %v", r)
				}
			}()
			process.Start()
			c <- true
		}(process)
	}

	// Wait for all processes exited
	for i := 0; i < len(processesToRun); i++ {
		<-c
	}

	core.Logger.Infof(logPrefix, "All processes exited")
}
