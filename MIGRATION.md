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

---

## Devnet Upgrade Rehearsal (meme-offline-0)

This section records the observed results of the live SDK 0.50 upgrade rehearsal executed on the
`meme-offline-0` devnet. It serves as the authoritative reference for the mainnet (`meme-1`)
upgrade runbook.

---

### Pre-Upgrade Environment

| Parameter | Value |
|---|---|
| Chain ID | `meme-offline-0` |
| Binary | v1.0.0 (Tendermint 0.34.x / SDK ~0.45.x) |
| Earliest block | height 1 — 2026-02-06T10:56:48Z |
| Consensus engine | Tendermint 0.34.x |

**Module versions at start (SDK 0.45 baseline):**

| Module | Version |
|---|---|
| auth | 2 |
| bank | 2 |
| gov | 2 |
| ibc | 2 |
| staking | 2 |
| wasm | 1 |

---

### Upgrade Execution

The upgrade followed the standard on-chain governance software upgrade proposal flow:

1. **Submit proposal** — a `MsgSoftwareUpgrade` proposal was submitted on-chain with upgrade
   name `sdk50` and plan height `1000`.
2. **Vote** — all active validators voted `YES`; proposal reached quorum and passed.
3. **Halt** — at block height **1000** the node halted with the expected upgrade-needed panic:
   ```
   UPGRADE "sdk50" NEEDED at height: 1000
   ```
4. **Binary swap** — the v1.0.0 binary was replaced with the v2.0.0 build (CometBFT 0.38.19,
   SDK 0.50.14).
5. **Restart** — the node was restarted with the new binary; the upgrade handler ran the
   in-place store migrations and the chain resumed block production.

---

### Post-Upgrade Verification

| Parameter | Value |
|---|---|
| Chain ID | `meme-offline-0` |
| App version | `v2.0.0` |
| CometBFT | `0.38.19` |
| Go | `1.25.7` |
| Block height at verification | 344,909+ (as of 2026-02-27) |
| Active validators | 3 (voting power 70 : 1 : 1) |
| Staking bond denom | `umeme` |
| Max validators | 100 |
| Unbonding time | 1,814,400 s (≈ 21 days) |

**Module versions before → after upgrade:**

| Module | Pre-upgrade | Post-upgrade |
|---|---|---|
| auth | 2 | 5 |
| bank | 2 | 4 |
| gov | 2 | 5 |
| ibc | 2 | 6 |
| staking | 2 | 5 |
| slashing | — | 4 |
| transfer | — | 5 |
| crisis | — | 2 |
| feegrant | — | 2 |
| mint | — | 2 |
| upgrade | — | 2 |
| consensus | — | 1 |
| wasm | 1 | 1 |

---

### Result

**PASS** — the chain did not panic during or after the upgrade handler execution. Block production
continued uninterrupted past height 1000. All store migrations applied cleanly and module versions
match the expected SDK 0.50.14 targets.

---

### Open Items

| # | Item | Detail |
|---|---|---|
| 1 | Wasm contracts not deployed | The `code_infos` endpoint is reachable and returns correctly, but 0 contracts have been uploaded to devnet post-upgrade. Contract deployment and execution smoke tests are pending. |
| 2 | `min_commission_rate` param discrepancy | The ante handler enforces a 5% minimum commission rate, but the on-chain staking param `min_commission_rate` is currently `0`. The two must be reconciled before mainnet upgrade (either update the param via governance or align the ante logic). |

---

### Next Steps

1. **Deploy `memeart` contract to devnet** — upload and instantiate the memeart CosmWasm
   contract on `meme-offline-0` to validate the wasm keeper end-to-end.
2. **Run full wasm smoke test** — execute store-code → instantiate → execute → query flow and
   confirm correct behaviour under SDK 0.50 / wasmvm v2.
3. Resolve the `min_commission_rate` discrepancy (open item #2) before mainnet runbook is
   finalised.
4. Promote rehearsal results to the mainnet upgrade runbook once smoke tests pass.
