package app

import (
	"bytes"
	"encoding/json"
	"net/http"
	"strings"
)

func LogMessage(message string, additionalData []string) {
	values := []string{message}
	for i := range additionalData {
		values = append(values, additionalData[i])
	}
	TelegramMessage(conf.TelegramNotifyChat, strings.Join(values, "\n"), "html")
}

func TelegramMessage(chatId string, message string, parseMode string) {
	requestBody, _ := json.Marshal(map[string]string{
		"chat_id":    chatId,
		"text":       message,
		"parse_mode": parseMode,
	})
	url := "https://api.telegram.org/bot" + conf.TelegramBotToken + "/sendMessage"
	_, _ = http.Post(url, "application/json", bytes.NewBuffer(requestBody))
}
