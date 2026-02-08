# wasmvm v2.2.1 Migration - Fixes Applied

## Summary

✅ **Completed:** 7 out of 12 breaking changes fixed
⚠️ **Remaining:** 5 critical changes (VM init, KVStore adapter, interface updates)

## Fixed Issues

### 1. ✅ StargateMsg → AnyMsg
**File:** `x/wasm/keeper/handler_plugin_encoders.go`
- Renamed `StargateEncoder` → `AnyEncoder`
- Renamed `MessageEncoders.Stargate` → `MessageEncoders.Any`
- Renamed `EncodeStargateMsg` → `EncodeAnyMsg`
- Updated `msg.Stargate` → `msg.Any` in encode logic
- Updated default encoder initialization

### 2. ✅ GoAPI Structure Changed
**File:** `x/wasm/keeper/api.go`
- Renamed `humanAddress` → `humanizeAddress`
- Renamed `canonicalAddress` → `canonicalizeAddress`
- Added `validateAddress` function (new requirement)
- Updated `cosmwasmAPI` initialization with new field names
- Added `DefaultGasCostValidateAddress` constant
- Added `costValidate` variable

### 3. ✅ Events Type Changed
**Files:** `x/wasm/keeper/events.go`, `x/wasm/keeper/events_test.go`, `x/wasm/keeper/keeper.go`
- Changed `wasmvmtypes.Events` → `wasmvmtypes.Array[wasmvmtypes.Event]`
- Updated function signatures in:
  - `newCustomEvents()`
  - `dispatchMessages()` parameter
- Updated all test cases in `events_test.go`

### 4. ✅ Type Aliases Removed
**File:** `x/wasm/keeper/query_plugins.go`
- Changed `wasmvmtypes.Delegations` → `wasmvmtypes.Array[wasmvmtypes.Delegation]`
- Changed `wasmvmtypes.Coins` → `wasmvmtypes.Array[wasmvmtypes.Coin]`
- Updated function signatures:
  - `sdkToDelegations()`
  - `ConvertSdkCoinsToWasmCoins()`

### 5. ✅ EventCosts Signature
**Files:** `x/wasm/keeper/gas_register.go`, `x/wasm/keeper/keeper.go`
- Updated `EventCosts()` to accept events parameter:
  - Old: `EventCosts(attrs []wasmvmtypes.EventAttribute)`
  - New: `EventCosts(attrs []wasmvmtypes.EventAttribute, events wasmvmtypes.Array[wasmvmtypes.Event])`
- Updated call site in keeper to pass `evts` parameter

### 6. ✅ VoteMsg Field Rename
**File:** `x/wasm/keeper/handler_plugin_encoders.go`
- Changed `msg.Vote.Vote` → `msg.Vote.Option`

### 7. ✅ Error Import Update
**File:** `x/wasm/keeper/api.go`
- Uses `wasmvmtypes.InvalidRequest{Err: "..."}` for validation errors

## Compilation Status

**Before fixes:** 13+ compilation errors
**After fixes:** 5 remaining errors (all related to VM initialization and KVStore interface)

## Remaining Issues

### 8. ⚠️ VM Initialization
**File:** `x/wasm/keeper/keeper.go` (line 99)
- Need to change from `wasmvm.NewVM()` to `wasmvm.NewVMWithConfig()`
- `supportedFeatures` must be converted from `string` to `[]string`

### 9. ⚠️ RequiredFeatures → RequiredCapabilities
**File:** `x/wasm/keeper/keeper.go` (lines 181, 193)
- Change `report.RequiredFeatures` → `report.RequiredCapabilities`

### 10. ⚠️ WasmerEngine Interface Mismatch
**File:** `x/wasm/types/wasmer_engine.go`
- `*wasmvm.VM` no longer implements `WasmerEngine` interface
- Consider using concrete type instead of interface

### 11. ⚠️ KVStore Iterator Interface
**Files:** Multiple in `x/wasm/keeper/keeper.go`
- Need `StoreAdapter` wrapper to bridge SDK store to wasmvm KVStore
- Affects all VM calls: Instantiate, Execute, Migrate, Sudo, Query, IBC methods

### 12. ⚠️ Additional VM Signature Changes
- Various other method signatures may have changed in wasmvm v2

## Files Modified

1. ✅ `x/wasm/keeper/handler_plugin_encoders.go`
2. ✅ `x/wasm/keeper/api.go`
3. ✅ `x/wasm/keeper/events.go`
4. ✅ `x/wasm/keeper/events_test.go`
5. ✅ `x/wasm/keeper/query_plugins.go`
6. ✅ `x/wasm/keeper/gas_register.go`
7. ✅ `x/wasm/keeper/keeper.go` (partial - events only)

## Next Steps

**Option A (Recommended):** Rebase on wasmd v0.54.5
- All wasmvm v2 changes already integrated
- Lower risk, less work
- See: `WASMVM_V2_COMPLETE_MIGRATION.md`

**Option B:** Continue surgical fixes
- Implement StoreAdapter
- Fix VM initialization
- Update ~15 more call sites
- Extensive testing required

## Testing

Once migration is complete:
```bash
# Build
make build

# Unit tests
go test ./x/wasm/...

# Integration tests (if available)
make test-integration
```

## References

- wasmvm v2 Changelog: https://github.com/CosmWasm/wasmvm/releases/tag/v2.0.0
- wasmd v0.54 Reference: https://github.com/CosmWasm/wasmd/tree/v0.54.5
- This Repo's Guide: `WASMVM_V2_MIGRATION_GUIDE.md`
- Complete Analysis: `WASMVM_V2_COMPLETE_MIGRATION.md`
