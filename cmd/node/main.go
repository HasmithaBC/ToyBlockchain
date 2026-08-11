package main

import (
	"fmt"
	"toyblockchain/internal/api"
	"toyblockchain/internal/core"
	"toyblockchain/internal/network"
)

// This is our Server (Node). It will run forever in the background.
func main() {
	fmt.Println("Initializing ToyBlockchain Node...")
	
	// 1. Create the core blockchain (The Kitchen)
	bc := core.NewBlockchain()
	
	// 2. Wrap it in a network Node (The Restaurant)
	node := network.NewNode("8080", bc)
	
	// 3. Start the HTTP server (The Waiter)
	api.StartServer("8080", node)
}
