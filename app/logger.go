package app

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
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
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(requestBody))
	if err != nil || (resp != nil && resp.StatusCode != 200) {
		println("========= BEGIN TELEGRAM MESSAGE ERROR =========")
		if resp != nil {
			println("Resp status ", resp.Status)
			bodyBytes, e := ioutil.ReadAll(resp.Body)
			if e == nil {
				bodyString := string(bodyBytes)
				println(bodyString)
			}
		}
		if err != nil {
			println(err.Error())
		}
		println("========= END TELEGRAM MESSAGE ERROR =========")
	}
}
