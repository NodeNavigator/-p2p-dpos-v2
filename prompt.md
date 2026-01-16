
Design and implement a production-oriented P2P blockchain system in Go that demonstrates Delegated Proof of Stake (DPoS) consensus. The system should be minimal but architected with production-grade extensibility, security, and maintainability in mind.

This project is intended to demonstrate deep understanding of blockchain networking, consensus design, and system architecture. While full production completeness is not required, the system must clearly separate concerns, follow Go module best practices, and document how it would scale to production.

────────────────────────────────────────
Core Requirements
────────────────────────────────────────

Peer-to-Peer Networking
Use libp2p for networking
Secure node identity using cryptographic keys
Peer discovery via bootnodes
Gossip-based message broadcasting
Network message versioning
Delegated Proof of Stake (DPoS)
Support two roles: Delegators and Validators
Delegators can delegate stake to validators
Validator voting power = total delegated stake
Validators ranked by voting power
Top N validators form the active validator set
Deterministic validator set updates per block height
Block Production & Consensus
Active validators participate in consensus
Block proposer selected via round-robin rotation
Proposer creates a block and broadcasts it
Other validators verify:
block structure
proposer legitimacy
state transition validity
Block finality achieved via majority validator acceptance
Reject invalid or unauthorized proposals
Blockchain State Machine
Maintain deterministic state including:
account balances
delegations
validator registry
active validator set
current block height
Apply state transitions sequentially per block
Hash blocks and chain state for integrity
Persistent Storage
Persist blockchain data using LevelDB or equivalent
Store:
blocks
application state
validator metadata
Support node restart without state loss
Transaction Model
Define transactions for:
validator registration
stake delegation
Transactions must be signed and verified
Replay protection via nonces
Command Line Interface (CLI)
Provide a CLI tool to:
start a node (with configurable ports and bootnodes)
generate and manage node keys
register a validator
delegate stake to a validator
query chain state, validator set, and block height
Go Module & Package Structure
Use go.mod and semantic versioning
Organize code using clean, production-grade packages
No monolithic files
Observability & Safety
Structured logging
Clear error handling
Context-aware goroutines
Graceful shutdown
────────────────────────────────────────
Documentation
────────────────────────────────────────

Include a README.md explaining:

Overall system architecture
P2P networking design
DPoS consensus mechanism
Validator lifecycle
Block proposal and validation flow
Security assumptions
Simplifications compared to production blockchains
Future improvements (slashing, governance, fast sync)
────────────────────────────────────────
Non-Goals (Explicit)
────────────────────────────────────────

No slashing implementation
No governance proposals
No smart contract execution
No cross-chain communication
This project is a technical assessment to demonstrate architectural and protocol understanding, not a fully production-hardened blockchain.