package dooray

import (
	"encoding/json"
	"net/http"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"

	"github.com/kwseo/prometheus-dooray/pkg/prom"
)

type AlertmanagerHandler struct {
	incomingURL string
	logger log.Logger
}

func NewAlertmanagerHandler(incomingURL string, logger log.Logger) *AlertmanagerHandler {
	return &AlertmanagerHandler{
		incomingURL: incomingURL,
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
		message := FromAlertmanager(alert)
		if err := Send(h.incomingURL, *message); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}

type WebhookHandler struct {
	incomingURL string
	logger log.Logger
}

func NewWebhookHandler(incomingURL string, logger log.Logger) *WebhookHandler {
	return &WebhookHandler{
		incomingURL: incomingURL,
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

	messages, err := FromWebhook(webhook)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	for _, message := range messages {
		if err := Send(h.incomingURL, *message); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}

