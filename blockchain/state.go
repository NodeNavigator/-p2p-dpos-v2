package blockchain

import (
	"fmt"
	"sync"

	"go.uber.org/zap"
)

// ValidatorInfo holds information about a validator
type ValidatorInfo struct {
	PublicKey       string // Peer ID
	StakedAmount    uint64
	DelegatedAmount uint64 // Sum of delegations from others
	IsActive        bool
	Rank            uint64 // For proposer selection
}

// ChainState manages the mutable state of the blockchain
type ChainState struct {
	Mu                 sync.RWMutex
	Height             uint64
	LastBlockHash      string
	Balances           map[string]uint64
	Nonces             map[string]uint64            // Track tx nonces per account to prevent replay
	Validators         map[string]*ValidatorInfo    // All validators (active + inactive)
	ActiveValidators   []*ValidatorInfo             // Sorted list of active validators
	Delegations        map[string]map[string]uint64 // delegator -> (validator -> amount)
	PendingUndelegates map[string]*UndelegateInfo   // Pending undelegates
	Logger             *zap.Logger
}

// UndelegateInfo tracks undelegations that must be finalized
type UndelegateInfo struct {
	From           string
	Validator      string
	Amount         uint64
	FinalizeHeight uint64
}

// NewChainState creates a new chain state
func NewChainState(logger *zap.Logger) *ChainState {
	return &ChainState{
		Height:             0,
		LastBlockHash:      "",
		Balances:           make(map[string]uint64),
		Nonces:             make(map[string]uint64),
		Validators:         make(map[string]*ValidatorInfo),
		ActiveValidators:   make([]*ValidatorInfo, 0),
		Delegations:        make(map[string]map[string]uint64),
		PendingUndelegates: make(map[string]*UndelegateInfo),
		Logger:             logger,
	}
}

// GetBalance returns an account's balance
func (cs *ChainState) GetBalance(account string) uint64 {
	cs.Mu.RLock()
	defer cs.Mu.RUnlock()
	return cs.Balances[account]
}

// GetNonce returns an account's current nonce
func (cs *ChainState) GetNonce(account string) uint64 {
	cs.Mu.RLock()
	defer cs.Mu.RUnlock()
	return cs.Nonces[account]
}

// IncrementNonce increments nonce for an account
func (cs *ChainState) IncrementNonce(account string) {
	cs.Mu.Lock()
	defer cs.Mu.Unlock()
	cs.Nonces[account]++
}

// Transfer moves tokens from one account to another
func (cs *ChainState) Transfer(from, to string, amount uint64) error {
	cs.Mu.Lock()
	defer cs.Mu.Unlock()

	if cs.Balances[from] < amount {
		return fmt.Errorf("insufficient balance: %d < %d", cs.Balances[from], amount)
	}

	cs.Balances[from] -= amount
	cs.Balances[to] += amount
	cs.Logger.Debug("transfer executed", zap.String("from", from), zap.String("to", to), zap.Uint64("amount", amount))
	return nil
}

// RegisterValidator registers a new validator
func (cs *ChainState) RegisterValidator(publicKey string, stakeAmount uint64) error {
	cs.Mu.Lock()
	defer cs.Mu.Unlock()

	if _, exists := cs.Validators[publicKey]; exists {
		return fmt.Errorf("validator already registered: %s", publicKey)
	}

	if cs.Balances[publicKey] < stakeAmount {
		return fmt.Errorf("insufficient balance for stake: %d < %d", cs.Balances[publicKey], stakeAmount)
	}

	cs.Balances[publicKey] -= stakeAmount

	cs.Validators[publicKey] = &ValidatorInfo{
		PublicKey:       publicKey,
		StakedAmount:    stakeAmount,
		DelegatedAmount: 0,
		IsActive:        false,
		Rank:            0,
	}

	cs.Logger.Info("validator registered", zap.String("publicKey", publicKey), zap.Uint64("stake", stakeAmount))
	return nil
}

// Delegate delegates tokens to a validator
func (cs *ChainState) Delegate(delegator, validator string, amount uint64) error {
	cs.Mu.Lock()
	defer cs.Mu.Unlock()

	if cs.Balances[delegator] < amount {
		return fmt.Errorf("insufficient balance: %d < %d", cs.Balances[delegator], amount)
	}

	v, exists := cs.Validators[validator]
	if !exists {
		return fmt.Errorf("validator not found: %s", validator)
	}

	cs.Balances[delegator] -= amount
	v.DelegatedAmount += amount

	if cs.Delegations[delegator] == nil {
		cs.Delegations[delegator] = make(map[string]uint64)
	}
	cs.Delegations[delegator][validator] += amount

	cs.Logger.Debug("delegation executed", zap.String("delegator", delegator), zap.String("validator", validator), zap.Uint64("amount", amount))
	return nil
}

// Undelegate initiates undelegation (2-block finalization)
func (cs *ChainState) Undelegate(delegator, validator string, amount uint64, finalizeHeight uint64) error {
	cs.Mu.Lock()
	defer cs.Mu.Unlock()

	if cs.Delegations[delegator][validator] < amount {
		return fmt.Errorf("insufficient delegation: %d < %d", cs.Delegations[delegator][validator], amount)
	}

	v, exists := cs.Validators[validator]
	if !exists {
		return fmt.Errorf("validator not found: %s", validator)
	}

	cs.Delegations[delegator][validator] -= amount
	v.DelegatedAmount -= amount

	key := fmt.Sprintf("%s:%s", delegator, validator)
	cs.PendingUndelegates[key] = &UndelegateInfo{
		From:           delegator,
		Validator:      validator,
		Amount:         amount,
		FinalizeHeight: finalizeHeight,
	}

	cs.Logger.Debug("undelegation initiated", zap.String("delegator", delegator), zap.String("validator", validator), zap.Uint64("amount", amount))
	return nil
}

// FinalizePendingUndelegations completes pending undelegations whose time has come
func (cs *ChainState) FinalizePendingUndelegations(currentHeight uint64) {
	cs.Mu.Lock()
	defer cs.Mu.Unlock()

	for key, undeleg := range cs.PendingUndelegates {
		if undeleg.FinalizeHeight <= currentHeight {
			cs.Balances[undeleg.From] += undeleg.Amount
			delete(cs.PendingUndelegates, key)
			cs.Logger.Debug("undelegation finalized", zap.String("delegator", undeleg.From), zap.Uint64("amount", undeleg.Amount))
		}
	}
}

// UpdateActiveValidators updates the set of active validators (called each block)
func (cs *ChainState) UpdateActiveValidators(maxActive int) {
	cs.Mu.Lock()
	defer cs.Mu.Unlock()

	// Sort validators by total stake (staked + delegated) descending
	var validatorList []*ValidatorInfo
	for _, v := range cs.Validators {
		if v.StakedAmount > 0 {
			validatorList = append(validatorList, v)
		}
	}

	// Simple bubble sort by total stake
	for i := 0; i < len(validatorList); i++ {
		for j := i + 1; j < len(validatorList); j++ {
			stake1 := validatorList[i].StakedAmount + validatorList[i].DelegatedAmount
			stake2 := validatorList[j].StakedAmount + validatorList[j].DelegatedAmount
			if stake2 > stake1 {
				validatorList[i], validatorList[j] = validatorList[j], validatorList[i]
			}
		}
	}

	// Deactivate all
	for _, v := range cs.Validators {
		v.IsActive = false
		v.Rank = 0
	}

	// Activate top validators
	cs.ActiveValidators = make([]*ValidatorInfo, 0)
	for i := 0; i < len(validatorList) && i < maxActive; i++ {
		validatorList[i].IsActive = true
		validatorList[i].Rank = uint64(i)
		cs.ActiveValidators = append(cs.ActiveValidators, validatorList[i])
	}

	cs.Logger.Debug("validators updated", zap.Int("activeCount", len(cs.ActiveValidators)))
}

// GetNextProposer returns the next block proposer based on round-robin
func (cs *ChainState) GetNextProposer() (string, error) {
	cs.Mu.RLock()
	defer cs.Mu.RUnlock()

	if len(cs.ActiveValidators) == 0 {
		return "", fmt.Errorf("no active validators")
	}

	// Round-robin: (height + 1) % activeValidators
	idx := (cs.Height + 1) % uint64(len(cs.ActiveValidators))
	return cs.ActiveValidators[idx].PublicKey, nil
}
