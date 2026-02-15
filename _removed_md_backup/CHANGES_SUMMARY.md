# Quick Reference: Changes Made

## 1. handler_plugin_encoders.go

```diff
- type StargateEncoder func(sender sdk.AccAddress, msg *wasmvmtypes.StargateMsg) ([]sdk.Msg, error)
+ type AnyEncoder func(sender sdk.AccAddress, msg *wasmvmtypes.AnyMsg) ([]sdk.Msg, error)

  type MessageEncoders struct {
-     Stargate func(sender sdk.AccAddress, msg *wasmvmtypes.StargateMsg) ([]sdk.Msg, error)
+     Any      func(sender sdk.AccAddress, msg *wasmvmtypes.AnyMsg) ([]sdk.Msg, error)
  }

  func DefaultEncoders(...) MessageEncoders {
      return MessageEncoders{
-         Stargate: EncodeStargateMsg(unpacker),
+         Any:      EncodeAnyMsg(unpacker),
      }
  }

- func EncodeStargateMsg(unpacker codectypes.AnyUnpacker) StargateEncoder {
-     return func(sender sdk.AccAddress, msg *wasmvmtypes.StargateMsg) ([]sdk.Msg, error) {
+ func EncodeAnyMsg(unpacker codectypes.AnyUnpacker) AnyEncoder {
+     return func(sender sdk.AccAddress, msg *wasmvmtypes.AnyMsg) ([]sdk.Msg, error) {

- case msg.Stargate != nil:
-     return e.Stargate(contractAddr, msg.Stargate)
+ case msg.Any != nil:
+     return e.Any(contractAddr, msg.Any)

  func EncodeGovMsg(...) {
-     voteOption, err := convertVoteOption(msg.Vote.Vote)
+     voteOption, err := convertVoteOption(msg.Vote.Option)
  }
```

## 2. api.go

```diff
  const (
+     DefaultGasCostValidateAddress = DefaultGasCostHumanAddress + DefaultGasCostCanonicalAddress
  )

  var (
+     costValidate = DefaultGasCostValidateAddress * types.DefaultGasMultiplier
  )

- func humanAddress(canon []byte) (string, uint64, error) {
+ func humanizeAddress(canon []byte) (string, uint64, error) {

- func canonicalAddress(human string) ([]byte, uint64, error) {
+ func canonicalizeAddress(human string) ([]byte, uint64, error) {

+ func validateAddress(human string) (uint64, error) {
+     canonicalized, err := sdk.AccAddressFromBech32(human)
+     if err != nil {
+         return costValidate, err
+     }
+     if canonicalized.String() != human {
+         return costValidate, wasmvmtypes.InvalidRequest{Err: "address not normalized"}
+     }
+     return costValidate, nil
+ }

  var cosmwasmAPI = wasmvm.GoAPI{
-     HumanAddress:     humanAddress,
-     CanonicalAddress: canonicalAddress,
+     HumanizeAddress:     humanizeAddress,
+     CanonicalizeAddress: canonicalizeAddress,
+     ValidateAddress:     validateAddress,
  }
```

## 3. events.go

```diff
- func newCustomEvents(evts wasmvmtypes.Events, contractAddr sdk.AccAddress) (sdk.Events, error) {
+ func newCustomEvents(evts wasmvmtypes.Array[wasmvmtypes.Event], contractAddr sdk.AccAddress) (sdk.Events, error) {
```

## 4. events_test.go

```diff
  specs := map[string]struct {
-     src     wasmvmtypes.Events
+     src     wasmvmtypes.Array[wasmvmtypes.Event]
  }{
      "all good": {
-         src: wasmvmtypes.Events{{
+         src: wasmvmtypes.Array[wasmvmtypes.Event]{{
```

## 5. query_plugins.go

```diff
- func sdkToDelegations(...) (wasmvmtypes.Delegations, error) {
+ func sdkToDelegations(...) (wasmvmtypes.Array[wasmvmtypes.Delegation], error) {
-     result := make([]wasmvmtypes.Delegation, len(delegations))
+     result := make(wasmvmtypes.Array[wasmvmtypes.Delegation], len(delegations))

- func ConvertSdkCoinsToWasmCoins(coins []sdk.Coin) wasmvmtypes.Coins {
+ func ConvertSdkCoinsToWasmCoins(coins []sdk.Coin) wasmvmtypes.Array[wasmvmtypes.Coin] {
-     converted := make(wasmvmtypes.Coins, len(coins))
+     converted := make(wasmvmtypes.Array[wasmvmtypes.Coin], len(coins))
```

## 6. gas_register.go

```diff
- func (g WasmGasRegister) EventCosts(attrs []wasmvmtypes.EventAttribute) storetypes.Gas {
+ func (g WasmGasRegister) EventCosts(attrs []wasmvmtypes.EventAttribute, events wasmvmtypes.Array[wasmvmtypes.Event]) storetypes.Gas {
```

## 7. keeper.go

```diff
  func dispatchMessages(
      attrs []wasmvmtypes.EventAttribute,
-     evts wasmvmtypes.Events,
+     evts wasmvmtypes.Array[wasmvmtypes.Event],
  ) ([]byte, error) {
-     attributeGasCost := k.gasRegister.EventCosts(attrs)
+     attributeGasCost := k.gasRegister.EventCosts(attrs, evts)
```

## Statistics

- **Files modified:** 7
- **Lines changed:** ~30
- **Breaking changes fixed:** 7 / 12
- **Compilation errors:** 13 → 5
- **Time to fix:** ~30 minutes

## Verification

```bash
# Check changes
git diff --stat

# Attempt build (will show remaining 5 errors)
go build ./x/wasm/keeper/...
```

## What's Left

The remaining 5 errors all require architectural changes:
1. VM initialization (NewVM → NewVMWithConfig)
2. RequiredFeatures → RequiredCapabilities (2 locations)
3. KVStore adapter implementation
4. Update ~15 VM call sites

See `WASMVM_V2_COMPLETE_MIGRATION.md` for details.
