// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2026 The semrel Authors

package main

import (
	"log"

	plugin "github.com/SemRels/hook-slack/internal/plugin"
)

func main() {
	notifier := plugin.NewSlackNotifier(plugin.SlackConfig{})
	log.Printf("hook-slack plugin ready: sends Slack release notifications (%T)", notifier)
}
