# CLI Compilation Fixes for SDK 0.50 Compatibility

## Summary

Fixed all SDK 0.50 compatibility issues in the `x/wasm/client/cli/` directory. All CLI code now compiles successfully.

## Files Modified

1. **x/wasm/client/cli/genesis_msg.go**
2. **x/wasm/client/cli/gov_tx.go**
3. **x/wasm/client/cli/new_tx.go**
4. **x/wasm/client/cli/tx.go**

## Detailed Changes

### 1. genesis_msg.go

#### Issue 1: GenesisDoc type mismatch (lines 389-404)
- **Problem**: `GenesisStateFromGenFile` now returns `*AppGenesis` instead of `*GenesisDoc` in SDK 0.50
- **Fix**: Updated `GenesisData` struct and `ReadWasmGenesis` function to use `*genutiltypes.AppGenesis`

```go
// Before
type GenesisData struct {
    GenesisFile     string
    GenDoc          *tmtypes.GenesisDoc
    AppState        map[string]json.RawMessage
    WasmModuleState *types.GenesisState
}

// After
type GenesisData struct {
    GenesisFile     string
    AppGenesis      *genutiltypes.AppGenesis
    AppState        map[string]json.RawMessage
    WasmModuleState *types.GenesisState
}
```

#### Issue 2: undefined sdkerrors.Wrap (lines 440, 446)
- **Problem**: `sdkerrors.Wrap` is deprecated in SDK 0.50
- **Fix**: Replaced with `fmt.Errorf` with `%w` verb for error wrapping
- **Removed import**: `sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"`

```go
// Before
return sdkerrors.Wrap(err, "marshal wasm genesis state")

// After
return fmt.Errorf("marshal wasm genesis state: %w", err)
```

#### Issue 3: ExportGenesisFile parameter type (line 450)
- **Problem**: `ExportGenesisFile` expects `*AppGenesis` not `*GenesisDoc`
- **Fix**: Updated to use `g.AppGenesis` instead of `g.GenDoc`

```go
// Before
return genutil.ExportGenesisFile(g.GenDoc, g.GenesisFile)

// After
return genutil.ExportGenesisFile(g.AppGenesis, g.GenesisFile)
```

#### Issue 4: keyring.New signature change (line 503)
- **Problem**: `keyring.New` requires codec and options parameters in SDK 0.50
- **Fix**: Added `clientCtx.Codec` parameter

```go
// Before
kb, err := keyring.New(sdk.KeyringServiceName(), keyringBackend, homeDir, inBuf)

// After
kb, err := keyring.New(sdk.KeyringServiceName(), keyringBackend, homeDir, inBuf, clientCtx.Codec)
```

#### Issue 5: GetAddress() return signature (line 512)
- **Problem**: `GetAddress()` now returns `(sdk.AccAddress, error)` instead of just `sdk.AccAddress`
- **Fix**: Added error handling

```go
// Before
return info.GetAddress(), nil

// After
addr, err := info.GetAddress()
if err != nil {
    return nil, fmt.Errorf("failed to get address from key info: %w", err)
}
return addr, nil
```

#### Issue 6: Removed unused import
- **Removed**: `tmtypes "github.com/cometbft/cometbft/types"`

### 2. gov_tx.go

#### Issue: govtypes.NewMsgSubmitProposal signature change (9 occurrences: lines 65, 142, 210, 287, 412, 474, 529, 588, 659)
- **Problem**: In SDK 0.50, governance proposals use the new v1 API. Legacy content-based proposals need to be wrapped using `v1.NewLegacyContent` before submission
- **Fix**: Updated all 9 proposal submission functions to:
  1. Wrap legacy proposal content with `v1.NewLegacyContent`
  2. Use `v1.NewMsgSubmitProposal` with new signature
  3. Remove `msg.ValidateBasic()` calls (not available in v1 API)

**New imports added**:
```go
authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
v1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
```

**Removed import**:
```go
"github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"
```

**Pattern applied to all proposal types**:
- StoreCodeProposal
- InstantiateContractProposal
- MigrateContractProposal
- ExecuteContractProposal
- SudoContractProposal
- UpdateAdminProposal
- ClearAdminProposal
- PinCodesProposal
- UnpinCodesProposal

```go
// Before
content := types.StoreCodeProposal{...}
msg, err := govtypes.NewMsgSubmitProposal(&content, deposit, clientCtx.GetFromAddress())
if err != nil {
    return err
}
if err = msg.ValidateBasic(); err != nil {
    return err
}

// After
content := types.StoreCodeProposal{...}

// Wrap legacy content for SDK 0.50
authority := authtypes.NewModuleAddress(govtypes.ModuleName).String()
legacyContent, err := v1.NewLegacyContent(&content, authority)
if err != nil {
    return err
}

msg, err := v1.NewMsgSubmitProposal(
    []sdk.Msg{legacyContent},
    deposit,
    clientCtx.GetFromAddress().String(),
    "", // metadata
    proposalTitle,
    proposalDescr, // summary
    false,        // expedited
)
if err != nil {
    return err
}
```

### 3. new_tx.go

#### Issue: undefined sdkerrors.Wrap (line 46)
- **Problem**: `sdkerrors.Wrap` is deprecated in SDK 0.50
- **Fix**: Replaced with `fmt.Errorf` with `%w` verb
- **Added import**: `"fmt"`
- **Removed import**: `sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"`

```go
// Before
return types.MsgMigrateContract{}, sdkerrors.Wrap(err, "code id")

// After
return types.MsgMigrateContract{}, fmt.Errorf("code id: %w", err)
```

### 4. tx.go

#### Issue: undefined sdkerrors.Wrap (line 106)
- **Problem**: `sdkerrors.Wrap` is deprecated in SDK 0.50
- **Fix**: Replaced with `fmt.Errorf` with `%w` verb
- **Removed import**: `sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"`

```go
// Before
return types.MsgStoreCode{}, sdkerrors.Wrap(err, flagInstantiateByAddress)

// After
return types.MsgStoreCode{}, fmt.Errorf("%s: %w", flagInstantiateByAddress, err)
```

## SDK 0.50 Key API Changes Applied

### 1. Error Wrapping
- **Old**: `sdkerrors.Wrap(err, "message")`
- **New**: `fmt.Errorf("message: %w", err)`

### 2. Genesis Handling
- **Old**: `genutiltypes.GenesisStateFromGenFile()` returns `(map, *GenesisDoc, error)`
- **New**: `genutiltypes.GenesisStateFromGenFile()` returns `(map, *AppGenesis, error)`
- **Old**: `genutil.ExportGenesisFile(*GenesisDoc, string)`
- **New**: `genutil.ExportGenesisFile(*AppGenesis, string)`

### 3. Keyring API
- **Old**: `keyring.New(name, backend, dir, input)`
- **New**: `keyring.New(name, backend, dir, input, codec, ...opts)`
- **Old**: `info.GetAddress()` returns `sdk.AccAddress`
- **New**: `info.GetAddress()` returns `(sdk.AccAddress, error)`

### 4. Governance Proposals
- **Old**: Content-based proposals with `govtypes.NewMsgSubmitProposal(content, deposit, proposer)`
- **New**: Message-based proposals with legacy content wrapper:
  1. Wrap: `v1.NewLegacyContent(content, authority)`
  2. Submit: `v1.NewMsgSubmitProposal(messages, deposit, proposer, metadata, title, summary, expedited)`
- **Removed**: `msg.ValidateBasic()` no longer exists on `v1.MsgSubmitProposal`

## Verification

All CLI packages now compile successfully:

```bash
$ go build -o /dev/null ./x/wasm/client/cli/...
# Success - no errors

$ go list -f '{{.Incomplete}}' ./x/wasm/client/cli
false
```

## Notes

- These fixes only address the CLI layer (`x/wasm/client/cli/`)
- Other SDK 0.50 compatibility issues exist in other parts of the codebase (e.g., `x/wasm/module.go`, `x/wasm/handler.go`, etc.)
- The governance proposal system in SDK 0.50 moved from content-based (v1beta1) to message-based (v1), but maintains backward compatibility through the `MsgExecLegacyContent` wrapper
- All legacy proposal types (StoreCode, Instantiate, Migrate, Execute, Sudo, UpdateAdmin, ClearAdmin, PinCodes, UnpinCodes) continue to work with the new wrapper approach
