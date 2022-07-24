package main

import (
	"fmt"
	"os"

	"github.com/thomasglachant/spooter/core"
)

const (
	notifierLogPrefix            = "notification"
	SubscriptionTypeEmail        = "email"
	SubscriptionTypeSlackWebhook = "slack_webhook"
)

type Notification interface {
	Subject() string
	TemplateName() string
	Data() map[string]interface{}
}

func SendNotification(notification Notification) error {
	for _, recipient := range agentConfig.Alerts.Recipients {
		if recipient.Type == SubscriptionTypeEmail {
			if err := sendEmail(recipient.Recipient, notification); err != nil {
				return fmt.Errorf("error sending email: %s", err.Error())
			}
		} else if recipient.Type == SubscriptionTypeSlackWebhook {
			if err := sendSlack(recipient.Recipient, notification); err != nil {
				return fmt.Errorf("error sending email: %s", err.Error())
			}
		}
	}
	return nil
}

func sendEmail(email string, notification Notification) error {
	core.Logger.Infof(notifierLogPrefix, "Send an email")

	err := core.SendEmail(
		&agentConfig.SMTP,
		email,
		notification.Subject(),
		notification.TemplateName(),
		notification.Data(),
	)
	if err != nil {
		return err
	}
	return nil
}

func sendSlack(webhookURL string, notification Notification) error {
	// check if template exists
	templateFile := fmt.Sprintf("templates/slack/%s.txt.tmpl", notification.TemplateName())
	if _, err := os.Stat(templateFile); err != nil {
		return fmt.Errorf("template file \"%s\" not found", templateFile)
	}

	core.Logger.Infof(notifierLogPrefix, "Send a slack message")

	err := core.SendSlackNotification(
		webhookURL,
		core.SlackRequest{
			Text: core.TemplateToString(
				[]string{templateFile},
				notification.Data(),
			),
		},
	)
	if err != nil {
		return err
	}
	return nil
}
