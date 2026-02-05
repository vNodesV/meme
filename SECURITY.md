# Security & Vulnerability Management

This repository uses layered checks to reduce supply-chain and dependency risk.

## govulncheck

CI runs `govulncheck ./...` on every push and PR. For local runs:

```bash
go install golang.org/x/vuln/cmd/govulncheck@latest
govulncheck ./...
```

### Mitigation workflow

1. **Confirm reachability**: `govulncheck` reports whether a vulnerable function is reachable. If it is not reachable, document the rationale and track for upgrade.
2. **Upgrade path**: Prefer upgrading the affected module to a fixed version in `go.mod`.
3. **Compensating controls**: If an upgrade is blocked, add tests or guardrails around the affected code path and document why the risk is acceptable.
4. **Follow-up**: Add a ticket to remove the workaround once the upstream fix is available.

## SBOM guidance

Generate an SBOM for release artifacts using `syft` or `cosign`:

```bash
syft packages dir:. -o spdx-json > sbom.spdx.json
```

Publish the SBOM alongside release artifacts.
