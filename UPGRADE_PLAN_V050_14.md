# Upgrade Plan: Cosmos SDK v0.50.14 with Custom Patches

## Executive Summary

This document provides a comprehensive upgrade plan for the MeMe Chain repository to align with the target dependency versions specified in the requirements. The repository is currently **ALREADY AT SDK v0.50.14** but requires dependency alignment and potential CometBFT configuration updates.

**Status**: Repository is at SDK v0.50.14-height-mismatch-iavl, but some dependencies need version alignment.

## Current State Analysis

### Repository Information
- **Branch**: copilot/prepare-upgrade-plan-for-cosmos-sdk
- **Current Commit**: c90b5e5 "Initial plan"
- **Base Commit** (per requirements): 5d4db2a "Refactor WASM module... Release/TAG v1.0.0-hop0"
- **Binary**: memed v1.1.0_vN
- **Build Status**: ✅ Successfully builds (app, cmd/memed, full binary)

### Version Comparison Matrix

| Component | Current | Target | Status |
|-----------|---------|--------|--------|
| Go Mod Requirement | 1.23.8 | 1.23.8 | ✅ Match |
| Go Installed | 1.24.12 | 1.23.8 | ⚠️ Newer (compatible) |
| Cosmos SDK | v0.50.14 (cheqd) | v0.50.14-height-mismatch-iavl | ✅ Match |
| CometBFT | v0.38.19 | v0.38.19 | ✅ Match |
| CometBFT Replace | None | /root/cometbft-sec-tachyon | ❌ Missing |
| CosmWasm | v2.2.1 | v2.2.1 | ✅ Match |
| IBC-go | v8.7.0 | v8.7.0 | ✅ Match |
| Store | v1.1.2 (cheqd) | v1.1.2 (cheqd) | ✅ Match |
| IAVL | v1.2.2 (cheqd) | v1.2.2 (cheqd) | ✅ Match |

### Dependency Gaps

#### Minor Version Differences

| Dependency | Current | Target | Impact |
|------------|---------|--------|--------|
| cosmossdk.io/api | v0.9.2 | v0.7.6 | ⚠️ Newer |
| cosmossdk.io/client/v2 | v2.0.0-beta.3 | v2.0.0-beta.5.0.20241121... | ⚠️ Older |
| cloud.google.com/go | v0.120.0 | v0.115.0 | ⚠️ Newer |

#### Missing Dependencies

| Dependency | Status |
|------------|--------|
| cosmossdk.io/tools/confix@v0.1.2 | ❌ Not in go.mod |
| github.com/noble-assets/globalfee@v1.0.1 | ❌ Not in go.mod |
| github.com/cosmos/ibc-apps/modules/async-icq/v8 | ❌ Not in go.mod |

#### Present Custom Forks (Good)

| Dependency | Current Fork |
|------------|--------------|
| osmosis-labs/fee-abstraction | ✅ github.com/cheqd/fee-abstraction/v8@v8.0.3-uneven-heights-formula |
| skip-mev/feemarket | ✅ github.com/cheqd/feemarket@v1.0.5-uneven-heights |

## Critical Issue: CometBFT Local Path Replace

### Problem
The target configuration specifies:
```go
github.com/cometbft/cometbft@v0.38.19 => /root/cometbft-sec-tachyon@(devel)
```

**Current go.mod** does NOT have this replace directive. This could be:
1. **Intentional** - The local path was used during development but removed for CI/distribution
2. **Required** - Security patches in the local fork that need to be applied

### Resolution Options

**Option A: Add Local Replace Directive** (if /root/cometbft-sec-tachyon exists)
```bash
# Check if the path exists
ls -la /root/cometbft-sec-tachyon/

# If exists, add to go.mod:
replace github.com/cometbft/cometbft => /root/cometbft-sec-tachyon
```

**Option B: Use GitHub Fork** (recommended for CI/CD)
```bash
# Create a GitHub fork with security patches
# Then replace in go.mod:
replace github.com/cometbft/cometbft v0.38.19 => github.com/vNodesV/cometbft v0.38.19-sec-tachyon
```

**Option C: Vendor the Changes**
```bash
go mod vendor
# Commit vendor/ directory with local changes
```

## Upgrade Plan

### Phase 1: Dependency Alignment (Low Risk)

**Goal**: Align all dependencies to target versions while maintaining build success.

#### Step 1.1: Update go.mod Versions
```bash
# Update specific versions to match target
go get cosmossdk.io/client/v2@v2.0.0-beta.5.0.20241121152743-3dad36d9a29e
go get cosmossdk.io/api@v0.7.6
go get cloud.google.com/go@v0.115.0
```

#### Step 1.2: Add Missing Optional Dependencies
```bash
# Add if needed by the application
go get cosmossdk.io/tools/confix@v0.1.2
go get github.com/noble-assets/globalfee@v1.0.1
go get github.com/cosmos/ibc-apps/modules/async-icq/v8@v8.0.1-0.20240820212149-6a9e98a8be6e
```

#### Step 1.3: Verify Build
```bash
go mod tidy
make build
make install
./build/memed version --long
```

### Phase 2: CometBFT Configuration (Medium Risk)

**Goal**: Resolve CometBFT local path vs GitHub fork discrepancy.

#### Step 2.1: Investigate Security Patches
```bash
# Check if /root/cometbft-sec-tachyon exists and document changes
diff -r /root/cometbft-sec-tachyon $GOPATH/pkg/mod/github.com/cometbft/cometbft@v0.38.19 > /tmp/cometbft-patches.diff

# Document the purpose of these patches
```

#### Step 2.2: Decision Point
- If patches are security-critical → Create GitHub fork and update replace directive
- If patches are dev-only → Document and keep current configuration
- If no local path exists → Current configuration is correct

### Phase 3: Application Code Review (Low Risk)

**Goal**: Ensure application code is compatible with all dependencies.

#### Step 3.1: Module Integration Check
```bash
# Verify all modules are properly wired
go build ./app
go build ./x/wasm
go build ./cmd/memed
```

#### Step 3.2: Test Suite
```bash
# Run existing tests
go test ./app/...
go test ./x/wasm/...
go test ./cmd/...
```

#### Step 3.3: Integration Tests
```bash
# Start local node
./build/memed init test --chain-id test-1
./build/memed start

# Verify:
# - Node starts without errors
# - Blocks are produced
# - Queries work
# - Transactions work
```

### Phase 4: Documentation and Validation

#### Step 4.1: Update Documentation
- [ ] Update BUILD_TEST_SUMMARY.md with current status
- [ ] Document all dependency decisions
- [ ] Create MIGRATION_NOTES.md for future reference

#### Step 4.2: Final Validation
```bash
# Full build
make clean
make build
make install

# Version check
memed version --long

# Quick smoke test
memed keys add test --keyring-backend test
memed init local --chain-id local-1
```

## Risk Assessment

### Low Risk Items ✅
- Repository already at SDK v0.50.14
- Build system working correctly
- Core dependencies aligned
- Custom cheqd forks in place

### Medium Risk Items ⚠️
- Minor version differences in indirect dependencies
- CometBFT local path configuration uncertainty
- Missing optional dependencies (may not be needed)

### High Risk Items ❌
- None identified

## Dependencies Decision Matrix

| Dependency | Action | Priority | Justification |
|------------|--------|----------|---------------|
| cosmossdk.io/client/v2 | Update to beta.5 | Medium | Target specifies newer version |
| cosmossdk.io/api | Downgrade to v0.7.6 | Low | Current v0.9.2 is compatible |
| cloud.google.com/go | Downgrade to v0.115.0 | Low | Current v0.120.0 is compatible |
| cosmossdk.io/tools/confix | Add if needed | Low | Optional dev tool |
| noble-assets/globalfee | Investigate | Medium | May be needed for fee functionality |
| ibc-apps/async-icq | Investigate | Medium | May be needed for IBC functionality |
| cometbft local replace | Investigate | **High** | Security implications |

## Implementation Checklist

### Pre-Implementation
- [x] Analyze current repository state
- [x] Compare with target versions
- [x] Identify gaps and discrepancies
- [x] Assess build status
- [ ] Investigate CometBFT local path requirement
- [ ] Determine if optional dependencies are needed

### Implementation
- [ ] Phase 1: Update dependencies to exact target versions
- [ ] Phase 2: Resolve CometBFT configuration
- [ ] Phase 3: Verify application code compatibility
- [ ] Phase 4: Update documentation

### Post-Implementation
- [ ] Run full test suite
- [ ] Perform integration testing
- [ ] Update all documentation
- [ ] Create migration guide
- [ ] Tag release version

## Testing Strategy

### Unit Tests
```bash
go test ./... -v -race -coverprofile=coverage.txt
```

### Integration Tests
```bash
# Local single node
./build/memed init test --chain-id test-1
./build/memed start --minimum-gas-prices="0.025umeme"

# Wait for blocks, then:
./build/memed query block 10
./build/memed query bank balances <address>
```

### Upgrade Tests (if applicable)
```bash
# Test upgrade path from v1.0.0 to current
# Create state at v1.0.0
# Upgrade to current version
# Verify state integrity
```

## Build Tags Confirmation

Target build tags: `netgo,ledger,goleveldb`

Current Makefile should use:
```makefile
build_tags = netgo,ledger,goleveldb
```

Verify with:
```bash
grep "build_tags" Makefile
```

## Next Steps

### Immediate Actions (Today)
1. **Investigate CometBFT local path** - Highest priority
   - Check if `/root/cometbft-sec-tachyon` exists
   - Document any patches/changes
   - Decide on fork strategy

2. **Test current build** - Already done ✅
   - Confirmed app builds successfully
   - Confirmed binary installs successfully

3. **Document findings** - In progress
   - This upgrade plan document
   - Dependency comparison complete

### Short Term (This Week)
1. Update dependencies to exact target versions
2. Resolve CometBFT configuration
3. Run full test suite
4. Update all documentation

### Medium Term (Next Sprint)
1. Set up CI/CD pipeline with target versions
2. Create reproducible build process
3. Document upgrade procedures
4. Create rollback procedures

## Conclusion

**The repository is in excellent shape:**
- ✅ Already at Cosmos SDK v0.50.14 with cheqd patches
- ✅ All core dependencies aligned
- ✅ Build system working
- ✅ Binary installs successfully
- ✅ Custom forks properly configured

**Minor actions needed:**
- ⚠️ Investigate and document CometBFT local path requirement
- ⚠️ Optionally update minor version differences
- ⚠️ Add optional dependencies if needed by application

**Estimated effort:** 1-2 days for complete alignment and documentation.

## References

- Target commit: 5d4db2a (v1.0.0-hop0)
- Current commit: c90b5e5
- Cosmos SDK: v0.50.14-height-mismatch-iavl
- Go version: 1.23.8 (required), 1.24.12 (installed)
- Build tags: netgo,ledger,goleveldb

## Appendix: Full Dependency List from Problem Statement

See attached file for complete list of 200+ dependencies with exact versions.

---

**Document Version**: 1.0  
**Date**: 2026-02-10  
**Author**: Copilot Agent (jarvis3.0)  
**Status**: Draft - Awaiting CometBFT investigation
