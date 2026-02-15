# SDK 0.50 CLI Migration - Quick Reference

## Quick Fix Guide

### Error: `undefined: sdkerrors.Wrap`
```go
// OLD ❌
return err, sdkerrors.Wrap(err, "message")

// NEW ✅
return err, fmt.Errorf("message: %w", err)
```

### Error: `GenesisDoc` type mismatch
```go
// OLD ❌
appState, genDoc, err := genutiltypes.GenesisStateFromGenFile(genFile)
type GenesisData struct {
    GenDoc *tmtypes.GenesisDoc
}
genutil.ExportGenesisFile(genDoc, file)

// NEW ✅
appState, appGenesis, err := genutiltypes.GenesisStateFromGenFile(genFile)
type GenesisData struct {
    AppGenesis *genutiltypes.AppGenesis
}
genutil.ExportGenesisFile(appGenesis, file)
```

### Error: `keyring.New` missing parameters
```go
// OLD ❌
kb, err := keyring.New(name, backend, dir, input)

// NEW ✅
kb, err := keyring.New(name, backend, dir, input, clientCtx.Codec)
```

### Error: `GetAddress()` returns multiple values
```go
// OLD ❌
return info.GetAddress(), nil

// NEW ✅
addr, err := info.GetAddress()
if err != nil {
    return nil, fmt.Errorf("failed to get address: %w", err)
}
return addr, nil
```

### Error: `govtypes.NewMsgSubmitProposal` undefined
```go
// OLD ❌
import govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"

content := types.SomeProposal{
    Title:       title,
    Description: description,
    // ... fields
}

msg, err := govtypes.NewMsgSubmitProposal(&content, deposit, proposer)
if err != nil {
    return err
}
if err = msg.ValidateBasic(); err != nil {
    return err
}

// NEW ✅
import (
    authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
    govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
    v1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
)

content := types.SomeProposal{
    Title:       title,
    Description: description,
    // ... fields
}

// Wrap legacy content
authority := authtypes.NewModuleAddress(govtypes.ModuleName).String()
legacyContent, err := v1.NewLegacyContent(&content, authority)
if err != nil {
    return err
}

// Submit with new v1 API
msg, err := v1.NewMsgSubmitProposal(
    []sdk.Msg{legacyContent},  // messages
    deposit,                    // initial deposit
    proposer.String(),          // proposer address as string
    "",                         // metadata
    title,                      // title
    description,                // summary
    false,                      // expedited
)
if err != nil {
    return err
}

// Note: ValidateBasic() no longer exists on v1.MsgSubmitProposal
```

## Import Changes

### Remove
```go
sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
tmtypes "github.com/cometbft/cometbft/types"
"github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"
```

### Add
```go
"fmt" // for error wrapping
authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
v1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
```

## Common Patterns

### Error Wrapping Pattern
```go
// Always use fmt.Errorf with %w for error wrapping
if err != nil {
    return fmt.Errorf("context message: %w", err)
}
```

### Governance Proposal Pattern
For ALL legacy content-based proposals:
1. Create the proposal content
2. Get gov module authority: `authtypes.NewModuleAddress(govtypes.ModuleName).String()`
3. Wrap: `v1.NewLegacyContent(&content, authority)`
4. Submit: `v1.NewMsgSubmitProposal([]sdk.Msg{wrapped}, deposit, proposer, "", title, summary, false)`
5. DO NOT call `msg.ValidateBasic()` - it doesn't exist in v1

### Genesis Handling Pattern
```go
// Read genesis
appState, appGenesis, err := genutiltypes.GenesisStateFromGenFile(genFile)

// Store in struct
type GenesisData struct {
    AppGenesis *genutiltypes.AppGenesis
    // ... other fields
}

// Export genesis
appGenesis.AppState = appStateJSON
return genutil.ExportGenesisFile(appGenesis, genFile)
```

## Testing

```bash
# Build CLI package
go build -o /dev/null ./x/wasm/client/cli/...

# Check if package is complete
go list -f '{{.Incomplete}}' ./x/wasm/client/cli
# Should output: false

# Verify old APIs removed
grep -r "sdkerrors\.Wrap" x/wasm/client/cli/*.go
# Should return nothing

# Verify new APIs present
grep -c "v1\.NewMsgSubmitProposal" x/wasm/client/cli/gov_tx.go
# Should return: 9 (or number of proposals you have)
```

## Reference Links

- SDK 0.50 Migration Guide: https://docs.cosmos.network/v0.50/build/migrations/upgrade-to-50
- Gov Module v1 API: https://docs.cosmos.network/v0.50/build/modules/gov
- Error Handling: https://go.dev/blog/go1.13-errors

## Files Changed in This Migration

- `x/wasm/client/cli/genesis_msg.go` - Genesis handling
- `x/wasm/client/cli/gov_tx.go` - All 9 proposal types  
- `x/wasm/client/cli/new_tx.go` - Error wrapping
- `x/wasm/client/cli/tx.go` - Error wrapping

For detailed explanations, see `CLI_FIXES_SDK_050.md`
