package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"time"
	"toyblockchain/internal/api"
	"toyblockchain/internal/core"
	"toyblockchain/internal/network"
)

// This is our Server (Node). It will run forever in the background.
func main() {
	// Allow the user to specify a port via command line flag (defaults to 8080)
	port := flag.String("port", "8080", "Port to run the node on")
	flag.Parse()

	fmt.Printf("Initializing ToyBlockchain Node on port %s...\n", *port)
	
	// 1. Create the core blockchain (The Kitchen)
	bc := core.NewBlockchain()
	
	// 2. Load from disk if available
	filename := fmt.Sprintf("chain_%s.json", *port)
	err := bc.LoadFromFile(filename)
	if err == nil {
		fmt.Printf("Loaded existing blockchain from %s (Height: %d)\n", filename, len(bc.Blocks))
	} else {
		fmt.Printf("Started new blockchain (Genesis Block created). Saving to %s\n", filename)
	}

	// 3. Start a background saver (Goroutine) that saves the chain every 5 seconds
	go func() {
		for {
			time.Sleep(5 * time.Second)
			bc.SaveToFile(filename)
		}
	}()
	
	// 4. Wrap it in a network Node (The Restaurant)
	node := network.NewNode(*port, bc)
	
	// 5. Auto-Discovery: If we are not the seed node (8080), register with it!
	if *port != "8080" {
		go func() {
			// Wait 2 seconds for our own server to start listening
			time.Sleep(2 * time.Second)
			
			seedURL := "http://localhost:8080"
			myURL := fmt.Sprintf("http://localhost:%s", *port)
			
			reqData := map[string]string{"peer_url": myURL}
			jsonData, _ := json.Marshal(reqData)
			
			resp, err := http.Post(seedURL+"/register", "application/json", bytes.NewBuffer(jsonData))
			if err == nil {
				defer resp.Body.Close()
				var peers []string
				if json.NewDecoder(resp.Body).Decode(&peers) == nil {
					node.AddPeer(seedURL)
					for _, p := range peers {
						// 1. Prevent the node from adding itself to its own address book!
						// (Because the seed node adds us to its list *before* returning it)
						if p == myURL {
							continue
						}
						
						// 2. Add the discovered peer to our own address book
						node.AddPeer(p)
						
						// 3. IMPORTANT: Tell that peer about US! Otherwise, they won't know we exist.
						go func(peerURL string) {
							http.Post(peerURL+"/register", "application/json", bytes.NewBuffer(jsonData))
						}(p)
					}
					fmt.Printf("\n[Auto-Discovery] Registered with seed %s and connected to %d peers.\n", seedURL, len(node.GetPeers()))
					
					fmt.Println("[Auto-Discovery] Starting automatic chain sync...")
					node.SyncChain()
				}
			} else {
				fmt.Printf("\n[Auto-Discovery] Failed to register with seed node: %v\n", err)
			}
		}()
	}
	
	// 6. Start the HTTP server (The Waiter)
	api.StartServer(*port, node)
}
