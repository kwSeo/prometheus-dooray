package dooray

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/pkg/errors"

	"github.com/kwseo/prometheus-dooray/pkg/prom"
)

type AlertmanagerHandler struct {
	cfg    Config
	logger log.Logger
}

func NewAlertmanagerHandler(cfg Config, logger log.Logger) *AlertmanagerHandler {
	return &AlertmanagerHandler{
		cfg:    cfg,
		logger: log.With(logger, "module", "AlertmanagerHandler"),
	}
}

func (h *AlertmanagerHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	level.Info(h.logger).Log("msg", "Received alerts by Prometheus Alertmanager endpoint.")
	var alerts []prom.Alert
	if err := json.NewDecoder(r.Body).Decode(&alerts); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	for _, alert := range alerts {
		message := h.FromAlertmanager(alert)
		if err := Send(h.cfg.IncomingURL, *message); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}

func (h *AlertmanagerHandler) FromAlertmanager(alert prom.Alert) *Message {
	return &Message{
		BotName:      h.cfg.BotName,
		BotIconImage: h.cfg.IconURL,
		Text:         CreateText(alert),
	}
}

type WebhookHandler struct {
	cfg    Config
	logger log.Logger
}

func NewWebhookHandler(cfg Config, logger log.Logger) *WebhookHandler {
	return &WebhookHandler{
		cfg:    cfg,
		logger: log.With(logger, "module", "WebhookHandler"),
	}
}

func (h *WebhookHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	level.Info(h.logger).Log("msg", "Received alerts from Alertmanager Webhook.")
	var webhook prom.Webhook
	if err := json.NewDecoder(r.Body).Decode(&webhook); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	messages, err := h.FromWebhook(webhook)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	for _, message := range messages {
		if err := Send(h.cfg.IncomingURL, *message); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}

func (h *WebhookHandler) FromWebhook(webhook prom.Webhook) ([]*Message, error) {
	if webhook.Version != prom.SupportedVersion {
		return nil, errors.Errorf("unsupported version: %s", webhook.Version)
	}
	var messages []*Message
	for _, alert := range webhook.Alerts {
		text := CreateText(prom.Alert{
			Labels:       alert.Labels,
			Annotations:  alert.Annotations,
			StartsAt:     alert.StartsAt,
			EndsAt:       alert.EndsAt,
			GeneratorURL: alert.GeneratorURL,
		})
		messages = append(messages, &Message{
			BotName:      h.cfg.BotName,
			BotIconImage: h.cfg.IconURL,
			Text:         fmt.Sprintf("---Status: %s---\n%s", alert.Status, text),
		})
	}
	return messages, nil
}
