package core

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"sync"
)

// Blockchain represents the chain of blocks.
type Blockchain struct {
	Blocks      []*Block
	PendingPool []Transaction
	Mutex       sync.RWMutex
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
	bc.Mutex.Lock()
	defer bc.Mutex.Unlock()
	
	// append() is a built-in Go function to add items to a slice
	bc.Blocks = append(bc.Blocks, b)
}

// ProcessIncomingBlock validates a block from a peer and appends it if valid.
func (bc *Blockchain) ProcessIncomingBlock(b *Block, difficulty int) error {
	bc.Mutex.Lock()
	defer bc.Mutex.Unlock()

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
	
	// 5. If valid, append the block directly (without calling AddBlock to avoid deadlock)
	bc.Blocks = append(bc.Blocks, b)
	
	// 6. Clean up the pending pool (remove transactions that are in this new block)
	var updatedPool []Transaction
	for _, pendingTx := range bc.PendingPool {
		found := false
		
		pendingID := pendingTx.Signature
		if pendingID == "" { pendingID = pendingTx.Payload() }
		
		for _, blockTx := range b.Transactions {
			blockID := blockTx.Signature
			if blockID == "" { blockID = blockTx.Payload() }
			
			if pendingID == blockID {
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
func (bc *Blockchain) AddTransaction(tx Transaction) error {
	bc.Mutex.Lock()
	defer bc.Mutex.Unlock()

	// 1. Validation: Signature MUST be valid
	if !tx.Verify() {
		return fmt.Errorf("invalid transaction signature")
	}

	// 2. Validation: Amount must be positive
	if tx.Amount <= 0 {
		return fmt.Errorf("transaction amount must be greater than zero")
	}

	// 3. Validation: Sender must have enough balance (unless it's the System)
	if tx.Sender != "System" {
		balances := bc.calculateBalancesUnsafe()
		
		// Subtract any amounts the sender has ALREADY committed to in the PendingPool
		for _, pendingTx := range bc.PendingPool {
			if pendingTx.Sender == tx.Sender {
				balances[tx.Sender] -= pendingTx.Amount
			}
		}

		if balances[tx.Sender] < tx.Amount {
			return fmt.Errorf("insufficient balance")
		}
	}

	// 4. Add to the pool
	bc.PendingPool = append(bc.PendingPool, tx)
	return nil
}

// ValidateChain checks the integrity of the entire blockchain using a Read-Lock.
func (bc *Blockchain) ValidateChain(difficulty int) (bool, int) {
	bc.Mutex.RLock()
	defer bc.Mutex.RUnlock()

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
	bc.Mutex.Lock()
	defer bc.Mutex.Unlock()

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
		
		pendingID := pendingTx.Signature
		if pendingID == "" { pendingID = pendingTx.Payload() }
		
		for i := forkIndex; i < len(peerBlocks); i++ {
			for _, minedTx := range peerBlocks[i].Transactions {
				minedID := minedTx.Signature
				if minedID == "" { minedID = minedTx.Payload() }
				
				if pendingID == minedID {
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

// SaveToFile saves the current blockchain state to a JSON file safely.
func (bc *Blockchain) SaveToFile(filename string) error {
	bc.Mutex.RLock()
	defer bc.Mutex.RUnlock()
	
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()
	
	// Create a temporary struct that doesn't include the Mutex
	data := struct {
		Blocks      []*Block
		PendingPool []Transaction
	}{
		Blocks:      bc.Blocks,
		PendingPool: bc.PendingPool,
	}
	
	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	return encoder.Encode(data)
}

// LoadFromFile loads the blockchain state from a JSON file safely.
func (bc *Blockchain) LoadFromFile(filename string) error {
	bc.Mutex.Lock()
	defer bc.Mutex.Unlock()
	
	file, err := os.Open(filename)
	if err != nil {
		return err 
	}
	defer file.Close()
	
	var data struct {
		Blocks      []*Block
		PendingPool []Transaction
	}
	
	if err := json.NewDecoder(file).Decode(&data); err != nil {
		return err
	}
	
	bc.Blocks = data.Blocks
	bc.PendingPool = data.PendingPool
	return nil
}
