# MeMe Chain Custom Agent Directives

This directory contains specialized agent directives for working on the MeMe Chain blockchain project.

## Available Agents

### 1. `jarvis2.agent.md` (PRIMARY - Use This)
**Status**: ✅ CURRENT & ACTIVE  
**Purpose**: SDK 0.50.14 migration expert agent  
**Use For**:
- Cosmos SDK 0.50.x migration work
- Keeper initialization and updates
- CosmWasm wasmvm v2.2.1 integration
- Build fixes and testing
- General development on the current codebase

**Key Features**:
- Complete SDK 0.50 pattern reference
- Keeper initialization templates
- Migration best practices
- Known issues and solutions
- Testing guidelines



## Which Agent Should I Use?

### For Current Development Work → Use `jarvis2.agent.md`
- ✅ Fixing build errors
- ✅ Updating keepers
- ✅ Working with SDK 0.50 patterns
- ✅ CosmWasm integration
- ✅ Test fixes
- ✅ Any code changes

### For Historical Reference Only → See `jarvis2.agent.md`
- 📚 Understanding original upgrade strategy
- 📚 Live chain version history
- 📚 Multi-hop upgrade planning context

## Current Project Status

**Migration Complete**: SDK 0.50.14 with wasmvm v2.2.1
- ✅ app/ package: 100% migrated
- ✅ x/wasm module: Builds successfully
- ✅ All keepers: Updated to SDK 0.50 patterns
- 🔄 External dependencies: Minor compatibility issues remain (wasmd interfaces)

## Quick Start

1. Read `jarvis2.agent.md` for current patterns
2. Check `APP_MIGRATION_COMPLETE.md` for migration status
3. Review `SDK_050_KEEPER_QUICK_REF.md` for quick reference
4. See `KEEPER_MIGRATION_SUMMARY.md` for detailed changes

## Documentation Structure

```
.github/agents/
├── jarvis2.agent.md    ← PRIMARY: Current working agent for SDK 0.50 migration
└── README.md                       ← This file

/
├── APP_MIGRATION_COMPLETE.md      ← Migration completion summary
├── KEEPER_MIGRATION_SUMMARY.md    ← Detailed keeper changes
├── SDK_050_KEEPER_QUICK_REF.md    ← Quick reference guide
└── BUILD_TEST_SUMMARY.md          ← Build/test status
```

## Contributing

When working on this project:
1. Use the `jarvis2.agent.md` directive
2. Follow SDK 0.50 patterns documented there
3. Update documentation when discovering new patterns
4. Test builds after changes
5. Document known issues and solutions

---

**Last Updated**: 2026-02-08  
**Current SDK Version**: 0.50.14  
**Current wasmvm Version**: v2.2.1  
**Project Status**: Active migration, app/ package complete
