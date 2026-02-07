#!/bin/bash
set -o errexit -o nounset -o pipefail

# Local/offline single-node setup for memed.
# Usage:
#   DENOM=umeme CHAIN_ID=meme-offline-0 HOME_DIR=$HOME/.memed bash contrib/local/setup_memed.sh

PASSWORD=${PASSWORD:-1234567890}
DENOM=${DENOM:-umeme}
CHAIN_ID=${CHAIN_ID:-meme-offline-0}
MONIKER=${MONIKER:-memeUpgrade}
HOME_DIR=${HOME_DIR:-$HOME/.memed}
KEYRING="--keyring-backend test --home ${HOME_DIR}"

mkdir -p "${HOME_DIR}"

patch_genesis_denom() {
  GENESIS_FILE="$1"
  DEN="$2"
  if command -v jq >/dev/null 2>&1; then
    tmp="${GENESIS_FILE}.tmp"
    jq --arg d "$DEN" '
      (.app_state.staking.params.bond_denom) = $d
      | (.app_state.mint.params.mint_denom) = $d
      | (.app_state.crisis.constant_fee.denom) = $d
      | (.app_state.gov.deposit_params.min_deposit[]?.denom) = $d
    ' "$GENESIS_FILE" > "$tmp" && mv "$tmp" "$GENESIS_FILE"
  else
    sed -i "s/\"bond_denom\": \"stake\"/\"bond_denom\": \"${DEN}\"/g" "$GENESIS_FILE" 2>/dev/null || true
    sed -i "s/\"mint_denom\": \"stake\"/\"mint_denom\": \"${DEN}\"/g" "$GENESIS_FILE" 2>/dev/null || true
    sed -i "s/\"denom\": \"stake\"/\"denom\": \"${DEN}\"/g" "$GENESIS_FILE" 2>/dev/null || true
  fi
}

memed init --chain-id "$CHAIN_ID" "$MONIKER" --home "$HOME_DIR"

GEN="$HOME_DIR/config/genesis.json"
patch_genesis_denom "$GEN" "$DENOM"

if ! memed keys show validator $KEYRING >/dev/null 2>&1; then
  (echo "$PASSWORD"; echo "$PASSWORD") | memed keys add validator $KEYRING
fi

echo "$PASSWORD" | memed add-genesis-account validator "100000000${DENOM}" $KEYRING
(echo "$PASSWORD"; echo "$PASSWORD"; echo "$PASSWORD") | memed gentx validator "70000000${DENOM}" --chain-id "$CHAIN_ID" --amount "70000000${DENOM}" $KEYRING
memed collect-gentxs --home "$HOME_DIR"
