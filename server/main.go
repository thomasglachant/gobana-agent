package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/thomasglachant/spooter/core"
)

var logPrefix = "server"

// auto-generated during build
var version string
var date string

func main() {
	fmt.Printf("# spooter-server v%s (%s)\n", version, date)

	var configFile string
	var checkConfig bool
	flag.StringVar(&configFile, "config", "", "Path to config file")
	flag.BoolVar(&checkConfig, "check-config", false, "Check config file is valid")
	flag.Parse()

	//
	// check config special case
	if checkConfig {
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
	if err := core.ReadConfig(configFile, config); err != nil {
		core.Logger.Criticalf(logPrefix, "unable to load agent config : %s", err)
	}

	core.Logger.DebugEnabled = config.Debug

	// start processes
	core.RunProcesses([]core.ProcessInterface{})
}
