#!/bin/bash
set -o errexit -o nounset -o pipefail

PASSWORD=${PASSWORD:-1234567890}
STAKE=${STAKE_TOKEN:-umeme}
FEE=${FEE_TOKEN:-umeme}
CHAIN_ID=${CHAIN_ID:-testing}
MONIKER=${MONIKER:-node001}

patch_genesis_denom() {
  # Patches denom fields in genesis to $1 (e.g. umeme) in a targeted way.
  # Uses jq when available; falls back to sed for common patterns.
  GENESIS_FILE="$1"
  DENOM="$2"

  if command -v jq >/dev/null 2>&1; then
    tmp="${GENESIS_FILE}.tmp"
    jq --arg d "$DENOM" '
      (.app_state.staking.params.bond_denom) = $d
      | (.app_state.mint.params.mint_denom) = $d
      | (.app_state.crisis.constant_fee.denom) = $d
      | (.app_state.gov.deposit_params.min_deposit[]?.denom) = $d
    ' "$GENESIS_FILE" > "$tmp" && mv "$tmp" "$GENESIS_FILE"
  else
    sed -i "s/\"bond_denom\": \"stake\"/\"bond_denom\": \"${DENOM}\"/g" "$GENESIS_FILE" 2>/dev/null || true
    sed -i "s/\"mint_denom\": \"stake\"/\"mint_denom\": \"${DENOM}\"/g" "$GENESIS_FILE" 2>/dev/null || true
    sed -i "s/\"denom\": \"stake\"/\"denom\": \"${DENOM}\"/g" "$GENESIS_FILE" 2>/dev/null || true
  fi
}

memed init --chain-id "$CHAIN_ID" "$MONIKER"
# staking/governance token is hardcoded in config, change this
## OSX requires: -i.
sed -i. "s/\"stake\"/\"$STAKE\"/" "$HOME"/.memed/config/genesis.json
if ! memed keys show validator; then
  (
    echo "$PASSWORD"
    echo "$PASSWORD"
  ) | memed keys add validator
fi
# hardcode the validator account for this instance
echo "$PASSWORD" | memed add-genesis-account validator "1000000000$STAKE,1000000000$FEE"
# (optionally) add a few more genesis accounts
for addr in "$@"; do
  echo "$addr"
  memed add-genesis-account "$addr" "1000000000$STAKE,1000000000$FEE"
done
# submit a genesis validator tx
## Workraround for https://github.com/cosmos/cosmos-sdk/issues/8251
(
  echo "$PASSWORD"
  echo "$PASSWORD"
  echo "$PASSWORD"
) | memed gentx validator "250000000$STAKE" --chain-id="$CHAIN_ID" --amount="250000000$STAKE"
## should be:
# (echo "$PASSWORD"; echo "$PASSWORD"; echo "$PASSWORD") | memed gentx validator "250000000$STAKE" --chain-id="$CHAIN_ID"
memed collect-gentxs
