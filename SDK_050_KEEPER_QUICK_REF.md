# SDK 0.50.14 Keeper Initialization - Quick Reference

## Key Changes Summary

### 1. Use runtime.NewKVStoreService() Instead of Raw Keys
**OLD (SDK 0.47)**:
```go
keeper.NewKeeper(codec, keys[types.StoreKey], ...)
```

**NEW (SDK 0.50)**:
```go
keeper.NewKeeper(codec, runtime.NewKVStoreService(keys[types.StoreKey]), ...)
```

### 2. Add Authority Parameter
Most keepers now require an authority address (typically the gov module):
```go
authority := authtypes.NewModuleAddress(govtypes.ModuleName).String()
```

### 3. Add Address Codecs
Create address codecs for different address types:
```go
addressCodec := addresscodec.NewBech32Codec(sdk.GetConfig().GetBech32AccountAddrPrefix())
validatorAddressCodec := addresscodec.NewBech32Codec(sdk.GetConfig().GetBech32ValidatorAddrPrefix())
consensusAddressCodec := addresscodec.NewBech32Codec(sdk.GetConfig().GetBech32ConsensusAddrPrefix())
```

### 4. Consensus Keeper Replaces SetParamStore
**OLD**:
```go
bApp.SetParamStore(app.paramsKeeper.Subspace(baseapp.Paramspace).WithKeyTable(paramskeeper.ConsensusParamsKeyTable()))
```

**NEW**:
```go
app.consensusKeeper = consensuskeeper.NewKeeper(
    appCodec,
    runtime.NewKVStoreService(keys[consensuskeeper.StoreKey]),
    authority,
    runtime.ProvideEventService(),
)
bApp.SetParamStore(app.consensusKeeper.ParamsStore)
```

### 5. Some Keepers Now Return Pointers
Crisis and Upgrade keepers return pointers - dereference when assigning:
```go
app.crisisKeeper = *crisiskeeper.NewKeeper(...)
app.upgradeKeeper = *upgradekeeper.NewKeeper(...)
app.govKeeper = *govkeeper.NewKeeper(...)
```

### 6. Staking Keeper Hooks Changed
**OLD**:
```go
app.stakingKeeper = *stakingKeeper.SetHooks(hooks)
```

**NEW**:
```go
stakingKeeper.SetHooks(hooks)  // Returns void now
app.stakingKeeper = *stakingKeeper
```

### 7. Bank Keeper Needs Logger
```go
app.bankKeeper = bankkeeper.NewBaseKeeper(
    appCodec,
    runtime.NewKVStoreService(keys[banktypes.StoreKey]),
    app.accountKeeper,
    app.ModuleAccountAddrs(),
    authority,
    logger,  // cosmossdk.io/log.Logger
)
```

### 8. IBC Requires Capability Keeper
Initialize capability keeper FIRST, create scoped keepers, then seal:
```go
app.capabilityKeeper = capabilitykeeper.NewKeeper(appCodec, keys[capabilitytypes.StoreKey], memKeys[capabilitytypes.MemStoreKey])
app.scopedIBCKeeper = app.capabilityKeeper.ScopeToModule("ibc")
app.scopedTransferKeeper = app.capabilityKeeper.ScopeToModule(ibctransfertypes.ModuleName)
app.capabilityKeeper.Seal()
```

### 9. Evidence Keeper Needs CometInfo Service
```go
evidenceKeeper := evidencekeeper.NewKeeper(
    appCodec,
    runtime.NewKVStoreService(keys[evidencetypes.StoreKey]),
    &app.stakingKeeper,
    app.slashingKeeper,
    addressCodec,
    runtime.ProvideCometInfoService(),  // NEW
)
```

### 10. Gov Keeper Needs Config and Distribution Keeper
```go
govConfig := govtypes.DefaultConfig()
app.govKeeper = *govkeeper.NewKeeper(
    appCodec,
    runtime.NewKVStoreService(keys[govtypes.StoreKey]),
    app.accountKeeper,
    app.bankKeeper,
    &app.stakingKeeper,
    app.distrKeeper,  // NEW
    app.BaseApp.MsgServiceRouter(),
    govConfig,  // NEW
    authority,
)
```

## Import Additions Required

```go
import (
    "cosmossdk.io/core/address"
    addresscodec "github.com/cosmos/cosmos-sdk/codec/address"
    "github.com/cosmos/cosmos-sdk/x/consensus"
    consensuskeeper "github.com/cosmos/cosmos-sdk/x/consensus/keeper"
    capabilitykeeper "github.com/cosmos/ibc-go/modules/capability/keeper"
    capabilitytypes "github.com/cosmos/ibc-go/modules/capability/types"
    govv1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
)
```

## Store Key Additions

```go
keys := storetypes.NewKVStoreKeys(
    // ... existing keys ...
    consensuskeeper.StoreKey,
    capabilitytypes.StoreKey,
)

memKeys := storetypes.NewMemoryStoreKeys(
    capabilitytypes.MemStoreKey,
)
```

## Common Mistakes to Avoid

1. ❌ Using `&stakingKeeper` when stakingKeeper is already a pointer
2. ❌ Forgetting to seal the capability keeper
3. ❌ Not initializing capability keeper before IBC/Transfer keepers
4. ❌ Using deprecated `app.getSubspace()` instead of authority param
5. ❌ Using `keys[...]` directly instead of `runtime.NewKVStoreService(keys[...])`
6. ❌ Forgetting to dereference keepers that now return pointers

## References

- Cosmos SDK v0.50 Migration Guide: https://docs.cosmos.network/v0.50/learn/beginner/00-app-anatomy
- IBC-Go v8 Migration: https://github.com/cosmos/ibc-go/blob/main/docs/migrations/v7-to-v8.md
