package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"toyblockchain/internal/core"
	"toyblockchain/internal/network"
)

// Server holds our node and handles HTTP requests.
type Server struct {
	node *network.Node
}

// StartServer launches our HTTP server to listen for requests.
// Think of this as the Waiter waiting for customers to arrive.
func StartServer(port string, n *network.Node) {
	s := &Server{node: n}

	// Register our route handlers
	http.HandleFunc("/blocks", s.handleGetBlocks)
	http.HandleFunc("/transaction", s.handlePostTransaction)
	http.HandleFunc("/block", s.handlePostBlock)
	http.HandleFunc("/peers", s.handleGetPeers)
	http.HandleFunc("/register", s.handlePostRegister)
	http.HandleFunc("/sync", s.handlePostSync)
	http.HandleFunc("/balances", s.handleGetBalances)
	http.HandleFunc("/mempool", s.handleGetMempool)
	http.HandleFunc("/mine", s.handleGetMine)
	http.HandleFunc("/validate", s.handleGetValidate)

	fmt.Printf("Node listening on port %s...\n", port)
	
	// Start the server (this blocks forever)
	err := http.ListenAndServe(":"+port, nil)
	if err != nil {
		fmt.Println("Server failed:", err)
	}
}

// handleGetBlocks returns the entire blockchain as JSON.
func (s *Server) handleGetBlocks(w http.ResponseWriter, r *http.Request) {
	// Set the header so the client knows this is JSON, not an HTML website
	w.Header().Set("Content-Type", "application/json")
	
	// We take the Go array of Blocks, convert it to JSON, and write it to the response!
	json.NewEncoder(w).Encode(s.node.Blockchain.Blocks)
}

// handlePostTransaction accepts a new transaction via JSON and adds it to the pool.
func (s *Server) handlePostTransaction(w http.ResponseWriter, r *http.Request) {
	// 1. Security Check: Ensure the user actually sent a POST request
	if r.Method != http.MethodPost {
		http.Error(w, "Only POST requests are allowed here", http.StatusMethodNotAllowed)
		return
	}

	// 2. Decode: Translate the incoming JSON string into a Go struct
	var tx core.Transaction
	err := json.NewDecoder(r.Body).Decode(&tx)
	if err != nil {
		http.Error(w, "Invalid JSON data provided", http.StatusBadRequest)
		return
	}

	// NEW: Cache Check
	if s.node.MarkTransactionAsSeen(tx) {
		// We already gossiped this. Ignore it to prevent infinite loops!
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, "Already seen this transaction. Ignoring.\n")
		return
	}

	// 3. Process: Hand it to the Kitchen!
	err = s.node.Blockchain.AddTransaction(tx)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// NEW: We added it successfully, now act like a megaphone!
	s.node.BroadcastTransaction(tx)

	// 4. Respond: Tell the user it was successful
	w.WriteHeader(http.StatusCreated)
	fmt.Fprintf(w, "Success! Transaction added to the pending pool and broadcasted to peers.\n")
}

// handlePostBlock accepts a newly mined block from a peer.
func (s *Server) handlePostBlock(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Only POST requests are allowed", http.StatusMethodNotAllowed)
		return
	}

	var b core.Block
	err := json.NewDecoder(r.Body).Decode(&b)
	if err != nil {
		http.Error(w, "Invalid JSON data", http.StatusBadRequest)
		return
	}

	// We pass the block to the Kitchen to validate it! (Hardcoding difficulty 3 for now)
	err = s.node.Blockchain.ProcessIncomingBlock(&b, 3)
	if err != nil {
		// If the block is invalid (e.g. its index is way ahead of us), it might mean
		// we are out of sync and on a shorter fork! Trigger a sync in the background.
		go s.node.SyncChain()

		// HTTP 409 Conflict means the block was rejected
		http.Error(w, err.Error(), http.StatusConflict)
		return
	}

	// NEW: We successfully accepted the block. Now act like a megaphone!
	s.node.BroadcastBlock(b)

	w.WriteHeader(http.StatusCreated)
	fmt.Fprintf(w, "Block successfully validated, added to the chain, and broadcasted!\n")
}

// handleGetPeers returns the list of all known friends in the network.
func (s *Server) handleGetPeers(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Only GET requests are allowed", http.StatusMethodNotAllowed)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(s.node.GetPeers())
}

// RegisterRequest represents the JSON data expected when a new peer says hello.
type RegisterRequest struct {
	PeerURL string `json:"peer_url"`
}

// handlePostRegister adds a new peer's URL to our address book.
func (s *Server) handlePostRegister(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Only POST requests are allowed", http.StatusMethodNotAllowed)
		return
	}

	var req RegisterRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil || req.PeerURL == "" {
		http.Error(w, "Invalid JSON data", http.StatusBadRequest)
		return
	}

	s.node.AddPeer(req.PeerURL)

	// Automatically sync with the network whenever a new peer connects!
	go s.node.SyncChain()

	// Instead of text, we return our entire address book back to the new peer!
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(s.node.GetPeers())
}

// handlePostSync forces the node to ask all peers for their latest blockchain.
func (s *Server) handlePostSync(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Only POST requests are allowed", http.StatusMethodNotAllowed)
		return
	}
	
	s.node.SyncChain()
	
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "Sync process completed! Check the node's terminal for logs.\n")
}

// handleGetBalances returns the calculated balances of all accounts.
func (s *Server) handleGetBalances(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Only GET requests are allowed", http.StatusMethodNotAllowed)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(s.node.Blockchain.CalculateBalances())
}

// handleGetMempool returns all transactions currently waiting in the pool.
func (s *Server) handleGetMempool(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Only GET requests are allowed", http.StatusMethodNotAllowed)
		return
	}
	
	// We need to briefly RLock the blockchain to safely read the pending pool
	s.node.Blockchain.Mutex.RLock()
	defer s.node.Blockchain.Mutex.RUnlock()
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(s.node.Blockchain.PendingPool)
}

// handleGetMine triggers the mining process on this node for all pending transactions
func (s *Server) handleGetMine(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Only GET requests are allowed", http.StatusMethodNotAllowed)
		return
	}

	s.node.Blockchain.Mutex.Lock()

	lastBlock := s.node.Blockchain.Blocks[len(s.node.Blockchain.Blocks)-1]
	
	// Copy pending pool to the new block
	poolCopy := make([]core.Transaction, len(s.node.Blockchain.PendingPool))
	copy(poolCopy, s.node.Blockchain.PendingPool)
	
	rewardAddr := r.URL.Query().Get("reward")
	if rewardAddr != "" {
		rewardTx := core.Transaction{
			Sender:    "System",
			Recipient: rewardAddr,
			Amount:    100, // 100 coins reward for mining!
		}
		poolCopy = append([]core.Transaction{rewardTx}, poolCopy...)
	}
	
	newBlock := core.NewBlock(lastBlock.Index+1, poolCopy, lastBlock.Hash)
	
	s.node.Blockchain.Mutex.Unlock() // Unlock early so mining doesn't freeze the API!

	difficulty := 3 // Hardcode difficulty for now
	fmt.Printf("Mining block %d with %d transactions...\n", newBlock.Index, len(newBlock.Transactions))
	newBlock.MineBlock(difficulty)

	// Validate and add the newly mined block
	err := s.node.Blockchain.ProcessIncomingBlock(newBlock, difficulty)
	if err != nil {
		http.Error(w, "Failed to process mined block: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Broadcast it to the network!
	s.node.BroadcastBlock(*newBlock)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(newBlock)
}

// handleGetValidate checks if the current node's blockchain is valid.
func (s *Server) handleGetValidate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Only GET requests are allowed", http.StatusMethodNotAllowed)
		return
	}

	// We use the same hardcoded difficulty of 3
	valid, badIndex := s.node.Blockchain.ValidateChain(3)

	w.Header().Set("Content-Type", "application/json")
	response := map[string]interface{}{
		"is_valid":  valid,
		"bad_index": badIndex,
	}
	
	json.NewEncoder(w).Encode(response)
}

