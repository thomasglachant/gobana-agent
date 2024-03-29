package agent

import (
	"fmt"

	"gobana-agent/core"
)

const (
	notifierLogPrefix            = "notification"
	subscriptionTypeEmail        = "email"
	subscriptionTypeSlackWebhook = "slack_webhook"
)

type Notification interface {
	Subject() string
	TemplateName() string
	Data() map[string]interface{}
}

func SendNotification(notification Notification) error {
	for _, recipient := range AppConfig.Alerts.Recipients {
		if recipient.Kind == subscriptionTypeEmail {
			if err := sendEmail(recipient.Recipient, notification); err != nil {
				return fmt.Errorf("error sending email: %w", err)
			}
		} else if recipient.Kind == subscriptionTypeSlackWebhook {
			if err := sendSlack(recipient.Recipient, notification); err != nil {
				return fmt.Errorf("error sending email: %w", err)
			}
		}
	}
	return nil
}

func sendEmail(email string, notification Notification) error {
	core.Logger.Infof(notifierLogPrefix, "Send an email")

	err := core.SendEmail(
		&AppConfig.SMTP,
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
	if !core.CheckTemplateExists(templateFile) {
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
