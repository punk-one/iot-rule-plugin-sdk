# Contributing

## Development checks

Use Go from `go.mod` and run the following before submitting a change:

```bash
go test ./...
go test -race ./...
go vet ./...
go build ./examples/...
go run ./internal/tools/public_api_export.go
go run ./internal/tools/contract_snapshot
git diff --exit-code -- PUBLIC_API.md contract_v1.json
```

## v1 compatibility review

The existing public interfaces are frozen. New capabilities must use a new,
small optional interface or a new RPC method with capability negotiation. A
change that intentionally updates `contract_v1.json` must explain:

- why it is backward compatible;
- how old plugin binaries behave with the new engine;
- how new plugins degrade on the oldest supported v1 engine;
- which source, wire and behavior fixtures were added;
- whether timeout, retry, Ack, queue or shutdown semantics changed.

Do not rewrite a published tag. Keep unrelated generated or local files out of
the change and never commit credentials.
