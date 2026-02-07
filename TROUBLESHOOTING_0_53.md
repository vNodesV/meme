# Troubleshooting SDK v0.53 upgrade `go mod tidy` errors

This note maps the reported `go mod tidy` errors to concrete fixes and the
underlying migration tasks required for a successful Cosmos SDK v0.53.x + IBC
v10 upgrade.

## Root cause summary

Most errors come from **API/middleware removals** and **module splits** in the
SDK/IBC v10 stack:

- Legacy REST packages were removed (auth/gov REST, `types/rest`, Swagger statik).
- IBC v10 removed the `capability` module and **ScopedKeeper** wiring.
- `tmservice` moved to `cmtservice` and CometBFT replaced Tendermint.
- SDK `x/*` modules are now split between `github.com/cosmos/cosmos-sdk/x/*` and
  `cosmossdk.io/x/*` depending on the module (e.g., `evidence`, `feegrant`,
  `upgrade` are split).
- Governance v1 uses **message routing** (not v1beta1 proposal handlers) for
  most modules; proposal handler lists shrink.
- `simapp` is external and now requires newer Go; tests must be refactored or
  gated/removed.

## Error-by-error fixes

### `cosmossdk.io/x/auth/client/rest` / `cosmossdk.io/x/gov/client/rest`
**Fix:** Remove REST handlers and legacy REST routes. SDK v0.53 does not expose
legacy REST endpoints. Replace with gRPC Gateway routes.

### `github.com/cosmos/cosmos-sdk/client/docs/statik`
**Fix:** Remove the blank import and Swagger statik wiring. The statik package
is no longer shipped in the SDK.

### `github.com/cosmos/cosmos-sdk/client/grpc/tmservice`
**Fix:** Replace with `github.com/cosmos/cosmos-sdk/client/grpc/cmtservice`.

### `github.com/cosmos/cosmos-sdk/x/capability` (and `ScopedKeeper`)
**Fix:** IBC v10 removes the capability module. Remove:
- `capability` module from ModuleBasics and module manager
- `capabilityKeeper` and all `scoped*` keepers from the app struct
- capability store keys and mem store keys

Refactor IBC/transfer/wasm keepers to the v10 signatures that no longer
require scoped keepers.

### `github.com/cosmos/ibc-go/v10/modules/core/02-client/client`
**Fix:** Update to `.../client/cli` if CLI handler is needed. IBC v10 no longer
exposes the old `client` package.

### `github.com/cosmos/cosmos-sdk/snapshots`
**Fix:** Import `cosmossdk.io/store/snapshots` instead.

### `github.com/cosmos/cosmos-sdk/types/rest`
**Fix:** Remove usage; legacy REST removed.

### `cosmossdk.io/x/staking/teststaking`
**Fix:** Replace with local helper that converts staking validators to CometBFT
validators using `cryptocodec.ToCmtPubKeyInterface`, or use the new
`cosmos-sdk/testutil` helpers if you refactor tests.

### `github.com/cosmos/ibc-go/v10/modules/light-clients/07-tendermint/types`
**Fix:** In IBC v10, the light client lives under
`github.com/cosmos/ibc-go/v10/modules/light-clients/07-tendermint` for the
module and `.../types` for types. Use the exact path from the v10 module.

### `github.com/gogo/protobuf/grpc`
**Fix:** Switch the protobuf replacement back to
`github.com/regen-network/protobuf` (which provides `grpc`), or regenerate
protobufs to use standard gRPC imports.

### `cosmossdk.io/core/legacy` / `cosmossdk.io/core/appconfig`
**Fix:** Pin `cosmossdk.io/core` to the **SDK v0.53.5** version
(`cosmossdk.io/core v0.11.3`) and align all SDK split modules to the same
version set. Avoid the `cosmossdk.io` meta-module.

### `cosmossdk.io/x/consensus/types`
**Fix:** Use the SDK v0.53 `x/consensus` module from the main SDK repository
(`github.com/cosmos/cosmos-sdk/x/consensus`).

### `github.com/cosmos/cosmos-sdk/types/authz` / `github.com/cosmos/cosmos-sdk/simsx`
**Fix:** Use the SDK v0.53 module paths and avoid importing split `cosmossdk.io`
modules where the SDK still provides packages.

### `github.com/cosmos/cosmos-sdk/simapp` / `simapp/helpers`
**Fix:** The external simapp requires Go >= 1.25.7. Either:
- Remove or refactor simapp tests, or
- Port tests to the new `testutil` packages used in upstream wasmd v0.55+.

## Recommended path forward

1. **Use upstream wasmd v0.55.1 as a reference** for SDK v0.53 wiring. It
   already includes the correct module splits, app wiring, and keeper
   signatures.
2. **Remove legacy REST** and update API registration to gRPC Gateway +
   `cmtservice`.
3. **Remove capability module & scoped keepers**, then update IBC/transfer/wasm
   keepers to v10 signatures.
4. **Align go.mod** to the SDK v0.53.5 dependency set:
   - `cosmossdk.io/core v0.11.3`
   - `github.com/cometbft/cometbft v0.38.21`
   - `github.com/cometbft/cometbft-db v0.14.1`
   - `github.com/gogo/protobuf` replaced by regen protobuf
5. **Refactor tests** to remove simapp or port to `testutil` patterns used by
   upstream wasmd.

## Useful upstream references

- Wasmd v0.55.1 app wiring and tests:
  - `https://raw.githubusercontent.com/CosmWasm/wasmd/v0.55.1/app/app.go`
  - `https://raw.githubusercontent.com/CosmWasm/wasmd/v0.55.1/app/test_helpers.go`
- IBC v10 migration (capability removal):
  - `docs/docs/05-migrations/13-v8_1-to-v10.md` in the ibc-go v10 repo.
