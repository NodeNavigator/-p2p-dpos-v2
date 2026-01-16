package cli

import (
	"fmt"

	"github.com/example/p2p-dpos/node"
	"github.com/urfave/cli/v2"
)

// Commands returns all CLI commands
func Commands(n *node.Node) []*cli.Command {
	return []*cli.Command{
		{
			Name:  "status",
			Usage: "show node status",
			Action: func(c *cli.Context) error {
				return StatusCommand(n)
			},
		},
		{
			Name:  "balance",
			Usage: "show account balance",
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:    "account",
					Aliases: []string{"a"},
					Usage:   "account public key",
				},
			},
			Action: func(c *cli.Context) error {
				return BalanceCommand(n, c.String("account"))
			},
		},
		{
			Name:  "validators",
			Usage: "list all validators",
			Action: func(c *cli.Context) error {
				return ValidatorsCommand(n)
			},
		},
		{
			Name:  "peers",
			Usage: "show connected peers",
			Action: func(c *cli.Context) error {
				return PeersCommand(n)
			},
		},
		{
			Name:  "register-validator",
			Usage: "register as a validator",
			Flags: []cli.Flag{
				&cli.Uint64Flag{
					Name:     "stake",
					Aliases:  []string{"s"},
					Usage:    "amount to stake",
					Required: true,
				},
			},
			Action: func(c *cli.Context) error {
				return RegisterValidatorCommand(n, c.Uint64("stake"))
			},
		},
		{
			Name:  "transfer",
			Usage: "transfer tokens to another account",
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:     "to",
					Usage:    "recipient public key",
					Required: true,
				},
				&cli.Uint64Flag{
					Name:     "amount",
					Aliases:  []string{"a"},
					Usage:    "amount to transfer",
					Required: true,
				},
			},
			Action: func(c *cli.Context) error {
				return TransferCommand(n, c.String("to"), c.Uint64("amount"))
			},
		},
		{
			Name:  "delegate",
			Usage: "delegate tokens to a validator",
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:     "validator",
					Aliases:  []string{"v"},
					Usage:    "validator public key",
					Required: true,
				},
				&cli.Uint64Flag{
					Name:     "amount",
					Aliases:  []string{"a"},
					Usage:    "amount to delegate",
					Required: true,
				},
			},
			Action: func(c *cli.Context) error {
				return DelegateCommand(n, c.String("validator"), c.Uint64("amount"))
			},
		},
	}
}

// StatusCommand shows node status
func StatusCommand(n *node.Node) error {
	state := n.GetState()

	state.Mu.RLock()
	height := state.Height
	hash := state.LastBlockHash
	validators := len(state.Validators)
	activeValidators := len(state.ActiveValidators)
	state.Mu.RUnlock()

	fmt.Println("=== Node Status ===")
	fmt.Printf("Peer ID: %s\n", n.GetKeypair().PublicKeyHex())
	fmt.Printf("Block Height: %d\n", height)
	fmt.Printf("Last Block Hash: %s\n", hash)
	fmt.Printf("Total Validators: %d\n", validators)
	fmt.Printf("Active Validators: %d\n", activeValidators)

	if n.GetHost() != nil {
		fmt.Printf("Connected Peers: %d\n", n.GetHost().PeerCount())
	}

	return nil
}

// BalanceCommand shows an account's balance
func BalanceCommand(n *node.Node, account string) error {
	if account == "" {
		account = n.GetKeypair().PublicKeyHex()
	}

	balance := n.GetState().GetBalance(account)
	fmt.Printf("Balance of %s: %d\n", account, balance)
	return nil
}

// ValidatorsCommand lists all validators
func ValidatorsCommand(n *node.Node) error {
	state := n.GetState()

	state.Mu.RLock()
	defer state.Mu.RUnlock()

	if len(state.Validators) == 0 {
		fmt.Println("No validators registered")
		return nil
	}

	fmt.Println("=== Validators ===")
	for pubKey, info := range state.Validators {
		status := "inactive"
		if info.IsActive {
			status = "active"
		}
		totalStake := info.StakedAmount + info.DelegatedAmount
		fmt.Printf("Public Key: %s\n", pubKey)
		fmt.Printf("  Stake: %d, Delegated: %d (Total: %d)\n", info.StakedAmount, info.DelegatedAmount, totalStake)
		fmt.Printf("  Status: %s (Rank: %d)\n", status, info.Rank)
	}

	return nil
}

// PeersCommand shows connected peers
func PeersCommand(n *node.Node) error {
	if n.GetHost() == nil {
		fmt.Println("Host not initialized")
		return nil
	}

	peers := n.GetHost().GetPeers()
	fmt.Printf("Connected Peers: %d\n", len(peers))
	for _, p := range peers {
		fmt.Printf("  %s\n", p.String())
	}

	return nil
}

// RegisterValidatorCommand registers the node as a validator
func RegisterValidatorCommand(n *node.Node, stake uint64) error {
	state := n.GetState()
	pubKey := n.GetKeypair().PublicKeyHex()

	state.Mu.RLock()
	defer state.Mu.RUnlock()

	if err := state.RegisterValidator(pubKey, stake); err != nil {
		return fmt.Errorf("failed to register validator: %w", err)
	}

	fmt.Printf("Registered as validator with stake: %d\n", stake)
	return nil
}

// TransferCommand transfers tokens
func TransferCommand(n *node.Node, to string, amount uint64) error {
	state := n.GetState()
	from := n.GetKeypair().PublicKeyHex()

	state.Mu.Lock()
	defer state.Mu.Unlock()

	if err := state.Transfer(from, to, amount); err != nil {
		return fmt.Errorf("failed to transfer: %w", err)
	}

	fmt.Printf("Transferred %d tokens to %s\n", amount, to)
	return nil
}

// DelegateCommand delegates tokens to a validator
func DelegateCommand(n *node.Node, validator string, amount uint64) error {
	state := n.GetState()
	delegator := n.GetKeypair().PublicKeyHex()

	state.Mu.Lock()
	defer state.Mu.Unlock()

	if err := state.Delegate(delegator, validator, amount); err != nil {
		return fmt.Errorf("failed to delegate: %w", err)
	}

	fmt.Printf("Delegated %d tokens to %s\n", amount, validator)
	return nil
}
