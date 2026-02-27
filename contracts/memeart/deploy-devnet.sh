#!/usr/bin/env bash
# Deploy memeart to devnet (meme-offline-0)
# Requires: memed binary in PATH, devnet key in keyring
# Usage: MEMED_KEY=mykey ./deploy-devnet.sh
set -euo pipefail
BINARY="${BINARY:-memed}"
KEY="${MEMED_KEY:-validator}"
NODE="https://meme.srvs.vnodesv.net/rpc"
CHAIN_ID="meme-offline-0"
WASM="$(dirname "$0")/target/wasm32-unknown-unknown/release/memeart.wasm"

[ -f "$WASM" ] || { echo "ERROR: Build first with ./build.sh"; exit 1; }

echo "Storing wasm code..."
STORE_TX=$($BINARY tx wasm store "$WASM" \
  --from "$KEY" --node "$NODE" --chain-id "$CHAIN_ID" \
  --gas auto --gas-adjustment 1.4 --fees 5000000umeme \
  -y --output json)
echo "$STORE_TX" | python3 -c "import sys,json; d=json.load(sys.stdin); print('TxHash:', d.get('txhash','?'))"
