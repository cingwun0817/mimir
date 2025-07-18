package config

import (
	"github.com/spf13/viper"
)

type TelegramConfig struct {
	BotToken string
	ChatID   string
}

type Config struct {
	Telegram TelegramConfig
}

var Cfg Config

func LoadFromViper() {
	Cfg = Config{
		Telegram: TelegramConfig{
			BotToken: viper.GetString("telegram.bot_token"),
			ChatID:   viper.GetString("telegram.chat_id"),
		},
	}
}
