package telegram

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"time"
)

// Client sends messages to Telegram channels using the Bot API.
type Client struct {
	token      string
	chatID     string
	apiBaseURL *url.URL
	httpClient *http.Client
}

// New creates a configured Telegram client. token and chatID must be provided.
func New(token, chatID string, opts ...Option) (*Client, error) {
	if token == "" {
		return nil, errors.New("telegram: bot token is required")
	}
	if chatID == "" {
		return nil, errors.New("telegram: chat id is required")
	}

	baseURL, _ := url.Parse("https://api.telegram.org")
	client := &Client{
		token:      token,
		chatID:     chatID,
		apiBaseURL: baseURL,
		httpClient: &http.Client{Timeout: 10 * time.Second},
	}

	for _, opt := range opts {
		opt(client)
	}

	return client, nil
}

// Option mutates client configuration.
type Option func(*Client)

// WithHTTPClient replaces the default HTTP client.
func WithHTTPClient(httpClient *http.Client) Option {
	return func(c *Client) {
		if httpClient != nil {
			c.httpClient = httpClient
		}
	}
}

// WithBaseURL overrides the Telegram API base URL.
func WithBaseURL(baseURL string) Option {
	return func(c *Client) {
		if baseURL == "" {
			return
		}
		if parsed, err := url.Parse(baseURL); err == nil {
			c.apiBaseURL = parsed
		}
	}
}

type sendMessageRequest struct {
	ChatID                string `json:"chat_id"`
	Text                  string `json:"text"`
	ParseMode             string `json:"parse_mode,omitempty"`
	DisableWebPagePreview bool   `json:"disable_web_page_preview,omitempty"`
}

// SendMessage posts text to the configured chat.
func (c *Client) SendMessage(ctx context.Context, text string) error {
	if text == "" {
		return errors.New("telegram: message text is empty")
	}

	endpoint, err := c.apiBaseURL.Parse(fmt.Sprintf("/bot%s/sendMessage", c.token))
	if err != nil {
		return fmt.Errorf("telegram: build url: %w", err)
	}

	payload := sendMessageRequest{
		ChatID:                c.chatID,
		Text:                  text,
		DisableWebPagePreview: true,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("telegram: marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint.String(), bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("telegram: create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("telegram: send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("telegram: unexpected status %s", resp.Status)
	}

	return nil
}
