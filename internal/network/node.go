package network

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"toyblockchain/internal/core"
)

// Node represents a single server in our decentralized network.
// It holds its own copy of the blockchain and a list of friends (peers).
type Node struct {
	Port       string
	Blockchain *core.Blockchain
	Peers      []string
	
	// We need a lock specifically for the Peers list, because the server
	// might try to add a new peer at the exact same time it is reading the list to broadcast!
	peerMutex  sync.RWMutex

	// Cache to prevent infinite gossip loops
	seenTxs    map[string]bool
	seenMutex  sync.RWMutex
}

// NewNode creates and returns a fresh Node.
func NewNode(port string, bc *core.Blockchain) *Node {
	return &Node{
		Port:       port,
		Blockchain: bc,
		Peers:      make([]string, 0), // Start with an empty address book
		seenTxs:    make(map[string]bool),
	}
}

// AddPeer safely adds a new friend to our address book using a Lock.
func (n *Node) AddPeer(peerURL string) {
	n.peerMutex.Lock()
	defer n.peerMutex.Unlock()
	
	// Simple check to prevent adding duplicates
	for _, p := range n.Peers {
		if p == peerURL {
			return
		}
	}
	n.Peers = append(n.Peers, peerURL)
}

// GetPeers safely returns a copy of the address book using a Read-Lock.
func (n *Node) GetPeers() []string {
	n.peerMutex.RLock()
	defer n.peerMutex.RUnlock()
	
	// We create a copy so the caller can't accidentally modify our internal slice
	peersCopy := make([]string, len(n.Peers))
	copy(peersCopy, n.Peers)
	return peersCopy
}

// MarkTransactionAsSeen returns true if it was ALREADY seen.
func (n *Node) MarkTransactionAsSeen(tx core.Transaction) bool {
	// A quick way to make a unique ID for our simple transaction
	txID := fmt.Sprintf("%s-%s-%d", tx.Sender, tx.Recipient, tx.Amount)

	n.seenMutex.Lock()
	defer n.seenMutex.Unlock()

	if n.seenTxs[txID] {
		return true // We already saw this!
	}
	
	// Mark it as seen
	n.seenTxs[txID] = true
	return false // It's brand new
}

// BroadcastTransaction sends the transaction to all friends.
func (n *Node) BroadcastTransaction(tx core.Transaction) {
	// 1. Convert the transaction to JSON
	jsonData, err := json.Marshal(tx)
	if err != nil {
		return
	}

	// 2. Loop through friends
	for _, peer := range n.GetPeers() {
		// Do this in a "goroutine" (in the background) so we don't slow down the server!
		go func(p string) {
			url := fmt.Sprintf("%s/transaction", p)
			http.Post(url, "application/json", bytes.NewBuffer(jsonData))
		}(peer)
	}
}

// BroadcastBlock sends a newly mined block to all friends.
func (n *Node) BroadcastBlock(b core.Block) {
	jsonData, err := json.Marshal(b)
	if err != nil {
		return
	}

	for _, peer := range n.GetPeers() {
		go func(p string) {
			url := fmt.Sprintf("%s/block", p)
			http.Post(url, "application/json", bytes.NewBuffer(jsonData))
		}(peer)
	}
}

// SyncChain asks all peers for their blockchain. If a longer, valid chain is found, it adopts it.
func (n *Node) SyncChain() {
	for _, peer := range n.GetPeers() {
		// 1. Ask the peer for their notebook
		url := fmt.Sprintf("%s/blocks", peer)
		resp, err := http.Get(url)
		if err != nil {
			continue // Peer is offline, skip them
		}
		
		// 2. Decode their notebook
		var peerBlocks []*core.Block
		err = json.NewDecoder(resp.Body).Decode(&peerBlocks)
		resp.Body.Close()
		if err != nil {
			continue
		}

		// 3. Attempt to resolve conflicts (Hardcoding difficulty 3)
		if n.Blockchain.ResolveConflict(peerBlocks, 3) {
			fmt.Printf("Successfully resolved fork and synced to %d blocks from %s\n", len(peerBlocks), peer)
		}
	}
}
