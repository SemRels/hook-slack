// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2026 The semrel Authors

package plugin

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (fn roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return fn(req)
}

func TestSlackNotifierNotifySuccess(t *testing.T) {
	t.Parallel()

	var payload slackPayload
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Cleanup(func() {
			if err := r.Body.Close(); err != nil {
				t.Errorf("close request body: %v", err)
			}
		})
		if r.Header.Get("Content-Type") != "application/json" {
			t.Fatalf("unexpected content type: %s", r.Header.Get("Content-Type"))
		}
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			t.Fatalf("decode payload: %v", err)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	n := NewSlackNotifier(SlackConfig{WebhookURL: srv.URL, Channel: "#releases", Username: "bot", IconEmoji: ":tada:"})
	if err := n.Notify(context.Background(), "v1.2.3", "- feature", "myapp"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if payload.Username != "bot" || payload.IconEmoji != ":tada:" || payload.Channel != "#releases" {
		t.Fatalf("unexpected payload: %+v", payload)
	}
	if len(payload.Blocks) != 3 {
		t.Fatalf("expected 3 blocks, got %d", len(payload.Blocks))
	}
}

func TestSlackNotifierNotifyRequiresWebhook(t *testing.T) {
	t.Parallel()

	err := NewSlackNotifier(SlackConfig{}).Notify(context.Background(), "v1.0.0", "", "")
	if err == nil || !strings.Contains(err.Error(), "webhook URL") {
		t.Fatalf("expected webhook error, got %v", err)
	}
}

func TestSlackNotifierNotifyRetriesOnServerError(t *testing.T) {
	var attempts int
	var logs bytes.Buffer
	oldWriter := retryLogWriter
	retryLogWriter = &logs
	defer func() { retryLogWriter = oldWriter }()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		if attempts < 3 {
			w.WriteHeader(http.StatusBadGateway)
			_, _ = io.WriteString(w, "bad gateway")
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	n := NewSlackNotifier(SlackConfig{WebhookURL: srv.URL, MaxRetries: 2, RetryDelay: time.Millisecond})
	if err := n.Notify(context.Background(), "v1.0.0", "", "repo"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if attempts != 3 {
		t.Fatalf("expected 3 attempts, got %d", attempts)
	}
	if got := logs.String(); !strings.Contains(got, "retry attempt 1/2") || !strings.Contains(got, "retry attempt 2/2") {
		t.Fatalf("expected retry logs, got %q", got)
	}
}

func TestSlackNotifierNotifyDoesNotRetryOnClientError(t *testing.T) {
	t.Parallel()

	var attempts int
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		w.WriteHeader(http.StatusBadRequest)
	}))
	defer srv.Close()

	err := NewSlackNotifier(SlackConfig{WebhookURL: srv.URL, MaxRetries: 3, RetryDelay: time.Millisecond}).Notify(context.Background(), "v1.0.0", "", "repo")
	if err == nil || !strings.Contains(err.Error(), "unexpected status 400") {
		t.Fatalf("expected status error, got %v", err)
	}
	if attempts != 1 {
		t.Fatalf("expected 1 attempt, got %d", attempts)
	}
}

func TestSlackNotifierNotifyRetriesOnNetworkError(t *testing.T) {
	t.Parallel()

	var attempts int
	n := NewSlackNotifier(SlackConfig{WebhookURL: "https://hooks.slack.test", MaxRetries: 2, RetryDelay: time.Millisecond})
	n.client = &http.Client{
		Timeout: DefaultTimeout,
		Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
			attempts++
			if attempts < 3 {
				return nil, &net.DNSError{Err: "temporary failure", IsTemporary: true}
			}
			return &http.Response{StatusCode: http.StatusOK, Body: io.NopCloser(strings.NewReader("ok")), Header: make(http.Header)}, nil
		}),
	}

	if err := n.Notify(context.Background(), "v1.0.0", "", "repo"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if attempts != 3 {
		t.Fatalf("expected 3 attempts, got %d", attempts)
	}
}

func TestSlackNotifierDefaultsAndBuildBlocks(t *testing.T) {
	t.Parallel()

	n := NewSlackNotifier(SlackConfig{WebhookURL: "https://hooks.slack.test"})
	if n.cfg.Username != "semrel" || n.cfg.IconEmoji != ":rocket:" {
		t.Fatalf("unexpected defaults: %+v", n.cfg)
	}
	if n.cfg.MaxRetries != DefaultMaxRetries || n.cfg.RetryDelay != DefaultRetryDelay {
		t.Fatalf("unexpected retry defaults: %+v", n.cfg)
	}

	blocks := n.buildBlocks("v1.0.0", strings.Repeat("x", 1000), "")
	if len(blocks) != 3 {
		t.Fatalf("expected 3 blocks, got %d", len(blocks))
	}
	if !strings.Contains(blocks[0].Text.Text, "v1.0.0") {
		t.Fatalf("expected version in header: %+v", blocks[0])
	}
	if !strings.HasSuffix(blocks[1].Text.Text, "...") {
		t.Fatalf("expected truncated changelog, got %q", blocks[1].Text.Text)
	}
}

func TestRetryDoDoesNotRetryContextErrors(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	attempts := 0

	_, err := retryDo(ctx, 3, time.Millisecond, func() (*http.Response, error) {
		attempts++
		return nil, context.Canceled
	})
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context canceled, got %v", err)
	}
	if attempts != 1 {
		t.Fatalf("expected 1 attempt, got %d", attempts)
	}
}
