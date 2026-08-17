package main

import (
	"bytes"
	"crypto/ed25519"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
	"toyblockchain/internal/core"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		return
	}

	command := os.Args[1]

	switch command {
	case "help":
		printUsage()
	case "generate-wallet":
		handleGenerateWallet()
	case "send":
		handleSend()
	case "mine":
		handleMine()
	case "balances":
		handleBalances()
	case "validate":
		handleValidate()
	case "print":
		handlePrint()
	case "sync":
		handleSync()
	case "peers":
		handlePeers()
	case "add-peer":
		handleAddPeer()
	default:
		fmt.Printf("Unknown command: %s\n", command)
		printUsage()
	}
}

func printUsage() {
	fmt.Println("ToyBlockchain P2P Client")
	fmt.Println("Usage:")
	fmt.Println("  client help")
	fmt.Println("  client generate-wallet <filename_prefix>")
	fmt.Println("  client send --from <private_key_file> --to <public_key_file> --amount <int> [--port 8080]")
	fmt.Println("  client mine [--reward <public_key_file>] [--port 8080]")
	fmt.Println("  client balances [--port 8080]")
	fmt.Println("  client validate [--port 8080]")
	fmt.Println("  client print [--port 8080]")
	fmt.Println("  client sync [--port 8080]")
	fmt.Println("  client peers [--port 8080]")
	fmt.Println("  client add-peer --port 8080 --peer http://localhost:8081")
	fmt.Println("")
	fmt.Println("To start a new P2P node, run:")
	fmt.Println("  go run cmd/node/main.go --port <port_number>")
}

func handleGenerateWallet() {
	if len(os.Args) < 3 {
		fmt.Println("Usage: client generate-wallet <filename_prefix>")
		return
	}
	prefix := os.Args[2]

	pubKey, privKey, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		fmt.Printf("Error generating keys: %v\n", err)
		return
	}

	privHex := hex.EncodeToString(privKey)
	pubHex := hex.EncodeToString(pubKey)

	// Ensure the Secret directory exists
	err = os.MkdirAll("Secret", 0755)
	if err != nil {
		fmt.Printf("Error creating Secret directory: %v\n", err)
		return
	}

	privPath := filepath.Join("Secret", prefix+".key")
	pubPath := filepath.Join("Secret", prefix+".pub")

	err = os.WriteFile(privPath, []byte(privHex), 0600)
	if err != nil {
		fmt.Printf("Error writing private key: %v\n", err)
		return
	}

	err = os.WriteFile(pubPath, []byte(pubHex), 0644)
	if err != nil {
		fmt.Printf("Error writing public key: %v\n", err)
		return
	}

	fmt.Printf("Successfully generated wallet!\n")
	fmt.Printf("Private key saved to: %s (KEEP THIS SECRET)\n", privPath)
	fmt.Printf("Public key saved to: %s (Share this to receive money)\n", pubPath)
}

func handleSend() {
	sendCmd := flag.NewFlagSet("send", flag.ExitOnError)
	fromFile := sendCmd.String("from", "", "File containing your private key (e.g. alice.key)")
	toFile := sendCmd.String("to", "", "File containing recipient's public key (e.g. bob.pub)")
	amount := sendCmd.Int64("amount", 0, "Amount to send")
	port := sendCmd.String("port", "8080", "Port of the node to interact with")

	sendCmd.Parse(os.Args[2:])

	if *fromFile == "" || *toFile == "" || *amount <= 0 {
		fmt.Println("Usage: client send --from <private_key_file> --to <public_key_file> --amount <int> [--port 8080]")
		return
	}

	fromPath := *fromFile
	if filepath.Dir(fromPath) == "." {
		fromPath = filepath.Join("Secret", fromPath)
	}

	toPath := *toFile
	if filepath.Dir(toPath) == "." {
		toPath = filepath.Join("Secret", toPath)
	}

	privKeyBytes, err := os.ReadFile(fromPath)
	if err != nil {
		fmt.Printf("Error reading sender private key: %v\n", err)
		return
	}
	senderPrivKeyHex := string(privKeyBytes)

	pubKeyBytes, err := os.ReadFile(toPath)
	if err != nil {
		fmt.Printf("Error reading recipient public key: %v\n", err)
		return
	}
	recipientPubKeyHex := string(pubKeyBytes)

	tx := core.Transaction{
		Timestamp: time.Now().UnixNano(),
		Recipient: recipientPubKeyHex,
		Amount:    *amount,
	}

	err = tx.Sign(senderPrivKeyHex)
	if err != nil {
		fmt.Printf("Error signing transaction: %v\n", err)
		return
	}

	jsonData, err := json.Marshal(tx)
	if err != nil {
		fmt.Printf("Error encoding JSON: %v\n", err)
		return
	}

	url := fmt.Sprintf("http://localhost:%s/transaction", *port)
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		fmt.Printf("Failed to connect to node on port %s: %v\n", *port, err)
		return
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	fmt.Printf("Node Response (%s): %s\n", resp.Status, string(body))
}

func handleMine() {
	mineCmd := flag.NewFlagSet("mine", flag.ExitOnError)
	port := mineCmd.String("port", "8080", "Port of the node to interact with")
	reward := mineCmd.String("reward", "", "Public key file to send mining reward to (e.g. Alice.pub)")
	mineCmd.Parse(os.Args[2:])

	url := fmt.Sprintf("http://localhost:%s/mine", *port)
	
	if *reward != "" {
		rewardPath := *reward
		if filepath.Dir(rewardPath) == "." {
			rewardPath = filepath.Join("Secret", rewardPath)
		}
		
		pubKeyBytes, err := os.ReadFile(rewardPath)
		if err != nil {
			fmt.Printf("Error reading reward public key: %v\n", err)
			return
		}
		url = fmt.Sprintf("http://localhost:%s/mine?reward=%s", *port, string(pubKeyBytes))
	}

	resp, err := http.Get(url)
	if err != nil {
		fmt.Printf("Failed to connect to node on port %s: %v\n", *port, err)
		return
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode == http.StatusOK {
		fmt.Printf("Successfully mined block!\n")
	} else {
		fmt.Printf("Node Error: %s\n", string(body))
	}
}

func handleBalances() {
	balCmd := flag.NewFlagSet("balances", flag.ExitOnError)
	port := balCmd.String("port", "8080", "Port of the node to interact with")
	balCmd.Parse(os.Args[2:])

	url := fmt.Sprintf("http://localhost:%s/balances", *port)
	resp, err := http.Get(url)
	if err != nil {
		fmt.Printf("Failed to connect to node on port %s: %v\n", *port, err)
		return
	}
	defer resp.Body.Close()

	var balances map[string]int64
	if err := json.NewDecoder(resp.Body).Decode(&balances); err != nil {
		fmt.Printf("Error decoding JSON: %v\n", err)
		return
	}

	fmt.Println("--- Account Balances ---")
	
	nameMap := make(map[string]string)
	files, err := os.ReadDir("Secret")
	if err == nil {
		for _, file := range files {
			if strings.HasSuffix(file.Name(), ".pub") {
				content, err := os.ReadFile(filepath.Join("Secret", file.Name()))
				if err == nil {
					name := strings.TrimSuffix(file.Name(), ".pub")
					nameMap[string(content)] = name
				}
			}
		}
	}

	for account, balance := range balances {
		if account == "System" && balance == 0 {
			continue
		}
		
		displayAccount := account
		if name, ok := nameMap[account]; ok {
			displayAccount = name
		} else if len(account) > 16 && account != "System" {
			displayAccount = account[:8] + "..." + account[len(account)-8:]
		}
		
		fmt.Printf("%s: %d\n", displayAccount, balance)
	}
}

func handleValidate() {
	valCmd := flag.NewFlagSet("validate", flag.ExitOnError)
	port := valCmd.String("port", "8080", "Port of the node to interact with")
	valCmd.Parse(os.Args[2:])

	url := fmt.Sprintf("http://localhost:%s/validate", *port)
	resp, err := http.Get(url)
	if err != nil {
		fmt.Printf("Failed to connect to node on port %s: %v\n", *port, err)
		return
	}
	defer resp.Body.Close()

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		fmt.Printf("Error decoding JSON: %v\n", err)
		return
	}

	isValid, ok := result["is_valid"].(bool)
	if ok && isValid {
		fmt.Println("Node reports its chain is VALID! No tampering detected.")
	} else {
		badIndex := result["bad_index"]
		fmt.Printf("Node reports its chain is INVALID! Tampering detected at block %v\n", badIndex)
	}
}

func handlePrint() {
	printCmd := flag.NewFlagSet("print", flag.ExitOnError)
	port := printCmd.String("port", "8080", "Port of the node to interact with")
	printCmd.Parse(os.Args[2:])

	url := fmt.Sprintf("http://localhost:%s/blocks", *port)
	resp, err := http.Get(url)
	if err != nil {
		fmt.Printf("Failed to connect to node on port %s: %v\n", *port, err)
		return
	}
	defer resp.Body.Close()

	var blocks []core.Block
	if err := json.NewDecoder(resp.Body).Decode(&blocks); err != nil {
		fmt.Printf("Error decoding JSON: %v\n", err)
		return
	}

	data, _ := json.MarshalIndent(blocks, "", "  ")
	fmt.Println(string(data))
}

func handleSync() {
	syncCmd := flag.NewFlagSet("sync", flag.ExitOnError)
	port := syncCmd.String("port", "8080", "Port of the node to interact with")
	syncCmd.Parse(os.Args[2:])

	url := fmt.Sprintf("http://localhost:%s/sync", *port)
	resp, err := http.Post(url, "application/json", nil)
	if err != nil {
		fmt.Printf("Failed to connect to node on port %s: %v\n", *port, err)
		return
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode == http.StatusOK {
		fmt.Printf("Successfully triggered sync! Node responded:\n%s", string(body))
	} else {
		fmt.Printf("Node Error: %s\n", string(body))
	}
}

func handlePeers() {
	peersCmd := flag.NewFlagSet("peers", flag.ExitOnError)
	port := peersCmd.String("port", "8080", "Port of the node to interact with")
	peersCmd.Parse(os.Args[2:])

	url := fmt.Sprintf("http://localhost:%s/peers", *port)
	resp, err := http.Get(url)
	if err != nil {
		fmt.Printf("Failed to connect to node on port %s: %v\n", *port, err)
		return
	}
	defer resp.Body.Close()

	var peers []string
	if err := json.NewDecoder(resp.Body).Decode(&peers); err != nil {
		fmt.Printf("Error decoding JSON: %v\n", err)
		return
	}

	fmt.Printf("--- Node on port %s Address Book ---\n", *port)
	if len(peers) == 0 {
		fmt.Println("No peers found (Lonely node...)")
	} else {
		for i, p := range peers {
			fmt.Printf("%d. %s\n", i+1, p)
		}
	}
}

func handleAddPeer() {
	peerCmd := flag.NewFlagSet("add-peer", flag.ExitOnError)
	port := peerCmd.String("port", "8080", "Port of the node you want to configure")
	peerURL := peerCmd.String("peer", "", "The URL of the peer to add (e.g. http://localhost:8081)")
	peerCmd.Parse(os.Args[2:])

	if *peerURL == "" {
		fmt.Println("Usage: client add-peer --port 8080 --peer http://localhost:8081")
		return
	}

	// 1. Tell the Target Node about the New Peer
	url1 := fmt.Sprintf("http://localhost:%s/register", *port)
	reqData1 := map[string]string{"peer_url": *peerURL}
	jsonData1, _ := json.Marshal(reqData1)

	resp1, err := http.Post(url1, "application/json", bytes.NewBuffer(jsonData1))
	if err != nil {
		fmt.Printf("Failed to connect to node on port %s: %v\n", *port, err)
	} else {
		defer resp1.Body.Close()
		if resp1.StatusCode == http.StatusCreated {
			fmt.Printf("Successfully added %s to Node %s's address book!\n", *peerURL, *port)
		} else {
			fmt.Printf("Node %s returned error: %s\n", *port, resp1.Status)
		}
	}

	// 2. Tell the New Peer about the Target Node (so they can talk to each other!)
	targetURL := fmt.Sprintf("http://localhost:%s", *port)
	reqData2 := map[string]string{"peer_url": targetURL}
	jsonData2, _ := json.Marshal(reqData2)

	url2 := fmt.Sprintf("%s/register", *peerURL)
	resp2, err := http.Post(url2, "application/json", bytes.NewBuffer(jsonData2))
	if err != nil {
		fmt.Printf("Failed to connect to peer %s: %v\n", *peerURL, err)
	} else {
		defer resp2.Body.Close()
		if resp2.StatusCode == http.StatusCreated {
			fmt.Printf("Successfully added %s to Peer %s's address book (Reciprocal Registration)!\n", targetURL, *peerURL)
			
			// NEW: Decode the peers list returned by the peer to find other nodes in the network!
			var returnedPeers []string
			if err := json.NewDecoder(resp2.Body).Decode(&returnedPeers); err == nil {
				for _, p := range returnedPeers {
					// Don't add ourselves or the target node again
					if p == targetURL || p == *peerURL {
						continue
					}
					
					fmt.Printf("Discovered secondary peer %s! Linking it to %s...\n", p, targetURL)
					
					// Tell Node 8080 to add Node 8083
					req3 := map[string]string{"peer_url": p}
					json3, _ := json.Marshal(req3)
					http.Post(fmt.Sprintf("%s/register", targetURL), "application/json", bytes.NewBuffer(json3))
					
					// Tell Node 8083 to add Node 8080
					req4 := map[string]string{"peer_url": targetURL}
					json4, _ := json.Marshal(req4)
					http.Post(fmt.Sprintf("%s/register", p), "application/json", bytes.NewBuffer(json4))
				}
			}
			
		} else {
			fmt.Printf("Peer %s returned error: %s\n", *peerURL, resp2.Status)
		}
	}
}
