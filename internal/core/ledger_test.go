package core

import "testing"

func TestCalculateBalances(t *testing.T) {
	bc := NewBlockchain() // Starts with Genesis block (System -> System : 0)
	
	// 1. Create some fake transactions
	tx1 := Transaction{Sender: "System", Recipient: "Alice", Amount: 100}
	tx2 := Transaction{Sender: "Alice", Recipient: "Bob", Amount: 30}
	
	// 2. Put them in a block and add to chain
	b := NewBlock(1, []Transaction{tx1, tx2}, bc.Blocks[0].Hash)
	bc.AddBlock(b)
	
	// 3. Calculate balances dynamically
	balances := bc.CalculateBalances()
	
	// Alice got 100, sent 30, should have 70
	if balances["Alice"] != 70 {
		t.Errorf("Expected Alice to have 70, got %d", balances["Alice"])
	}
	
	// Bob received 30, should have 30
	if balances["Bob"] != 30 {
		t.Errorf("Expected Bob to have 30, got %d", balances["Bob"])
	}
}
