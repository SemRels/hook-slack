package main

import (
	"bytes"
	"context"
	"errors"
	"testing"

	plugin "github.com/SemRels/hook-slack/internal/plugin"
)

type fakeNotifier struct {
	version string
	repo    string
	err     error
}

func (f *fakeNotifier) Notify(_ context.Context, version, _ string, repository string) error {
	f.version = version
	f.repo = repository
	return f.err
}

func env(kv map[string]string) func(string) string {
	return func(key string) string { return kv[key] }
}

func TestRunSuccess(t *testing.T) {

	fake := &fakeNotifier{}
	old := newNotifier
	newNotifier = func(cfg plugin.SlackConfig) notifier {
		if cfg.WebhookURL != "https://hooks.slack.test" {
			t.Fatalf("unexpected webhook: %s", cfg.WebhookURL)
		}
		return fake
	}
	defer func() { newNotifier = old }()

	var stderr bytes.Buffer
	code := run(context.Background(), env(map[string]string{
		"SEMREL_PLUGIN_WEBHOOK_URL": "https://hooks.slack.test",
		"SEMREL_VERSION":            "v1.2.3",
		"SEMREL_PLUGIN_REPOSITORY":  "SemRels/semrel",
	}), &stderr)

	if code != 0 || stderr.Len() != 0 {
		t.Fatalf("unexpected result: code=%d stderr=%q", code, stderr.String())
	}
	if fake.version != "v1.2.3" || fake.repo != "SemRels/semrel" {
		t.Fatalf("unexpected notify args: %+v", fake)
	}
}

func TestRunDryRun(t *testing.T) {

	called := false
	old := newNotifier
	newNotifier = func(plugin.SlackConfig) notifier {
		called = true
		return &fakeNotifier{}
	}
	defer func() { newNotifier = old }()

	var stderr bytes.Buffer
	code := run(context.Background(), env(map[string]string{
		"SEMREL_PLUGIN_WEBHOOK_URL": "https://hooks.slack.test",
		"SEMREL_VERSION":            "v1.2.3",
		"SEMREL_DRY_RUN":            "true",
	}), &stderr)

	if code != 0 || called {
		t.Fatalf("unexpected result: code=%d called=%v", code, called)
	}
}

func TestRunValidationError(t *testing.T) {

	var stderr bytes.Buffer
	code := run(context.Background(), env(map[string]string{}), &stderr)
	if code != 1 || stderr.Len() == 0 {
		t.Fatalf("unexpected result: code=%d stderr=%q", code, stderr.String())
	}
}

func TestRunNotifyError(t *testing.T) {

	old := newNotifier
	newNotifier = func(plugin.SlackConfig) notifier {
		return &fakeNotifier{err: errors.New("boom")}
	}
	defer func() { newNotifier = old }()

	var stderr bytes.Buffer
	code := run(context.Background(), env(map[string]string{
		"SEMREL_PLUGIN_WEBHOOK_URL": "https://hooks.slack.test",
		"SEMREL_VERSION":            "v1.2.3",
	}), &stderr)
	if code != 1 || stderr.Len() == 0 {
		t.Fatalf("unexpected result: code=%d stderr=%q", code, stderr.String())
	}
}

func TestRunMissingVersion(t *testing.T) {
	var stderr bytes.Buffer
	code := run(context.Background(), env(map[string]string{
		"SEMREL_PLUGIN_WEBHOOK_URL": "https://hooks.slack.test",
	}), &stderr)
	if code != 1 || stderr.Len() == 0 {
		t.Fatalf("unexpected result: code=%d stderr=%q", code, stderr.String())
	}
}

func TestFirstNonEmpty(t *testing.T) {
	if got := firstNonEmpty("", "v1.2.3", "v1.2.4"); got != "v1.2.3" {
		t.Fatalf("unexpected value: %s", got)
	}
}
