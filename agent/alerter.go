package agent

import (
	"fmt"
	"regexp"
	"strings"
	"sync"
	"time"

	"gobana-agent/core"
)

const (
	alerterLogPrefix = "alerter"

	triggerTypeRegex        = "match_regex"
	triggerTypeEqual        = "is"
	triggerTypeNotEqual     = "is_not"
	triggerTypeContains     = "contains"
	triggerTypeNotContains  = "not_contains"
	triggerTypeStartWith    = "start_with"
	triggerTypeNotStartWith = "not_start_with"
)

type Alert struct {
	Date        time.Time
	Application string
	Server      string
	Filename    string
	ParserName  string
	TriggerName string
	Fields      map[string]string
	Raw         string
}

type Alerts []*Alert

func (alerts Alerts) Subject() string {
	return fmt.Sprintf("%s - %d new alert(s)", AppConfig.Application, len(alerts))
}

func (alerts Alerts) TemplateName() string {
	return "alert"
}

func (alerts Alerts) Data() map[string]interface{} {
	return map[string]interface{}{
		"Alerts": alerts,
	}
}

type AlerterProcess struct {
	mu       sync.Mutex
	exitChan chan bool

	alertBuffer Alerts
}

func (alerter *AlerterProcess) Name() string {
	return alerterLogPrefix
}

func (alerter *AlerterProcess) Run() error {
	alerter.exitChan = make(chan bool)

	// subscribe events
	subscriptionID := core.EventDispatcher.Subscribe(core.EventDescription{
		Name:     eventNameEntryDiscover,
		Priority: 0,
		Callback: HandleParserTrigger,
	})
	defer core.EventDispatcher.Unsubscribe(subscriptionID)

	core.ProcessInfiniteLoop(time.Duration(AppConfig.Alerts.Frequency)*time.Second, alerter.exitChan, func() {
		// flush pending Alerts
		alerter.flush()
	})
	// execute last flush before exiting
	alerter.flush()

	return nil
}

func (alerter *AlerterProcess) HandleStop() {
	alerter.exitChan <- true
}

func (alerter *AlerterProcess) addAlert(alert *Alert) {
	alerter.mu.Lock()
	alerter.alertBuffer = append(alerter.alertBuffer, alert)
	alerter.mu.Unlock()

	core.Logger.Debugf(alerter.Name(), "Add alert to pool")
}

func (alerter *AlerterProcess) flush() {
	if len(alerter.alertBuffer) == 0 {
		return
	}

	core.Logger.Debugf(alerterLogPrefix, "Flush %d pending alerts", len(alerter.alertBuffer))

	alerter.mu.Lock()
	alerts := alerter.alertBuffer
	alerter.alertBuffer = Alerts{}
	alerter.mu.Unlock()

	err := SendNotification(alerts)
	if err != nil {
		core.Logger.Errorf(alerterLogPrefix, "Error sending notification: %s", err)
	}
}

func HandleParserTrigger(entryObj interface{}) {
	entry := entryObj.(*core.Entry)

	for _, trigger := range AppConfig.Alerts.Triggers {
		go func(trigger TriggerConfigStruct) {
			allFieldsMatch := true
			for _, triggerValue := range trigger.Values {
				fieldValue := ""
				switch {
				case triggerValue.Field == "_parser":
					fieldValue = entry.Metadata.Parser
				case triggerValue.Field == "_filename":
					fieldValue = entry.Metadata.Filename
				default:
					if _, ok := entry.Fields[triggerValue.Field]; ok {
						fieldValue = entry.Fields[triggerValue.Field]
					} else {
						core.Logger.Errorf(alerterLogPrefix, "unable to check field value (field \"%s\" not exists)", triggerValue.Field)
					}
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
				core.Logger.Debugf(alerterLogPrefix, "Line match with trigger \"%s\"", trigger.Name)
				Alerter.addAlert(&Alert{
					Date:        time.Now(),
					Application: AppConfig.Application,
					Server:      AppConfig.Server,
					Filename:    entry.Metadata.Filename,
					ParserName:  entry.Metadata.Parser,
					TriggerName: trigger.Name,
					Fields:      entry.Fields,
					Raw:         entry.Raw,
				})
			}
		}(trigger)
	}
}

//nolint:gocyclo
func checkTriggerValueMatch(fieldValue, operator, operatorValue string) (bool, error) {
	// special case : ignore empty values
	if fieldValue == "" {
		return false, nil
	}

	lowerFieldValue := strings.ToLower(fieldValue)
	lowerOperatorValue := strings.ToLower(operatorValue)

	switch operator {
	case triggerTypeRegex:
		r := regexp.MustCompile(operatorValue)
		if r.MatchString(fieldValue) {
			return true, nil
		}
	case triggerTypeEqual:
		if lowerFieldValue == lowerOperatorValue {
			return true, nil
		}
	case triggerTypeNotEqual:
		if lowerFieldValue != lowerOperatorValue {
			return true, nil
		}
	case triggerTypeContains:
		if strings.Contains(lowerFieldValue, lowerOperatorValue) {
			return true, nil
		}
	case triggerTypeNotContains:
		if !strings.Contains(lowerFieldValue, lowerOperatorValue) {
			return true, nil
		}
	case triggerTypeStartWith:
		if lowerFieldValue[0:len(operatorValue)] == lowerOperatorValue {
			return true, nil
		}
	case triggerTypeNotStartWith:
		if lowerFieldValue[0:len(operatorValue)] != lowerOperatorValue {
			return true, nil
		}
	default:
		return false, fmt.Errorf("unknown trigger operator: %s", operator)
	}
	return false, nil
}
