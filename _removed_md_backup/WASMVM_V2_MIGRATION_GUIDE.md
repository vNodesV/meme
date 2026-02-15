# WasmVM v2 Migration Guide

## Overview

This guide documents the systematic migration from wasmvm v1.x to wasmvm v2.2.1 for the meme chain's wasmd fork.

**Target Versions:**
- wasmvm: v1.5.9 → v2.2.1
- cosmos-sdk: 0.53.5 → 0.50.14 (cheqd fork)
- ibc-go: v10 → v8.7.0

## Breaking Changes Summary

### 1. StargateMsg → AnyMsg (BREAKING)

**Change:** `wasmvmtypes.StargateMsg` has been removed and replaced with `wasmvmtypes.AnyMsg`

**Migration:**
- Replace all references to `StargateMsg` with `AnyMsg`
- Replace `EncodeStargateMsg` with `EncodeAnyMsg`
- Update `MessageEncoders` struct field from `Stargate` to `Any`

**Files Affected:**
- `x/wasm/keeper/handler_plugin_encoders.go`

**Before:**
```go
type StargateEncoder func(sender sdk.AccAddress, msg *wasmvmtypes.StargateMsg) ([]sdk.Msg, error)

type MessageEncoders struct {
    Stargate func(sender sdk.AccAddress, msg *wasmvmtypes.StargateMsg) ([]sdk.Msg, error)
}

func EncodeStargateMsg(unpacker codectypes.AnyUnpacker) StargateEncoder {
    return func(sender sdk.AccAddress, msg *wasmvmtypes.StargateMsg) ([]sdk.Msg, error) {
        codecAny := codectypes.Any{
            TypeUrl: msg.TypeURL,
            Value:   msg.Value,
        }
        // ...
    }
}
```

**After:**
```go
type AnyEncoder func(sender sdk.AccAddress, msg *wasmvmtypes.AnyMsg) ([]sdk.Msg, error)

type MessageEncoders struct {
    Any func(sender sdk.AccAddress, msg *wasmvmtypes.AnyMsg) ([]sdk.Msg, error)
}

func EncodeAnyMsg(unpacker codectypes.AnyUnpacker) AnyEncoder {
    return func(sender sdk.AccAddress, msg *wasmvmtypes.AnyMsg) ([]sdk.Msg, error) {
        codecAny := codectypes.Any{
            TypeUrl: msg.TypeURL,
            Value:   msg.Value,
        }
        // ...
    }
}
```

### 2. GoAPI Structure Changed (BREAKING)

**Change:** The `wasmvm.GoAPI` struct fields have been renamed:
- `HumanAddress` → `HumanizeAddress`
- `CanonicalAddress` → `CanonicalizeAddress`
- Added: `ValidateAddress` (new required field)

**Migration:**
- Rename function references in GoAPI initialization
- Add `ValidateAddress` implementation

**Files Affected:**
- `x/wasm/keeper/api.go`

**Before:**
```go
var cosmwasmAPI = wasmvm.GoAPI{
    HumanAddress:     humanAddress,
    CanonicalAddress: canonicalAddress,
}
```

**After:**
```go
var cosmwasmAPI = wasmvm.GoAPI{
    HumanizeAddress:     humanizeAddress,
    CanonicalizeAddress: canonicalizeAddress,
    ValidateAddress:     validateAddress,
}
```

**Required New Function:**
```go
func validateAddress(human string) (uint64, error) {
    canonicalized, err := sdk.AccAddressFromBech32(human)
    if err != nil {
        return costValidate, err
    }
    // AccAddressFromBech32 already calls VerifyAddressFormat
    if canonicalized.String() != human {
        return costValidate, errors.New("address not normalized")
    }
    return costValidate, nil
}
```

### 3. Events Type Changed (BREAKING)

**Change:** `wasmvmtypes.Events` is now `wasmvmtypes.Array[wasmvmtypes.Event]`

**Migration:**
- Replace `wasmvmtypes.Events` with `wasmvmtypes.Array[wasmvmtypes.Event]`
- The usage remains the same; only the type name changes

**Files Affected:**
- `x/wasm/keeper/events.go`
- `x/wasm/keeper/events_test.go`
- `x/wasm/keeper/keeper.go`

**Before:**
```go
func newCustomEvents(evts wasmvmtypes.Events, contractAddr sdk.AccAddress) (sdk.Events, error) {
    events := make(sdk.Events, 0, len(evts))
    for _, e := range evts {
        // ...
    }
}
```

**After:**
```go
func newCustomEvents(evts wasmvmtypes.Array[wasmvmtypes.Event], contractAddr sdk.AccAddress) (sdk.Events, error) {
    events := make(sdk.Events, 0, len(evts))
    for _, e := range evts {
        // ...
    }
}
```

### 4. Type Aliases Removed (BREAKING)

**Change:** Several type aliases have been removed:
- `wasmvmtypes.Delegations` → `wasmvmtypes.Array[wasmvmtypes.Delegation]`
- `wasmvmtypes.Coins` → `wasmvmtypes.Array[wasmvmtypes.Coin]`

**Migration:**
- Replace with explicit generic array types

**Files Affected:**
- `x/wasm/keeper/query_plugins.go`

**Before:**
```go
func sdkToDelegations(ctx sdk.Context, keeper types.StakingKeeper, delegations []stakingtypes.Delegation) (wasmvmtypes.Delegations, error) {
    result := make([]wasmvmtypes.Delegation, len(delegations))
    // ...
    return result, nil
}

func ConvertSdkCoinsToWasmCoins(coins []sdk.Coin) wasmvmtypes.Coins {
    converted := make(wasmvmtypes.Coins, len(coins))
    // ...
    return converted
}
```

**After:**
```go
func sdkToDelegations(ctx sdk.Context, keeper types.StakingKeeper, delegations []stakingtypes.Delegation) (wasmvmtypes.Array[wasmvmtypes.Delegation], error) {
    result := make(wasmvmtypes.Array[wasmvmtypes.Delegation], len(delegations))
    // ...
    return result, nil
}

func ConvertSdkCoinsToWasmCoins(coins []sdk.Coin) wasmvmtypes.Array[wasmvmtypes.Coin] {
    converted := make(wasmvmtypes.Array[wasmvmtypes.Coin], len(coins))
    // ...
    return converted
}
```

## Migration Strategy

### Recommended Approach: Surgical Fixes

Given the constraints:
1. Must preserve custom functionality
2. Must maintain SDK 0.50.14 compatibility
3. Need minimal changes

**We recommend surgical fixes over rebasing** because:
- The changes are well-defined and localized
- Custom functionality is preserved
- Lower risk of introducing new bugs
- Easier to review and test

### Alternative: Rebase on wasmd v0.54.5

If you have significant divergence from upstream wasmd, consider:
1. Creating a feature branch with your custom changes
2. Rebasing onto wasmd v0.54.5 (which uses wasmvm v2)
3. Reapplying custom patches

**However**, this is only necessary if:
- You have extensive custom wasm module changes
- The surgical approach becomes too complex
- You want to stay closer to upstream for future upgrades

## Migration Checklist

### Phase 1: Update Type References
- [ ] Replace `StargateMsg` → `AnyMsg` in handler_plugin_encoders.go
- [ ] Replace `Stargate` → `Any` in MessageEncoders struct
- [ ] Replace `EncodeStargateMsg` → `EncodeAnyMsg`
- [ ] Update GoAPI field names in api.go
- [ ] Add ValidateAddress function in api.go
- [ ] Replace `wasmvmtypes.Events` → `wasmvmtypes.Array[wasmvmtypes.Event]`
- [ ] Replace `wasmvmtypes.Delegations` → `wasmvmtypes.Array[wasmvmtypes.Delegation]`
- [ ] Replace `wasmvmtypes.Coins` → `wasmvmtypes.Array[wasmvmtypes.Coin]`

### Phase 2: Update Tests
- [ ] Update events_test.go with new Array type
- [ ] Update handler_plugin_encoders_test.go
- [ ] Run all wasm keeper tests

### Phase 3: Verify
- [ ] Build succeeds: `make build`
- [ ] All tests pass: `make test`
- [ ] Integration tests pass (if any)

## Reference Implementation

All changes follow the official wasmd v0.54.5 implementation:
- https://github.com/CosmWasm/wasmd/tree/v0.54.5/x/wasm/keeper

## Testing Strategy

1. **Unit Tests:** Ensure all keeper tests pass
2. **Integration Tests:** Test contract execution end-to-end
3. **Local Testnet:** Deploy and execute actual contracts
4. **Upgrade Test:** If possible, test upgrade path from v1 to v2

## Rollback Plan

If issues are discovered:
1. Git revert the migration commits
2. Return to wasmvm v1.5.9
3. Document the specific failure
4. Investigate and retry with fixes

## Additional Notes

### Cosmos SDK Compatibility

The migration is compatible with Cosmos SDK 0.50.14. The cheqd fork patches should not interfere with wasmvm changes as they target the SDK core, not the wasm module.

### IBC Considerations

The downgrade from IBC v10 → v8 is unusual but should not affect wasmvm functionality directly. Ensure IBC-enabled contracts are tested thoroughly.

### Custom Functionality

If you have custom message encoders or query plugins:
- Review each custom implementation
- Update type signatures to match new wasmvm v2 types
- Test custom functionality extensively

## Support Resources

- wasmvm v2 release notes: https://github.com/CosmWasm/wasmvm/releases/tag/v2.0.0
- wasmd v0.54.0 changelog: https://github.com/CosmWasm/wasmd/blob/main/CHANGELOG.md
- CosmWasm migration guide: https://docs.cosmwasm.com/
