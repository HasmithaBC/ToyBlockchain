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

// ProcessIncomingBlock validates a block from a peer and appends it if valid.
func (bc *Blockchain) ProcessIncomingBlock(b *Block, difficulty int) error {
	lastBlock := bc.Blocks[len(bc.Blocks)-1]
	
	// 1. Validate Index
	if b.Index != lastBlock.Index+1 {
		return fmt.Errorf("invalid index: expected %d, got %d", lastBlock.Index+1, b.Index)
	}
	
	// 2. Validate PrevHash
	if b.PrevHash != lastBlock.Hash {
		return fmt.Errorf("invalid prevhash")
	}
	
	// 3. Validate Hash recomputation
	if b.Hash != b.CalculateHash() {
		return fmt.Errorf("invalid hash: recomputation failed")
	}
	
	// 4. Validate Proof of Work
	target := strings.Repeat("0", difficulty)
	if !strings.HasPrefix(b.Hash, target) {
		return fmt.Errorf("invalid proof of work")
	}
	
	// 5. If valid, append the block
	bc.AddBlock(b)
	
	// 6. Clean up the pending pool (remove transactions that are in this new block)
	var updatedPool []Transaction
	for _, pendingTx := range bc.PendingPool {
		found := false
		for _, blockTx := range b.Transactions {
			// A simple check to see if the transaction is already in the block
			if pendingTx.Sender == blockTx.Sender && pendingTx.Recipient == blockTx.Recipient && pendingTx.Amount == blockTx.Amount {
				found = true
				break
			}
		}
		if !found {
			updatedPool = append(updatedPool, pendingTx)
		}
	}
	bc.PendingPool = updatedPool

	return nil
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
		
		// Subtract any amounts the sender has ALREADY committed to in the PendingPool
		for _, pendingTx := range bc.PendingPool {
			if pendingTx.Sender == sender {
				balances[sender] -= pendingTx.Amount
			}
		}

		if balances[sender] < amount {
			return fmt.Errorf("insufficient balance: %s only has %d available (including pending transactions)", sender, balances[sender])
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

	// Validate Genesis Block (Block 0)
	if len(bc.Blocks) > 0 {
		genesis := bc.Blocks[0]
		if genesis.Index != 0 {
			return false, 0
		}
		if genesis.Hash != genesis.CalculateHash() {
			return false, 0 // Recomputation failed on Genesis
		}
		expectedGenesisPrevHash := strings.Repeat("0", 64)
		if genesis.PrevHash != expectedGenesisPrevHash {
			return false, 0 // Genesis block has invalid PrevHash
		}
	}

	// Start at 1 for the rest of the chain
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

// ResolveConflict handles a chain reorganisation (Fork Resolution).
// It takes a longer chain from a peer, validates it, and if valid, replaces our chain.
// It safely returns any orphaned transactions to the pending pool.
func (bc *Blockchain) ResolveConflict(peerBlocks []*Block, difficulty int) bool {
	// 1. Is it actually longer?
	if len(peerBlocks) <= len(bc.Blocks) {
		return false
	}

	// 2. Validate the peer's chain
	tempChain := &Blockchain{Blocks: peerBlocks}
	isValid, _ := tempChain.ValidateChain(difficulty)
	if !isValid {
		return false
	}

	// 3. Find where the chains split (the fork point)
	forkIndex := 0
	for i := 0; i < len(bc.Blocks) && i < len(peerBlocks); i++ {
		if bc.Blocks[i].Hash != peerBlocks[i].Hash {
			forkIndex = i
			break
		}
	}
	if forkIndex == 0 {
		forkIndex = len(bc.Blocks) // No fork, they are just ahead of us
	}

	// 4. Gather transactions from OUR orphaned blocks (blocks we are about to throw away)
	for i := forkIndex; i < len(bc.Blocks); i++ {
		for _, tx := range bc.Blocks[i].Transactions {
			bc.PendingPool = append(bc.PendingPool, tx)
		}
	}

	// 5. Adopt the peer's longer chain
	bc.Blocks = peerBlocks

	// 6. Clean the PendingPool (remove txs that are already in the new chain)
	var cleanedPool []Transaction
	for _, pendingTx := range bc.PendingPool {
		isMined := false
		for i := forkIndex; i < len(peerBlocks); i++ {
			for _, minedTx := range peerBlocks[i].Transactions {
				if pendingTx.Sender == minedTx.Sender && pendingTx.Recipient == minedTx.Recipient && pendingTx.Amount == minedTx.Amount {
					isMined = true
					break
				}
			}
			if isMined {
				break
			}
		}
		if !isMined {
			cleanedPool = append(cleanedPool, pendingTx)
		}
	}
	bc.PendingPool = cleanedPool

	return true
}
