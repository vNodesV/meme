# Project Plan

This plan consolidates the immediate and mid-term actions for the SDK v0.50.14 migration and CosmWasm/wasmvm v2 upgrade, focusing on devnet validation and production readiness.

## Priorities

### P0 — Verification & Safety (Immediate)

- Run full build + unit tests.
- Devnet upgrade rehearsal (single-node first, then multi-validator).
- Verify IBC + wasm invariants post-upgrade.

### P1 — wasmvm v2 Completion (Near-Term)

- Choose one path:
  - **Rebase on wasmd v0.54.x** (recommended), or
  - **Complete surgical fixes** (StoreAdapter + VM config + call sites).

### P2 — CI/CD & Security (Short-Term)

- Ensure `govulncheck` and SBOM generation for releases.
- Add upgrade rehearsal to CI (if available).
- Multi-arch build validation (linux/amd64, linux/arm64).

### P3 — Documentation & Operability (Ongoing)

- Keep migration and troubleshooting docs updated.
- Document upgrade rehearsal results in `MIGRATION.md`.

## Suggested Timeline

| Phase | Target | Output |
|------|--------|--------|
| P0 | This week | Verified builds + devnet upgrade log |
| P1 | This week | wasmvm v2 completion path chosen |
| P2 | Next sprint | CI checks + SBOM + security scan |
| P3 | Ongoing | Updated docs + operator guidance |

## Success Criteria

- `go build ./...` and `make install` pass.
- Devnet upgrade completes without panics.
- wasm contract operations (store/instantiate/execute/query/migrate) succeed.
- IBC client + transfer checks pass after upgrade.

## Attribution

- Consolidation and edits: [CP]
