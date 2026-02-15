# app/export.go SDK 0.50 Migration Patterns Reference

Quick reference guide for the patterns used in the app/export.go migration.

## Pattern 1: Context Creation

**SDK 0.47:**
```go
ctx := app.NewContext(true, tmproto.Header{Height: app.LastBlockHeight()})
```

**SDK 0.50:**
```go
ctx := app.NewContext(true)
```

**Why:** SDK 0.50 simplified context creation. Height is managed internally.

---

## Pattern 2: Module Manager Export

**SDK 0.47:**
```go
genState := app.mm.ExportGenesis(ctx, app.appCodec)
```

**SDK 0.50:**
```go
genState, err := app.mm.ExportGenesis(ctx, app.appCodec)
if err != nil {
    return servertypes.ExportedApp{}, err
}
```

**Why:** Added error return for better error propagation.

---

## Pattern 3: Keeper Pointer Requirement

**SDK 0.47:**
```go
validators, err := staking.WriteValidators(ctx, app.stakingKeeper)
```

**SDK 0.50:**
```go
validators, err := staking.WriteValidators(ctx, &app.stakingKeeper)
```

**Why:** Function signature requires pointer to keeper.

---

## Pattern 4: String to ValAddress Conversion

**SDK 0.47:**
```go
app.distrKeeper.WithdrawValidatorCommission(ctx, val.GetOperator())
```

**SDK 0.50:**
```go
valAddr, err := sdk.ValAddressFromBech32(val.GetOperator())
if err != nil {
    panic(err)
}
_, _ = app.distrKeeper.WithdrawValidatorCommission(ctx, valAddr)
```

**Why:** Type safety - `GetOperator()` returns string, methods need `sdk.ValAddress`.

---

## Pattern 5: Collection-Based Keeper Methods

**SDK 0.47:**
```go
dels := app.stakingKeeper.GetAllDelegations(ctx)
```

**SDK 0.50:**
```go
dels, err := app.stakingKeeper.GetAllDelegations(ctx)
if err != nil {
    panic(err)
}
```

**Why:** Collections API returns errors for better handling.

---

## Pattern 6: FeePool Collections API

**SDK 0.47:**
```go
feePool := app.distrKeeper.GetFeePool(ctx)
feePool.CommunityPool = feePool.CommunityPool.Add(scraps...)
app.distrKeeper.SetFeePool(ctx, feePool)
```

**SDK 0.50:**
```go
feePool, err := app.distrKeeper.FeePool.Get(ctx)
if err != nil {
    panic(err)
}
feePool.CommunityPool = feePool.CommunityPool.Add(scraps...)
if err := app.distrKeeper.FeePool.Set(ctx, feePool); err != nil {
    panic(err)
}
```

**Why:** FeePool is now a collection field with Get/Set methods.

---

## Pattern 7: Hook Error Returns

**SDK 0.47:**
```go
app.distrKeeper.Hooks().AfterValidatorCreated(ctx, valAddr)
```

**SDK 0.50:**
```go
if err := app.distrKeeper.Hooks().AfterValidatorCreated(ctx, valAddr); err != nil {
    panic(err)
}
```

**Why:** All hooks now return errors for proper error handling.

---

## Pattern 8: Store Iterator via Keeper

**SDK 0.47:**
```go
store := ctx.KVStore(app.keys[stakingtypes.StoreKey])
iter := sdk.KVStoreReversePrefixIterator(store, stakingtypes.ValidatorsKey)
defer iter.Close()

for ; iter.Valid(); iter.Next() {
    addr := sdk.ValAddress(iter.Key()[1:])
    // ...
}
```

**SDK 0.50:**
```go
iter, err := app.stakingKeeper.ValidatorsPowerStoreIterator(ctx)
if err != nil {
    panic(err)
}
defer iter.Close()

for ; iter.Valid(); iter.Next() {
    addr := sdk.ValAddress(stakingtypes.ParseValidatorPowerRankKey(iter.Key()))
    // ...
}
```

**Why:** 
- Encapsulation - use keeper methods instead of raw store access
- Proper key parsing with `ParseValidatorPowerRankKey()`
- Error handling for iterator creation

---

## Pattern 9: GetValidator Error Return

**SDK 0.47:**
```go
validator, found := app.stakingKeeper.GetValidator(ctx, addr)
if !found {
    panic("expected validator, not found")
}
```

**SDK 0.50:**
```go
validator, err := app.stakingKeeper.GetValidator(ctx, addr)
if err != nil {
    panic("expected validator, not found")
}
```

**Why:** Changed from `(Validator, bool)` to `(Validator, error)`.

---

## Pattern 10: SetValidator Error Return

**SDK 0.47:**
```go
app.stakingKeeper.SetValidator(ctx, validator)
```

**SDK 0.50:**
```go
if err := app.stakingKeeper.SetValidator(ctx, validator); err != nil {
    panic(err)
}
```

**Why:** Now returns error for state storage issues.

---

## Pattern 11: Dual Return Values

**SDK 0.47:**
```go
_, err := app.stakingKeeper.ApplyAndReturnValidatorSetUpdates(ctx)
```

**SDK 0.50:**
```go
if _, err = app.stakingKeeper.ApplyAndReturnValidatorSetUpdates(ctx); err != nil {
    log.Fatal(err)
}
```

**Why:** Returns both validator updates and error. Existing `err` variable allows `=` instead of `:=`.

---

## Common Error Handling Pattern

Throughout the file, we use this pattern for non-recoverable errors:

```go
result, err := keeper.Method(ctx, params)
if err != nil {
    panic(err)  // Or log.Fatal(err) for top-level functions
}
```

**Why:** Export is critical for chain state; failures should be loud.

---

## Iterator Best Practices

1. **Always check error on creation:**
   ```go
   iter, err := keeper.Iterator(ctx)
   if err != nil {
       return err
   }
   ```

2. **Always defer Close():**
   ```go
   defer iter.Close()
   ```

3. **Use proper key parsing:**
   ```go
   addr := sdk.ValAddress(stakingtypes.ParseValidatorPowerRankKey(iter.Key()))
   ```

---

## Migration Checklist

When migrating similar code:

- [ ] Update context creation (remove Header)
- [ ] Add error handling to all keeper methods
- [ ] Convert GetOperator() strings to sdk.ValAddress
- [ ] Replace GetFeePool/SetFeePool with FeePool.Get/Set
- [ ] Use keeper iterator methods instead of raw store
- [ ] Update GetValidator to expect error not bool
- [ ] Add error returns to SetValidator calls
- [ ] Check for dual return values on state update methods
- [ ] Add error handling to all hooks
- [ ] Use proper key parsing functions

---

## Testing After Migration

```bash
# Build test
go build ./app

# Syntax check
gofmt -l app/export.go

# Full build
make install
```

---

## References

- SDK 0.50 Upgrade Guide: https://github.com/cosmos/cosmos-sdk/blob/release/v0.50.x/UPGRADING.md
- Collections API: https://docs.cosmos.network/main/build/packages/collections
- Store Service: https://docs.cosmos.network/main/build/packages/store
