package client

import (
	"fmt"
	"sync"
	"time"

	"spooter/core"
)

const (
	alerterTimer     = 10 * time.Second
	alerterDelay     = 100 * time.Millisecond
	alerterLogPrefix = "alerter"
)

type Alert struct {
	Date     time.Time
	Filename string
	Patterns []string
	Content  string
}

type Alerter struct {
	mu            sync.Mutex
	askForStop    bool
	lastExecution time.Time

	pendingAlerts []*Alert
}

func (process *Alerter) Run() error {
	process.askForStop = false
	for !process.askForStop {
		if time.Now().Before(process.lastExecution.Add(alerterTimer)) {
			time.Sleep(alerterDelay)

			continue
		}
		process.lastExecution = time.Now()

		// flush pending alerts
		process.flush()
	}

	return nil
}

func (process *Alerter) HandleStop() {
	process.askForStop = true
}

func (process *Alerter) addAlert(alert *Alert) {
	process.mu.Lock()
	process.pendingAlerts = append(process.pendingAlerts, alert)
	process.mu.Unlock()

	core.Logger.Debugf(alerterLogPrefix, "Add alert to pool")
}

func (process *Alerter) flush() {
	process.mu.Lock()
	alerts := process.pendingAlerts
	process.pendingAlerts = []*Alert{} // clear pending alerts
	process.mu.Unlock()

	if len(alerts) == 0 {
		return
	}

	core.Logger.Infof(alerterLogPrefix, "Flush %d alerts", len(alerts))

	for _, recipient := range config.Alerts.Subscriptions {
		if recipient.Type == SubscriptionTypeEmail {
			core.Logger.Infof(alerterLogPrefix, "Send an email")
			core.SendEmail(
				&config.SMTP,
				recipient.Value,
				fmt.Sprintf("New alert on %s", config.Metadata.Application),
				"alert",
				map[string]interface{}{
					"Metadata": config.Metadata,
					"Alerts":   alerts,
				})
		} else if recipient.Type == SubscriptionTypeSlack {
			core.Logger.Infof(alerterLogPrefix, "Send a slack message")
			core.SendSlackString(
				recipient.Value,
				core.TemplateToString([]string{"templates/slack/alert.txt.tmpl"}, map[string]interface{}{
					"Metadata": config.Metadata,
					"Alerts":   alerts,
				}))
		}
	}
}
