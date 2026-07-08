package core

import (
	"fmt"
	"strings"
)

// Blockchain represents the chain of blocks.
type Blockchain struct {
	Blocks      []*Block
	PendingPool []Transaction
}

// NewBlockchain initializes a new blockchain with the genesis block.
func NewBlockchain() *Blockchain {
	genesisBlock := NewGenesisBlock()
	
	return &Blockchain{
		Blocks:      []*Block{genesisBlock},
		PendingPool: []Transaction{},
	}
}

// AddBlock safely appends a newly mined block to the end of the chain.
func (bc *Blockchain) AddBlock(b *Block) {
	// append() is a built-in Go function to add items to a slice
	bc.Blocks = append(bc.Blocks, b)
}

// AddTransaction adds a new transaction to the pending pool after validating it.
func (bc *Blockchain) AddTransaction(sender, recipient string, amount int64) error {
	// 1. Validation: Amount must be positive
	if amount <= 0 {
		return fmt.Errorf("transaction amount must be greater than zero")
	}

	// 2. Validation: Sender must have enough balance (unless it's the System)
	if sender != "System" {
		balances := bc.CalculateBalances()
		if balances[sender] < amount {
			return fmt.Errorf("insufficient balance: %s only has %d", sender, balances[sender])
		}
	}

	// 3. Add to the waiting room
	tx := Transaction{
		Sender:    sender,
		Recipient: recipient,
		Amount:    amount,
	}
	bc.PendingPool = append(bc.PendingPool, tx)
	
	return nil
}

// ValidateChain checks the integrity of the entire blockchain.
// Returns true if valid. If false, it returns the Index of the broken block.
func (bc *Blockchain) ValidateChain(difficulty int) (bool, int) {
	target := strings.Repeat("0", difficulty)

	// Start at 1 because block 0 is the Genesis block
	for i := 1; i < len(bc.Blocks); i++ {
		currentBlock := bc.Blocks[i]
		previousBlock := bc.Blocks[i-1]

		// 1. The Recomputation Rule
		if currentBlock.Hash != currentBlock.CalculateHash() {
			return false, int(currentBlock.Index)
		}

		// 2. The Link Rule
		if currentBlock.PrevHash != previousBlock.Hash {
			return false, int(currentBlock.Index)
		}

		// 3. The Proof-of-Work Rule
		if !strings.HasPrefix(currentBlock.Hash, target) {
			return false, int(currentBlock.Index)
		}

		// 4. The Index Rule
		if currentBlock.Index != previousBlock.Index+1 {
			return false, int(currentBlock.Index)
		}

		// 5. The Time Rule
		if currentBlock.Timestamp < previousBlock.Timestamp {
			return false, int(currentBlock.Index)
		}
	}

	return true, -1
}
