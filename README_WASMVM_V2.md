# wasmvm v2.2.1 Migration Status

## TL;DR

✅ **Fixed 7/12 breaking changes** - All issues you reported are resolved + 3 more
⚠️ **5 issues remaining** - All require architectural refactoring (VM init, KVStore adapter)
📊 **Compilation:** 13 errors → 5 errors
🎯 **Recommendation:** Rebase on wasmd v0.54.5 (saves time, reduces risk)

---

## What You Asked For

You reported 4 compilation errors:
1. ❌ `StargateMsg` undefined
2. ❌ `GoAPI` field names changed
3. ❌ `Events` type removed
4. ❌ `Delegations`, `Coins` undefined

## What I Delivered

### ✅ All 4 Issues Fixed + 3 Bonus Fixes

| # | Issue | Status | File(s) |
|---|-------|--------|---------|
| 1 | StargateMsg → AnyMsg | ✅ Fixed | handler_plugin_encoders.go |
| 2 | GoAPI structure | ✅ Fixed | api.go |
| 3 | Events → Array[Event] | ✅ Fixed | events.go, events_test.go, keeper.go |
| 4 | Delegations, Coins | ✅ Fixed | query_plugins.go |
| 5 | EventCosts signature | ✅ Fixed | gas_register.go, keeper.go |
| 6 | Vote field rename | ✅ Fixed | handler_plugin_encoders.go |
| 7 | Address validation | ✅ Fixed | api.go |

### ⚠️ 5 Additional Issues Discovered

During compilation, wasmvm v2 revealed deeper architectural changes:

| # | Issue | Impact | Effort |
|---|-------|--------|--------|
| 8 | VM init changed | Critical | High |
| 9 | Features→Capabilities | Medium | Low |
| 10 | Engine interface | Critical | High |
| 11 | KVStore adapter | Critical | High |
| 12 | VM method signatures | High | Medium |

**These require ~10-12 hours of additional work if fixed surgically.**

---

## Option A: Rebase on wasmd v0.54.5 ⭐ RECOMMENDED

### Why?
- ✅ All wasmvm v2 changes already integrated
- ✅ Production-tested by CosmWasm team
- ✅ Compatible with SDK 0.50.x (you're on 0.50.14, they're on 0.50.10)
- ✅ Saves ~6 hours of work
- ✅ Lower risk of bugs
- ✅ Easier to maintain

### How?
```bash
# 1. Backup
git branch backup-$(date +%Y%m%d)

# 2. Fetch wasmd v0.54.5
git remote add wasmd https://github.com/CosmWasm/wasmd.git
git fetch wasmd v0.54.5

# 3. Create migration branch
git checkout -b migrate-v054 wasmd/v0.54.5

# 4. Identify custom changes (from your old branch)
git log --oneline <your-old-branch> ^wasmd/v0.54.5 > custom_changes.txt

# 5. Reapply custom changes
# (Review custom_changes.txt and cherry-pick or manually apply)

# 6. Update go.mod for cheqd patches
go mod edit -replace github.com/cosmos/cosmos-sdk=github.com/cheqd/cosmos-sdk@v0.50.14-height-mismatch-iavl.0.20250808071119-3b33570d853b
go mod edit -replace cosmossdk.io/store=github.com/cheqd/cosmos-sdk/store@v1.1.2-0.20250808071119-3b33570d853b
go mod tidy

# 7. Test
make build
make test
```

**Estimated time:** 4-6 hours

---

## Option B: Continue Surgical Fixes

### What's Required

1. **Implement StoreAdapter** (~1 hour)
   - Create wrapper in `x/wasm/types/wasmer_engine.go`
   - Adapts SDK Iterator to wasmvm Iterator

2. **Fix VM Initialization** (~1 hour)
   ```go
   // Before
   wasmvm.NewVM(homeDir, features, limit, debug, cache)
   
   // After
   wasmvm.NewVMWithConfig(wasmvmtypes.VMConfig{
       Cache: wasmvmtypes.CacheOptions{
           BaseDir: homeDir,
           AvailableCapabilities: strings.Split(features, ","),
           MemoryCacheSizeBytes: wasmvmtypes.NewSizeMebi(cache),
           InstanceMemoryLimitBytes: wasmvmtypes.NewSizeMebi(limit),
       },
   }, debug)
   ```

3. **Update All VM Calls** (~3 hours)
   - Instantiate, Execute, Migrate, Sudo, Query
   - All IBC methods
   - ~15 call sites total
   - Each needs `types.NewStoreAdapter(store)`

4. **Testing** (~4-6 hours)
   - Unit tests
   - Integration tests
   - Manual contract testing

**Estimated time:** 10-12 hours

**Risk:** High - easy to miss subtle breaking changes

---

## Documentation Created

I've created comprehensive guides for you:

1. **`MIGRATION_SUMMARY.md`** ⭐ START HERE
   - Executive overview
   - Decision matrix
   - Next steps

2. **`WASMVM_V2_MIGRATION_GUIDE.md`**
   - Detailed guide for initial 4 fixes
   - Migration patterns
   - Code examples

3. **`WASMVM_V2_COMPLETE_MIGRATION.md`**
   - All 12 breaking changes documented
   - Complete surgical fix instructions
   - Complexity assessment

4. **`WASMVM_V2_FIXES_APPLIED.md`**
   - What's been fixed
   - What remains
   - Testing checklist

5. **`CHANGES_SUMMARY.md`**
   - Quick diff reference
   - Before/after code
   - Statistics

6. **`README_WASMVM_V2.md`** (this file)
   - Quick reference
   - Decision guide

---

## Current Status

### Build Status
```
Compilation errors: 5 remaining
All errors in: x/wasm/keeper/keeper.go
All related to: VM initialization & KVStore interface
```

### Modified Files
```
x/wasm/keeper/handler_plugin_encoders.go  ✅ Complete
x/wasm/keeper/api.go                      ✅ Complete
x/wasm/keeper/events.go                   ✅ Complete
x/wasm/keeper/events_test.go              ✅ Complete
x/wasm/keeper/query_plugins.go            ✅ Complete
x/wasm/keeper/gas_register.go             ✅ Complete
x/wasm/keeper/keeper.go                   ⚠️  Partial (events only)
x/wasm/types/wasmer_engine.go             ⚠️  Needs StoreAdapter
```

---

## Quick Decision Guide

**Choose Option A (Rebase) if:**
- ✅ You want the fastest path to completion
- ✅ You value production-tested code
- ✅ Your custom changes are minimal
- ✅ You want easier future maintenance

**Choose Option B (Surgical) if:**
- ⚠️ You have extensive custom wasm module changes
- ⚠️ You need to understand every change
- ⚠️ You have time for thorough testing
- ⚠️ Rebasing is not an option for your workflow

**My recommendation:** Option A saves 4-6 hours and has lower risk.

---

## Next Steps

### If Choosing Option A (Rebase)
1. Review `custom_changes.txt` to identify what needs reapplying
2. Follow the rebase steps above
3. Test thoroughly
4. I can help with any conflicts or issues

### If Choosing Option B (Surgical)
1. Review `WASMVM_V2_COMPLETE_MIGRATION.md` sections 7-12
2. Implement StoreAdapter
3. Fix VM initialization
4. Update all VM call sites
5. Extensive testing required
6. I can complete these fixes if needed

---

## Testing After Migration

### Minimum Testing
```bash
# Build
make build

# Unit tests
go test ./x/wasm/...

# Integration (if available)
make test-integration
```

### Contract Testing
1. Deploy simple contract (cw20-base)
2. Instantiate with init msg
3. Execute message
4. Query state
5. Migrate to new code
6. Test IBC (if applicable)

### Upgrade Testing (Live Chain)
1. Run old binary
2. Submit upgrade proposal
3. Wait for upgrade height
4. Halt naturally
5. Swap to new binary
6. Verify blocks continue
7. Test existing contracts work

---

## Questions?

**About the fixes I made:**
- See `CHANGES_SUMMARY.md` for exact diffs
- See `WASMVM_V2_FIXES_APPLIED.md` for detailed explanations

**About remaining work:**
- See `WASMVM_V2_COMPLETE_MIGRATION.md` sections 7-12
- Includes code examples and full explanation

**About the recommendation:**
- See `MIGRATION_SUMMARY.md` decision matrix
- Includes time/risk/effort comparison

---

## References

- **wasmvm v2 Release:** https://github.com/CosmWasm/wasmvm/releases/tag/v2.0.0
- **wasmd v0.54.5:** https://github.com/CosmWasm/wasmd/tree/v0.54.5
- **CosmWasm Docs:** https://docs.cosmwasm.com/

---

**Created by:** AI Assistant analyzing your wasmvm v1 → v2.2.1 migration
**Date:** Based on your current compilation errors
**Status:** 7/12 fixes complete, recommendation provided
