package network

import (
	"sync"
	"testing"
	"toyblockchain/internal/core"
)

// TestNodeConcurrencyRace proves that the node can handle massive concurrent
// reads and writes across the Blockchain, Peers list, and Seen Transactions
// map without triggering a data race when run with `go test -race`.
func TestNodeConcurrencyRace(t *testing.T) {
	bc := core.NewBlockchain()
	node := NewNode("8080", bc)

	var wg sync.WaitGroup

	// 1. Simulate 50 goroutines constantly updating and reading the Peer list
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			node.AddPeer("http://localhost:8081")
			_ = node.GetPeers()
		}()
	}

	// 2. Simulate 50 goroutines blasting the node with transactions
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			// We use a System transaction to bypass cryptographic signature validation 
			// for the sake of a fast concurrency test.
			tx := core.Transaction{Sender: "System", Recipient: "Bob", Amount: int64(10)}
			
			// This safely hits the seenTxs RWMutex
			if !node.MarkTransactionAsSeen(tx) {
				// This safely hits the Blockchain RWMutex
				_ = node.Blockchain.AddTransaction(tx)
			}
		}()
	}

	// 3. Simulate 10 goroutines trying to "Mine" blocks simultaneously
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			node.Blockchain.Mutex.Lock()
			
			poolCopy := make([]core.Transaction, len(node.Blockchain.PendingPool))
			copy(poolCopy, node.Blockchain.PendingPool)
			
			node.Blockchain.Mutex.Unlock()
			
			if len(poolCopy) > 0 {
				_ = core.NewBlock(1, poolCopy, "testhash")
			}
		}()
	}

	// Wait for all 110 concurrent operations to smash into the node
	wg.Wait()
}
