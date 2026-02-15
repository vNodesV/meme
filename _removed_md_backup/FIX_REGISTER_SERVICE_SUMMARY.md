# Fix Summary: RegisterService Panic Resolution

## Problem

`memed start` was panicking with the error:
```
panic: type_url / has not been registered yet. Before calling RegisterService, you must register all interfaces by calling the `RegisterInterfaces` method on module.BasicManager.
```

## Root Causes

### Primary Issue: Missing RegisterInterfaces Call
The app initialization was creating a BasicManager but never calling `RegisterInterfaces()` on it before calling `RegisterServices()` on the Module Manager. In SDK 0.50, this is a strict requirement.

### Secondary Issue: Proto Registry Incompatibility
The wasm module's protobuf-generated files (*.pb.go) were using the old `github.com/gogo/protobuf` package, while the rest of the codebase uses `github.com/cosmos/gogoproto`. These maintain separate proto type registries, causing:
- `proto.MessageName()` to return empty strings for wasm messages
- All wasm messages to register with type URL "/" instead of proper URLs like "/cosmwasm.wasm.v1.MsgStoreCode"
- MsgServiceRouter unable to resolve type URLs during service registration

## Solution

### 1. Added RegisterInterfaces Call (app/app.go)
```go
// Line 636 in app/app.go
app.basicManager.RegisterInterfaces(interfaceRegistry)
```

This ensures all interfaces are registered before `app.mm.RegisterServices(app.configurator)` is called.

### 2. Fixed Proto Import Compatibility (x/wasm/types/*.pb.go)
Updated all protobuf-generated files to use `cosmos/gogoproto` instead of `gogo/protobuf`:

**Files Modified:**
- x/wasm/types/tx.pb.go
- x/wasm/types/query.pb.go  
- x/wasm/types/genesis.pb.go
- x/wasm/types/ibc.pb.go
- x/wasm/types/proposal.pb.go
- x/wasm/types/types.pb.go

**Change Made:**
```go
// OLD (incorrect)
import proto "github.com/gogo/protobuf/proto"
import _ "github.com/gogo/protobuf/gogoproto"

// NEW (correct)
import proto "github.com/cosmos/gogoproto/proto"
import _ "github.com/cosmos/gogoproto/gogoproto"
```

### 3. Added Custom Type URL Registration Fallback (x/wasm/types/codec.go)
Added explicit type URL registration in RegisterInterfaces as a safety measure, though this became unnecessary once the proto imports were fixed.

## Verification

```bash
# Build succeeds
cd /home/runner/work/meme/meme
go build ./cmd/memed

# Node initializes
./memed init test --chain-id test-chain

# Node starts without panic
./memed start --minimum-gas-prices="0stake"
# Output: "9:15AM INF starting node with ABCI CometBFT in-process module=server"
# (No panic - success!)
```

## Technical Details

### Why This Happened
1. **SDK 0.50 Requirement**: The BaseApp MsgServiceRouter checks that all message types are registered before allowing service registration. This is enforced at startup.

2. **Proto Registry Split**: The old `gogo/protobuf` and new `cosmos/gogoproto` packages maintain completely separate type registries. When proto files call `proto.RegisterType()` in init() using one package, but application code calls `proto.MessageName()` using the other package, the lookups fail silently (returning empty strings).

3. **Cascading Failure**: Empty type URLs (/) cause duplicate registration errors and prevent proper service registration, ultimately causing the BaseApp to panic.

### Why The Fix Works
1. **RegisterInterfaces Call**: Ensures the InterfaceRegistry knows about all message types before the MsgServiceRouter tries to validate them.

2. **Unified Proto Package**: Using `cosmos/gogoproto` consistently ensures that:
   - Type registration (in pb.go init functions) and type lookup (in SDK code) use the same registry
   - proto.MessageName() returns proper type URLs like "/cosmwasm.wasm.v1.MsgStoreCode"
   - InterfaceRegistry can successfully resolve type URLs during service registration

## Future Considerations

### Proto File Regeneration
When regenerating proto files, ensure:
- Use cosmos/gogoproto tools, not gogo/protobuf tools
- Verify imports in generated *.pb.go files
- Test that proto.MessageName() returns non-empty strings

### SDK Upgrades
This pattern (RegisterInterfaces before RegisterServices) is now standard in Cosmos SDK 0.50+. All future modules must follow this pattern.

## Files Changed
- app/app.go (added RegisterInterfaces call)
- x/wasm/types/codec.go (added custom type URL registration)
- x/wasm/types/*.pb.go (6 files - updated proto imports)

## Commits
- 72713b2: Initial RegisterInterfaces investigation
- 1ca8f69: Complete fix with proto import updates
