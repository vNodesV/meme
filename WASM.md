# CosmWasm / wasmvm v2 Status

This document summarizes wasmvm v2.2.1 migration status and recommended next steps.

## Current Status

- ✅ 7/12 breaking changes fixed (type renames, GoAPI updates, events, vote field).
- ⚠️ 5 critical issues remain if following the surgical path.

## Recommended Path

**Rebase on wasmd v0.54.x** (recommended):
- Production-tested integration of wasmvm v2
- Less risk than a full surgical migration

**Surgical path** (if rebasing is not possible):
1. Implement StoreAdapter for SDK ↔ wasmvm iterator compatibility.
2. Update VM initialization to `NewVMWithConfig`.
3. Rename RequiredFeatures → RequiredCapabilities.
4. Update engine interface and VM call sites.

## Files Involved

- `x/wasm/keeper/keeper.go` (VM init + call sites)
- `x/wasm/types/wasmer_engine.go` (StoreAdapter + engine interface)
- `x/wasm/keeper/*` (type fixes already applied)

## Testing Expectations

- Build: `go build ./x/wasm/...`
- Unit tests: `go test ./x/wasm/...`
- Contract smoke tests: store → instantiate → execute → query → migrate
- IBC wasm tests if enabled

## Decision Checklist

- Do you need to preserve custom wasm changes? (If no → rebase)
- Do you need to stay close to upstream? (If yes → rebase)
- Do you have time for 10–12 hours of surgical fixes? (If no → rebase)

## Attribution

- Consolidation and edits: [CP]
