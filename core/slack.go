package core

import (
	"bytes"
	"encoding/json"
	"net/http"
	"time"
)

type SlackRequest struct {
	Text string `json:"text"`
}

func SendSlackNotification(webhookURL string, slackReq SlackRequest) {
	slackBody, _ := json.Marshal(slackReq)
	req, err := http.NewRequest(http.MethodPost, webhookURL, bytes.NewBuffer(slackBody))
	if err != nil {
		Logger.Errorf(logPrefix, "Error during call Slack : %s", err)

		return
	}

	req.Header.Add("Content-Type", "application/json")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		Logger.Errorf(logPrefix, "Error during call Slack : %s", err)

		return
	}
	defer resp.Body.Close()

	buf := new(bytes.Buffer)
	_, _ = buf.ReadFrom(resp.Body)
	if buf.String() != "ok" {
		Logger.Errorf(logPrefix, "Error during call Slack : non-ok response returned from Slack : %s", buf.String())

		return
	}
}

func SendSlackString(webhookURL, str string) {
	SendSlackNotification(webhookURL, SlackRequest{
		Text: str,
	})
}
