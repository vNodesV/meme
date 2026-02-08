# Keeper Initialization Migration to SDK 0.50.14 - Summary

## Changes Made to app/app.go

### 1. Added Required Imports
- `cosmossdk.io/core/address` - for address codecs
- `addresscodec "github.com/cosmos/cosmos-sdk/codec/address"` - for Bech32 codec implementation
- `cometlog "github.com/cometbft/cometbft/libs/log"` - aliased to avoid conflict
- `consensus` and `consensuskeeper` - for consensus params management
- `govv1` - for gov v1 types
- `capabilitykeeper` and `capabilitytypes` - for IBC capability management

### 2. Updated WasmApp Struct
- Added `consensusKeeper consensuskeeper.Keeper`
- Added `capabilityKeeper *capabilitykeeper.Keeper`
- Added scoped keepers: `scopedIBCKeeper`, `scopedTransferKeeper`, `scopedWasmKeeper`

### 3. Store Keys
- Added `consensuskeeper.StoreKey` to KV store keys
- Added `capabilitytypes.StoreKey` to KV store keys
- Added `capabilitytypes.MemStoreKey` to memory store keys
- Replaced `ibchost.StoreKey` with custom `IBCStoreKey` constant ("ibc")

### 4. Keeper Initializations - All Use runtime.NewKVStoreService()

#### Capability Keeper (First - Before All Others)
```go
app.capabilityKeeper = capabilitykeeper.NewKeeper(
    appCodec,
    keys[capabilitytypes.StoreKey],
    memKeys[capabilitytypes.MemStoreKey],
)
// Scope keepers for IBC modules
app.scopedIBCKeeper = app.capabilityKeeper.ScopeToModule(IBCStoreKey)
app.scopedTransferKeeper = app.capabilityKeeper.ScopeToModule(ibctransfertypes.ModuleName)
app.scopedWasmKeeper = app.capabilityKeeper.ScopeToModule(wasm.ModuleName)
app.capabilityKeeper.Seal()
```

#### Consensus Keeper (Replaces SetParamStore)
```go
app.consensusKeeper = consensuskeeper.NewKeeper(
    appCodec,
    runtime.NewKVStoreService(keys[consensuskeeper.StoreKey]),
    authority,
    runtime.ProvideEventService(),
)
bApp.SetParamStore(app.consensusKeeper.ParamsStore)
```

#### Account Keeper
```go
app.accountKeeper = authkeeper.NewAccountKeeper(
    appCodec,
    runtime.NewKVStoreService(keys[authtypes.StoreKey]),
    authtypes.ProtoBaseAccount,
    maccPerms,
    addressCodec,  // NEW
    sdk.GetConfig().GetBech32AccountAddrPrefix(),  // NEW
    authority,  // NEW
)
```

#### Bank Keeper
```go
app.bankKeeper = bankkeeper.NewBaseKeeper(
    appCodec,
    runtime.NewKVStoreService(keys[banktypes.StoreKey]),
    app.accountKeeper,
    app.ModuleAccountAddrs(),
    authority,  // NEW
    logger,  // NEW - cosmossdk.io/log.Logger
)
```

#### Authz Keeper
```go
app.AuthzKeeper = authzkeeper.NewKeeper(
    runtime.NewKVStoreService(keys[authzkeeper.StoreKey]),
    appCodec,
    app.BaseApp.MsgServiceRouter(),
    app.accountKeeper,  // NEW - 4th parameter
)
```

#### Fee Grant Keeper
```go
app.FeeGrantKeeper = feegrantkeeper.NewKeeper(
    appCodec,
    runtime.NewKVStoreService(keys[feegrant.StoreKey]),
    app.accountKeeper,
)
```

#### Staking Keeper
```go
stakingKeeper := stakingkeeper.NewKeeper(
    appCodec,
    runtime.NewKVStoreService(keys[stakingtypes.StoreKey]),
    app.accountKeeper,
    app.bankKeeper,
    authority,  // NEW
    validatorAddressCodec,  // NEW
    consensusAddressCodec,  // NEW
)
```

#### Mint Keeper
```go
app.mintKeeper = mintkeeper.NewKeeper(
    appCodec,
    runtime.NewKVStoreService(keys[minttypes.StoreKey]),
    stakingKeeper,  // Changed from &stakingKeeper
    app.accountKeeper,
    app.bankKeeper,
    authtypes.FeeCollectorName,
    authority,  // NEW
)
```

#### Distribution Keeper
```go
app.distrKeeper = distrkeeper.NewKeeper(
    appCodec,
    runtime.NewKVStoreService(keys[distrtypes.StoreKey]),
    app.accountKeeper,
    app.bankKeeper,
    stakingKeeper,  // Changed from &stakingKeeper
    authtypes.FeeCollectorName,
    authority,  // NEW - removed ModuleAccountAddrs()
)
```

#### Slashing Keeper
```go
app.slashingKeeper = slashingkeeper.NewKeeper(
    appCodec,
    legacyAmino,  // NEW
    runtime.NewKVStoreService(keys[slashingtypes.StoreKey]),
    stakingKeeper,  // Changed from &stakingKeeper
    authority,  // NEW
)
```

#### Crisis Keeper
```go
app.crisisKeeper = *crisiskeeper.NewKeeper(  // Returns pointer now
    appCodec,
    runtime.NewKVStoreService(keys[crisistypes.StoreKey]),
    invCheckPeriod,
    app.bankKeeper,
    authtypes.FeeCollectorName,
    authority,  // NEW
    addressCodec,  // NEW
)
```

#### Upgrade Keeper
```go
app.upgradeKeeper = *upgradekeeper.NewKeeper(  // Returns pointer now
    skipUpgradeHeights,
    runtime.NewKVStoreService(keys[upgradetypes.StoreKey]),
    appCodec,
    homePath,
    app.BaseApp,
    authority,  // NEW
)
```

#### Staking Hooks
```go
// SetHooks returns void now, not *Keeper
stakingKeeper.SetHooks(
    stakingtypes.NewMultiStakingHooks(app.distrKeeper.Hooks(), app.slashingKeeper.Hooks()),
)
app.stakingKeeper = *stakingKeeper
```

#### IBC Keeper
```go
app.ibcKeeper = ibckeeper.NewKeeper(
    appCodec,
    keys[IBCStoreKey],
    app.getSubspace(IBCStoreKey),
    app.stakingKeeper,
    app.upgradeKeeper,
    app.scopedIBCKeeper,  // NEW
    authority,  // NEW
)
```

#### Transfer Keeper
```go
app.transferKeeper = ibctransferkeeper.NewKeeper(
    appCodec,
    keys[ibctransfertypes.StoreKey],
    app.getSubspace(ibctransfertypes.ModuleName),
    app.ibcKeeper.ChannelKeeper,  // ICS4 wrapper
    app.ibcKeeper.ChannelKeeper,
    app.ibcKeeper.PortKeeper,  // Changed from &app.ibcKeeper.PortKeeper
    app.accountKeeper,
    app.bankKeeper,
    app.scopedTransferKeeper,  // NEW
    authority,  // NEW
)
```

#### Evidence Keeper
```go
evidenceKeeper := evidencekeeper.NewKeeper(
    appCodec,
    runtime.NewKVStoreService(keys[evidencetypes.StoreKey]),  // Changed
    &app.stakingKeeper,
    app.slashingKeeper,
    addressCodec,  // NEW
    runtime.ProvideCometInfoService(),  // NEW
)
app.evidenceKeeper = *evidenceKeeper
```

#### Gov Keeper
```go
govConfig := govtypes.DefaultConfig()  // NEW - required config
app.govKeeper = *govkeeper.NewKeeper(  // Returns pointer now
    appCodec,
    runtime.NewKVStoreService(keys[govtypes.StoreKey]),
    app.accountKeeper,
    app.bankKeeper,
    &app.stakingKeeper,
    app.distrKeeper,  // NEW - distribution keeper added
    app.BaseApp.MsgServiceRouter(),  // NEW
    govConfig,  // NEW
    authority,  // NEW
)
app.govKeeper.SetLegacyRouter(govRouter)  // NEW - for v1beta1 compatibility
```

### 5. Deprecated Items Removed/Updated
- Removed `bApp.SetParamStore(app.paramsKeeper.Subspace(...))` - replaced with consensus keeper
- Removed deprecated proposal handlers:
  - `params.NewParamChangeProposalHandler`
  - `distr.NewCommunityPoolSpendProposalHandler`
  - `upgrade.NewSoftwareUpgradeProposalHandler`
  - `ibcclient.NewClientProposalHandler`
- Changed `govtypes.NewRouter()` to `govv1beta1.NewRouter()` for legacy support
- Changed `app.BaseApp.DeliverTx` to `app` in genutil.NewAppModule()

### 6. Authority Address
Created a single authority address used across all keepers:
```go
authority := authtypes.NewModuleAddress(govtypes.ModuleName).String()
```

### 7. Address Codecs
Created three address codecs for different address types:
```go
addressCodec := addresscodec.NewBech32Codec(sdk.GetConfig().GetBech32AccountAddrPrefix())
validatorAddressCodec := addresscodec.NewBech32Codec(sdk.GetConfig().GetBech32ValidatorAddrPrefix())
consensusAddressCodec := addresscodec.NewBech32Codec(sdk.GetConfig().GetBech32ConsensusAddrPrefix())
```

## Remaining Issues

### 1. Database Type Mismatch
```
cannot use db (variable of interface type "github.com/cometbft/cometbft-db".DB) as "github.com/cosmos/cosmos-db".DB
```
**Solution Required**: The app needs to migrate from cometbft-db to cosmos-db, OR a wrapper/adapter needs to be created.

### 2. Wasmd Module Compatibility
The wasmd module itself (not just the app) needs updates to be fully compatible with SDK 0.50.14:
- Context type changes (sdk.Context → context.Context)
- Keeper interface mismatches
- IBC module interface changes

**Solution Required**: Either:
- Upgrade to a wasmd version compatible with SDK 0.50.14
- OR fork and update wasmd's x/wasm module code
- OR wait for official wasmd SDK 0.50 support

### 3. IBC Module Interface Changes
```
transfer.AppModule does not implement porttypes.IBCModule (missing methods)
```
**Solution Required**: This is related to IBC-Go v8 and how modules expose IBC capabilities. May need to use IBCCoreKeeper or adjust module wiring.

### 4. Ante Handler Issues
```
app/ante.go:100:38: cannot use options.IBCChannelkeeper...
```
**Solution Required**: The ante handler setup in ante.go needs updating for SDK 0.50 patterns.

## Migration Status
✅ **COMPLETE**: All keeper initialization signatures updated for SDK 0.50.14
✅ **COMPLETE**: Consensus keeper integration
✅ **COMPLETE**: Capability keeper for IBC
✅ **COMPLETE**: Address codecs
✅ **COMPLETE**: Authority addresses
⚠️ **BLOCKED**: Database migration (cometbft-db → cosmos-db)
⚠️ **BLOCKED**: Wasmd module compatibility
⚠️ **BLOCKED**: Full build success

## Recommendations

1. **Short Term**: The keeper initializations in app/app.go are now SDK 0.50.14 compliant
2. **Medium Term**: Address the DB migration
3. **Long Term**: Either upgrade to wasmd SDK 0.50 branch or complete the wasmd module migration

## Testing After Complete Migration

Once remaining issues are resolved, test:
1. `memed init` - Initialize chain
2. `memed start` - Start single node
3. Verify all modules load correctly
4. Test wasm contract upload/instantiate
5. Test IBC transfers
