package plugin

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestSlackNotifierNotifySuccess(t *testing.T) {
	t.Parallel()

	var payload slackPayload
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
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

func TestSlackNotifierNotifyHTTPError(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadGateway)
		_, _ = io.WriteString(w, "bad gateway")
	}))
	defer srv.Close()

	err := NewSlackNotifier(SlackConfig{WebhookURL: srv.URL}).Notify(context.Background(), "v1.0.0", "", "repo")
	if err == nil || !strings.Contains(err.Error(), "unexpected status") {
		t.Fatalf("expected status error, got %v", err)
	}
}

func TestSlackNotifierDefaultsAndBuildBlocks(t *testing.T) {
	t.Parallel()

	n := NewSlackNotifier(SlackConfig{WebhookURL: "https://hooks.slack.test"})
	if n.cfg.Username != "semrel" || n.cfg.IconEmoji != ":rocket:" {
		t.Fatalf("unexpected defaults: %+v", n.cfg)
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
