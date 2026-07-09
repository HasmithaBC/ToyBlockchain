package core

import "testing"

func TestNewBlockchain(t *testing.T) {
	bc := NewBlockchain()
	
	if len(bc.Blocks) != 1 {
		t.Errorf("Expected blockchain to have exactly 1 block (the genesis block), got %d", len(bc.Blocks))
	}
	
	if bc.Blocks[0].Index != 0 {
		t.Errorf("Expected first block to be genesis block with index 0")
	}
}

// Scenario: An overspending transaction is rejected (FR-4)
// Given an account whose balance is 100
// When a transaction attempts to send 150 from that account
// Then the transaction is rejected
// And the account balance is unchanged
func TestAddTransaction(t *testing.T) {
	bc := NewBlockchain()
	
	// Give Alice some money to start with (System -> Alice)
	bc.AddTransaction("System", "Alice", 100)
	
	// Since mining hasn't happened, the transaction is just sitting in the pool!
	if len(bc.PendingPool) != 1 {
		t.Errorf("Expected 1 transaction in pending pool, got %d", len(bc.PendingPool))
	}
	
	// Test Validation: Alice tries to send Bob 50. 
	// This should fail because Alice's 100 is still pending and hasn't been mined into a block yet!
	err := bc.AddTransaction("Alice", "Bob", 50)
	if err == nil {
		t.Errorf("Expected an error because Alice's balance is currently 0 in the actual ledger!")
	}
	
	// Test Validation: Negative amount
	err = bc.AddTransaction("System", "Bob", -10)
	if err == nil {
		t.Errorf("Expected an error for negative amount")
	}
}

// Scenarios tested here:
// 1. An honest chain validates successfully (FR-6)
// 2. Tampering with a block is detected (FR-6)
func TestValidateChain(t *testing.T) {
	bc := NewBlockchain()
	difficulty := 2
	
	// Mine a valid block
	tx1 := Transaction{Sender: "Alice", Recipient: "Bob", Amount: 10}
	b1 := NewBlock(1, []Transaction{tx1}, bc.Blocks[0].Hash)
	b1.MineBlock(difficulty)
	bc.AddBlock(b1)
	
	// Scenario: An honest chain validates successfully (FR-6)
	// Given a chain of several mined blocks
	// When the chain is validated
	// Then validation reports the chain as valid
	valid, _ := bc.ValidateChain(difficulty)
	if !valid {
		t.Errorf("Honest chain failed validation!")
	}
	
	// Scenario: Tampering with a block is detected (FR-6)
	// Given a valid chain of several blocks
	// When a transaction inside an earlier block is modified
	// And the chain is validated
	// Then validation fails
	// And it identifies the first block whose hash no longer matches
	bc.Blocks[1].Transactions[0].Amount = 9000
	
	valid, badIndex := bc.ValidateChain(difficulty)
	if valid {
		t.Errorf("Validation failed to catch the tampered block!")
	}
	if badIndex != 1 {
		t.Errorf("Validation should have flagged block 1, but flagged %d", badIndex)
	}
}
