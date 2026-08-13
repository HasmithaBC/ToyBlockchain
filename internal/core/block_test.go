package core

import "testing"

func TestNewBlock(t *testing.T) {
	txs := []Transaction{
		{Sender: "System", Recipient: "Bob", Amount: 50},
	}
	
	b := NewBlock(1, txs, "prevhash_123")

	if b.Index != 1 {
		t.Errorf("Expected index 1, got %d", b.Index)
	}
	if b.PrevHash != "prevhash_123" {
		t.Errorf("Expected PrevHash prevhash_123, got %s", b.PrevHash)
	}
	if len(b.Transactions) != 1 {
		t.Errorf("Expected 1 transaction, got %d", len(b.Transactions))
	}
	if b.Transactions[0].Amount != 50 {
		t.Errorf("Expected transaction amount 50, got %d", b.Transactions[0].Amount)
	}
}

func TestCalculateHash_IsDeterministic(t *testing.T) {
	txs := []Transaction{
		{Sender: "System", Recipient: "Bob", Amount: 10},
	}
	b := NewBlock(1, txs, "prevhash_abc")
	
	b.Timestamp = 1600000000

	hash1 := b.CalculateHash()
	hash2 := b.CalculateHash()

	if hash1 != hash2 {
		t.Errorf("Hashing is not deterministic!\nHash1: %s\nHash2: %s", hash1, hash2)
	}
	if hash1 == "" {
		t.Errorf("Hash should not be empty")
	}
}

func TestNewGenesisBlock(t *testing.T) {
	b := NewGenesisBlock()
	
	if b.Index != 0 {
		t.Errorf("Genesis block must have index 0")
	}
	
	expectedPrevHash := "0000000000000000000000000000000000000000000000000000000000000000"
	if b.PrevHash != expectedPrevHash {
		t.Errorf("Genesis block must have 64 zeros as PrevHash")
	}
	if b.Hash == "" {
		t.Errorf("Genesis block must have a pre-calculated hash")
	}
}

func TestMineBlock(t *testing.T) {
	txs := []Transaction{
		{Sender: "System", Recipient: "Bob", Amount: 50},
	}
	b := NewBlock(1, txs, "prevhash_xyz")
	
	difficulty := 2 
	b.MineBlock(difficulty)
	
	expectedPrefix := "00"
	if b.Hash[:difficulty] != expectedPrefix {
		t.Errorf("Expected hash to start with %s, got %s", expectedPrefix, b.Hash)
	}
	if b.Hash != b.CalculateHash() {
		t.Errorf("Winning hash does not match a recalculation! Something is wrong.")
	}
}
