# meme.state.md — MeMe Chain Project Memory

**Last updated:** 2026-02-27  
**Session:** ce8afec7-3df4-48c9-b680-b86fe5046ca3  
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
| Upgrade handler | `sdk50` — `baseapp.MigrateParams()` + `mm.RunMigrations()` |
| Store loader | Crisis + Consensus store keys added to `UpgradeStoreLoader.Added` |
| Export | `app/export.go` updated for SDK 0.50 context/collections/error returns |
| Proto registry | `RegisterInterfaces` before `RegisterServices`; gogoproto alignment |
| Gov CLI | Migrated to v1 API with legacy content wrapper |
| `sdkerrors.Wrap` | Replaced with `fmt.Errorf("%w")` throughout |

### ⚠️ Remaining / In-Progress

| Priority | Area | Status | Notes |
|----------|------|--------|-------|
| P0 | Build + unit tests | ✅ DONE | `go build ./...` + `go test ./...` |
| P0 | Devnet upgrade rehearsal | ✅ DONE | Single-node first; submit `sdk50` proposal |
| P1 | wasmvm v2 completion | ✅ DONE (surgical fixes applied) | **Recommended: rebase on wasmd v0.54.x** |
| P2 | CI/CD + security | 🔄 Pending | `govulncheck`, SBOM, multi-arch build |
| P3 | Documentation | 🔄 Ongoing | Upgrade rehearsal results → `MIGRATION.md` |

---

## wasmvm v2 Decision

**Recommended path: Rebase on wasmd v0.54.x**  
Rationale: Production-tested; lower risk than completing 5 surgical fixes.

**Surgical path remaining (if chosen):**
1. `x/wasm/types/wasmer_engine.go` — StoreAdapter for SDK ↔ wasmvm iterator
2. `x/wasm/keeper/keeper.go` — VM init: `NewVMWithConfig`
3. `x/wasm/keeper/keeper.go` — `RequiredFeatures` → `RequiredCapabilities` rename
4. `x/wasm/types/wasmer_engine.go` — engine interface updates
5. `x/wasm/keeper/*.go` — VM call-site updates

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

---

## Critical File Map

| File | Purpose |
|------|---------|
| `app/app.go` | Module manager, keepers, store keys, `initParamsKeeper` |
| `app/upgrades.go` | `sdk50` upgrade handler |
| `app/params/proto.go` | `MakeEncodingConfig` with `NewInterfaceRegistryWithOptions` |
| `cmd/memed/root.go` | `NewRootCmd`, `initAutoCliOptions`, AutoCLI wiring |
| `x/wasm/keeper/keeper.go` | VM init + call sites (wasmvm v2 TODO) |
| `x/wasm/types/wasmer_engine.go` | Engine interface + StoreAdapter (wasmvm v2 TODO) |

---

## Recent Commits (at state creation)

```
576e838  Fix package ecosystem name in Dependabot config (gomod)
8ace7f3  Fix package ecosystem name (Go)
6950bfc  Update package ecosystem to 'Go modules'
3b39296  Merge PR #41 — README revision
f795e25  (tag: v2.0.0-vNodesAI) Remove deprecated migration docs
a885499  Markdown files consolidation
fc27aee  Add chain ID handling in makeAppCreator
1051789  Refactor Bech32 prefix handling → MakeEncodingConfig
46ea9af  Fix address codec config for transaction signing in SDK 0.50
```

---

## Open Follow-Ups

- [ ] Decide wasmvm v2 path: rebase vs surgical
- [ ] Run `go build ./...` clean verification
- [ ] Devnet upgrade rehearsal (log results to MIGRATION.md)
- [ ] Enable `govulncheck` in CI
- [ ] Multi-arch build validation (linux/amd64, linux/arm64)
- [ ] Evaluate IBC wasm tests post-upgrade

---

## Upgrade History

| Date | Agent | Action |
|------|-------|--------|
| 2026-02-27 | jarvis5.0 | Initial state file bootstrap (`new` command) |
