package config

import (
	"os"
	"strconv"
)

type Config struct {
	LogLevel       string
	TelegramBotId  string
	TelegramChatId int64
	DatabaseDSN    string
}

func getDefaults() Config {
	return Config{
		LogLevel:       "debug",
		TelegramBotId:  "",
		TelegramChatId: 0,
		DatabaseDSN:    "operations.db",
	}
}

func GetFromEnv() Config {
	conf := getDefaults()

	if logLevel := os.Getenv("LOG_LEVEL"); logLevel != "" {
		conf.LogLevel = logLevel
	}

	if botID := os.Getenv("TELEGRAM_BOT_ID"); botID != "" {
		conf.TelegramBotId = botID
	}

	if chatID := os.Getenv("TELEGRAM_CHAT_ID"); chatID != "" {
		conf.TelegramChatId, _ = strconv.ParseInt(chatID, 10, 64)
	}

	if dsn := os.Getenv("DATABASE_DSN"); dsn != "" {
		conf.DatabaseDSN = dsn
	}

	return conf
}
