package app

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
)

func LogMessage(message string, additionalData []string) {
	values := []string{message}
	for i := range additionalData {
		values = append(values, additionalData[i])
	}
	messagePrepared := strings.Join(values, "\n")
	if conf.TelegramBotToken != "" && conf.TelegramNotifyChat != "" {
		go TelegramMessage(messagePrepared, "html")
	}
	if conf.SlackToken != "" && conf.SlackChannel != "" {
		go SlackMessage(messagePrepared)
	}
}

func SlackMessage(message string) {
	data := url.Values{}
	data.Set("token", conf.SlackToken)
	data.Set("channel", conf.SlackChannel)
	data.Set("text", message)

	apiUrl := "https://slack.com/api/chat.postMessage"
	resp, err := http.Post(apiUrl, "application/x-www-form-urlencoded", strings.NewReader(data.Encode()))
	logError("SLACK", resp, err)
}

func TelegramMessage(message string, parseMode string) {
	requestBody, _ := json.Marshal(map[string]string{
		"chat_id":    conf.TelegramNotifyChat,
		"text":       message,
		"parse_mode": parseMode,
	})
	apiUrl := "https://api.telegram.org/bot" + conf.TelegramBotToken + "/sendMessage"
	resp, err := http.Post(apiUrl, "application/json", bytes.NewBuffer(requestBody))
	logError("TELEGRAM", resp, err)
}

func logError(method string, resp *http.Response, err error) {
	if err != nil || (resp != nil && resp.StatusCode != 200) {
		println("========= BEGIN " + method + " MESSAGE ERROR =========")
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
		println("========= END " + method + " MESSAGE ERROR =========")
	}
}
