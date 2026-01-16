# Running Multiple Nodes - Step-by-Step Guide

This guide shows you exactly how to run 2, 3, or 4 nodes and verify they work together.

---

## Quick Overview

| Nodes | Setup Time | Complexity | Best For |
|-------|-----------|-----------|----------|
| 2 nodes | 15 min | Easy | Learning consensus |
| 3 nodes | 20 min | Medium | Testing validator rotation |
| 4 nodes | 30 min | Medium | Stress testing |

---

## Setup 1: Two-Node Network (15 minutes)

Perfect for verifying basic P2P consensus between two nodes.

### Step 1.1: Prepare (1 minute)

```bash
cd /data/projects/projects/2025/BlockMaze/finvasiaGitlab/p2p-dpos-v2

# Ensure binary is built
make clean
make build
```

**Expected output:**
```
=== Cleaning build artifacts ===
✓ Clean complete
=== Building node ===
✓ Build complete: build/node
-rwxr-xr-x 1 user user 36M build/node
```

### Step 1.2: Start Node 1 (Terminal 1)

Open **Terminal 1** and run:

```bash
cd /data/projects/projects/2025/BlockMaze/finvasiaGitlab/p2p-dpos-v2
./build/node interactive --initial-balance 5000 --port 30333
```

**Wait for startup** (~3 seconds). You should see:

```
Your public key (peer ID): 12D3KooWg3xZ7mXxjZ8X... [COPY THIS]
Initial balance: 5000 tokens
Block Height: 0
Listening on: /ip4/127.0.0.1/tcp/30333/p2p/12D3KooWg3xZ7mXxjZ8X...
```

✅ **Action**: Copy the full peer ID (everything after `/p2p/`)

### Step 1.3: Start Node 2 (Terminal 2)

Open **Terminal 2** and run:

```bash
cd /data/projects/projects/2025/BlockMaze/finvasiaGitlab/p2p-dpos-v2
./build/node interactive --initial-balance 5000 --port 30334 \
  --bootstrap /ip4/127.0.0.1/tcp/30333/p2p/12D3KooWg3xZ7mXxjZ8X...
```

Replace `12D3KooWg3xZ7mXxjZ8X...` with the peer ID from Step 1.2.

**Expected output:**
```
Your public key (peer ID): 12D3KooTjK9pQqL2nB4Y...
Initial balance: 5000 tokens
Connecting to bootstrap peer...
✓ Connected to 1 peer
Block Height: 0
```

✅ **Verification**: Both nodes should show they're connected.

### Step 1.4: Register Both as Validators (30 seconds)

**In Terminal 1**, type:
```
> register-validator --stake 2000
✓ Registered as validator with stake: 2000
```

**In Terminal 2**, type:
```
> register-validator --stake 2000
✓ Registered as validator with stake: 2000
```

### Step 1.5: Verify Both See Same Validators (1 minute)

**In Terminal 1**, type:
```
> validators
=== Validators ===
Validator A (12D3KooWg3xZ7...):
  Stake: 2000, Delegated: 0 (Total: 2000)
  Status: active (Rank: 0)

Validator B (12D3KooTjK9pQq...):
  Stake: 2000, Delegated: 0 (Total: 2000)
  Status: active (Rank: 1)
```

**In Terminal 2**, type:
```
> validators
=== Validators ===
Validator A (12D3KooWg3xZ7...):
  Stake: 2000, Delegated: 0 (Total: 2000)
  Status: active (Rank: 0)

Validator B (12D3KooTjK9pQq...):
  Stake: 2000, Delegated: 0 (Total: 2000)
  Status: active (Rank: 1)
```

✅ **Success**: Both nodes see the same 2 validators!

### Step 1.6: Watch Block Production (10 seconds)

**In Terminal 1**, type multiple times (10 seconds apart):
```
> status
Block Height: 1
Active Validators: 2
Connected Peers: 1

> status
Block Height: 2
Active Validators: 2
Connected Peers: 1

> status
Block Height: 3
Active Validators: 2
Connected Peers: 1
```

✅ **Success**: Blocks increasing, consensus working!

### Step 1.7: Test Token Transfer (2 minutes)

**Get Node 2's address** (in Terminal 2):
```
> balance
Your address: 1a2b3c4d5e6f...
Balance: 3000 (5000 - 2000 stake)
```

**In Terminal 1**, send tokens:
```
> transfer --to 1a2b3c4d5e6f... --amount 500
✓ Transfer of 500 tokens accepted
```

**In Terminal 2**, verify:
```
> balance
Your address: 1a2b3c4d5e6f...
Balance: 3500 (3000 + 500 received)
```

✅ **Success**: Transfer worked across nodes!

### Step 1.8: Test Delegation (2 minutes)

**In Terminal 2**, delegate to Node 1:
```
> delegate --validator 12D3KooWg3xZ7mXxjZ8X... --amount 500
✓ Delegation of 500 tokens accepted
```

**In Terminal 1**, verify delegation was received:
```
> validators
=== Validators ===
Validator A (12D3KooWg3xZ7...):
  Stake: 2000, Delegated: 500 (Total: 2500)  ← Increased!
  Status: active (Rank: 0)

Validator B (12D3KooTjK9pQq...):
  Stake: 2000, Delegated: 0 (Total: 2000)
  Status: active (Rank: 1)
```

✅ **Success**: Delegation propagated across network!

### Step 1.9: Cleanup

In either terminal, type:
```
> exit
```

Both nodes will shut down gracefully.

---

## Setup 2: Three-Node Network (20 minutes)

For testing validator rotation across 3 nodes.

### Step 2.1: Start Node 1 (Terminal 1)

```bash
cd /data/projects/projects/2025/BlockMaze/finvasiaGitlab/p2p-dpos-v2
./build/node interactive --initial-balance 5000 --port 30333
```

**Copy peer ID**: `PEER_ID_1`

### Step 2.2: Start Node 2 (Terminal 2)

```bash
./build/node interactive --initial-balance 5000 --port 30334 \
  --bootstrap /ip4/127.0.0.1/tcp/30333/p2p/PEER_ID_1
```

**Copy peer ID**: `PEER_ID_2`

### Step 2.3: Start Node 3 (Terminal 3)

```bash
./build/node interactive --initial-balance 5000 --port 30335 \
  --bootstrap /ip4/127.0.0.1/tcp/30333/p2p/PEER_ID_1
```

**Copy peer ID**: `PEER_ID_3`

### Step 2.4: Register All Three (Terminal 1, 2, 3)

**Terminal 1**:
```
> register-validator --stake 1500
✓ Registered
```

**Terminal 2**:
```
> register-validator --stake 1500
✓ Registered
```

**Terminal 3**:
```
> register-validator --stake 1500
✓ Registered
```

### Step 2.5: Verify All See Same State (Terminal 1)

```
> validators
=== Validators ===
Validator A: Stake: 1500, Status: active (Rank: 0)
Validator B: Stake: 1500, Status: active (Rank: 1)
Validator C: Stake: 1500, Status: active (Rank: 2)

> status
Connected Peers: 2
Active Validators: 3
Block Height: 0
```

✅ **Success**: All 3 validators visible to Node 1!

### Step 2.6: Watch Block Production (20 seconds)

Run in **Terminal 1**, **Terminal 2**, and **Terminal 3** simultaneously:

```bash
watch -n 3 'grep -i "Block Height:" status.txt'
```

Or manually check:

**Terminal 1**:
```
> status
Block Height: 1

[Wait 10 seconds]

> status
Block Height: 2

[Wait 10 seconds]

> status
Block Height: 3
```

Watch which node produces each block (proposer alternates):
- Node A: Block 1 ← Rank 0
- Node B: Block 2 ← Rank 1
- Node C: Block 3 ← Rank 2
- Node A: Block 4 ← Rank 0 (cycle repeats)

✅ **Success**: Round-robin block production working!

### Step 2.7: Test Multi-Transfer (Terminal 1 → 2 → 3)

**Terminal 1 to Terminal 2**:
```
> transfer --to <NODE_2_ADDR> --amount 300
✓ Accepted
```

**Terminal 2 to Terminal 3**:
```
> transfer --to <NODE_3_ADDR> --amount 300
✓ Accepted
```

**Terminal 3 to Terminal 1**:
```
> transfer --to <NODE_1_ADDR> --amount 300
✓ Accepted
```

**Verify in Terminal 1**:
```
> balance <NODE_1_ADDR>
Balance: 4800 (5000 - 1500 stake + 300 received)
```

✅ **Success**: Full mesh token transfers working!

---

## Setup 3: Four-Node Stress Test (30 minutes)

For comprehensive testing with maximum complexity.

### Step 3.1: Build Fresh

```bash
make clean && make build
```

### Step 3.2: Start Nodes (Open 4 terminals)

**Terminal 1 - Node A**:
```bash
./build/node interactive --initial-balance 5000 --port 30333
# Copy: PEER_ID_A
```

**Terminal 2 - Node B**:
```bash
./build/node interactive --initial-balance 5000 --port 30334 \
  --bootstrap /ip4/127.0.0.1/tcp/30333/p2p/PEER_ID_A
# Copy: PEER_ID_B
```

**Terminal 3 - Node C**:
```bash
./build/node interactive --initial-balance 5000 --port 30335 \
  --bootstrap /ip4/127.0.0.1/tcp/30333/p2p/PEER_ID_A
```

**Terminal 4 - Node D**:
```bash
./build/node interactive --initial-balance 5000 --port 30336 \
  --bootstrap /ip4/127.0.0.1/tcp/30333/p2p/PEER_ID_A
```

### Step 3.3: Register All Four

In each terminal, type:
```
> register-validator --stake 1200
```

### Step 3.4: Monitor Dashboard (Main Terminal - Terminal 5)

Create a monitoring script:

```bash
cat > /tmp/monitor.sh << 'EOF'
#!/bin/bash
echo "=== 4-Node Network Status ==="
echo "Terminal 1 (Node A): Type: status"
echo "Terminal 2 (Node B): Type: status"
echo "Terminal 3 (Node C): Type: status"
echo "Terminal 4 (Node D): Type: status"
echo ""
echo "Monitoring metrics:"
echo "  ✓ Block Height (should increment)"
echo "  ✓ Validators count (should be 4)"
echo "  ✓ Connected Peers (should be 3)"
echo "  ✓ No error messages"
EOF
chmod +x /tmp/monitor.sh
/tmp/monitor.sh
```

### Step 3.5: Run Tests Simultaneously

**In Terminal 1**:
```
> register-validator --stake 1200
> wait 30 seconds
> validators
> status
```

**In Terminal 2**:
```
> register-validator --stake 1200
> transfer --to <NODE_3_ADDR> --amount 400
> status
```

**In Terminal 3**:
```
> register-validator --stake 1200
> delegate --validator <NODE_1_ADDR> --amount 600
> status
```

**In Terminal 4**:
```
> register-validator --stake 1200
> wait 30 seconds
> validators
> status
```

### Step 3.6: Verify Results

After 30 seconds, check in **Terminal 1**:

```
> status
=== Node Status ===
Block Height: 3 (should be increasing)
Active Validators: 4 (all registered)
Connected Peers: 3 (fully connected mesh)
No errors

> validators
=== Validators ===
Validator A: Stake: 1200, Status: active
Validator B: Stake: 1200, Status: active
Validator C: Stake: 1200, Delegated: 600, Status: active
Validator D: Stake: 1200, Status: active
```

✅ **Success**: 4-node network fully operational!

---

## Verification Checklist

### For Any Multi-Node Setup

- [ ] **Connectivity**: All nodes show "Connected Peers" = (n-1)
- [ ] **Validators**: All nodes see same validator list
- [ ] **Block Height**: Synchronized within 1 block between nodes
- [ ] **Transfers**: Cross-node transfers succeed
- [ ] **Delegation**: Delegations visible on all nodes
- [ ] **No Crashes**: No panic/error messages in any terminal
- [ ] **Memory**: Stays stable, no rapid growth
- [ ] **Block Time**: ~10 seconds between blocks consistently

### Expected Success Indicators

| Check | 2-Node | 3-Node | 4-Node |
|-------|--------|--------|--------|
| Peer Count | 1 | 2 | 3 |
| Validators | 2 | 3 | 4 |
| Block Production | ✅ | ✅ | ✅ |
| Transfers | ✅ | ✅ | ✅ |
| Delegation | ✅ | ✅ | ✅ |
| Stable Runtime | ✅ | ✅ | ✅ |

---

## Troubleshooting

### "Nodes won't connect"

**Problem**: Node 2+ show "Connected Peers: 0"

**Solution**:
```bash
# Verify Node 1 is actually running
# In Terminal 1, type:
> status
# Should show peer ID and listening address

# Verify bootstrap address format:
# Correct: /ip4/127.0.0.1/tcp/30333/p2p/12D3KooWxxx...
# Wrong:   12D3KooWxxx... (missing address part)
```

### "Port already in use"

**Problem**: `Error: listen EADDRINUSE :::30333`

**Solution**:
```bash
# Use different ports:
Terminal 1: --port 30333
Terminal 2: --port 30334
Terminal 3: --port 30335
Terminal 4: --port 30336
```

### "Blocks not being produced"

**Problem**: Block Height stays at 0

**Solution**:
```
# In each node:
> validators
# Should show at least 1 validator registered

# If empty:
> register-validator --stake 1000

# Wait 10-15 seconds and check:
> status
# Block Height should increase
```

### "Validators not showing on other nodes"

**Problem**: Node 1 registered, but Node 2 doesn't see it

**Solution**:
```
# Give network 5 seconds to propagate
# Then in Node 2:
> validators

# If still not showing, check:
# 1. Node 2 is connected (status → Connected Peers > 0)
# 2. Node 1 has confirmed registration (type status in Node 1)
# 3. Try restarting Node 2
```

### "Memory growing quickly"

**Problem**: Memory increases >50MB during test

**This is normal for 30+ minute runs.** No action needed for MVP testing.

---

## Copy-Paste Ready Commands

### Two-Node Quick Start

**Terminal 1**:
```bash
cd /data/projects/projects/2025/BlockMaze/finvasiaGitlab/p2p-dpos-v2
./build/node interactive --initial-balance 5000 --port 30333
```

**Terminal 2** (replace PEER_ID):
```bash
cd /data/projects/projects/2025/BlockMaze/finvasiaGitlab/p2p-dpos-v2
./build/node interactive --initial-balance 5000 --port 30334 \
  --bootstrap /ip4/127.0.0.1/tcp/30333/p2p/PEER_ID_HERE
```

**Both terminals**:
```
> register-validator --stake 2000
> wait 10 seconds
> validators
> status
```

### Three-Node Quick Start

**Terminal 1**:
```bash
./build/node interactive --initial-balance 5000 --port 30333
```

**Terminal 2** (replace PEER_ID_1):
```bash
./build/node interactive --initial-balance 5000 --port 30334 \
  --bootstrap /ip4/127.0.0.1/tcp/30333/p2p/PEER_ID_1
```

**Terminal 3** (replace PEER_ID_1):
```bash
./build/node interactive --initial-balance 5000 --port 30335 \
  --bootstrap /ip4/127.0.0.1/tcp/30333/p2p/PEER_ID_1
```

**All terminals**:
```
> register-validator --stake 1500
> wait 20 seconds
> validators
> transfer --to <address> --amount 300
> status
```

---

## Performance Monitoring

### What to Watch For

**Block Production**:
```bash
# Every 10 seconds:
> status
# Block Height should increment by 1
```

**Peer Connectivity**:
```bash
# Should be stable:
> status
Connected Peers: 1  (for 2-node, should be 1)
Connected Peers: 2  (for 3-node, should be 2)
Connected Peers: 3  (for 4-node, should be 3)
```

**Validator Rotation** (3+ nodes):
```bash
# Block producers should rotate:
Block 1: Producer A
Block 2: Producer B
Block 3: Producer C
Block 4: Producer A (cycles)
```

---

## Next Steps

After testing multi-node setup:

1. ✅ **Verify consensus works** ← You're here
2. Run stress test with `make test-stress` (full automation)
3. Share results with LEAD_BRIEFING.md
4. Plan production deployment

---

## Questions?

- **Commands not working?** → See MAKEFILE_GUIDE.md
- **Need automation?** → Use `make test-multi` or `make test-stress`
- **For your lead?** → Share LEAD_BRIEFING.md
- **Architecture details?** → See README.md

---

**Status**: ✅ Ready to run multi-node networks!
