package server

import (
	"context"
	"errors"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

type fakeMessenger struct {
	lastMessage string
	err         error
}

func (f *fakeMessenger) SendMessage(_ context.Context, text string) error {
	f.lastMessage = text
	return f.err
}

func TestWebhookHandlerSuccess(t *testing.T) {
	messenger := &fakeMessenger{}
	handler := NewWebhookHandler(messenger, log.New(io.Discard, "", 0))

	body := `{"message":"Heavy volume","tick":"BTCUSD","time":"2024-06-03T10:00:00Z","interval":"1h"}`
	req := httptest.NewRequest(http.MethodPost, "/webhook", strings.NewReader(body))
	res := httptest.NewRecorder()

	handler.ServeHTTP(res, req)

	if res.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", res.Code)
	}

	expected := "Heavy volume\nSymbol: BTCUSD\nInterval: 1h\nTime: 2024-06-03T10:00:00Z"
	if messenger.lastMessage != expected {
		t.Fatalf("unexpected message: %q", messenger.lastMessage)
	}
}

func TestWebhookHandlerMessengerError(t *testing.T) {
	messenger := &fakeMessenger{err: errors.New("boom")}
	handler := NewWebhookHandler(messenger, log.New(io.Discard, "", 0))

	body := `{"message":"Test"}`
	req := httptest.NewRequest(http.MethodPost, "/webhook", strings.NewReader(body))
	res := httptest.NewRecorder()

	handler.ServeHTTP(res, req)

	if res.Code != http.StatusBadGateway {
		t.Fatalf("expected status 502, got %d", res.Code)
	}
}

func TestWebhookHandlerRejectsInvalidJSON(t *testing.T) {
	messenger := &fakeMessenger{}
	handler := NewWebhookHandler(messenger, log.New(io.Discard, "", 0))

	req := httptest.NewRequest(http.MethodPost, "/webhook", strings.NewReader("{"))
	res := httptest.NewRecorder()

	handler.ServeHTTP(res, req)

	if res.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", res.Code)
	}
}
