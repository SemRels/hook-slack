package plugin

import (
	"context"
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"
)

type closeErrorBody struct {
	io.Reader
	closed bool
}

func (b *closeErrorBody) Close() error {
	b.closed = true
	return errors.New("close failed")
}

func TestNotifyIgnoresCloseErrorAfterAcceptedResponse(t *testing.T) {
	body := &closeErrorBody{Reader: strings.NewReader("ok")}
	notifier := NewSlackNotifier(SlackConfig{WebhookURL: "https://slack.example.test"})
	notifier.client = &http.Client{Transport: roundTripFunc(func(*http.Request) (*http.Response, error) {
		return &http.Response{StatusCode: http.StatusOK, Body: body, Header: make(http.Header)}, nil
	})}

	if err := notifier.Notify(context.Background(), "v1.0.0", "", "repo"); err != nil {
		t.Fatalf("Notify() error = %v", err)
	}
	if !body.closed {
		t.Fatal("response body was not closed")
	}
}
