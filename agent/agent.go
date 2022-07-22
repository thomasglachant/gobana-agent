package agent

import (
	"spooter/core"
)

var (
	alerter *Alerter
	watcher *Watcher
)

func GetProcesses() ([]core.RunningProcessInterface, error) {
	// Create processes
	alerter = &Alerter{}
	watcher = &Watcher{}

	return []core.RunningProcessInterface{
		watcher,
		alerter,
	}, nil
}
