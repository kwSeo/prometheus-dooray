package dooray

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"net/http"

	"github.com/pkg/errors"

	"github.com/kwseo/prometheus-dooray/pkg/prom"
)

const ContentType = "application/json"

type Config struct {
	IncomingURL string
	IconURL     string
	BotName     string
}

func (c *Config) RegisterFlags(set *flag.FlagSet) {
	set.StringVar(&c.IncomingURL, "dooray.incoming-url", "", "incoming URL of Dooray Messanger")
	set.StringVar(&c.IconURL, "dooray.icon-url", "https://static.dooray.com/static_images/dooray-bot.png", "Icon URL of Dooray Messanger")
	set.StringVar(&c.BotName, "dooray.bot-name", "Prometheus", "Bot name of Dooray messanger")
}

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

func CreateText(alert prom.Alert) string {
	text := "## Labels\n"
	for k, v := range alert.Labels {
		text += fmt.Sprintf("%s: %s\n", k, v)
	}
	text += "\n## Annotations\n"
	for k, v := range alert.Annotations {
		text += fmt.Sprintf("%s: %s\n", k, v)
	}
	return text
}
