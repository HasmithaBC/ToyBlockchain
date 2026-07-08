package core

// Transaction represents a transfer of value from a sender to a recipient.
type Transaction struct {
	Sender    string
	Recipient string
	Amount    int64
}
