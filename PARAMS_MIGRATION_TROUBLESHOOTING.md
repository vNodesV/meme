# Quick Troubleshooting Guide: SDK 0.50 Params Migration

## Error: "parameter X not registered"

### Symptom
```
panic: parameter X not registered
goroutine 1 [running]:
github.com/cosmos/cosmos-sdk/x/params/types.Subspace.checkType(...)
```

### Root Cause
Module's params subspace missing `.WithKeyTable()` call in `initParamsKeeper()`

### Solution
1. Identify which module is failing (look at stack trace)
2. Find the module's `ParamKeyTable()` function
3. Add `.WithKeyTable()` call to that module's subspace

### Example - IBC Client Fix
```go
// Before (WRONG)
paramsKeeper.Subspace(IBCStoreKey)

// After (CORRECT)
paramsKeeper.Subspace(IBCStoreKey).WithKeyTable(ibcclienttypes.ParamKeyTable())
```

---

## Error: "collections: not found: key 'no_key'"

### Symptom
```
ERR failed to get consensus params err="collections: not found: key 'no_key'"
```

### Analysis
- Appears **before** upgrade starts
- **Non-fatal** - this is expected behavior
- Consensus params don't exist in collections store yet
- They get migrated during upgrade execution

### Action
**No action needed** - this is normal

---

## Checklist: Adding New Module with Params

When adding a new module that has params to your app:

- [ ] Check if module has `params_legacy.go` or similar
- [ ] Check if module has `ParamKeyTable()` function
- [ ] Check if module has `ParamSetPairs()` method
- [ ] If any above are true, add to `initParamsKeeper`:
  ```go
  paramsKeeper.Subspace(moduletypes.ModuleName).WithKeyTable(moduletypes.ParamKeyTable())
  ```
- [ ] Add module types import if needed
- [ ] Build and test

---

## Common Modules That Need WithKeyTable

### Core SDK Modules (0.50)
```go
paramsKeeper.Subspace(authtypes.ModuleName).WithKeyTable(authtypes.ParamKeyTable())
paramsKeeper.Subspace(banktypes.ModuleName).WithKeyTable(banktypes.ParamKeyTable())
paramsKeeper.Subspace(stakingtypes.ModuleName).WithKeyTable(stakingtypes.ParamKeyTable())
paramsKeeper.Subspace(minttypes.ModuleName).WithKeyTable(minttypes.ParamKeyTable())
paramsKeeper.Subspace(distrtypes.ModuleName).WithKeyTable(distrtypes.ParamKeyTable())
paramsKeeper.Subspace(slashingtypes.ModuleName).WithKeyTable(slashingtypes.ParamKeyTable())
paramsKeeper.Subspace(govtypes.ModuleName).WithKeyTable(govv1.ParamKeyTable())
paramsKeeper.Subspace(crisistypes.ModuleName).WithKeyTable(crisistypes.ParamKeyTable())
```

### IBC Modules
```go
// IBC client - needs WithKeyTable for AllowedClients param
paramsKeeper.Subspace(IBCStoreKey).WithKeyTable(ibcclienttypes.ParamKeyTable())

// IBC transfer - NO WithKeyTable (no legacy params)
paramsKeeper.Subspace(ibctransfertypes.ModuleName)
```

### CosmWasm
```go
// Wasm - NO WithKeyTable (handles params internally)
paramsKeeper.Subspace(wasm.ModuleName)
```

### Consensus
```go
// Baseapp consensus params
paramsKeeper.Subspace(baseapp.Paramspace).WithKeyTable(paramstypes.ConsensusParamsKeyTable())
```

---

## How to Find ParamKeyTable Location

### Method 1: Go Doc
```bash
go doc github.com/cosmos/MODULE_PATH/types ParamKeyTable
```

### Method 2: Find Files
```bash
find /path/to/module -name "*params*.go" | xargs grep -l "ParamKeyTable"
```

### Method 3: Search in Dependencies
```bash
# List module location
go list -f '{{.Dir}}' github.com/cosmos/MODULE_PATH/types

# Check files in that directory
ls MODULE_DIR/*params*.go
```

---

## Import Reference

### IBC Client Types
```go
import (
    ibcclienttypes "github.com/cosmos/ibc-go/v8/modules/core/02-client/types"
)
```

### Core SDK Types
```go
import (
    authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
    banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
    stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
    minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"
    distrtypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
    slashingtypes "github.com/cosmos/cosmos-sdk/x/slashing/types"
    govv1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
    crisistypes "github.com/cosmos/cosmos-sdk/x/crisis/types"
    paramstypes "github.com/cosmos/cosmos-sdk/x/params/types"
)

import (
    "github.com/cosmos/cosmos-sdk/baseapp"
)
```

---

## Testing After Fix

### 1. Build Test
```bash
go build ./app
# Should succeed without errors
```

### 2. Install Test
```bash
make install
# Should complete successfully
```

### 3. Binary Check
```bash
ls -lh ~/go/bin/memed
~/go/bin/memed version
# Should show binary and version
```

### 4. Upgrade Test (Devnet)
```bash
# Start node with upgrade
memed start --home ~/.meme

# Watch logs for:
# - "applying upgrade..." message
# - "migrating module X from version Y to version Z"
# - No panics
# - Successful completion
```

---

## Quick Reference: initParamsKeeper Template

```go
func initParamsKeeper(appCodec codec.BinaryCodec, legacyAmino *codec.LegacyAmino, key, tkey storetypes.StoreKey) paramskeeper.Keeper {
	paramsKeeper := paramskeeper.NewKeeper(appCodec, legacyAmino, key, tkey)

	// Core SDK modules WITH legacy params
	paramsKeeper.Subspace(authtypes.ModuleName).WithKeyTable(authtypes.ParamKeyTable())
	paramsKeeper.Subspace(banktypes.ModuleName).WithKeyTable(banktypes.ParamKeyTable())
	paramsKeeper.Subspace(stakingtypes.ModuleName).WithKeyTable(stakingtypes.ParamKeyTable())
	paramsKeeper.Subspace(minttypes.ModuleName).WithKeyTable(minttypes.ParamKeyTable())
	paramsKeeper.Subspace(distrtypes.ModuleName).WithKeyTable(distrtypes.ParamKeyTable())
	paramsKeeper.Subspace(slashingtypes.ModuleName).WithKeyTable(slashingtypes.ParamKeyTable())
	paramsKeeper.Subspace(govtypes.ModuleName).WithKeyTable(govv1.ParamKeyTable())
	paramsKeeper.Subspace(crisistypes.ModuleName).WithKeyTable(crisistypes.ParamKeyTable())
	
	// IBC modules
	paramsKeeper.Subspace(ibctransfertypes.ModuleName)  // No legacy params
	paramsKeeper.Subspace(IBCStoreKey).WithKeyTable(ibcclienttypes.ParamKeyTable())  // IBC client params
	
	// CosmWasm
	paramsKeeper.Subspace(wasm.ModuleName)  // Handles params internally
	
	// Consensus params
	paramsKeeper.Subspace(baseapp.Paramspace).WithKeyTable(paramstypes.ConsensusParamsKeyTable())

	return paramsKeeper
}
```

---

## When to Use This Guide

Use this guide when you encounter:
- ✅ "parameter X not registered" panic during upgrade
- ✅ Module migration failures
- ✅ Adding new modules with params to your app
- ✅ Upgrading from SDK 0.47 to SDK 0.50+
- ✅ Upgrading IBC modules

---

## Related Documentation

- `IBC_PARAMS_FIX.md` - Detailed analysis of IBC client fix
- `SESSION_SDK50_UPGRADE_FIX.md` - Complete session summary
- `.github/agents/jarvis3.0.agent.md` - Agent directive with patterns
- `APP_MIGRATION_COMPLETE.md` - SDK 0.50 migration guide

---

**Last Updated**: February 12, 2026  
**Status**: Verified and tested
