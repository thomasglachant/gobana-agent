package core

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type SlackRequest struct {
	Text string `json:"text"`
}

func SendSlackNotification(webhookURL string, slackReq SlackRequest) error {
	slackBody, _ := json.Marshal(slackReq)
	req, err := http.NewRequestWithContext(context.Background(), http.MethodPost, webhookURL, bytes.NewBuffer(slackBody))
	if err != nil {
		return fmt.Errorf("error during call Slack : %s", err)
	}

	req.Header.Add("Content-Type", "application/json")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("error during call Slack : %s", err)
	}
	defer resp.Body.Close()

	buf := new(bytes.Buffer)
	_, _ = buf.ReadFrom(resp.Body)
	if buf.String() != "ok" {
		return fmt.Errorf("error during call Slack : non-ok response returned from Slack : %s", buf.String())
	}

	return nil
}
