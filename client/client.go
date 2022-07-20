package client

import (
	"fmt"

	"spooter/core"

	"golang.org/x/exp/slices"
)

const (
	PatternTypeRegex      = "regex"
	SubscriptionTypeEmail = "email"
	SubscriptionTypeSlack = "slack"
	ModeStandalone        = "standalone"
)

const logPrefix = "client"

var config *core.ConfigClient

var (
	alerter *Alerter
	monitor *Monitor
)

func GetProcesses(cnf *core.ConfigStruct) ([]core.RunningProcessInterface, error) {
	config = &cnf.Client

	core.Logger.Infof(logPrefix, "Check config")
	validationErr := checkConfig()
	if validationErr != nil {
		return nil, fmt.Errorf("InvalidConfig : %s", validationErr)
	}
	core.Logger.Infof(logPrefix, "Config is valid")

	// Create processes
	alerter = &Alerter{}
	monitor = &Monitor{}

	return []core.RunningProcessInterface{
		alerter,
		monitor,
	}, nil
}

//nolint:gocyclo
func checkConfig() error {
	// check metadata
	if config.Metadata.Application == "" {
		return fmt.Errorf("metadata.application is required")
	}

	// check mode
	if config.Mode != ModeStandalone {
		return fmt.Errorf("mode should be \"standalone\"")
	}

	// Standalone mode
	if config.Mode == ModeStandalone {
		// check lookups
		if len(config.Lookups) == 0 {
			return fmt.Errorf("you must provide at least one lookup")
		}
		lookupNames := []string{}
		for i, lookup := range config.Lookups {
			if lookup.Name == "" {
				return fmt.Errorf("lookup #%d : you must provide a name for this lookup", i)
			}
			if slices.Contains(lookupNames, lookup.Name) {
				return fmt.Errorf("lookup #%d : name %s is already used", i, lookup.Name)
			}
			lookupNames = append(lookupNames, lookup.Name)

			// patterns
			if len(lookup.Patterns) == 0 {
				return fmt.Errorf("lookup #%d : you must provide at least one pattern", i)
			}
			patternNames := []string{}
			for j, pattern := range lookup.Patterns {
				if pattern.Name == "" {
					return fmt.Errorf("lookup #%d | pattern #%d : you must provide a name for this pattern", i, j)
				}
				if slices.Contains(patternNames, pattern.Name) {
					return fmt.Errorf("lookup #%d | pattern #%d : name %s is already used", i, j, pattern.Name)
				}
				patternNames = append(patternNames, pattern.Name)
				if pattern.Type != PatternTypeRegex {
					return fmt.Errorf("lookup #%d | pattern #%d : you must provide a type for this pattern", i, j)
				}
				if pattern.Value == "" {
					return fmt.Errorf("lookup #%d | pattern #%d : you must provide a value for this pattern", i, j)
				}
			}

			// files
			if len(lookup.Files) == 0 {
				return fmt.Errorf("lookup #%d : you must provide at least one file to monitor", i)
			}
			for j, file := range lookup.Files {
				if file == "" {
					return fmt.Errorf("lookup #%d | file #%d : you must provide a file to monitor", i, j)
				}
			}
		}

		// Check alerts
		if len(config.Alerts.Subscriptions) == 0 {
			return fmt.Errorf("you must provide at least one alert subscription")
		}
		for i, subscription := range config.Alerts.Subscriptions {
			if subscription.Type != SubscriptionTypeEmail && subscription.Type != SubscriptionTypeSlack {
				return fmt.Errorf("alert subscription #%d : you must provide type \"email\" or \"slack\" (%s given)", i, subscription.Type)
			}
			if subscription.Value == "" {
				return fmt.Errorf("alert subscription #%d : you must provide a value for this subscription", i)
			}
			if len(subscription.Lookups) == 0 {
				return fmt.Errorf("alert subscription #%d : you must provide at least one lookup for this subscription", i)
			}
		}
	}

	return nil
}
