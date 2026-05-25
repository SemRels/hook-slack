# hook-slack

Slack notification hook plugin for SemRel.

Publishes release notifications and status updates to Slack channels after SemRel runs.

## Documentation

- SemRel docs (planned): <https://github.com/SemRels/semrel/tree/main/docs/plugins/hook-slack>
- Plugin template: <https://github.com/SemRels/plugin-template>
- Registry: <https://registry.semrel.io>

## Repository Layout

~~~text
cmd/plugin/              Plugin entry point
internal/plugin/         Business logic scaffold
internal/grpc/           gRPC transport scaffold
proto/v1                 Symlink to the SemRel protobuf contract
.github/workflows/       CI, release, and security automation
~~~

## Development

~~~bash
go build ./cmd/plugin
go test ./...
~~~

## Configuration Example

~~~yaml
plugins:
  - name: hook-slack
    type: hook
    config:
      webhook_url: ${SLACK_WEBHOOK_URL}
      channel: '#releases'
      notify_on:
        - success
        - failure
~~~

## Status

This repository is bootstrapped from SemRels/plugin-template and is ready for implementation.
