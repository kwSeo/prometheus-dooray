package dooray

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/pkg/errors"

	"github.com/kwseo/prometheus-dooray/pkg/prom"
)

const ContentType = "application/json"

type Message struct {
	BotName      string       `json:"botName"`
	BotIconImage string       `json:"botIconImage"`
	Text         string       `json:"text"`
	Attachments  []Attachment `json:"attachments"`
}

type Attachment struct {
	Title     string `json:"title"`
	TitleLink string `json:"titleLink"`
	Text      string `json:"text"`
	Color     string `json:"color"`
}

func FromAlertmanager(alert prom.Alert) *Message {
	text := "## Labels\n"
	for k, v := range alert.Labels {
		text += fmt.Sprintf("%s: %s\n", k, v)
	}
	text += "\n## Annotations\n"
	for k, v := range alert.Annotations {
		text += fmt.Sprintf("%s: %s\n", k, v)
	}
	return &Message{
		BotName:      "Prometheus",
		BotIconImage: "https://static.dooray.com/static_images/dooray-bot.png",
		Text:         text,
	}
}

func FromWebhook(webhook prom.Webhook) ([]*Message, error) {
	if webhook.Version != prom.SupportedVersion {
		return nil, errors.Errorf("unsupported version: %s", webhook.Version)
	}
	var messages []*Message
	for _, alert := range webhook.Alerts {
		message := FromAlertmanager(prom.Alert{
			Labels:       alert.Labels,
			Annotations:  alert.Annotations,
			StartsAt:     alert.StartsAt,
			EndsAt:       alert.EndsAt,
			GeneratorURL: alert.GeneratorURL,
		})
		message.Text = fmt.Sprintf("---Status: %s---\n%s", alert.Status, message.Text)
		messages = append(messages, message)
	}
	return messages, nil
}

func Send(incomingURL string, msg Message) error {
	buf := new(bytes.Buffer)
	if err := json.NewEncoder(buf).Encode(msg); err != nil {
		return errors.Wrap(err, "failed to encode the message")
	}
	resp, err := http.Post(incomingURL, ContentType, buf)
	if err != nil {
		return errors.Wrap(err, "failed to request")
	}
	if resp.StatusCode != http.StatusOK {
		return errors.Errorf("Unsuccessful response code: %d", resp.StatusCode)
	}
	return nil
}
