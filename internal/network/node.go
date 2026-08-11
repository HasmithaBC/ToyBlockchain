package network

import (
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
}

// NewNode creates and returns a fresh Node.
func NewNode(port string, bc *core.Blockchain) *Node {
	return &Node{
		Port:       port,
		Blockchain: bc,
		Peers:      make([]string, 0), // Start with an empty address book
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
