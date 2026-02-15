# Migration Summary (SDK v0.50.14)

This document consolidates the SDK v0.50.14 migration work across app wiring, CLI, export, and wasmvm v2 readiness. It references the concrete files updated in the repo.

## Scope

- Cosmos SDK: v0.50.14 (cheqd fork with height-mismatch patches)
- CometBFT: v0.38.19
- IBC-go: v8.7.0
- wasmvm: v2.2.1

## Completed Work

### App Wiring & Keepers

- Store services migrated to `runtime.NewKVStoreService`.
- Authority strings and address codecs added.
- Consensus params keeper integrated.
- Capability keeper wired for IBC.

**Key files:**
- `app/app.go`
- `app/ante.go`
- `app/encoding.go`
- `app/genesis.go`
- `app/keeper_adapters.go`

### Params Migration Fix

- IBC client params subspace now registers `ParamKeyTable()` to avoid upgrade panic.

**Key files:**
- `app/app.go` (initParamsKeeper)

### CLI & Server Entry

- Updated cmd/memed for SDK 0.50 signature changes.
- Updated wasm CLI commands to new gov v1 submission pattern.

**Key files:**
- `cmd/memed/root.go`
- `cmd/memed/main.go`
- `cmd/memed/genaccounts.go`
- `x/wasm/client/cli/*.go`

### Export Flow

- `app/export.go` updated for SDK 0.50 context, collections, and error returns.

**Key files:**
- `app/export.go`

### RegisterInterfaces / Proto Compatibility

- Ensured `RegisterInterfaces` is called before `RegisterServices`.
- Resolved protobuf registry mismatch by aligning gogoproto usage.

**Key files:**
- `app/app.go`
- `x/wasm/types/*.pb.go`
- `x/wasm/types/codec.go`

## wasmvm v2 Migration Status

- 7/12 breaking changes fixed.
- 5 issues remain if using surgical path: VM initialization, RequiredCapabilities rename, engine interface, StoreAdapter, and VM call-site updates.

**Key files:**
- `x/wasm/keeper/keeper.go`
- `x/wasm/types/wasmer_engine.go`

See `WASM.md` for the recommended path and remaining tasks.

## Testing Status (Recommended)

- Build: `go build ./...`
- Unit tests: `go test ./...`
- Integration: start local node + wasm contract smoke tests
- Devnet upgrade rehearsal (single-node, then multi-validator)

## Remaining Work

- Choose wasmvm v2 completion path (rebase vs surgical).
- Run devnet upgrade rehearsal and document results.
- Full test suite + security scans.

## Attribution

- Consolidation and edits: [CP]
