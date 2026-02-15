# Security Policy

This repository follows a layered security approach for dependencies, build integrity, and chain operations.

## Reporting Vulnerabilities

If you discover a security issue, **do not open a public issue**. Instead, contact the maintainers through a private channel agreed by the team. Include:

- A clear description of the issue
- Affected version/commit (if known)
- Steps to reproduce
- Impact assessment

## Dependency & Supply-Chain Checks

### govulncheck

Run locally before releases:

```bash
go install golang.org/x/vuln/cmd/govulncheck@latest
govulncheck ./...
```

**Mitigation workflow:**

1. **Confirm reachability** — If a vulnerable function is not reachable, document the rationale.
2. **Upgrade path** — Prefer upgrading the affected module in `go.mod`.
3. **Compensating controls** — Add guardrails/tests if upgrade is blocked.
4. **Follow-up** — Track and remove workarounds when upstream fixes land.

### SBOM

Generate an SBOM for releases:

```bash
syft packages dir:. -o spdx-json > sbom.spdx.json
```

Publish SBOMs alongside release artifacts.

## Secrets & Keys

- **Never** commit private keys, mnemonics, or validator keys.
- Keep `priv_validator_key.json` and keyring data outside git.
- Use environment variables or secret managers for credentials.

## Build & Release Hardening

- Prefer pinned Go toolchains (see `go.mod` toolchain version).
- Use reproducible builds where possible.
- For containers, use pinned base images and non-root runtime users.

## Operational Security

- Follow staged upgrade rehearsals (see `UPGRADE.md`).
- Validate wasm contract invariants before/after upgrades.
- Monitor consensus and IBC logs during upgrades.

## Attribution

- Consolidation and edits: [CP]
