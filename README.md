# MeMe Chain

Cosmos SDK chain with CosmWasm smart contracts powering an NFT marketplace and art services platform. This repository tracks the SDK v0.50.14 migration and the wasmvm v2.2.1 upgrade path.

## Snapshot

- **Chain IDs**: `meme-1` (mainnet), `meme-offline-0` (devnet)
- **Denom**: `umeme`
- **SDK**: v0.50.14 (cheqd fork with height-mismatch patches)
- **CometBFT**: v0.38.19
- **CosmWasm**: wasmvm v2.2.1
- **IBC**: ibc-go/v8 v8.7.0

## Status at a Glance

- ✅ App wiring, keeper initialization, CLI updates, and export paths migrated to SDK 0.50 patterns.
- ✅ Critical params migration fix added for IBC client subspace registration.
- ⚠️ wasmvm v2 migration has two paths: **rebase on wasmd v0.54.x** (recommended) or **finish surgical fixes** (see `WASM.md`).
- 🔄 Full integration testing and devnet upgrade rehearsal are the next operational steps.

## Quick Start (Dev)

go1.24.12+ is required.
1. git pull origin main && git checkout v2.0.0-vNodesAI
2. make install
3. memed init <moniker> --chain-id meme-offline-0
4. Download devnet genesis: `curl -o ~/.memed/config/genesis.json https://raw.githubusercontent.com/memecoin/meme/main/devnet/genesis.json`
5. Snapshots will be availalble "soon". In the meantime, you can initialize a local node and wait for it to sync from genesis.

For production/mainnet setup (genesis, peers, gas settings, and systemd), see `UPGRADE.md`.

## Documentation Map

- `PLAN.md` — prioritized action plan and roadmap.
- `MIGRATION.md` — consolidated SDK 0.50 migration summary + file map.
- `UPGRADE.md` — devnet/mainnet upgrade runbooks.
- `WASM.md` — wasmvm v2 migration status and options.
- `TROUBLESHOOTING.md` — error lookup and fixes.
- `SECURITY.md` — security posture and reporting guidance.

## Repo Layout (Top-Level)

- `app/` — chain wiring, keepers, params, upgrades
- `cmd/memed/` — CLI and server entrypoint
- `x/wasm/` — CosmWasm module and keeper logic
- `proto/` — protobuf definitions
- `scripts/` — build and tooling scripts

## Contributing

Please open issues or PRs with clear reproduction steps and logs. For security concerns, follow `SECURITY.md`.

## Attribution

- Consolidation and edits: [CP]



