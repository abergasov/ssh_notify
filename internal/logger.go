package internal

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const (
	notificationTimeout = 10 * time.Second
	maxOutboundRequests = 8
)

var (
	notificationClient = &http.Client{Timeout: notificationTimeout}
	outboundRequests   = make(chan struct{}, maxOutboundRequests)
	slackAPIURL        = "https://slack.com/api/chat.postMessage"
	telegramAPIURL     = "https://api.telegram.org/bot%s/sendMessage"
)

func LogMessage(message string, additionalData ...string) {
	values := []string{message}
	for i := range additionalData {
		values = append(values, additionalData[i])
	}
	messagePrepared := strings.Join(values, "\n")
	if conf.TelegramBotToken != "" && conf.TelegramNotifyChat != "" {
		sendAsync(func() error { return TelegramMessage(messagePrepared, "html") })
	}
	if conf.SlackToken != "" && conf.SlackChannel != "" {
		sendAsync(func() error { return SlackMessage(messagePrepared) })
	}
}

func sendAsync(send func() error) {
	outboundRequests <- struct{}{}
	go func() {
		defer func() { <-outboundRequests }()
		_ = send()
	}()
}

func SlackMessage(message string) error {
	data := url.Values{}
	data.Set("token", conf.SlackToken)
	data.Set("channel", conf.SlackChannel)
	data.Set("text", message)

	resp, err := notificationClient.Post(slackAPIURL, "application/x-www-form-urlencoded", strings.NewReader(data.Encode()))
	if err != nil {
		return logError("SLACK", nil, nil, err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return logError("SLACK", resp, nil, err)
	}
	if resp.StatusCode != http.StatusOK {
		return logError("SLACK", resp, body, fmt.Errorf("unexpected response status %s", resp.Status))
	}

	var result struct {
		OK    bool   `json:"ok"`
		Error string `json:"error"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return logError("SLACK", resp, body, fmt.Errorf("decode response: %w", err))
	}
	if !result.OK {
		return logError("SLACK", resp, body, fmt.Errorf("API response was not ok: %s", result.Error))
	}
	return nil
}

func TelegramMessage(message string, parseMode string) error {
	requestBody, err := json.Marshal(map[string]string{
		"chat_id":    conf.TelegramNotifyChat,
		"text":       message,
		"parse_mode": parseMode,
	})
	if err != nil {
		return logError("TELEGRAM", nil, nil, err)
	}

	apiURL := fmt.Sprintf(telegramAPIURL, conf.TelegramBotToken)
	resp, err := notificationClient.Post(apiURL, "application/json", bytes.NewReader(requestBody))
	if err != nil {
		return logError("TELEGRAM", nil, nil, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusOK {
		return nil
	}

	body, readErr := ioutil.ReadAll(resp.Body)
	if readErr != nil {
		return logError("TELEGRAM", resp, nil, readErr)
	}
	return logError("TELEGRAM", resp, body, fmt.Errorf("unexpected response status %s", resp.Status))
}

func logError(method string, resp *http.Response, body []byte, err error) error {
	if err == nil {
		return nil
	}

	log.Printf("========= BEGIN %s MESSAGE ERROR =========", method)
	if resp != nil {
		log.Printf("response status: %s", sanitizeLogText(resp.Status))
	}
	if len(body) > 0 {
		log.Print(sanitizeLogText(string(body)))
	}
	log.Print(sanitizeLogText(err.Error()))
	log.Printf("========= END %s MESSAGE ERROR =========", method)
	return fmt.Errorf("%s notification: %s", strings.ToLower(method), sanitizeLogText(err.Error()))
}

func sanitizeLogText(value string) string {
	if conf == nil {
		return value
	}
	for _, token := range []string{conf.TelegramBotToken, conf.SlackToken} {
		if token == "" {
			continue
		}
		value = strings.ReplaceAll(value, token, "[REDACTED]")
		value = strings.ReplaceAll(value, url.QueryEscape(token), "[REDACTED]")
		value = strings.ReplaceAll(value, url.PathEscape(token), "[REDACTED]")
	}
	return value
}
