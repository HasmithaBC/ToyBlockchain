package core

// CalculateBalances safely reads the entire blockchain using a Read-Lock.
func (bc *Blockchain) CalculateBalances() map[string]int64 {
	bc.Mutex.RLock()
	defer bc.Mutex.RUnlock()
	return bc.calculateBalancesUnsafe()
}

// calculateBalancesUnsafe does the math without locking (used internally by AddTransaction).
func (bc *Blockchain) calculateBalancesUnsafe() map[string]int64 {
	balances := make(map[string]int64)

	for _, block := range bc.Blocks {
		for _, tx := range block.Transactions {
			if tx.Sender != "System" {
				balances[tx.Sender] -= tx.Amount
			}
			balances[tx.Recipient] += tx.Amount
		}
	}

	return balances
}
