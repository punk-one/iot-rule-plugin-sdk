# AGENTS.md

This file is operational guidance for AI coding agents working in `iot-rule-plugin-sdk`.

## Mission

This repository is the stable public boundary between the engine and all plugins. Keep it small, generic,
transport-safe, and independent of any business domain. Engine internals belong in `iot-rule-engine`; S7,
MQTT, database, transformation, and customer logic belong in `iot-rule-plugins`.

## v1 compatibility is the primary constraint

- Version `v1.0.0` starts a minimum three-year no-breaking-change commitment.
- Do not remove, rename, reorder, or reinterpret exported API, serialized fields, RPC methods, error codes,
  handshake values, dispense keys, or defaults covered by `contract_v1.json`.
- Additive changes must remain optional and must preserve old engine/plugin interoperability.
- Prefer new optional interfaces, capability flags, or version-negotiated RPC methods over widening an
  existing interface method set.
- Never regenerate and accept contract files merely to make tests pass. Review every diff semantically.

## Ownership and concurrency

- `Event` values crossing ownership boundaries must be cloned. `Clone` must deep-copy every supported
  mutable JSON-compatible container and byte representation without JSON round-tripping numeric values.
- Plugin publishers retain ownership of their input after `Publish` returns; engine and transport layers
  must not alias caller-owned maps or slices.
- Runtime callbacks, logging, health, and metrics must be bounded and safe during concurrent stop.
- Every transport call requires a finite deadline. Timeout and stop paths must not leak goroutines.
- Keep public DTOs serializable and deterministic across supported Go versions.

## Repository map

- Root `.go` files: public API; changes here require contract review.
- `internal/core`: canonical types behind public aliases.
- `internal/transport`: RPC and go-plugin implementation.
- `internal/runtime`: context, emitter, logger, state, and metrics implementations.
- `internal/healthruntime`: asynchronous health delivery.
- `schema`: rule-flow and shared schema types.
- `examples`: compiling API examples.
- `tests`: black-box public contract tests.
- `docs`: normative runtime and compatibility documentation.

## Dependencies and security

- Avoid new dependencies in the SDK. Explain size, maintenance, license, and compatibility impact first.
- Never expose arbitrary filesystem, process, network, or environment access as a generic SDK convenience.
- Never log payloads, secrets, connector credentials, or authentication material by default.
- Keep all examples synthetic and free of real infrastructure addresses.
- Do not change license terms, tag, commit, or push unless explicitly authorized.
- The repository is `AGPL-3.0-only`. Preserve license notices and call out compatibility implications for
  applications or plugins that incorporate the SDK.

## Verification

Run all of the following for public or transport changes:

```bash
go test ./...
go test -race ./...
go vet ./...
go build ./examples/...
go run ./internal/tools/public_api_export.go
go run ./internal/tools/contract_snapshot
git diff --exit-code -- PUBLIC_API.md contract_v1.json
git diff --check
```

Keep English as the default `README.md` and Chinese in `README.zh-CN.md`. Update changelog and normative
docs when behavior changes, even when the Go signature does not.
