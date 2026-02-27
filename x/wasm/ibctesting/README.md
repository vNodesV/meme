# testing package for ibc
Customized version of cosmos-sdk x/ibc/testing

## Build Tag

This package is gated behind the `ibctesting` build tag because it has
41 compile errors remaining from the SDK 0.50 / ABCI 2.0 / IBC v8 migration.
The production binary (`memed`) is unaffected.

To compile this package (e.g., when working on ibctesting fixes):

```bash
go build -tags ibctesting ./x/wasm/ibctesting/
go test -tags ibctesting ./x/wasm/...
```

### Remaining migration work (Phase 2)

| Category | Count | Example |
|----------|-------|---------|
| ABCI 2.0 (BeginBlock → FinalizeBlock) | 3 | coordinator.go, chain.go |
| Query signature (context.Context, *Request) | 5 | chain.go, wasm.go |
| IBC v8 API (SendPacket, handshake msgs) | 4 | endpoint.go |
| sdk.Int → math.Int | 3 | chain.go |
| Missing helpers (SetupWithGenesisValSet, SignAndDeliver) | 2 | chain.go |
| CometBFT v0.38 (MakeCommit removed) | 1 | chain.go |
| Events type mismatch (sdk.Events vs []abci.Event) | 5 | endpoint.go |
| Other (Validators struct, GetHistoricalInfo, exported.Header) | 6 | chain.go, endpoint.go |