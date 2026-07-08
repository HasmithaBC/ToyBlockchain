package core

import (
	"crypto/sha256"
	"fmt"
	"strings"
	"time"
)

// Block represents a single batch of transactions in the blockchain.
type Block struct {
	Index        int64
	Timestamp    int64
	Transactions []Transaction
	PrevHash     string
	Nonce        int64
	Hash         string
}

// NewBlock creates a new block in the chain.
func NewBlock(index int64, transactions []Transaction, prevHash string) *Block {
	return &Block{
		Index:        index,
		Timestamp:    time.Now().Unix(),
		Transactions: transactions,
		PrevHash:     prevHash,
		Nonce:        0,
		Hash:         "", // We will calculate this right after creation!
	}
}

// CalculateHash generates a deterministic SHA-256 hash of the block's contents.
func (b *Block) CalculateHash() string {
	// 1. Serialize the block deterministically (excluding the Hash field itself)
	// Order: Index, Timestamp, Transactions, PrevHash, Nonce
	record := fmt.Sprintf("%d%d%v%s%d", b.Index, b.Timestamp, b.Transactions, b.PrevHash, b.Nonce)
	
	// 2. Hash the serialized string
	h := sha256.New()
	h.Write([]byte(record))
	
	// 3. Return the hash as a hex string
	return fmt.Sprintf("%x", h.Sum(nil))
}

// NewGenesisBlock creates the deterministic first block in the blockchain.
func NewGenesisBlock() *Block {
	txs := []Transaction{
		{Sender: "System", Recipient: "System", Amount: 0},
	}
	
	// A string of 64 zeros represents the PrevHash
	genesisPrevHash := "0000000000000000000000000000000000000000000000000000000000000000"
	
	b := NewBlock(0, txs, genesisPrevHash)
	
	// We MUST hardcode the timestamp so the Genesis block's hash is identical 
	// every single time we start the program!
	b.Timestamp = 1600000000
	b.Hash = b.CalculateHash()
	
	return b
}

// MineBlock performs the proof of work to find a valid hash for the block.
func (b *Block) MineBlock(difficulty int) {
	// Create a string with the required number of leading zeros (e.g., "0000")
	target := strings.Repeat("0", difficulty)
	
	startTime := time.Now()
	
	// Keep hashing until the first 'difficulty' characters match our target
	for {
		b.Hash = b.CalculateHash()
		
		if strings.HasPrefix(b.Hash, target) {
			break // We won the lottery!
		}
		
		// If it didn't match, increment the nonce and try again
		b.Nonce++
	}
	
	elapsed := time.Since(startTime)
	fmt.Printf("Block mined! Nonce: %d, Time: %v, Hash: %s\n", b.Nonce, elapsed, b.Hash)
}
