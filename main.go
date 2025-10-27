package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/BurntSushi/toml"

	"telegram-alerts/internal/server"
	"telegram-alerts/internal/telegram"
)

const defaultAddr = ":8080"

type Config struct {
	Telegram TelegramConfig `toml:"telegram"`
	Server   ServerConfig   `toml:"server"`
}

type TelegramConfig struct {
	BotToken string `toml:"bot_token"`
	ChatID   string `toml:"chat_id"`
}

type ServerConfig struct {
	Addr string `toml:"addr"`
}

func main() {
	configPath := flag.String("config", "config.toml", "Path to the configuration file")
	flag.Parse()

	logger := log.New(os.Stdout, "", log.LstdFlags)

	cfg, err := loadConfig(*configPath)
	if err != nil {
		logger.Fatalf("failed to load config: %v", err)
	}

	if cfg.Telegram.BotToken == "" {
		logger.Fatal("telegram bot token is required in config")
	}
	if cfg.Telegram.ChatID == "" {
		logger.Fatal("telegram chat id is required in config")
	}

	addr := cfg.Server.Addr
	if addr == "" {
		addr = defaultAddr
	}

	telegramClient, err := telegram.New(cfg.Telegram.BotToken, cfg.Telegram.ChatID)
	if err != nil {
		logger.Fatalf("failed to configure telegram client: %v", err)
	}

	webhookHandler := server.NewWebhookHandler(telegramClient, logger)

	mux := http.NewServeMux()
	mux.Handle("/webhook", webhookHandler)
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})

	srv := &http.Server{
		Addr:    addr,
		Handler: mux,
	}

	go func() {
		logger.Printf("listening on %s", addr)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Fatalf("server error: %v", err)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop

	logger.Println("shutting down...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logger.Printf("graceful shutdown failed: %v", err)
	}
}

func loadConfig(path string) (*Config, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open config: %w", err)
	}
	defer file.Close()

	var cfg Config
	if _, err := toml.NewDecoder(file).Decode(&cfg); err != nil {
		return nil, fmt.Errorf("decode config: %w", err)
	}

	return &cfg, nil
}
