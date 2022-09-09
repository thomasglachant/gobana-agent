package main

import (
	"embed"
	"flag"
	"fmt"
	"time"

	"spooter-agent/agent"
	"spooter-agent/core"
)

// Filesystem which contains templates
//
//go:embed templates/**/*
var templateFs embed.FS

// Filesystem which contains assets
//
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
		agent.CheckConfig(configFile)
	}

	// setup embedded fs to core
	core.TemplateFs = templateFs
	core.AssetFs = assetFs
	core.Logger.DebugEnabled = agent.AppConfig.Debug
	agent.AppVersion = version

	// start app
	agent.StartAgent(configFile)
}
