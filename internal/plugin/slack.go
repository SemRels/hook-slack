// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2026 The semrel Authors

package plugin

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

type SlackConfig struct {
	WebhookURL string
	Channel    string
	Username   string
	IconEmoji  string
}

type SlackNotifier struct {
	cfg    SlackConfig
	client *http.Client
}

func NewSlackNotifier(cfg SlackConfig) *SlackNotifier {
	if cfg.Username == "" {
		cfg.Username = "semrel"
	}
	if cfg.IconEmoji == "" {
		cfg.IconEmoji = ":rocket:"
	}
	return &SlackNotifier{
		cfg:    cfg,
		client: &http.Client{Timeout: DefaultTimeout},
	}
}

type slackPayload struct {
	Channel   string       `json:"channel,omitempty"`
	Username  string       `json:"username,omitempty"`
	IconEmoji string       `json:"icon_emoji,omitempty"`
	Blocks    []slackBlock `json:"blocks"`
}

type slackBlock struct {
	Type string     `json:"type"`
	Text *slackText `json:"text,omitempty"`
}

type slackText struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

func (n *SlackNotifier) Notify(ctx context.Context, version, changelog, repository string) error {
	if n.cfg.WebhookURL == "" {
		return fmt.Errorf("slack: webhook URL is required")
	}

	payload := slackPayload{
		Channel:   n.cfg.Channel,
		Username:  n.cfg.Username,
		IconEmoji: n.cfg.IconEmoji,
		Blocks:    n.buildBlocks(version, changelog, repository),
	}

	b, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("slack: marshal payload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, n.cfg.WebhookURL, bytes.NewReader(b))
	if err != nil {
		return fmt.Errorf("slack: create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := n.client.Do(req)
	if err != nil {
		return fmt.Errorf("slack: send notification: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("slack: unexpected status %d", resp.StatusCode)
	}
	return nil
}

func (n *SlackNotifier) buildBlocks(version, changelog, repository string) []slackBlock {
	header := fmt.Sprintf(":rocket: *%s released*", version)
	if repository != "" {
		header = fmt.Sprintf(":rocket: *%s %s released*", repository, version)
	}

	blocks := []slackBlock{{Type: "section", Text: &slackText{Type: "mrkdwn", Text: header}}}
	if changelog != "" {
		notes := strings.TrimSpace(changelog)
		if len(notes) > 900 {
			notes = notes[:900] + "..."
		}
		blocks = append(blocks, slackBlock{Type: "section", Text: &slackText{Type: "mrkdwn", Text: notes}})
	}
	blocks = append(blocks, slackBlock{Type: "divider"})
	return blocks
}
