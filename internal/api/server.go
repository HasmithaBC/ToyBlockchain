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

	// 3. Process: Hand it to the Kitchen!
	err = s.node.Blockchain.AddTransaction(tx.Sender, tx.Recipient, tx.Amount)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// 4. Respond: Tell the user it was successful
	w.WriteHeader(http.StatusCreated)
	fmt.Fprintf(w, "Success! Transaction added to the pending pool.\n")
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
		// HTTP 409 Conflict means the block was rejected (invalid hash, index, etc)
		http.Error(w, err.Error(), http.StatusConflict)
		return
	}

	w.WriteHeader(http.StatusCreated)
	fmt.Fprintf(w, "Block successfully validated and added to the chain!\n")
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

	// Instead of text, we return our entire address book back to the new peer!
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(s.node.GetPeers())
}
