package config

import (
	"github.com/spf13/viper"
)

type TelegramConfig struct {
	BotToken string
	ChatID   string
}

type DbConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	DBName   string
}

type Config struct {
	Telegram TelegramConfig
	Db       DbConfig
}

var Cfg Config

func LoadFromViper() {
	Cfg = Config{
		Telegram: TelegramConfig{
			BotToken: viper.GetString("telegram.bot_token"),
			ChatID:   viper.GetString("telegram.chat_id"),
		},
		Db: DbConfig{
			Host:     viper.GetString("db.host"),
			Port:     viper.GetInt("db.port"),
			User:     viper.GetString("db.user"),
			Password: viper.GetString("db.password"),
			DBName:   viper.GetString("db.dbname"),
		},
	}
}
