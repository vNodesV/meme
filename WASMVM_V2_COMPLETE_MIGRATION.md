# Complete wasmvm v2.2.1 Migration Guide

## Executive Summary

Your migration from wasmvm v1.5.9 to v2.2.1 has **significant breaking changes** beyond the basic type renames. After fixing the initial 4 issues you reported, there are **8 additional critical changes** required.

**Status of Initial 4 Issues:** ✅ **FIXED**
1. ✅ StargateMsg → AnyMsg  
2. ✅ GoAPI structure (HumanAddress → HumanizeAddress, etc.)
3. ✅ Events → Array[Event]
4. ✅ Type renames (Delegations, Coins)

**Additional Issues Discovered:** ⚠️ **REQUIRES ACTION**
5. ⚠️ EventCosts signature change
6. ⚠️ msg.Vote.Vote → msg.Vote.Option
7. ⚠️ VM initialization (NewVM → NewVMWithConfig)
8. ⚠️ RequiredFeatures → RequiredCapabilities
9. ⚠️ supportedFeatures string → []string
10. ⚠️ WasmerEngine.Create() signature change
11. ⚠️ KVStore interface mismatch (needs StoreAdapter)
12. ⚠️ Multiple other VM method signature changes

## Recommendation: Use Official wasmd v0.54.5 as Base

**The scope of changes is too large for surgical fixes.** I strongly recommend:

### Option A: Rebase on wasmd v0.54.5 (RECOMMENDED)
**Pros:**
- All wasmvm v2 changes already integrated
- Compatible with SDK 0.50.x (wasmd v0.54 uses SDK 0.50.10)
- Battle-tested in production
- Easier to maintain going forward
- Less risk of missing subtle breaking changes

**Cons:**
- Need to identify and reapply your custom changes
- More work upfront

**How to execute:**
```bash
# 1. Identify your custom changes
git log --oneline --no-merges origin/main ^upstream/main > custom_commits.txt

# 2. Create backup branch
git branch backup-pre-v054-rebase

# 3. Fetch wasmd v0.54.5
git remote add wasmd-upstream https://github.com/CosmWasm/wasmd.git
git fetch wasmd-upstream v0.54.5

# 4. Create migration branch from wasmd v0.54.5
git checkout -b migrate-to-v054 wasmd-upstream/v0.54.5

# 5. Cherry-pick or manually reapply custom changes
# Review each custom commit and apply it

# 6. Update go.mod for cheqd SDK patches
go mod edit -replace github.com/cosmos/cosmos-sdk=github.com/cheqd/cosmos-sdk@v0.50.14-height-mismatch-iavl.0.20250808071119-3b33570d853b
go mod edit -replace cosmossdk.io/store=github.com/cheqd/cosmos-sdk/store@v1.1.2-0.20250808071119-3b33570d853b
go mod tidy
```

### Option B: Complete Surgical Migration (NOT RECOMMENDED)
If you must continue with surgical fixes, here's what remains:

---

## Remaining Breaking Changes (If Doing Surgical Migration)

### 5. EventCosts Signature Change
**File:** `x/wasm/keeper/gas_register.go` ✅ ALREADY FIXED

### 6. VoteMsg Field Rename  
**File:** `x/wasm/keeper/handler_plugin_encoders.go` ✅ ALREADY FIXED

### 7. VM Initialization Complete Rewrite
**File:** `x/wasm/keeper/keeper.go` (line 99)

**Before:**
```go
wasmer, err := wasmvm.NewVM(
    filepath.Join(homeDir, "wasm"),
    supportedFeatures,
    contractMemoryLimit,
    wasmConfig.ContractDebugMode,
    wasmConfig.MemoryCacheSize,
)
```

**After:**
```go
wasmer, err := wasmvm.NewVMWithConfig(wasmvmtypes.VMConfig{
    Cache: wasmvmtypes.CacheOptions{
        BaseDir:                  filepath.Join(homeDir, "wasm"),
        AvailableCapabilities:    strings.Split(supportedFeatures, ","), // Convert string to []string
        MemoryCacheSizeBytes:     wasmvmtypes.NewSizeMebi(wasmConfig.MemoryCacheSize),
        InstanceMemoryLimitBytes: wasmvmtypes.NewSizeMebi(contractMemoryLimit),
    },
}, wasmConfig.ContractDebugMode)
```

**Note:** `supportedFeatures` must be converted from `string` to `[]string`

### 8. RequiredFeatures → RequiredCapabilities
**Files:** `x/wasm/keeper/keeper.go` (lines 181, 193)

**Before:**
```go
report.RequiredFeatures
```

**After:**
```go
report.RequiredCapabilities
```

### 9. WasmerEngine.Create Signature Changed
**File:** `x/wasm/types/wasmer_engine.go` (line 19)

The interface signature doesn't match wasmvm v2. The actual `wasmvm.VM` struct has changed methods.

**Issue:** Your `Keeper.wasmVM` is typed as `types.WasmerEngine` (interface), but `*wasmvm.VM` no longer implements it.

**Solution:** Update the interface or use `*wasmvm.VM` directly.

**Best approach:**
```go
// In keeper.go
type Keeper struct {
    wasmVM *wasmvm.VM  // Use concrete type instead of interface
    // ... other fields
}
```

### 10. KVStore Iterator Interface Mismatch
**Files:** Multiple locations in `x/wasm/keeper/keeper.go`

**Problem:**
```
prefix.Store does not implement "github.com/CosmWasm/wasmvm/v2/types".KVStore
    have Iterator([]byte, []byte) "cosmossdk.io/store/types".Iterator
    want Iterator([]byte, []byte) "github.com/CosmWasm/wasmvm/v2/types".Iterator
```

**Cause:** SDK's `storetypes.Iterator` ≠ wasmvm's `types.Iterator`

**Solution:** Use `StoreAdapter` wrapper

**Add to `x/wasm/types/wasmer_engine.go`:**
```go
import (
    storetypes "cosmossdk.io/store/types"
    wasmvmtypes "github.com/CosmWasm/wasmvm/v2/types"
)

// StoreAdapter bridges SDK KVStore to wasmvm KVStore
type StoreAdapter struct {
    parent storetypes.KVStore
}

// NewStoreAdapter constructor
func NewStoreAdapter(s storetypes.KVStore) *StoreAdapter {
    if s == nil {
        panic("store must not be nil")
    }
    return &StoreAdapter{parent: s}
}

func (s StoreAdapter) Get(key []byte) []byte {
    return s.parent.Get(key)
}

func (s StoreAdapter) Set(key, value []byte) {
    s.parent.Set(key, value)
}

func (s StoreAdapter) Delete(key []byte) {
    s.parent.Delete(key)
}

func (s StoreAdapter) Iterator(start, end []byte) wasmvmtypes.Iterator {
    return s.parent.Iterator(start, end)
}

func (s StoreAdapter) ReverseIterator(start, end []byte) wasmvmtypes.Iterator {
    return s.parent.ReverseIterator(start, end)
}
```

**Then update all VM calls in `keeper.go`:**

**Before:**
```go
k.wasmVM.Instantiate(
    checksum, env, info, initMsg,
    prefixStore,  // ❌ Wrong type
    cosmwasmAPI, querier, gasMeter, gasLimit, costJSONDeserialization,
)
```

**After:**
```go
k.wasmVM.Instantiate(
    checksum, env, info, initMsg,
    types.NewStoreAdapter(prefixStore),  // ✅ Wrapped
    cosmwasmAPI, querier, gasMeter, gasLimit, costJSONDeserialization,
)
```

**Applies to:**
- `Instantiate` (line 283)
- `Execute` (line 351)
- `Migrate` (line 412)
- `Sudo` (line 458)
- All IBC methods
- `Query` methods

---

## Files Requiring Changes

### Already Modified ✅
1. ✅ `x/wasm/keeper/handler_plugin_encoders.go` - StargateMsg → AnyMsg, Vote field
2. ✅ `x/wasm/keeper/api.go` - GoAPI structure
3. ✅ `x/wasm/keeper/events.go` - Events type
4. ✅ `x/wasm/keeper/events_test.go` - Events type
5. ✅ `x/wasm/keeper/query_plugins.go` - Delegations, Coins
6. ✅ `x/wasm/keeper/gas_register.go` - EventCosts signature
7. ✅ `x/wasm/keeper/keeper.go` - Events type, EventCosts call

### Requires Additional Changes ⚠️
8. ⚠️ `x/wasm/keeper/keeper.go` - VM initialization, RequiredFeatures→Capabilities, StoreAdapter
9. ⚠️ `x/wasm/types/wasmer_engine.go` - Add StoreAdapter, possibly update interface

---

## Complexity Assessment

| Task | Complexity | Risk | Effort |
|------|-----------|------|--------|
| Initial 4 fixes (completed) | Low | Low | ✅ Done |
| EventCosts, Vote fixes | Low | Low | ✅ Done |
| VM initialization | Medium | Medium | 1-2 hours |
| StoreAdapter implementation | Medium | Medium | 2-3 hours |
| Testing all changes | High | High | 4-8 hours |
| **Surgical Total** | **High** | **High** | **~12 hours** |
| **Rebase on v0.54.5** | **Medium** | **Low** | **~4-6 hours** |

---

## Testing Checklist

After completing migration (either approach):

### Unit Tests
```bash
make test-unit
go test ./x/wasm/...
```

### Integration Tests
```bash
# If you have integration tests
make test-integration
```

### Manual Contract Testing
1. Deploy a simple contract (e.g., cw20-base)
2. Instantiate it
3. Execute a message
4. Query state
5. Migrate to new code
6. Test IBC if applicable

### Upgrade Testing (for live chain)
1. Run old binary with existing state
2. Submit upgrade proposal
3. Halt at upgrade height
4. Swap binary
5. Restart and verify blocks continue
6. Verify all existing contracts still work

---

## Decision Matrix

| Factor | Surgical Fixes | Rebase v0.54.5 |
|--------|---------------|----------------|
| **Time to complete** | ~12 hours | ~6 hours |
| **Risk of bugs** | High | Low |
| **Future maintainability** | Hard | Easy |
| **Stay close to upstream** | No | Yes |
| **Custom code preservation** | Automatic | Manual |
| **Testing burden** | Heavy | Moderate |
| **Recommended?** | ❌ No | ✅ Yes |

---

## My Recommendation

**Rebase on wasmd v0.54.5.** Here's why:

1. **Lower Risk:** wasmd v0.54 is production-tested with wasmvm v2
2. **Less Work:** ~6 hours vs ~12 hours
3. **Maintainable:** Easier to track upstream changes
4. **Complete:** You won't miss subtle breaking changes
5. **SDK Compatible:** v0.54.5 uses SDK 0.50.10, close to your 0.50.14

The only downside is you need to manually reapply custom changes, but:
- You can identify them with `git log`
- Most custom changes are likely in `app/` and config, not `x/wasm/keeper/`
- You already have cheqd SDK patches in `go.mod`, so that's easy

---

## Next Steps

**If you choose to rebase (recommended):**
1. I can help identify your custom changes
2. Guide you through the rebase process
3. Help reapply custom patches
4. Update go.mod for cheqd patches

**If you choose surgical fixes:**
1. I'll implement the StoreAdapter
2. Fix VM initialization
3. Update all VM call sites
4. Help with testing

**What would you like to do?**
