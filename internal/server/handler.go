package server

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"
)

// Messenger sends messages to chat clients.
type Messenger interface {
	SendMessage(ctx context.Context, text string) error
}

// WebhookHandler handles TradingView alerts and forwards them to Telegram.
type WebhookHandler struct {
	messenger Messenger
	logger    *log.Logger
	timeout   time.Duration
}

// NewWebhookHandler constructs the handler.
func NewWebhookHandler(m Messenger, logger *log.Logger) *WebhookHandler {
	if logger == nil {
		logger = log.New(io.Discard, "", 0)
	}
	return &WebhookHandler{
		messenger: m,
		logger:    logger,
		timeout:   5 * time.Second,
	}
}

// ServeHTTP implements http.Handler.
func (h *WebhookHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var payload tradingViewPayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, "invalid json payload", http.StatusBadRequest)
		return
	}

	if err := payload.validate(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	message := formatMessage(payload)

	ctx, cancel := context.WithTimeout(r.Context(), h.timeout)
	defer cancel()

	if err := h.messenger.SendMessage(ctx, message); err != nil {
		h.logger.Printf("telegram send failed: %v", err)
		http.Error(w, "failed to deliver alert", http.StatusBadGateway)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	io.WriteString(w, `{"status":"ok"}`)
}

type tradingViewPayload struct {
	Message  string `json:"message"`
	Ticker   string `json:"tick"`
	Time     string `json:"time"`
	Interval string `json:"interval"`
}

func (p tradingViewPayload) validate() error {
	if strings.TrimSpace(p.Message) == "" {
		return errors.New("message is required")
	}
	return nil
}

func formatMessage(p tradingViewPayload) string {
	var builder strings.Builder
	builder.WriteString(p.Message)

	details := make([]string, 0, 3)
	if p.Ticker != "" {
		details = append(details, fmt.Sprintf("Symbol: %s", p.Ticker))
	}
	if p.Interval != "" {
		details = append(details, fmt.Sprintf("Interval: %s", p.Interval))
	}
	if p.Time != "" {
		details = append(details, fmt.Sprintf("Time: %s", p.Time))
	}

	if len(details) > 0 {
		builder.WriteString("\n")
		builder.WriteString(strings.Join(details, "\n"))
	}

	return builder.String()
}
