# hook-slack

Slack hook plugin for Semantic Release.

Publishes Semantic Release notifications to Slack.

## Documentation

- Docs (coming soon): <https://github.com/SemRels/semrel/tree/main/docs/plugins/hook-slack>
- Template source: <https://github.com/SemRels/plugin-template>

## Repository Layout

`	ext
cmd/plugin/              Plugin entry point
internal/plugin/         Business logic scaffold
internal/grpc/           gRPC transport scaffold
proto/v1                 Symlink to the SemRel protobuf contract
.github/workflows/       CI, release, and security automation
`

## Development

`ash
go build ./cmd/plugin
go test ./...
`

## Configuration Example

`yaml
plugins:
  - name: hook-slack
    type: hook
    config:
      webhook_url: ${SLACK_WEBHOOK_URL}
      channel: '#releases'
      notify_on: success
`

## Status

This repository is bootstrapped from SemRels/plugin-template and is ready for implementation.