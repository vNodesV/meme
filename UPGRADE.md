# Upgrade Runbook

This runbook covers devnet and mainnet upgrade rehearsals for SDK v0.50.14 and wasmvm v2.

## Pre-Upgrade Checklist

- Build passes: `go build ./...`
- Binary version verified: `memed version --long`
- Genesis and app config validated
- wasm contracts smoke-tested (store/instantiate/execute/query)
- IBC client params subspace registered (IBC client `ParamKeyTable`)

## Devnet Upgrade (Single Node)

1. Start node with the **old** binary and create state.
2. Submit software-upgrade proposal with a future height.
3. Let the node halt at the upgrade height.
4. Swap to the **new** binary.
5. Restart and verify blocks resume.
6. Run wasm + IBC checks.

## Devnet Upgrade (Multi-Validator (3))

    Phase 1:
    1. Initialize 3 validator homes using SDK 0.45.1 (memed v1.0.0, forked from CosmosMEME's github) and produce a shared genesis.
    2. Start all validators with the old binary.
    3. Submit upgrade proposal and vote yes.
    4. At upgrade height, stop all nodes.
    5. Swap binaries (cosmovisor recommended).
    6. Restart and verify consensus.

    Phase 2:
    1. Have multiple active mainnet validators reviewing and testing the upgraded binary on devnet.
    2. Collect feedback and address any issues before mainnet upgrade.
    3. Document devnet upgrade outcomes and lessons learned in `MIGRATION.md` for mainnet reference.
    4. Prepare testnet upgrade announcement and checklist based on devnet experience.

## Testnet Upgrade (Mult-Validator)
1. Prepare proposal and voting process similar to mainnet.
2. Coordinate with testnet validators for upgrade timing and support.
3. Monitor testnet upgrade closely and document any issues or observations.
4. Use testnet upgrade as a final rehearsal before mainnet, ensuring all critical paths are validated.

## Mainnet Upgrade (High-Level)

- Announce upgrade window with validator checklist.
- Provide checksums and binary release notes.
- Use cosmovisor paths for automatic swap.
- Verify:
  - Block production
  - wasm code IDs and contract history
  - IBC client status

## Post-Upgrade Validation

- Verify code IDs and contract history via wasm queries.
- Ensure IBC clients and channels are active.
- Monitor logs for param migration errors.

## Operational Notes

- Do not expose keys in logs or config files.
- Keep a rollback binary available.
- Document all upgrade outcomes in `MIGRATION.md`.

## Attribution

- Consolidation and edits: [CP]
