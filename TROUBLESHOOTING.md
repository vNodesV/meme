# Troubleshooting Guide

Common errors encountered during SDK v0.50 migration and wasmvm v2 upgrade.

## Params Migration

### Error: `parameter X not registered`

**Cause:** Missing `.WithKeyTable()` for a module that has legacy params.

**Fix:** Register the module’s ParamKeyTable in `initParamsKeeper`.

**Example:**
```go
paramsKeeper.Subspace(IBCStoreKey).WithKeyTable(ibcclienttypes.ParamKeyTable())
```

### Error: `collections: not found: key 'no_key'`

**Cause:** Consensus params not yet migrated during startup.

**Action:** No action needed. This is expected before the upgrade handler runs.

## Service Registration

### Error: `type_url / has not been registered yet`

**Cause:** `RegisterInterfaces` was not called before `RegisterServices`, or protobuf registries are mismatched.

**Fixes:**

1. Call `basicManager.RegisterInterfaces(interfaceRegistry)` before `mm.RegisterServices`.
2. Ensure wasm protobuf files import `github.com/cosmos/gogoproto/proto` (not `github.com/gogo/protobuf/proto`).

## ModuleBasics nil panics

### Error: `invalid memory address or nil pointer` in `AppModuleBasic.GetTxCmd`

**Cause:** `ModuleBasics` created as empty structs without codecs in SDK 0.50.

**Fix:** Use `module.NewBasicManagerFromManager` and/or `MakeBasicManager()` helper.

## IBC Transfer Module Wiring

### Error: `transfer.AppModule does not implement IBCModule`

**Cause:** SDK/IBC interface changes in v8.

**Fix:** Use `transfer.NewIBCModule(keeper)` in the IBC router.

## wasmvm v2 Issues

### Error: KVStore iterator mismatch

**Cause:** SDK store iterator types differ from wasmvm iterator types.

**Fix:** Implement and use a StoreAdapter wrapper (see `WASM.md`).

### Error: `NewVM` signature mismatch

**Cause:** wasmvm v2 uses `NewVMWithConfig`.

**Fix:** Update VM initialization to the new config-based API.

## SDK v0.53 (Historical)

If you see `go mod tidy` errors from v0.53-era attempts, they often point to:

- Legacy REST packages removed
- Capability module removed in IBC v10
- `tmservice` moved to `cmtservice`

Use the v0.53 troubleshooting notes only if you are explicitly targeting that stack.

## Attribution

- Consolidation and edits: [CP]
