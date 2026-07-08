package core

// CalculateBalances reads the entire blockchain and returns a map of everyone's balance.
func (bc *Blockchain) CalculateBalances() map[string]int64 {
	// make() is a built-in Go function to initialize a map
	balances := make(map[string]int64)

	// Loop through every block in the chain
	for _, block := range bc.Blocks {
		// Loop through every transaction in the block
		for _, tx := range block.Transactions {
			// A "System" sender means it's newly minted money (like the Genesis block).
			// We don't deduct money from the "System".
			if tx.Sender != "System" {
				balances[tx.Sender] -= tx.Amount
			}
			
			// Add the money to the recipient's balance
			balances[tx.Recipient] += tx.Amount
		}
	}

	return balances
}
