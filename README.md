# IoT Rule Engine Plugin SDK

[English](README.md) | [简体中文](README.zh-CN.md)

The IoT Rule Engine Plugin SDK defines the stable process, event, health, state, connector, and RPC
contracts used by IoT Rule Engine plugins. Version `v1.0.0` is the first public contract baseline.

## Stability commitment

The v1 public API and wire contract are intended to remain backward compatible for at least three years.
Compatible additions may be released in v1 minor versions, but existing exported interfaces, fields,
constants, handshake values, RPC methods, and behavior must not be removed or reinterpreted.

Compatibility is enforced by:

- `contract_v1.json`: complete API and RPC contract snapshot.
- `public_api.json` and `PUBLIC_API.md`: exported Go API baseline.
- Black-box compatibility, handshake, ownership, race, and transport tests.
- Generated-document checks in CI.

See [`docs/v1-compatibility-policy.md`](docs/v1-compatibility-policy.md) for the normative policy.

## Features

- A common `Plugin` interface for sources, processors, and sinks.
- `ServeFactory` for one OS process with isolated per-rule-node plugin instances.
- A connector resource API with unary and streaming operations.
- JSON-compatible events with explicit clone and ownership semantics.
- Buffered logging, metrics, state, acknowledgements, and health reporting.
- Deadline-aware RPC transport with legacy handshake compatibility.
- Rule/schema types shared with the engine.

## Install

```bash
go get github.com/punk-one/iot-rule-plugin-sdk@v1.0.0
```

## Minimal processor

```go
package main

import sdk "github.com/punk-one/iot-rule-plugin-sdk"

type Processor struct {
	sdk.BasePlugin
	ctx sdk.RuntimeContext
}

func (p *Processor) Info() sdk.PluginInfo {
	return sdk.PluginInfo{
		ID: "example-processor", Name: "Example Processor", Version: "1.0.0",
		Type: sdk.PluginTypeProcessor,
	}
}

func (p *Processor) Start(ctx sdk.RuntimeContext) error {
	p.ctx = ctx
	return ctx.Health().Healthy("processor started", nil)
}

func (p *Processor) OnEvent(event sdk.Event) error {
	owned := event.Clone()
	owned.Payload["processed"] = true
	return p.ctx.Emitter().Publish(owned)
}

func main() {
	sdk.ServeFactory(func() sdk.Plugin { return &Processor{} }, sdk.ServeOptions{})
}
```

Each rule node receives a distinct `Processor` instance. Do not use package globals for rule-specific
configuration or mutable state.

## Event ownership

Treat incoming events as borrowed values. Clone before retaining or mutating an event, and create an
independent clone for every concurrent downstream branch. Rule payloads should use JSON-compatible data;
mutable byte slices and maps must never be shared across asynchronous owners.

Detailed runtime rules are in [`docs/runtime-contracts.md`](docs/runtime-contracts.md).

## Connector plugins

Connector plugins manage reusable resource records through a `provider_handle`. Protocol sessions are an
implementation detail and should follow demand: persistent for protocols such as MQTT, scheduled reuse for
high-frequency PLC scans, and immediate close when broken, stopped, or no longer needed.

Start with [`docs/plugin_template.md`](docs/plugin_template.md) and the runnable examples under `examples/`.

## Repository layout

- Root Go files: stable exported API.
- `internal/core`: canonical internal models.
- `internal/transport`: go-plugin transport, RPC DTOs, and handshake.
- `internal/runtime`: runtime context implementations.
- `schema`: rule and message schema types.
- `examples`: compilable source, processor, and health examples.
- `tests`: public API and compatibility black-box tests.
- `docs`: compatibility policy and integration contracts.

## Verification

```bash
go test ./...
go test -race ./...
go vet ./...
go build ./examples/...
go run ./internal/tools/public_api_export.go
go run ./internal/tools/contract_snapshot
git diff --exit-code -- PUBLIC_API.md contract_v1.json
```

Any generated contract diff requires an explicit compatibility review. A passing compiler is not enough
to prove v1 compatibility.

## Security

The SDK does not grant plugins unrestricted access to the engine process. Plugins run out of process and
receive only the runtime capabilities explicitly exposed by the contract. Report vulnerabilities according
to [`SECURITY.md`](SECURITY.md).

## License

The SDK is licensed under the [GNU Affero General Public License v3.0 only](LICENSE)
(`AGPL-3.0-only`). Projects that incorporate or distribute the SDK should review their AGPL compliance
obligations before release. The AGPL permits commercial use; it does not grant trademark rights.
