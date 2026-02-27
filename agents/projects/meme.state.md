# meme.state.md — MeMe Chain Project Memory

**Last updated:** 2026-02-27  
**Session:** 4fd08761-e52c-4507-b7eb-c8811168c857  
**Agent:** jarvis5.0

---

## Project Identity

| Field | Value |
|-------|-------|
| Module | `github.com/CosmWasm/wasmd` |
| Chain ID (mainnet) | `meme-1` |
| Chain ID (devnet) | `meme-offline-0` |
| Denom | `umeme` |
| Binary | `memed` |
| Go version | 1.23.2 |
| Toolchain | go1.25.7 |
| SDK | v0.50.14 (cheqd fork: `github.com/cheqd/cosmos-sdk v0.50.14-height-mismatch-iavl.*`) |
| CometBFT | v0.38.19 |
| IBC | ibc-go/v8 v8.7.0 |
| CosmWasm | wasmvm v2.2.1 |
| Upgrade name | `sdk50` |

---

## Architecture Overview

```
meme/
├── app/           — chain wiring, keepers, params, upgrades
│   ├── app.go     — module manager, store keys, keepers
│   ├── upgrades.go — upgrade handler (sdk50)
│   ├── params/proto.go — MakeEncodingConfig with NewInterfaceRegistryWithOptions
│   ├── ante.go    — ante handler
│   ├── export.go  — genesis export
│   └── keeper_adapters.go
├── cmd/memed/     — CLI entrypoint, root.go, AutoCLI integration
│   ├── root.go    — NewRootCmd, initAutoCliOptions
│   └── genaccounts.go
├── x/wasm/        — CosmWasm module fork
│   ├── keeper/keeper.go — VM init + call sites
│   └── types/wasmer_engine.go — engine interface + StoreAdapter
├── proto/         — protobuf definitions
└── agents/        — jarvis agent files
    └── projects/  — per-project state (this file)
```

---

## SDK 0.50 Migration Status

### ✅ Completed

| Area | Details |
|------|---------|
| App wiring | Store services via `runtime.NewKVStoreService`; authority strings; address codecs |
| Consensus keeper | Integrated; `PreBlocker` calls `mm.PreBlock()` |
| Params migration | All module subspaces call `.WithKeyTable()`; IBC client fixed; baseapp.Paramspace registered |
| CLI | AutoCLI via `autocli.AppOptions.EnhanceRootCommand()`; 19 query subcommands |
| Tx CLI | Direct module CLI imports with Bech32 codecs (accCodec, valCodec) |
| Encoding | `NewInterfaceRegistryWithOptions` with proper AddressCodec/ValidatorAddressCodec in SigningOptions |
| Upgrade handler | `sdk50` — `baseapp.MigrateParams()` + `mm.RunMigrations()` + nil guard on `cp.Block` |
| Store loader | Crisis + Consensus store keys added to `UpgradeStoreLoader.Added` |
| Export | `app/export.go` updated for SDK 0.50 context/collections/error returns |
| Proto registry | `RegisterInterfaces` before `RegisterServices`; gogoproto alignment |
| Gov CLI | Migrated to v1 API with legacy content wrapper |
| `sdkerrors.Wrap` | Replaced with `fmt.Errorf("%w")` throughout |
| Test suite | `go test ./...` all green (app + benchmarks + x/wasm) |
| wasmvm v2 | Surgical fixes complete: StoreAdapter, NewVMWithConfig, RequiredCapabilities, engine interface, VM call sites |
| GenTxCmd codec | `validatorAddressCodec` (memevaloper prefix) passed — not accountAddressCodec |
| Devnet rehearsal | Two-binary rehearsal passed: V1 halts "UPGRADE NEEDED", V2 applies handler, chain resumes |

### 🔄 Remaining

| Priority | Area | Status | Notes |
|----------|------|--------|-------|
| P2 | CI/CD + security | 🔄 Pending | `govulncheck`, SBOM, multi-arch build |
| P2 | MIGRATION.md | 🔄 Pending | Document rehearsal results |
| P3 | IBC wasm tests | 🔄 Pending | Post-upgrade validation (ibctesting build tag) |

---

## wasmvm v2 Status

**Decision: Surgical path — COMPLETE.**
All 5 fixes applied in prior + current sessions:
1. ✅ `x/wasm/types/wasmer_engine.go` — `StoreAdapter` + `NewStoreAdapter` bridging SDK KVStore ↔ wasmvm iterator
2. ✅ `x/wasm/keeper/keeper.go` — VM init via `NewVMWithConfig`
3. ✅ `x/wasm/keeper/keeper.go` — `RequiredCapabilities` (was `RequiredFeatures`)
4. ✅ `x/wasm/types/wasmer_engine.go` — engine interface updated
5. ✅ `x/wasm/keeper/*.go` — all VM call sites use `types.NewStoreAdapter(prefixStore)`

`go build ./x/wasm/...` and `go test ./x/wasm/...` pass cleanly.

---

## Key Conventions

| Convention | Detail |
|------------|--------|
| Encoding config | `NewInterfaceRegistryWithOptions` with `SigningOptions{AddressCodec, ValidatorAddressCodec}` |
| Address codecs | Configured in `autocli.AppOptions`; NOT in `client.Context` (SDK 0.50 no such fields) |
| Params subspaces | All must call `.WithKeyTable()` in `initParamsKeeper` |
| Error wrapping | `fmt.Errorf("%w", err)` — no `sdkerrors.Wrap` |
| Upgrade name | `"sdk50"` — must match on-chain governance proposal exactly |
| Crisis store key | New in SDK 0.50; add to `UpgradeStoreLoader.Added` |
| Consensus store key | `consensuskeeper.StoreKey = "Consensus"` (capital C) |
| AutoCLI | Use `ModuleOptions` map (not `Modules`) for core SDK modules to avoid nil panic |
| Build | `make install` or `go build ./cmd/memed` |
| Test | `go test ./...` |
| GenTxCmd | Pass `validatorAddressCodec` (memevaloper prefix), NOT accountAddressCodec |
| Upgrade handler | Guard `cp.Block != nil` before `cp.Block.MaxBytes` — GetConsensusParams returns partial struct when legacy subspace has no BlockParams |
| Upgrade genesis flags | Requires `--title`, `--upgrade-info "{}"`, `--no-validate`, `--summary` for SDK 0.50 `tx upgrade software-upgrade` |
| Governance genesis | Set `expedited_voting_period < voting_period`; default expedited (24h) must be < test voting period |
| Two-binary rehearsal | V1 = stub (no handler); V2 = real binary; patch upgrade-info.json name to `_bypass` to skip StoreLoader when both binaries share same store keys |

---

## Critical File Map

| File | Purpose |
|------|---------|
| `app/app.go` | Module manager, keepers, store keys, `initParamsKeeper` |
| `app/upgrades.go` | `sdk50` upgrade handler — nil guard on `cp.Block` required |
| `app/params/proto.go` | `MakeEncodingConfig` with `NewInterfaceRegistryWithOptions` |
| `cmd/memed/root.go` | `NewRootCmd`, `initAutoCliOptions`, AutoCLI wiring; GenTxCmd uses validatorAddressCodec |
| `x/wasm/keeper/keeper.go` | VM init + call sites (wasmvm v2 complete) |
| `x/wasm/types/wasmer_engine.go` | Engine interface + StoreAdapter (wasmvm v2 complete) |
| `app/app_test.go` | Integration tests — uses `t.TempDir()` per WasmApp + `sims.GenesisStateWithValSet` |
| `benchmarks/app.go` | Benchmark helpers — `AppInfo`, `InitializeWasmApp`, `GenSequenceOfTxs` (SDK 0.50) |
| `benchmarks/bench_test.go` | Benchmarks migrated to ABCI 2.0 `FinalizeBlock` pattern |

---

## Recent Commits (at state creation)

```
61db04d  chore: update project state — all P0 tasks complete
dff6ee3  fix(cli+upgrade): validator address codec for GenTxCmd and nil guard in upgrade handler
82bc2d8  test: fix app + benchmarks test suite for SDK 0.50 / ABCI 2.0
5dd1ef3  chore: bump version to v2.1.0
4b799c0  fix(wasm/keeper): fund validator address in addValidator for SDK 0.50
064c111  fix(x/wasm): complete SDK 0.50 test suite migration
```

---

## Open Follow-Ups

- [ ] Add `govulncheck` to CI (P2)
- [ ] Multi-arch build validation: linux/amd64, linux/arm64 (P2)
- [ ] Document upgrade rehearsal results in `MIGRATION.md` (P2)
- [ ] IBC wasm tests post-upgrade (ibctesting build tag — 41 known errors) (P3)
- [ ] Consider PR to main from dev/v2.1.0 once CI passes

---

## Upgrade History

| Date | Session | Agent | Action |
|------|---------|-------|--------|
| 2026-02-27 | ce8afec7 | jarvis5.0 | Initial state file bootstrap (`new` command) |
| 2026-02-27 | 4fd08761 | jarvis5.0 | P0 sprint: test suite, devnet rehearsal, wasmvm v2 complete. Commits: 82bc2d8, dff6ee3, 61db04d |
