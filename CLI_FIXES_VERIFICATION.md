# CLI Compilation Fixes - Verification Report

## Compilation Status: ✅ SUCCESS

All CLI code in `x/wasm/client/cli/` now compiles successfully with SDK 0.50.

## Files Fixed

1. ✅ `x/wasm/client/cli/genesis_msg.go`
2. ✅ `x/wasm/client/cli/gov_tx.go`
3. ✅ `x/wasm/client/cli/new_tx.go`
4. ✅ `x/wasm/client/cli/tx.go`

## Issues Fixed

### ✅ Issue 1: GenesisDoc type mismatch (genesis_msg.go:401)
- Changed from `*tmtypes.GenesisDoc` to `*genutiltypes.AppGenesis`
- Updated `GenesisData` struct field from `GenDoc` to `AppGenesis`

### ✅ Issue 2: undefined sdkerrors.Wrap (genesis_msg.go:440, 446)
- Replaced `sdkerrors.Wrap(err, "msg")` with `fmt.Errorf("msg: %w", err)`
- Removed `sdkerrors` import from genesis_msg.go

### ✅ Issue 3: GenesisDoc type mismatch in ExportGenesisFile (genesis_msg.go:450)
- Changed from `g.GenDoc` to `g.AppGenesis`

### ✅ Issue 4: keyring.New needs additional parameters (genesis_msg.go:503)
- Added `clientCtx.Codec` parameter to `keyring.New()` call

### ✅ Issue 5: GetAddress() returns multiple values (genesis_msg.go:512)
- Added error handling for `info.GetAddress()` return value

### ✅ Issue 6: govtypes.NewMsgSubmitProposal undefined (gov_tx.go - 9 occurrences)
Lines fixed: 65, 142, 210, 287, 412, 474, 529, 588, 659
- Wrapped legacy proposals with `v1.NewLegacyContent(content, authority)`
- Used `v1.NewMsgSubmitProposal()` with new signature (7 parameters)
- Removed all `msg.ValidateBasic()` calls (not available in v1 API)

## Additional Fixes

### ✅ new_tx.go
- Replaced `sdkerrors.Wrap` with `fmt.Errorf` (line 46)
- Added `"fmt"` import

### ✅ tx.go
- Replaced `sdkerrors.Wrap` with `fmt.Errorf` (line 106)
- Removed `sdkerrors` import

### ✅ Unused Imports
- Removed `tmtypes` from genesis_msg.go
- Removed `v1beta1` from gov_tx.go

## Verification Commands

```bash
# Verify CLI package compiles
go build -o /dev/null ./x/wasm/client/cli/...
# ✅ Exit code: 0

# Check package completeness
go list -f '{{.Incomplete}}' ./x/wasm/client/cli
# ✅ Output: false

# Verify old APIs are removed
grep -r "sdkerrors\.Wrap" x/wasm/client/cli/*.go
# ✅ Count: 0

grep -r "govtypes\.NewMsgSubmitProposal" x/wasm/client/cli/*.go
# ✅ Count: 0

grep -r "msg\.ValidateBasic" x/wasm/client/cli/gov_tx.go
# ✅ Count: 0

# Verify new APIs are in place
grep -c "v1\.NewLegacyContent" x/wasm/client/cli/gov_tx.go
# ✅ Count: 9 (all proposals wrapped)

grep -c "v1\.NewMsgSubmitProposal" x/wasm/client/cli/gov_tx.go
# ✅ Count: 9 (all proposals using new API)

grep -c "AppGenesis" x/wasm/client/cli/genesis_msg.go
# ✅ Count: 5 (proper usage throughout)
```

## New Imports Added

### genesis_msg.go
- (No new imports, only removed `sdkerrors`)

### gov_tx.go
```go
authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
v1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
```

### new_tx.go
```go
"fmt"
```

### tx.go
- (No new imports, only removed `sdkerrors`)

## Proposal Types Fixed (9 total)

All legacy governance proposals now use the SDK 0.50 v1 API:

1. ✅ StoreCodeProposal
2. ✅ InstantiateContractProposal
3. ✅ MigrateContractProposal
4. ✅ ExecuteContractProposal
5. ✅ SudoContractProposal
6. ✅ UpdateAdminProposal
7. ✅ ClearAdminProposal
8. ✅ PinCodesProposal
9. ✅ UnpinCodesProposal

## Key Pattern Applied

For all governance proposals:

```go
// Create legacy proposal content
content := types.SomeProposal{
    Title:       proposalTitle,
    Description: proposalDescr,
    // ... other fields
}

// Wrap for SDK 0.50
authority := authtypes.NewModuleAddress(govtypes.ModuleName).String()
legacyContent, err := v1.NewLegacyContent(&content, authority)
if err != nil {
    return err
}

// Submit with new API
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

return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
```

## Status

**All CLI compilation errors have been fixed. The x/wasm/client/cli package builds successfully.**

Note: Other parts of the codebase (e.g., x/wasm/module.go, x/wasm/handler.go) have SDK 0.50 compatibility issues that are outside the scope of this CLI-focused fix.
