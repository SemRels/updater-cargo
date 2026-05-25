# updater-cargo

`updater-cargo` is a SemRel plugin that updates a Rust crate version and publishes it with Cargo.

## Configuration

Environment variables:

- `CARGO_TOKEN` (exported to Cargo as `CARGO_REGISTRY_TOKEN`)

## Behavior

The plugin runs:

1. `cargo set-version <version>`
2. `cargo publish --no-verify`

## Development

```bash
go mod tidy
go build ./...
go test ./...
```
