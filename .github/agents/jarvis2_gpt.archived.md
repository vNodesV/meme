---
# Fill in the fields below to create a basic custom agent for your repository.
# The Copilot CLI can be used for local testing: https://gh.io/customagents/cli
# To make this agent available, merge this file into the default repository branch.
# For format details, see: https://gh.io/customagents/config

name: jarvisGPT
description: jarvis2.0 from GPT
---

# My Agent

# Copilot Agent Directives — meme chain (Cosmos SDK 0.50.14 + wasmvm v2)

You are working in a Cosmos SDK / wasmd fork repository. Your job is to complete the remaining migration work with minimal risk to chain state compatibility.

## Non-negotiables
- Do NOT introduce “cleanup refactors” unrelated to the task. Keep diffs small and reviewable.
- Preserve state compatibility: avoid changes that alter store keys, module names, protobuf field numbers, or genesis formats unless explicitly required.
- Follow existing patterns in this repo’s migration docs and in upstream wasmd where applicable.
- Always run `gofmt` on touched files and keep imports clean.
- Prefer changes that are easy to revert.

## Target stack constraints (do not change unless the task explicitly requests it)
- cosmos-sdk: v0.50.14 (with repo-specific/cheqd replacements as currently pinned)
- cometbft: v0.38.19 (as pinned)
- wasmvm: v2.2.1
- ibc-go: v8.7.0
- The repo already includes partial wasmvm v2 fixes (AnyMsg, GoAPI changes, Events type, etc.). Do not re-break those.

## Primary objective (Phase 1)
Unblock compilation and binary build by completing the remaining migration blockers:
1) wasmvm v2 migration completion (keeper + adapter)
2) cmd/memed SDK 0.50 wiring fixes (so `make install` works)

Stop after Phase 1 is complete and provide a concise “what changed / how to verify” summary.

## wasmvm v2 migration — required work items
Implement and integrate the remaining breaking changes (see repo docs: `MIGRATION_SUMMARY.md`, `WASMVM_V2_FIXES_APPLIED.md`):

### 1) KVStore iterator adapter
- Implement an adapter in `x/wasm/types/` (or the most appropriate existing package) that bridges Cosmos SDK store iterators to wasmvm v2 iterator interfaces.
- Update all call sites that currently pass SDK iterators into wasmvm v2 APIs.
- Ensure iterators are closed in all paths (defer close) to prevent leaks.

### 2) VM initialization
- Replace `NewVM()` usage with `NewVMWithConfig()` and create/configure the required config struct.
- Keep configuration explicit and minimal; avoid “magic defaults” unless they match upstream wasmd v0.54.x behavior.

### 3) WasmerEngine / interface mismatch
- If the local interface expects methods no longer provided by wasmvm v2, update the interface and introduce wrappers where needed.
- Prefer mirroring upstream `wasmd v0.54.5` patterns for engine creation and lifecycle.

### 4) RequiredFeatures → RequiredCapabilities
- Replace any remaining usage of `RequiredFeatures` with `RequiredCapabilities`.
- Update any related logic, tests, and comments.

### 5) VM call sites and method signatures
- Sweep remaining compilation errors in `x/wasm/keeper/keeper.go` and related files.
- Update signatures exactly to wasmvm v2 expectations; do not change behavior beyond what’s required.

## cmd/memed — required work items
- Fix `cmd/memed/` to compile under SDK 0.50:
  - root command setup
  - genesis/genaccounts command signature updates
  - main.go error handling and initialization wiring
- Goal: `make install` produces a working `memed` binary.

## Verification requirements (must be included in your PR description)
Provide commands and expected outcomes:
- `gofmt -w` (implicitly by formatting)
- `go build ./...` (or repo build command if used)
- `make install` then `memed version`
- `go test ./app/...` and any relevant `x/wasm/...` tests

If tests cannot be run in the environment, explicitly list what should be run by CI and what compilation checks you relied on.

## How to work (process)
- Work in small commits grouped by concern:
  1) Store adapter + iterator call sites
  2) VM initialization
  3) Engine/interface updates
  4) cmd/memed fixes
- Each commit message should explain the “why” in one sentence.
- If you borrow patterns from upstream wasmd v0.54.5, mirror them closely and add a short comment referencing “upstream wasmd v0.54.x pattern”.

## Safety & code quality
- No new dependencies unless absolutely necessary.
- Avoid panics unless there is an existing adapter pattern in `app/keeper_adapters.go` that already uses “panic on impossible error” and it is truly unreachable.
- Add/adjust tests only when they directly validate the migration behavior or prevent regressions.

## Deliverable
At the end, output:
- A list of files changed
- A short summary of remaining known gaps (if any)
- A “How to verify” block with commands
