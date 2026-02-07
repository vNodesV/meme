# Hop0 build with modern Go (local-only)

This branch (`hop0-modern-go`) is based on the live-chain commit
`3d3bb097154af6a8eaa83f43e8e47dc91dcdb8b2` (SDK v0.45.1 / Tendermint v0.34.16 /
IBC-Go v2.2.0 / wasmvm v1.0.0-beta10).

It keeps the *dependency stack* identical to the live chain while allowing you to
compile with a modern toolchain.

## Preferred: Go 1.23.2 (no downloads)

```bash
export GOTOOLCHAIN=local
go version
make clean && make build
./build/memed version --long
```

## Fallback: Go 1.22.10

```bash
export GOTOOLCHAIN=local
# ensure your PATH points to go1.22.10
go version
make clean && make build
```

## Offline single-node boot (no outside connections)

```bash
HOME0=$PWD/_offline/hop0
rm -rf "$HOME0"

./build/memed init hop0 --chain-id meme-offline-0 --home "$HOME0"
./build/memed keys add validator --home "$HOME0" --keyring-backend test
ADDR=$(./build/memed keys show validator -a --home "$HOME0" --keyring-backend test)

./build/memed add-genesis-account "$ADDR" 100000000umeme --home "$HOME0"
./build/memed gentx validator 70000000umeme --chain-id meme-offline-0 --home "$HOME0" --keyring-backend test
./build/memed collect-gentxs --home "$HOME0"
```

Disable peer discovery in `$HOME0/config/config.toml`:

- `persistent_peers = ""`
- `seeds = ""`
- `pex = false`

Then start:

```bash
./build/memed start --home "$HOME0"
```

Verify:

```bash
curl -s localhost:26657/status | jq '.result.sync_info.latest_block_height'
./build/memed query wasm params --home "$HOME0" --node tcp://127.0.0.1:26657
```

## If you see errors like `../go/src/internal/runtime/...`

That indicates a broken or mismatched `GOROOT`. Fix by unsetting it:

```bash
unset GOROOT
export GOTOOLCHAIN=local
make clean && make build
```
