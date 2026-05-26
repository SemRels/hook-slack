# hook-slack

Posts a release announcement to Slack using an incoming webhook.

This plugin is distributed as the standalone Go binary `semrel-plugin-hook-slack`. Semrel executes the binary as a subprocess, provides plugin configuration through `SEMREL_PLUGIN_*` environment variables, provides release context through `SEMREL_*` environment variables, reads standard output, and treats exit code `0` as success and any non-zero exit code as failure. Install the binary in `~/.semrel/plugins/` or anywhere on your `$PATH`.

## Installation

```bash
go install github.com/SemRels/hook-slack/cmd/plugin@latest
```

## Configuration

```yaml
plugins:
  - name: hook-slack
    path: ~/.semrel/plugins/semrel-plugin-hook-slack
    env:
      SEMREL_PLUGIN_WEBHOOK_URL: "https://hooks.slack.com/services/xxx/yyy/zzz"
      SEMREL_PLUGIN_CHANNEL: "#releases"
      SEMREL_PLUGIN_USERNAME: "semrel"
      SEMREL_PLUGIN_ICON_EMOJI: ":rocket:"
```

## `SEMREL_PLUGIN_*` variables

| Name | Required | Description | Default |
| --- | --- | --- | --- |
| `SEMREL_PLUGIN_WEBHOOK_URL` | Required | Slack incoming webhook URL. | None |
| `SEMREL_PLUGIN_CHANNEL` | Optional | Slack channel override for the message. | Webhook default |
| `SEMREL_PLUGIN_USERNAME` | Optional | Display name for the webhook sender. | semrel |
| `SEMREL_PLUGIN_ICON_EMOJI` | Optional | Emoji icon for the webhook sender. | :rocket: |

## `SEMREL_*` release context used

| Variable | Description |
| --- | --- |
| `SEMREL_VERSION` | Resolved release version for the current run. |
| `SEMREL_TAG_NAME` | Git tag name semrel will create or publish. |
| `SEMREL_NEXT_VERSION` | Next version computed by semrel for the release. |
| `SEMREL_CHANGELOG` | Generated changelog text for the release. |
| `SEMREL_DRY_RUN` | Whether semrel is running in dry-run mode. |

## Example behavior

The plugin sends a formatted Slack message with the release tag and changelog summary. In dry-run mode it prints the payload instead of calling the webhook.

## License

Apache-2.0
