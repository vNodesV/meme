# Quick Start Guide: Building and Running memed with go1.22.10

This guide provides step-by-step instructions for building, installing, and running the MeMe Chain daemon (memed) using go1.22.10 or later.

---

## Prerequisites

### System Requirements
- **Operating System:** Linux, macOS, or WSL2 (Windows)
- **Go Version:** go1.22.10 or later (go1.23.8 recommended)
- **GCC:** Required for ledger support
- **Git:** For cloning the repository
- **Disk Space:** ~500 MB for dependencies and binary

### Installing Go 1.22.10

If you need to install go1.22.10 specifically:

```bash
# Option 1: Using go toolchain manager
go install golang.org/dl/go1.22.10@latest
go1.22.10 download

# Option 2: Direct download
wget https://go.dev/dl/go1.22.10.linux-amd64.tar.gz
sudo tar -C /usr/local -xzf go1.22.10.linux-amd64.tar.gz

# Verify installation
go version  # Should show go1.22.10 or later
```

---

## Quick Installation (5 Minutes)

### 1. Clone the Repository
```bash
git clone https://github.com/vNodesV/meme
cd meme
```

### 2. Download Dependencies
```bash
go mod download
```

### 3. Build and Install
```bash
make install
```

**This will:**
- Build the memed binary with optimizations
- Install to `~/go/bin/memed` (or `$GOPATH/bin/memed`)
- Include build tags: netgo, ledger, goleveldb
- Take approximately 2-3 minutes

### 4. Verify Installation
```bash
# Add ~/go/bin to PATH if not already
export PATH=$PATH:~/go/bin

# Check version
memed version

# View full build info
memed version --long
```

**Expected Output:**
```
name: meme
server_name: memed
version: v1.1.0_vN
commit: <git-commit-hash>
build_tags: netgo,ledger,goleveldb
go: go version go1.22.10 linux/amd64
```

---

## Alternative Build Methods

### Method 1: Build Without Installing
```bash
# Build binary in current directory
go build -o ./memed ./cmd/memed

# Run from current directory
./memed version
```

### Method 2: Build Specific Package
```bash
# Build app package only
go build ./app

# Build main binary only
go build ./cmd/memed
```

### Method 3: Custom Build with go1.22.10
```bash
# If using go toolchain manager
go1.22.10 build -o memed ./cmd/memed

# Install with specific go version
go1.22.10 install ./cmd/memed
```

---

## First Time Setup

### 1. Initialize Node
```bash
# Initialize a new node
memed init <your-moniker> --chain-id meme-1

# Example
memed init mynode --chain-id meme-1
```

**This creates:**
- `~/.memed/config/config.toml` - Node configuration
- `~/.memed/config/app.toml` - Application configuration
- `~/.memed/config/genesis.json` - Genesis file
- `~/.memed/data/` - Data directory

### 2. Create or Import Keys
```bash
# Create a new key
memed keys add <key-name>

# Example
memed keys add mykey

# Import existing key
memed keys add mykey --recover
```

### 3. View Configuration
```bash
# List all keys
memed keys list

# View node configuration
cat ~/.memed/config/config.toml

# View app configuration
cat ~/.memed/config/app.toml
```

---

## Basic Commands

### Version Information
```bash
# Short version
memed version

# Detailed version with dependencies
memed version --long
```

### Help System
```bash
# Main help
memed --help

# Command-specific help
memed query --help
memed tx --help
memed keys --help
memed start --help
```

### Key Management
```bash
# Add new key
memed keys add <name>

# List all keys
memed keys list

# Show key details
memed keys show <name>

# Delete key
memed keys delete <name>

# Export key
memed keys export <name>

# Import key
memed keys import <name> <keyfile>
```

### Node Management
```bash
# Start node
memed start

# Start with specific home directory
memed start --home /path/to/.memed

# Check node status
memed status

# Query blockchain height
memed query block
```

---

## Development Builds

### Build with Debug Symbols
```bash
go build -gcflags="all=-N -l" -o memed ./cmd/memed
```

### Build with Specific Tags
```bash
# Without ledger support
go build -tags "netgo,goleveldb" ./cmd/memed

# With cleveldb instead of goleveldb
WITH_CLEVELDB=yes make install
```

### Build with Custom Linker Flags
```bash
go build -ldflags "\
  -X github.com/cosmos/cosmos-sdk/version.Name=meme \
  -X github.com/cosmos/cosmos-sdk/version.AppName=memed \
  -X github.com/cosmos/cosmos-sdk/version.Version=v1.1.0_vN" \
  ./cmd/memed
```

---

## Troubleshooting

### Issue: "memed: command not found"
**Solution:** Add `~/go/bin` to PATH
```bash
export PATH=$PATH:~/go/bin

# Make permanent (add to ~/.bashrc or ~/.zshrc)
echo 'export PATH=$PATH:~/go/bin' >> ~/.bashrc
source ~/.bashrc
```

### Issue: Build fails with ibctesting errors
**Solution:** This is expected. Build specific packages:
```bash
# Build working packages only
go build ./app
go build ./cmd/memed
```

**Background:** The `x/wasm/ibctesting` package has known SDK migration issues documented in CODE_REVIEW_FINDINGS.md. These don't affect the main binary.

### Issue: "go.mod specifies go 1.23.8"
**Warning (not error):** The project officially requires go1.23.8, but go1.22.10 is compatible.

**Options:**
1. Continue building (will work for most use cases)
2. Upgrade to go1.23.8 for full compatibility
3. Update go.mod to require go1.22.10 (if needed)

### Issue: Ledger support fails
**Solution:** Install build dependencies
```bash
# Ubuntu/Debian
sudo apt-get install gcc

# macOS
xcode-select --install

# Or disable ledger
LEDGER_ENABLED=false make install
```

### Issue: Out of memory during build
**Solution:** Increase build parallelism or available memory
```bash
# Reduce parallel builds
go build -p 1 ./cmd/memed

# Or increase system swap space
```

---

## Testing Your Build

### 1. Verify Version
```bash
memed version --long | grep -E "name|version|go:"
```

**Expected:**
```
name: meme
version: v1.1.0_vN
go: go version go1.22.10 linux/amd64
```

### 2. Test CLI Commands
```bash
# Test help system
memed --help
memed query --help
memed tx --help

# All should show command lists without errors
```

### 3. Test Initialization
```bash
# Initialize test node
memed init testnode --chain-id test-1 --home /tmp/test-memed

# Verify files created
ls -la /tmp/test-memed/config/

# Clean up
rm -rf /tmp/test-memed
```

### 4. Test Key Generation
```bash
# Create test key (use test keyring)
memed keys add testkey --keyring-backend test

# Should show address and mnemonic
```

---

## Performance Benchmarks

### Build Times (Typical Hardware)
- **First Build:** 2-3 minutes (downloads dependencies)
- **Incremental Build:** 10-30 seconds
- **Make Install:** 2-3 minutes
- **Binary Size:** ~148 MB

### Runtime Performance
- **Startup Time:** <1 second
- **Memory Usage:** ~50-100 MB idle
- **CPU Usage:** Minimal when idle

---

## Next Steps

### For Developers
1. Review [CODE_REVIEW_FINDINGS.md](CODE_REVIEW_FINDINGS.md) for known issues
2. Read [APP_MIGRATION_COMPLETE.md](APP_MIGRATION_COMPLETE.md) for SDK 0.50 patterns
3. Check [SDK_050_KEEPER_QUICK_REF.md](SDK_050_KEEPER_QUICK_REF.md) for development guidelines

### For Node Operators
1. Configure your node in `~/.memed/config/config.toml`
2. Set up systemd service for automatic startup
3. Configure firewall for P2P port (26656) and RPC port (26657)
4. Join the appropriate network (mainnet: meme-1, devnet: meme-offline-0)

### For Testing
1. Set up a local testnet
2. Run integration tests (when test issues are resolved)
3. Test smart contract deployment via wasm module

---

## Additional Resources

### Documentation
- [Go Version Compatibility Report](GO_VERSION_COMPATIBILITY_TEST.md)
- [Code Review Findings](CODE_REVIEW_FINDINGS.md)
- [Review Executive Summary](REVIEW_EXECUTIVE_SUMMARY.md)
- [Cosmos SDK 0.50 Upgrade Guide](https://github.com/cosmos/cosmos-sdk/blob/release/v0.50.x/UPGRADING.md)

### Support
- **Repository:** https://github.com/vNodesV/meme
- **Issues:** https://github.com/vNodesV/meme/issues
- **Chain ID (Mainnet):** meme-1
- **Chain ID (Devnet):** meme-offline-0

---

## Summary

This guide covered:
- ✅ Installing go1.22.10
- ✅ Building and installing memed
- ✅ First-time node setup
- ✅ Basic commands and operations
- ✅ Troubleshooting common issues
- ✅ Performance expectations

**Build Status:**
- Core functionality: ✅ Working
- CLI system: ✅ Working
- Known issues: Limited to test infrastructure (documented separately)

**Compatibility:**
- go1.22.10: ✅ Compatible
- go1.23.8: ✅ Recommended (official requirement)
- go1.24+: ✅ Forward compatible

For questions or issues, please consult the documentation or open an issue on GitHub.

---

**Last Updated:** 2026-02-09  
**Version:** 1.0  
**Compatible with:** memed v1.1.0_vN
