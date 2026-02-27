#!/usr/bin/env bash
# Build memeart CosmWasm contract for deployment
# Requires: Rust toolchain with wasm32-unknown-unknown target
# Install: curl --proto '=https' --tlsv1.2 -sSf https://sh.rustup.rs | sh
#          rustup target add wasm32-unknown-unknown
set -euo pipefail
cd "$(dirname "$0")"
echo "Building memeart v$(cargo metadata --no-deps --format-version 1 | python3 -c 'import sys,json; print(json.load(sys.stdin)[\"packages\"][0][\"version\"])')"
cargo build --target wasm32-unknown-unknown --release --lib
echo "Binary: target/wasm32-unknown-unknown/release/memeart.wasm"
ls -lh target/wasm32-unknown-unknown/release/memeart.wasm 2>/dev/null || echo "Build complete (check above for errors)"
