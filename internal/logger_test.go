package internal

import (
	"bytes"
	"errors"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"
	"testing"
	"time"
)

type closeTrackingBody struct {
	io.Reader
	closed bool
}

func (b *closeTrackingBody) Close() error {
	b.closed = true
	return nil
}

type roundTripperFunc func(*http.Request) (*http.Response, error)

func (f roundTripperFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

func TestSlackMessageClosesResponseBody(t *testing.T) {
	oldConf, oldClient, oldURL := conf, notificationClient, slackAPIURL
	defer func() { conf, notificationClient, slackAPIURL = oldConf, oldClient, oldURL }()
	conf = &Config{SlackToken: "token", SlackChannel: "channel"}
	body := &closeTrackingBody{Reader: strings.NewReader(`{"ok":true}`)}
	notificationClient = &http.Client{Timeout: notificationTimeout, Transport: roundTripperFunc(func(*http.Request) (*http.Response, error) {
		return &http.Response{StatusCode: http.StatusOK, Status: "200 OK", Body: body, Header: make(http.Header)}, nil
	})}
	slackAPIURL = "http://slack.test"

	if err := SlackMessage("message"); err != nil {
		t.Fatalf("SlackMessage() error = %v", err)
	}
	if !body.closed {
		t.Fatal("SlackMessage() did not close the response body")
	}
}

func TestSlackMessageReturnsErrorWhenAPISaysNotOK(t *testing.T) {
	oldConf, oldClient, oldURL := conf, notificationClient, slackAPIURL
	defer func() { conf, notificationClient, slackAPIURL = oldConf, oldClient, oldURL }()
	conf = &Config{SlackToken: "token", SlackChannel: "channel"}
	notificationClient = &http.Client{Timeout: notificationTimeout, Transport: roundTripperFunc(func(*http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: http.StatusOK,
			Status:     "200 OK",
			Body:       io.NopCloser(strings.NewReader(`{"ok":false,"error":"invalid_auth"}`)),
			Header:     make(http.Header),
		}, nil
	})}
	slackAPIURL = "http://slack.test"

	if err := SlackMessage("message"); err == nil || !strings.Contains(err.Error(), "invalid_auth") {
		t.Fatalf("SlackMessage() error = %v, want API ok:false error", err)
	}
}

func TestTelegramMessageUsesClientTimeout(t *testing.T) {
	oldConf, oldClient, oldURL := conf, notificationClient, telegramAPIURL
	defer func() { conf, notificationClient, telegramAPIURL = oldConf, oldClient, oldURL }()
	conf = &Config{TelegramBotToken: "token", TelegramNotifyChat: "chat"}
	notificationClient = &http.Client{Timeout: 25 * time.Millisecond, Transport: roundTripperFunc(func(req *http.Request) (*http.Response, error) {
		<-req.Context().Done()
		return nil, req.Context().Err()
	})}
	telegramAPIURL = "http://telegram.test/bot%s/sendMessage"

	started := time.Now()
	if err := TelegramMessage("message", "html"); err == nil {
		t.Fatal("TelegramMessage() error = nil, want client timeout")
	}
	if elapsed := time.Since(started); elapsed >= 150*time.Millisecond {
		t.Fatalf("TelegramMessage() returned after %s, want client timeout", elapsed)
	}
}

func TestTelegramMessageClosesSuccessfulResponseBody(t *testing.T) {
	oldConf, oldClient, oldURL := conf, notificationClient, telegramAPIURL
	defer func() { conf, notificationClient, telegramAPIURL = oldConf, oldClient, oldURL }()
	conf = &Config{TelegramBotToken: "token", TelegramNotifyChat: "chat"}
	body := &closeTrackingBody{Reader: strings.NewReader(`{"ok":true}`)}
	notificationClient = &http.Client{Timeout: notificationTimeout, Transport: roundTripperFunc(func(*http.Request) (*http.Response, error) {
		return &http.Response{StatusCode: http.StatusOK, Status: "200 OK", Body: body, Header: make(http.Header)}, nil
	})}
	telegramAPIURL = "http://telegram.test/bot%s/sendMessage"

	if err := TelegramMessage("message", "html"); err != nil {
		t.Fatalf("TelegramMessage() error = %v", err)
	}
	if !body.closed {
		t.Fatal("TelegramMessage() did not close the response body")
	}
}

func TestLogErrorRedactsConfiguredTokens(t *testing.T) {
	oldConf, oldOutput, oldFlags := conf, log.Writer(), log.Flags()
	defer func() {
		conf = oldConf
		log.SetOutput(oldOutput)
		log.SetFlags(oldFlags)
	}()
	conf = &Config{TelegramBotToken: "telegram-secret", SlackToken: "slack-secret"}
	var output bytes.Buffer
	log.SetOutput(&output)
	log.SetFlags(0)

	logError("TELEGRAM", nil, []byte("telegram-secret slack-secret"), errors.New("request to telegram-secret failed with slack-secret"))
	if got := output.String(); strings.Contains(got, "telegram-secret") || strings.Contains(got, "slack-secret") {
		t.Fatalf("log output exposed a configured token: %s", got)
	}
	if !strings.Contains(output.String(), "[REDACTED]") {
		t.Fatalf("log output = %q, want redacted token", output.String())
	}
}

func TestTelegramMessageRedactsTokenInRequestError(t *testing.T) {
	oldConf, oldClient, oldURL := conf, notificationClient, telegramAPIURL
	oldOutput, oldFlags := log.Writer(), log.Flags()
	defer func() {
		conf, notificationClient, telegramAPIURL = oldConf, oldClient, oldURL
		log.SetOutput(oldOutput)
		log.SetFlags(oldFlags)
	}()
	conf = &Config{TelegramBotToken: "telegram-secret", TelegramNotifyChat: "chat"}
	notificationClient = &http.Client{Timeout: notificationTimeout, Transport: roundTripperFunc(func(*http.Request) (*http.Response, error) {
		return nil, &url.Error{
			Op:  "Post",
			URL: "https://api.telegram.org/bottelegram-secret/sendMessage",
			Err: errors.New("network failure"),
		}
	})}
	telegramAPIURL = "https://api.telegram.org/bot%s/sendMessage"
	var output bytes.Buffer
	log.SetOutput(&output)
	log.SetFlags(0)

	err := TelegramMessage("message", "html")
	if err == nil {
		t.Fatal("TelegramMessage() error = nil, want request error")
	}
	if strings.Contains(err.Error(), "telegram-secret") {
		t.Fatalf("returned error exposed token: %s", err)
	}
	if strings.Contains(output.String(), "telegram-secret") {
		t.Fatalf("log output exposed token: %s", output.String())
	}
}
