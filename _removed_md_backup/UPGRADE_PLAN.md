# Meme Chain Upgrade Plan (meme-1)

## 1) Executive overview (<= 1 page)

This plan upgrades meme-1 in two hops to reach a modern, security-patched Cosmos SDK v0.53.x + CometBFT >= 0.38.21 stack while preserving all CosmWasm code and storage (including code IDs 1–5 and the “MEME Art Service” contract). Hop 0 anchors the live chain binary at commit `3d3bb097…` and corresponds to the published release `v1.0.0_hop0`, ensuring deterministic replay and a safe starting point. Hop 1 upgrades to a 0.47-era stack (IBC v7, wasmvm 1.0.1) that is closest to the current repo wiring, minimizing store migration risk. Hop 2 then moves to SDK 0.53.x with IBC v10 and CometBFT 0.38.21+ (security-patched). Operationally, the plan includes explicit upgrade handlers, full devnet rehearsal, cosmovisor and manual upgrade paths, post-upgrade acceptance tests focused on wasm invariants, and CI hardening (govulncheck, multi-arch builds, pinned non-root Docker runtime). The goal is a safe, auditable, staged migration with actionable change points that map directly to repo files.

## 2) Version matrix table

| Hop | Go | Cosmos SDK | Consensus (Tendermint/CometBFT) | IBC-Go | Wasmd | WasmVM | DB (cometbft-db/tm-db) | gRPC/Proto toolchain |
| --- | --- | --- | --- | --- | --- | --- | --- | --- |
| Hop 1 (live → intermediate) | 1.22.10 (matches live build) | 0.47.13 | CometBFT 0.37.5 | ibc-go/v7 7.10.x | wasmd 0.45–0.50 range (SDK 0.47 compatible) | wasmvm 1.0.1 | cometbft-db 0.9.5 | buf + gogoproto + grpc-gateway v1 |
| Hop 2 (intermediate → final) | 1.22.10+ (or 1.23.x) | 0.53.x | CometBFT >= 0.38.21 | ibc-go/v10.x | wasmd 0.54+ | wasmvm 2.x (matching wasmd) | cometbft-db 0.9.x/0.10.x | buf + protoc-gen-go v1.34+ + grpc-gateway v2 |

## 3) Staged migration plan (commands + ordering)

### Hop 0: Identify/build the exact live binary

**Goal:** Reproduce the live binary and ensure it replays chain state deterministically.

1. Confirm live commit: `/rest/node_info` reports `commit=3d3bb097154af6a8eaa83f43e8e47dc91dcdb8b2`.
2. Use the `v1.0.0_hop0` release as the binary baseline; verify `memed version --long` matches the commit and Go toolchain.
3. Build with matching tags (`netgo,ledger`) and Go 1.22.10 (or the exact Go version reported by `version --long`).
4. Replay against a snapshot/state-sync in a local environment and confirm no divergence.

If you cannot find the commit in the repo history, reconstruct from build metadata and validate replay on a separate node before scheduling any on-chain upgrade.

### Hop 1: migrate live chain to SDK 0.47-era compatible stack

**Goal:** Move to SDK 0.47.13 + CometBFT 0.37.5 + IBC v7 + wasmvm 1.0.1.

1. Create upgrade handler `v1.1.0-hop1` (new x/upgrade handler module).
2. Wire migrations in `app/app.go` and `cmd/memed`:
   - Cosmos SDK 0.46/0.47 store migrations (auth, bank, staking, gov v1beta1, params).
   - IBC v2 → v7 module migrations (core, transfer, interchain accounts if enabled).
   - Wasm module migration for wasmvm 1.0.1 compatibility.
3. Compile the hop1 binary and run a snapshot-based migration test.
4. Governance:
   - Submit software upgrade with name `v1.1.0-hop1` at height H.
   - Validators halt at height H and swap binaries (cosmovisor or manual).
5. Validate blocks resume and wasm invariants hold.

### Hop 2: migrate to Cosmos SDK v0.53.x + CometBFT >= 0.38.21

**Goal:** land on a modern, supported stack with patched consensus.

1. Create hop2 upgrade handler `v2.0.0` and bump dependencies:
   - SDK v0.53.x
   - CometBFT >= 0.38.21
   - ibc-go/v10.x
   - wasmd 0.54+ with wasmvm 2.x
2. Update protobuf tooling (buf config and protoc-gen-go) and re-generate types.
3. Run devnet rehearsal and contract-state verification (code IDs 1–5 preserved).
4. Submit upgrade proposal `v2.0.0`, halt at H2, deploy new binary, resume.

## 4) Code hotspot checklist (repo paths + what edits)

- `go.mod` / `go.sum`: bump SDK, IBC, CometBFT, wasmvm/wasmd, toolchain versions; ensure replace directives are still required.
- `app/` (`app/app.go`): register hop1/hop2 upgrade handlers and module migration ordering.
- `cmd/` (`cmd/memed`): wire upgrade handlers, expose version metadata.
- `x/` modules: check custom modules for `MigrateStore` compatibility and protobuf updates.
- `x/upgrade`: add `v1.1.0-hop1` and `v2.0.0` handlers and migrations.
- `x/wasm`: ensure wasm module config and state migrations preserve code IDs 1–5.
- `proto/`, `buf.work.yaml`: bump buf/protoc plugins and regenerate.
- `Makefile`, `scripts/`: add govulncheck target, multi-arch build targets.
- `.github/workflows/*`: ensure govulncheck, multi-arch docker build, pinned Go.
- `Dockerfile`/`docker/`: harden runtime (non-root), pin base images/tags.

## 5) Offline sanity runbook

```bash
make build
./build/memed init local --chain-id meme-local
./build/memed start --home ~/.meme
```

Expected: node produces blocks locally; no panic at startup.

## 6) Single-server devnet runbook (multi-validator on one host)

1. Create N validator homes and unique ports.
2. `memed init` for each validator, update `app.toml`/`config.toml` ports.
3. Create validator keys, `gentx`, and `collect-gentxs`.
4. Start all validators with `persistent_peers` set.

Upgrade rehearsal:

```bash
# submit upgrade proposal with name v1.1.0-hop1 at height H
memed tx gov submit-proposal software-upgrade v1.1.0-hop1 --upgrade-height H ...
memed tx gov vote <proposal-id> yes ...
```

At height H, stop old binary, switch to new binary (cosmovisor or manual), and resume. Repeat for Hop 2 (`v2.0.0`).

## 7) In-place upgrade plan for meme-1

- **Upgrade names**: `v1.1.0-hop1` and `v2.0.0`.
- **Height selection**: choose a future height with >= 7 days lead time; confirm with validators.
- **Governance template**:
  - `title`: “meme-1 Hop1 upgrade to SDK 0.47” / “meme-1 Hop2 upgrade to SDK 0.53”.
  - `summary`: include upgrade name, binary checksum, and rollback steps.
  - `upgrade_height`: H (Hop 1), H2 (Hop 2).
  - `deposit`: standard network minimum.
- **Validator checklist**:
  - Cosmovisor: install new binary under `cosmovisor/upgrades/<name>/bin/memed`.
  - Non-cosmovisor: download binary, verify checksum, stop service, swap, restart at height.
  - Confirm `memed version --long` after restart.

## 8) Post-upgrade acceptance tests (explicit)

### Wasm invariants

1. Verify code IDs 1..5 exist:
   - `curl -s "<REST>/cosmwasm/wasm/v1/code?pagination.limit=100"` includes 1..5.
2. Golden contract info/history:
   - `curl -s "<REST>/cosmwasm/wasm/v1/contract/meme14hj2tavq8fpesdwxxcu44rty3hh90vhujrvcmstl4zr3txmfvw9svav74l/history"` returns entries pre/post upgrade.
3. Safe query (schema unknown):
   - Use `memed query wasm contract-info <contract>` and `memed query wasm contract-state all <contract>` to inspect available keys.
   - Derive a query from available keys (avoid guessing).

### Chain invariants

1. Blocks continue for >= 200 blocks.
2. No unintended params drift (compare `memed query params` pre/post upgrade).

## 9) Risks & mitigations (ranked)

1. **State migration failures**: mitigate with snapshot-based rehearsal and explicit upgrade handlers per hop.
2. **CosmWasm state corruption**: verify code IDs and contract history; run golden contract queries in devnet.
3. **IBC channel/connection regressions**: run IBC v7 → v10 migration rehearsal and client/state checks.
4. **Binary mismatch during Hop 0**: confirm build metadata and replay state before scheduling upgrade.
5. **Operational coordination**: enforce upgrade window, height, and cosmovisor readiness checks.

## 10) Security verification checklist

- CometBFT >= 0.38.21 included in Hop 2 dependency bump (security-patched consensus).
- Run `govulncheck ./...` in CI and pre-release.
- SBOM guidance: generate via `syft packages dir:. -o spdx-json` and publish with release artifacts.
- Docker hardening: pinned base images/tags, non-root runtime user, minimal packages.
- Multi-arch builds: linux/amd64 + linux/arm64 in CI.
