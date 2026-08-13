package core

import (
	"crypto/ed25519"
	"crypto/rand"
	"encoding/hex"
	"testing"
)

func TestNewBlockchain(t *testing.T) {
	bc := NewBlockchain()
	if len(bc.Blocks) != 1 {
		t.Errorf("Expected blockchain to have exactly 1 block (the genesis block), got %d", len(bc.Blocks))
	}
	if bc.Blocks[0].Index != 0 {
		t.Errorf("Expected first block to be genesis block with index 0")
	}
}

func TestAddTransaction(t *testing.T) {
	bc := NewBlockchain()
	
	// Give Alice some money (System tx doesn't need signature)
	tx1 := Transaction{Sender: "System", Recipient: "Alice", Amount: 100}
	bc.AddTransaction(tx1)
	
	if len(bc.PendingPool) != 1 {
		t.Errorf("Expected 1 transaction in pending pool, got %d", len(bc.PendingPool))
	}
	
	// Create Alice keys for testing
	pubKey, privKey, _ := ed25519.GenerateKey(rand.Reader)
	privHex := hex.EncodeToString(privKey)
	pubHex := hex.EncodeToString(pubKey)
	
	tx2 := Transaction{Sender: pubHex, Recipient: "Bob", Amount: 50}
	tx2.Sign(privHex)
	
	// This should fail because Alice's 100 is still pending
	err := bc.AddTransaction(tx2)
	if err == nil {
		t.Errorf("Expected an error because Alice's balance is currently 0 in the actual ledger!")
	}
	
	// Negative amount
	tx3 := Transaction{Sender: "System", Recipient: "Bob", Amount: -10}
	err = bc.AddTransaction(tx3)
	if err == nil {
		t.Errorf("Expected an error for negative amount")
	}
}

func TestValidateChain(t *testing.T) {
	bc := NewBlockchain()
	difficulty := 2
	
	tx1 := Transaction{Sender: "System", Recipient: "Bob", Amount: 10}
	b1 := NewBlock(1, []Transaction{tx1}, bc.Blocks[0].Hash)
	b1.MineBlock(difficulty)
	bc.AddBlock(b1)
	
	valid, _ := bc.ValidateChain(difficulty)
	if !valid {
		t.Errorf("Honest chain failed validation!")
	}
	
	bc.Blocks[1].Transactions[0].Amount = 9000
	
	valid, badIndex := bc.ValidateChain(difficulty)
	if valid {
		t.Errorf("Validation failed to catch the tampered block!")
	}
	if badIndex != 1 {
		t.Errorf("Validation should have flagged block 1, but flagged %d", badIndex)
	}
}
