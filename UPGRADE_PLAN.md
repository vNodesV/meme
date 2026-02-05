# Meme Chain Upgrade Plan (meme-1)

## 1) Executive overview

This plan delivers a staged, in-place upgrade for meme-1 that preserves CosmWasm state while modernizing the stack to Cosmos SDK v0.53.x and CometBFT >= v0.38.21. It uses a two-hop upgrade to safely traverse store migrations from the live 0.45-era chain to a 0.47-era intermediate (aligned with the current repository), and then to a 0.53-era final target with IBC v10 and up-to-date wasmvm/wasmd. The plan emphasizes deterministic binary reproduction, explicit upgrade handlers, and operational guardrails (devnet rehearsal, cosmovisor support, and post-upgrade acceptance tests). It also adds CI security checks (govulncheck), multi-arch builds, and Docker hardening to reduce operational risk and supply-chain exposure. This is modeled on practices seen in cheqd-node and elys-network CI/security workflows while staying aligned to this repo’s structure and requirements.

## 2) Version matrix table

| Hop | Go | Cosmos SDK | Consensus (Tendermint/CometBFT) | IBC-Go | Wasmd | WasmVM | DB (cometbft-db) | gRPC/Proto toolchain |
| --- | --- | --- | --- | --- | --- | --- | --- | --- |
| Hop 1 (live → intermediate) | 1.23.2 (repo toolchain requirement) | 0.47.13 (repo baseline) | CometBFT 0.37.5 | ibc-go/v7 7.10.0 | wasmd 0.45.x-compatible (current repo) | wasmvm 1.0.1 | cometbft-db 0.9.5 | buf + gogoproto + grpc-gateway v1 |
| Hop 2 (intermediate → final) | 1.22.10+ | 0.53.x | CometBFT >= 0.38.21 | ibc-go/v10.x | wasmd 0.54+ | wasmvm 2.x (matching wasmd) | cometbft-db 0.9.x/0.10.x | buf + protoc-gen-go v1.34+ + grpc-gateway v2 |

## 3) Staged migration plan (commands + ordering)

### Hop 0: Identify/build the exact live binary

1. Pull the live commit hash from `/rest/node_info` (reported as `3d3bb097154af6a8eaa83f43e8e47dc91dcdb8b2`).
2. In this repo (or a tagged release), locate the commit or release that matches that SHA (e.g., `git log --all --grep 3d3bb097` or matching a tag to the commit).
3. Build a deterministic binary with matching build flags (record `memed version --long` including build tags `netgo,ledger`, commit, and Go version `go1.22.10`).
4. Replay the chain in a local state-sync or snapshot-based test to ensure the binary replays current state without divergence.

If the exact commit is unavailable, reproduce a binary using the build metadata from existing nodes (Go version, build tags, and module versions) and validate using state sync on a test node before upgrading.

### Hop 1: Upgrade live chain to SDK 0.47-era compatible stack

Goal: Move from SDK 0.45.1/Tendermint 0.34.16/ibc-go v2.2.0/wasmvm v1.0.0-beta10 to the current repo baseline (SDK 0.47.13/CometBFT 0.37.5/ibc-go v7/wasmvm 1.0.1).

1. Implement an x/upgrade handler for the Hop 1 name (e.g., `v1.1.0-hop1`) and register required store migrations in app wiring (`app/app.go` and `cmd/memed`). 
2. Ensure wasm module config is compatible with wasmvm 1.0.1 and retains existing code IDs (1–5).
3. Compile the Hop 1 binary and test against a snapshot or exported state.
4. On-chain governance proposal: submit a software upgrade with height H and name `v1.1.0-hop1`.
5. Validators halt at height H, swap binary, run `memed start` (or via cosmovisor), and verify blocks resume.

### Hop 2: Upgrade to Cosmos SDK v0.53.x + CometBFT >= 0.38.21

1. Create a new branch for v0.53 migration; bump module versions (SDK v0.53.x, ibc-go/v10, cometbft >= 0.38.21, wasmd/wasmvm pair). 
2. Implement the SDK 0.53 store migrations and IBC v10 migrations in a new upgrade handler (e.g., `v2.0.0`).
3. Update protobuf tooling (buf/protoc plugins) and regen types if required.
4. Rebuild, run migration tests on devnet, and ensure CosmWasm code and storage survive.
5. Submit upgrade proposal `v2.0.0`, halt at H2, deploy new binary, and resume.

## 4) Code hotspot checklist (repo paths + what edits)

- `go.mod` / `go.sum`: version bumps for SDK, IBC, CometBFT, wasmvm, and toolchain alignment. 
- `app/` (e.g., `app/app.go`): register upgrade handlers, module wiring, store migrations, and params changes.
- `cmd/` (e.g., `cmd/memed`): upgrade handler registration, version info output, and CLI flags that affect startup.
- `x/` modules: custom modules must implement `MigrateStore` logic or update protobuf types as needed.
- `x/upgrade` handlers: add Hop 1 and Hop 2 upgrade handlers with explicit migration ordering.
- `x/wasm` integration: ensure wasm module config (code, pinning, gas) preserves code IDs 1–5 and contract state.
- `proto/` and `buf.work.yaml`: update buf/proto tooling, gogo/gateway versioning, and regenerate.
- `Makefile`/`scripts/`: update build targets for multi-arch, include govulncheck.
- `.github/workflows/*`: add govulncheck, multi-arch docker build, update Go tooling.
- `Dockerfile` + `docker/`: harden runtime (non-root) and pin base images/tags.

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

At height H, stop old binary, switch to new binary (cosmovisor or manual), and resume. Repeat for Hop 2 (v2.0.0).

## 7) In-place upgrade plan for meme-1

- **Upgrade names**: `v1.1.0-hop1` (SDK 0.47) and `v2.0.0` (SDK 0.53).
- **Height selection**: choose a future height with at least 7 days lead time; confirm with validators.
- **Governance template**: include title, summary, `upgrade_height`, `binary`, and rollback plan.
- **Validator checklist**:
  - Cosmovisor: install new binary under `cosmovisor/upgrades/<name>/bin/memed`.
  - Non-cosmovisor: download binary, verify checksum, stop service, swap, restart at height H.
  - Confirm `memed version --long` after restart.

## 8) Post-upgrade acceptance tests (explicit)

### Wasm invariants

1. Verify code IDs 1..5 exist:
   - `curl -s "<REST>/cosmwasm/wasm/v1/code?pagination.limit=100"` includes 1..5.
2. Golden contract info/history:
   - `curl -s "<REST>/cosmwasm/wasm/v1/contract/meme14hj2tavq8fpesdwxxcu44rty3hh90vhujrvcmstl4zr3txmfvw9svav74l/history"` returns entries pre/post upgrade.
3. Safe query: if schema unknown, derive it from `contract_info` and the contract’s `schema` or `raw` storage keys. Use `memed query wasm contract-state all <contract>` to list and then re-query with known keys (avoid guessing).

### Chain invariants

1. Blocks continue for >= 200 blocks.
2. No unintended params drift (compare `app params` pre/post upgrade).

## 9) Risks & mitigations (ranked)

1. **State migration failures**: mitigate with snapshot-based rehearsal and explicit upgrade handlers for each module.
2. **CosmWasm state corruption**: verify code IDs and contract history; test golden contract queries in devnet.
3. **IBC channel/connection regressions**: use IBC v7 → v10 migration steps and run IBC e2e in devnet.
4. **Binary mismatch during Hop 0**: confirm binary build metadata and replay state before scheduling upgrade.
5. **Operational coordination**: enforce upgrade window, height, and cosmovisor readiness checks.

## 10) Security verification checklist

- CometBFT >= 0.38.21 included in Hop 2 dependency bump.
- Run `govulncheck ./...` in CI (and locally before releases).
- Generate SBOM with `syft`/`cosign` and publish with release artifacts.
- Docker hardening: pinned base images/tags and non-root runtime user.
- Multi-arch builds: linux/amd64 + linux/arm64 in CI.
