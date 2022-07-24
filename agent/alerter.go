package main

import (
	"fmt"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/thomasglachant/spooter/core"
)

const (
	alerterLogPrefix = "alerter"
	alerterTimer     = 10 * time.Second

	TriggerTypeRegex        = "match_regex"
	TriggerTypeEqual        = "is"
	TriggerTypeNotEqual     = "is_not"
	TriggerTypeContains     = "contains"
	TriggerTypeNotContains  = "not_contains"
	TriggerTypeStartWith    = "start_with"
	TriggerTypeNotStartWith = "not_start_with"
)

type Alert struct {
	Date        time.Time
	Filename    string
	ParserName  string
	TriggerName string
	Content     string
}

type Alerts []*Alert

func (alerts Alerts) Subject() string {
	return fmt.Sprintf("%d alert(s) on %s", len(alerts), agentConfig.Application)
}

func (alerts Alerts) TemplateName() string {
	return "alert"
}

func (alerts Alerts) Data() map[string]interface{} {
	return map[string]interface{}{
		"Alerts": alerts,
	}
}

type Alerter struct {
	mu       sync.Mutex
	exitChan chan bool

	pendingAlerts Alerts
}

func (alerter *Alerter) Run() error {
	alerter.exitChan = make(chan bool)
	subscriptionID := core.EventDispatcher.Subscribe(core.EventDescription{
		Name:     eventNameLogDiscover,
		Priority: 0,
		Callback: HandleParserTrigger,
	})
	defer core.EventDispatcher.Unsubscribe(subscriptionID)

	core.ProcessInfiniteLoop(alerterTimer, alerter.exitChan, func() {
		// flush pending alerts
		alerter.flush()
	})

	return nil
}

func (alerter *Alerter) HandleStop() {
	alerter.exitChan <- true
}

func (alerter *Alerter) addAlert(alert *Alert) {
	alerter.mu.Lock()
	alerter.pendingAlerts = append(alerter.pendingAlerts, alert)
	alerter.mu.Unlock()

	core.Logger.Debugf(alerterLogPrefix, "Add alert to pool")
}

func (alerter *Alerter) flush() {
	if len(alerter.pendingAlerts) == 0 {
		return
	}

	core.Logger.Debugf(alerterLogPrefix, "Flush pending alerts")

	alerter.mu.Lock()
	err := SendNotification(&alerter.pendingAlerts)
	if err != nil {
		core.Logger.Errorf(alerterLogPrefix, "Error sending notification: %s", err)
	}
	alerter.pendingAlerts = Alerts{}
	alerter.mu.Unlock()
}

func HandleParserTrigger(data core.EventData) {
	line := data["logLine"].(*LogLine)

	for _, trigger := range agentConfig.Alerts.Triggers {
		go func(trigger TriggerConfig) {
			allFieldsMatch := true
			for _, triggerValue := range trigger.Values {
				fieldValue := ""
				switch {
				case triggerValue.Field == "_parser":
					fieldValue = line.Metadata.Parser
				case triggerValue.Field == "_filename":
					fieldValue = line.Metadata.Filename
				default:
					if _, ok := line.Fields[triggerValue.Field]; !ok {
						core.Logger.Errorf(alerterLogPrefix, "unable to check field value (field \"%s\" not exists)", triggerValue.Field)
						continue
					}
					fieldValue = line.Fields[triggerValue.Field]
				}

				match, err := checkTriggerValueMatch(fieldValue, triggerValue.Operator, triggerValue.Value)
				if err != nil {
					core.Logger.Errorf(alerterLogPrefix, "unable to check field value : %s", err)
					continue
				}
				if !match {
					allFieldsMatch = false
					break
				}
			}

			if allFieldsMatch {
				core.Logger.Infof(alerterLogPrefix, "Line match with trigger \"%s\"", trigger.Name)
				alerter.addAlert(&Alert{
					Date:        time.Now(),
					Filename:    line.Metadata.Filename,
					ParserName:  line.Metadata.Parser,
					TriggerName: trigger.Name,
					Content:     line.Raw,
				})
			}
		}(trigger)
	}
}

func checkTriggerValueMatch(fieldValue, operator, operatorValue string) (bool, error) {
	lowerFieldValue := strings.ToLower(fieldValue)
	lowerOperatorValue := strings.ToLower(operatorValue)

	switch operator {
	case TriggerTypeRegex:
		r := regexp.MustCompile(operatorValue)
		if r.MatchString(fieldValue) {
			return true, nil
		}
	case TriggerTypeEqual:
		if lowerFieldValue == lowerOperatorValue {
			return true, nil
		}
	case TriggerTypeNotEqual:
		if lowerFieldValue != lowerOperatorValue {
			return true, nil
		}
	case TriggerTypeContains:
		if strings.Contains(lowerFieldValue, lowerOperatorValue) {
			return true, nil
		}
	case TriggerTypeNotContains:
		if !strings.Contains(lowerFieldValue, lowerOperatorValue) {
			return true, nil
		}
	case TriggerTypeStartWith:
		if lowerFieldValue[0:len(operatorValue)] == lowerOperatorValue {
			return true, nil
		}
	case TriggerTypeNotStartWith:
		if lowerFieldValue[0:len(operatorValue)] != lowerOperatorValue {
			return true, nil
		}
	default:
		return false, fmt.Errorf("unknown trigger operator: %s", operator)
	}
	return false, nil
}
