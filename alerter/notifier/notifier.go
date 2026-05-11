package notifier

import (
	"encoding/json"
	"fmt"
)

type Notifier interface {
	Send(subject, message string) error
}

type SlackConfig struct {
	WebhookURL string `json:"webhook_url"`
}

type DiscordConfig struct {
	WebhookURL string `json:"webhook_url"`
}

type TelegramConfig struct {
	BotToken string `json:"bot_token"`
	ChatID   string `json:"chat_id"`
}

type EmailConfig struct {
	SMTPHost string `json:"smtp_host"`
	SMTPPort string `json:"smtp_port"`
	Username string `json:"username"`
	Password string `json:"password"`
	From     string `json:"from"`
	To       string `json:"to"`
}

func New(channel, configJSON string) (Notifier, error) {
	switch channel {
	case "slack":
		var cfg SlackConfig
		if err := json.Unmarshal([]byte(configJSON), &cfg); err != nil {
			return nil, err
		}
		return &Slack{webhookURL: cfg.WebhookURL}, nil

	case "discord":
		var cfg DiscordConfig
		if err := json.Unmarshal([]byte(configJSON), &cfg); err != nil {
			return nil, err
		}
		return &Discord{webhookURL: cfg.WebhookURL}, nil

	case "telegram":
		var cfg TelegramConfig
		if err := json.Unmarshal([]byte(configJSON), &cfg); err != nil {
			return nil, err
		}
		return &Telegram{botToken: cfg.BotToken, chatID: cfg.ChatID}, nil

	case "email":
		var cfg EmailConfig
		if err := json.Unmarshal([]byte(configJSON), &cfg); err != nil {
			return nil, err
		}
		return &Email{cfg: cfg}, nil

	default:
		return nil, fmt.Errorf("unknown channel: %s", channel)
	}
}
