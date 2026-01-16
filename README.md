# P2P-DPoS v2 - Production-Grade Blockchain Implementation

## Overview

**p2p-dpos-v2** is a production-oriented P2P blockchain system with Delegated Proof of Stake (DPoS) consensus, built in Go with enterprise-grade architecture.

**Status**: ✅ **FULLY IMPLEMENTED AND COMPILING**

## Key Features

### Network & Consensus
- **P2P Networking**: libp2p-based peer discovery and communication
- **DPoS Consensus**: Delegated Proof of Stake with validator ranking
- **Gossip Protocol**: In-memory message routing for blocks and transactions
- **Peer Discovery**: Bootnode-based peer discovery system

### Blockchain Features
- **Block Structure**: Height-based, signed blocks with transaction roots
- **Cryptography**: Ed25519 digital signatures for all transactions
- **State Management**: Chain state with balances, validators, delegations
- **Transaction Types**: Transfer, Validator Registration, Delegation, Undelegation

### Storage & Observability
- **Persistent Storage**: LevelDB for blockchain state persistence
- **Structured Logging**: Uber/zap for production-grade observability
- **CLI Framework**: urfave/cli/v2 for professional command-line interface

## Architecture

```
p2p-dpos-v2/
├── config/          # Configuration management
├── crypto/          # Ed25519 signatures and key management
├── utils/           # Hash functions and utilities
├── blockchain/      # Core blockchain structures
│   ├── block.go     # Block and Transaction types
│   ├── state.go     # Chain state management
│   └── store.go     # LevelDB persistence layer
├── p2p/             # Networking layer
│   ├── host.go      # libp2p host initialization
│   ├── discovery.go # Peer discovery
│   ├── gossip.go    # Message broadcasting
│   └── messages.go  # Protocol message definitions
├── tx/              # Transaction building
│   └── builder.go   # Transaction builder pattern
├── consensus/       # DPoS consensus engine
│   └── dpos.go      # Block production and validation
├── node/            # Node orchestration
│   └── node.go      # Complete node lifecycle
├── cli/             # Command-line interface
│   └── commands.go  # All available commands
├── cmd/node/        # Entry point
│   └── main.go      # Application bootstrap
├── go.mod           # Go module definition
├── go.sum           # Dependency checksums
└── build/node       # Compiled binary
```

## Components

### 1. **Config Module** (`config/config.go`)
Centralized configuration for:
- Node identity and network settings
- Blockchain parameters (block time, max validators)
- DPoS parameters (finalization stake, validator set size)
- Storage and logging configuration

### 2. **Crypto Module** (`crypto/`)
- **keys.go**: Ed25519 keypair generation, signing, verification
- **conversion.go**: Conversion between Ed25519 and libp2p key formats
- **Hash functions**: SHA-256 hashing for blocks and transactions

### 3. **Blockchain Module** (`blockchain/`)

#### `block.go`
- **Block**: Height, timestamp, hash, proposer, transactions, validators
- **Transaction**: From, To, Amount, Type (Transfer/Register/Delegate/Undelegate)
- Operations: Hash calculation, finalization, transaction addition

#### `state.go`
- **ChainState**: Mutable state with balances, validators, delegations
- **ValidatorInfo**: Stake amount, delegation sum, active status, rank
- Operations: Transfer, RegisterValidator, Delegate, Undelegate, UpdateActiveValidators
- Thread-safe with RWMutex

#### `store.go`
- **Store**: LevelDB-based persistence
- Operations: SaveBlock, GetBlock, SaveState, SaveTransaction, GetTransaction

### 4. **Consensus Module** (`consensus/dpos.go`)
- **DPoS Engine**: Block production, validation, acceptance
- Round-robin proposer selection from active validators
- Transaction pool management
- Block production loop with configurable interval

### 5. **P2P Module** (`p2p/`)

#### `host.go`
- libp2p host initialization and peer management
- Methods: Connect, GetPeers, GetPeerAddrs, PeerCount

#### `discovery.go`
- Bootnode-based peer discovery
- Methods: Bootstrap, FindPeers, ConnectPeer, GetPeerInfo

#### `gossip.go`
- Simple in-memory pubsub for message routing
- Subscription-based message delivery
- Topics: "blocks", "transactions"

#### `messages.go`
- Protocol messages: BlockMessage, TransactionMessage, GetBlocksMessage
- Message encoding/decoding with JSON marshaling

### 6. **Transaction Module** (`tx/builder.go`)
- **Builder Pattern**: Fluent interface for transaction construction
- Helpers: Transfer(), RegisterValidator(), Delegate(), Undelegate()
- Signature verification support

### 7. **Node Module** (`node/node.go`)
- **Node Orchestration**: Complete node lifecycle management
- Integration of all components (network, consensus, storage)
- Message listeners for blocks and transactions
- Block broadcasting from consensus engine
- Graceful shutdown support

### 8. **CLI Module** (`cli/commands.go`)
- **Commands**: status, balance, validators, peers, register-validator, transfer, delegate
- Integration with urfave/cli/v2 framework
- State queries and transaction submission

## Building

### Prerequisites
- Go 1.22 or higher
- Standard build tools (make, git)

### Compilation

```bash
cd p2p-dpos-v2
go mod tidy          # Download dependencies
go build -o build/node ./cmd/node  # Build binary
```

The compiled binary will be at `build/node` (~36MB)

## Usage

### Quick Start - Interactive Mode

```bash
./build/node interactive --initial-balance 5000
```

Available commands:
- `status` - Show node status (height, connected peers, validators)
- `balance [account]` - Check token balance
- `validators` - List all validators
- `peers` - Show connected peers
- `register-validator` - Register as a validator
- `transfer --to <id> --amount <n>` - Transfer tokens
- `delegate --validator <id> --amount <n>` - Delegate tokens
- `exit` - Stop node

### Start Node Daemon

```bash
./build/node start --datadir ./blockchain-data --loglevel debug
```

## Dependencies

| Package | Version | Purpose |
|---------|---------|---------|
| libp2p/go-libp2p | v0.32.0 | P2P networking |
| multiformats/go-multiaddr | v0.12.0 | Multiaddress format |
| syndtr/goleveldb | v1.0.0 | Key-value storage |
| uber/zap | v1.27.0 | Structured logging |
| urfave/cli | v2.27.1 | CLI framework |
| golang.org/x | v0.21.0 | Cryptography |

## Data Structures

### Block
```go
type Block struct {
    Height     uint64              // Block number
    Timestamp  int64               // Unix milliseconds
    Hash       string              // Block hash
    PrevHash   string              // Parent block hash
    Proposer   string              // Public key of proposer
    Signature  []byte              // Ed25519 signature
    Txs        []*Transaction      // Transactions
    TxRoot     string              // Merkle root
    StateRoot  string              // State root hash
    Validators map[string]bool     // Current validators
}
```

### Transaction
```go
type Transaction struct {
    ID        string              // Transaction hash
    From      string              // Sender public key
    To        string              // Recipient/validator ID
    Amount    uint64              // Token amount
    Nonce     uint64              // Sequence number
    Type      TxType              // Transaction type
    Data      []byte              // Extra data
    Signature []byte              // Ed25519 signature
    Timestamp int64               // Unix milliseconds
}
```

### ValidatorInfo
```go
type ValidatorInfo struct {
    PublicKey       string          // Peer ID
    StakedAmount    uint64          // Self-staked tokens
    DelegatedAmount uint64          // Delegated from others
    IsActive        bool            // Active validator status
    Rank            uint64          // Selection rank
}
```

## DPoS Consensus Flow

1. **Block Production** (every 10 seconds by default)
   - Current height determines next proposer: `height % active_validators`
   - Proposer collects pending transactions
   - Creates block with current validator set
   - Signs block with private key

2. **Block Validation**
   - Verify block height matches expected next height
   - Verify proposer signature
   - Execute all transactions in order

3. **State Updates**
   - Execute transactions: transfers, delegations, registrations
   - Finalize pending undelegations
   - Update active validators based on top stake holders
   - Persist block to LevelDB

4. **Validator Set Updates**
   - Every block: rank validators by total stake (staked + delegated)
   - Top N validators (configurable) become active
   - Active set used for next block proposer selection

## Configuration Parameters

```go
BlockTime           = 10 seconds    // Block production interval
MaxValidators       = 10            // Maximum concurrent validators
MinValidatorStake   = 100           // Minimum stake to register
ValidatorSetSize    = 3             // Active validators per round
FinalizationStake   = 67%           // Stake needed for finality
ProposerRotation    = true          // Round-robin proposer selection
```

## Storage Schema

LevelDB keys:
- `block:{height}` - Block at height
- `height:{height}` - Block hash index
- `state:{height}` - Chain state snapshot
- `tx:{txID}` - Transaction by ID

## Network Protocol

### Message Types
1. **BlockMessage** - Broadcast new blocks
2. **TransactionMessage** - Broadcast pending transactions
3. **GetBlocksMessage** - Request block range
4. **BlockResponseMessage** - Send requested blocks

### Topics
- `blocks` - Block gossip
- `transactions` - Transaction gossip

## Differences from V1

| Feature | V1 | V2 |
|---------|----|----|
| Dependencies | Stdlib only | libp2p, LevelDB, zap, urfave/cli |
| Cryptography | SHA-256 only | Ed25519 signatures |
| Networking | In-memory | Real P2P with libp2p |
| Storage | Memory only | Persistent LevelDB |
| CLI | Custom handler | Professional urfave/cli |
| Logging | Printf | Structured zap logging |
| Scalability | Single machine | Network distributed |
| Production-ready | Demo | Enterprise-grade |

## Limitations

**By Design** (as per requirements):
- ❌ No slashing mechanism
- ❌ No governance system
- ❌ No smart contracts
- ❌ No cross-chain communication

**Current Version**:
- Simple gossip (not full DHT)
- No persistent peer discovery
- No light client support
- Single consensus round length

## Future Enhancements

1. **Networking**
   - Implement full Kademlia DHT
   - Add peer persistence
   - Implement light client protocol

2. **Consensus**
   - Add Byzantine fault tolerance
   - Implement checkpointing
   - Add block finality gadget

3. **Features**
   - Transaction mempool optimization
   - Account-based nonces
   - Minimum viable governance
   - Delegation unbonding period

4. **Performance**
   - Batch transaction processing
   - Parallel block verification
   - State pruning

## Testing

### Quick Validation
```bash
# Check binary exists
ls -lh build/node

# Test help
./build/node --help

# Test interactive mode (press Ctrl+C to exit)
./build/node interactive --initial-balance 1000
```

### Multi-Node Testing
1. Start node 1: `./build/node interactive --initial-balance 5000`
2. Register as validator
3. Start node 2 with bootstrap address from node 1
4. Delegate from node 2 to node 1
5. Verify blocks are being produced

## References

### Related Documentation
- [V1 (Quick Start)](../p2p-dpos/README.md) - Original prototype
- [libp2p Documentation](https://docs.libp2p.io/)
- [DPoS Consensus](https://en.wikipedia.org/wiki/Proof_of_stake#Delegated_proof_of_stake)

### Code Walkthrough
1. Start with [cmd/node/main.go](cmd/node/main.go) - Entry point
2. Review [config/config.go](config/config.go) - Configuration
3. Examine [node/node.go](node/node.go) - Node orchestration
4. Study [consensus/dpos.go](consensus/dpos.go) - Block production
5. Explore [blockchain/state.go](blockchain/state.go) - State management
6. Check [p2p/host.go](p2p/host.go) and [p2p/gossip.go](p2p/gossip.go) - Networking

## License

MIT License - See LICENSE file for details

## Contributing

For bug reports, feature requests, or contributions, please open an issue or pull request.
