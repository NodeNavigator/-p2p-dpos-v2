# Getting Started with P2P-DPoS v2

## Installation & Build

### Step 1: Ensure Dependencies
```bash
# Check Go version (must be 1.22+)
go version

# Navigate to project
cd p2p-dpos-v2

# Download dependencies
go mod download
go mod tidy
```

### Step 2: Build
```bash
# Create build directory
mkdir -p build

# Compile binary
go build -o build/node ./cmd/node

# Verify binary was created
file build/node  # Should show "ELF 64-bit executable"
ls -lh build/node  # Should show ~36MB
```

### Step 3: Verify Build
```bash
# Test binary runs
./build/node --help

# Should see:
# NAME:
#    p2p-dpos - P2P Blockchain with Delegated Proof of Stake
```

## Usage Modes

### Mode 1: Interactive Node (Development/Testing)

Perfect for testing and learning:

```bash
./build/node interactive --initial-balance 5000
```

This starts a node with 5000 initial tokens. Then at the prompt:

```
> status                          # Check node status
> balance                          # Check your balance
> register-validator --stake 1000  # Register as validator
> validators                       # List all validators
> transfer --to <peer-id> --amount 100
> delegate --validator <peer-id> --amount 500
> exit                             # Stop node
```

### Mode 2: Daemon Mode (Production)

Run as a background service:

```bash
./build/node start \
  --datadir ./blockchain \
  --loglevel debug
```

**Flags**:
- `--datadir` - Where to store blockchain data (LevelDB)
- `--loglevel` - Log level: debug, info, warn, error
- `--port` - Network port (default: auto)
- `--bootstrap` - Bootstrap peer addresses (multiaddr format)
- `--initial-balance` - Initial token balance

## Configuration

### Default Settings (in `config/config.go`)
```go
BlockTime         = 10 seconds   // New block every 10 seconds
MaxValidators     = 10           // Up to 10 validators
MinValidatorStake = 100          // Minimum 100 tokens to register
ValidatorSetSize  = 3            // Top 3 validators produce blocks
```

### Custom Configuration

Edit `config/config.go` and rebuild:
```go
func DefaultConfig(dataDir string) *Config {
    return &Config{
        BlockTime:           5 * time.Second,  // Change here
        ValidatorSetSize:    5,                 // Or here
        MinValidatorStake:   50,
        // ... other settings
    }
}
```

## Tutorial: Single Node

### Step 1: Start Node
```bash
./build/node interactive --initial-balance 1000

# Output:
# Your public key (peer ID): 1234567890abcdef...
# Initial balance: 1000 tokens
```

### Step 2: Check Status
```
> status
=== Node Status ===
Peer ID: 1234567890abcdef...
Block Height: 0
Last Block Hash: 
Total Validators: 0
Active Validators: 0
Connected Peers: 0
```

### Step 3: Register as Validator
```
> register-validator --stake 500
Registered as validator with stake: 500

# Now you have 500 - 500 = 0 balance
```

### Step 4: Wait for Blocks
Blocks are produced every 10 seconds. After ~10s:
```
> status
Block Height: 1
Active Validators: 1
```

Your node is now the sole validator and producing blocks!

## Tutorial: Multi-Node (Same Machine)

### Step 1: Start Node 1
```bash
# Terminal 1
./build/node interactive --initial-balance 5000
# Save the peer ID: peerA_id
```

### Step 2: Start Node 2
```bash
# Terminal 2
# Point to node 1 as bootstrap peer (use libp2p multiaddr)
./build/node interactive --initial-balance 5000 \
  --bootstrap /ip4/127.0.0.1/tcp/30333/p2p/peerA_id
```

### Step 3: Register Both as Validators
```
# Terminal 1
> register-validator --stake 2000

# Terminal 2
> register-validator --stake 2000
```

### Step 4: Verify Block Production
Both nodes should start producing blocks in round-robin fashion:

```
# Terminal 1
> status
Block Height: 2
Active Validators: 2

# Terminal 2
> validators
=== Validators ===
Public Key: ...
  Stake: 2000, Delegated: 0 (Total: 2000)
  Status: active (Rank: 0)
```

### Step 5: Test Transfers
```
# Terminal 1 - Send tokens to Node 2
> transfer --to <node2-pubkey> --amount 100
Transferred 100 tokens to <node2-pubkey>

# Terminal 2
> balance
Balance of <node2-pubkey>: 4900  # 5000 + 100 - 2000 (stake)
```

### Step 6: Test Delegation
```
# Terminal 2 - Delegate to Node 1
> delegate --validator <node1-pubkey> --amount 500
Delegated 500 tokens to <node1-pubkey>

# Terminal 1 - Verify delegation was received
> validators
Public Key: <node2-pubkey>
  Stake: 2000, Delegated: 500 (Total: 2500)
  Status: active (Rank: 0)  # Moved up in ranking
```

## Troubleshooting

### "Build output 'node' already exists and is a directory"
The `cmd/node` directory conflicts with build output. Fix:
```bash
go build -o build/node ./cmd/node  # Specify full output path
```

### "missing go.sum entry"
Dependencies are incomplete:
```bash
go mod tidy
go mod download
```

### Port Already in Use
The node auto-selects a free port. If needed, explicitly set:
```go
// In config/config.go
ListenAddrStrings("/ip4/127.0.0.1/tcp/0")  // 0 = auto-select
```

### No Blocks Being Produced
- Check `status` - is the validator count > 0?
- Wait 10+ seconds (one block interval)
- Check logs for errors: `--loglevel debug`

### Peer Connection Issues
Ensure bootstrap addresses are correct format:
```
/ip4/127.0.0.1/tcp/30333/p2p/QmXxxx...
```

## File Structure

```
build/
└── node              ← Compiled binary (run this)

config/
└── config.go         ← Configuration structs

crypto/
├── keys.go           ← Ed25519 key management
└── conversion.go     ← Key format conversion

blockchain/
├── block.go          ← Block and transaction types
├── state.go          ← Chain state (balances, validators)
└── store.go          ← LevelDB persistence

p2p/
├── host.go           ← libp2p host setup
├── discovery.go      ← Peer discovery
├── gossip.go         ← Message broadcasting
└── messages.go       ← Protocol messages

consensus/
└── dpos.go           ← DPoS block production

node/
└── node.go           ← Node orchestration

cli/
└── commands.go       ← CLI commands

cmd/
└── node/main.go      ← Entry point
```

## Next Steps

1. **Explore Code**
   - Read [README.md](README.md) for architecture details
   - Study [blockchain/state.go](blockchain/state.go) for state management
   - Review [consensus/dpos.go](consensus/dpos.go) for consensus logic

2. **Extend Features**
   - Add new transaction types in [blockchain/block.go](blockchain/block.go)
   - Implement new CLI commands in [cli/commands.go](cli/commands.go)
   - Enhance state transitions in [blockchain/state.go](blockchain/state.go)

3. **Optimize Performance**
   - Implement transaction pool limits
   - Add block caching
   - Optimize state root computation

4. **Deploy**
   - Set up proper config files
   - Configure logging to files
   - Use with systemd or Docker

## Performance Characteristics

- **Block Time**: 10 seconds (configurable)
- **Validators**: Up to 10 concurrent
- **Transactions/Block**: Up to 100
- **Throughput**: ~10 tx/sec (depends on config)
- **Binary Size**: ~36MB
- **Memory Usage**: ~100-500MB (varies with blockchain height)
- **Storage**: ~1MB per 100 blocks (LevelDB)

## Security Notes

⚠️ **Development Version**: This is a prototype implementation.

For production use:
- [ ] Implement proper key management (HSM/vault)
- [ ] Add signature verification before tx acceptance
- [ ] Implement transaction mempool limits
- [ ] Add rate limiting and DoS protection
- [ ] Secure peer discovery (authenticated bootstrap)
- [ ] TLS for network communication
- [ ] Database encryption at rest

## Related Resources

- [P2P-DPoS V1](../p2p-dpos/README.md) - Original prototype
- [libp2p Docs](https://docs.libp2p.io/)
- [Go Cryptography](https://pkg.go.dev/golang.org/x/crypto)
- [LevelDB Docs](https://github.com/syndtr/goleveldb)

---

Happy building! 🚀
