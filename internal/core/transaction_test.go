package core

import "testing"

func TestTransactionCreation(t *testing.T) {
	tx := Transaction{
		Sender:    "Alice",
		Recipient: "Bob",
		Amount:    100,
	}

	if tx.Sender != "Alice" {
		t.Errorf("Expected sender Alice, got %s", tx.Sender)
	}
	
	if tx.Amount != 100 {
		t.Errorf("Expected amount 100, got %d", tx.Amount)
	}
}
