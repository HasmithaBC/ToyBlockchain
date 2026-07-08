package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"toyblockchain/internal/core"
	"toyblockchain/internal/storage"
)

func main() {
	// If the user doesn't type a command, show help
	if len(os.Args) < 2 {
		printHelp()
		return
	}

	// os.Args[1] is the command the user typed
	command := os.Args[1]
	dataFile := "chain.json"

	// 1. Try to load the blockchain from the file
	bc, err := storage.LoadFromFile(dataFile)
	if err != nil {
		// If the file doesn't exist, we start a brand new chain!
		fmt.Println("No existing chain found, starting a new one...")
		bc = core.NewBlockchain()
	}

	switch command {
	case "help":
		printHelp()
	
	case "addtx":
		// We expect: go run main.go addtx Sender Recipient Amount
		// Which means we need at least 5 arguments in total
		if len(os.Args) < 5 {
			fmt.Println("Usage: addtx <sender> <recipient> <amount>")
			return
		}
		
		sender := os.Args[2]
		recipient := os.Args[3]
		amountStr := os.Args[4]

		// The amount comes in as a string, we must convert it to int64
		amount, err := strconv.ParseInt(amountStr, 10, 64)
		if err != nil {
			fmt.Println("Invalid amount:", amountStr)
			return
		}

		// Add it to our pending pool
		err = bc.AddTransaction(sender, recipient, amount)
		if err != nil {
			fmt.Println("Error adding transaction:", err)
		} else {
			fmt.Printf("Success! %s sent %d to %s (Pending in pool)\n", sender, amount, recipient)
			// 2. SAVE to the file so we don't forget it!
			storage.SaveToFile(bc, dataFile)
		}

	case "mine":
		if len(bc.PendingPool) == 0 {
			fmt.Println("No pending transactions to mine!")
			return
		}
		lastBlock := bc.Blocks[len(bc.Blocks)-1]
		newBlock := core.NewBlock(lastBlock.Index+1, bc.PendingPool, lastBlock.Hash)
		
		difficulty := 3 // Hardcode difficulty for CLI
		fmt.Printf("Mining block %d with %d transactions (Difficulty: %d)...\n", newBlock.Index, len(newBlock.Transactions), difficulty)
		newBlock.MineBlock(difficulty)
		
		bc.AddBlock(newBlock)
		bc.PendingPool = []core.Transaction{} // Clear the waiting room
		storage.SaveToFile(bc, dataFile)
		fmt.Println("Block successfully mined and saved!")

	case "balances":
		balances := bc.CalculateBalances()
		fmt.Println("--- Account Balances ---")
		for account, balance := range balances {
			fmt.Printf("%s: %d\n", account, balance)
		}
		
	case "validate":
		difficulty := 3 // Must match the mining difficulty!
		valid, badIndex := bc.ValidateChain(difficulty)
		if valid {
			fmt.Println("Chain is valid! No tampering detected.")
		} else {
			fmt.Printf("Chain is INVALID! Tampering detected at block %d\n", badIndex)
		}
		
	case "print":
		data, _ := json.MarshalIndent(bc, "", "  ")
		fmt.Println(string(data))

	default:
		fmt.Println("Unknown command:", command)
		printHelp()
	}
}

func printHelp() {
	fmt.Println("ToyBlockchain CLI")
	fmt.Println("Commands:")
	fmt.Println("  help      - Show this message")
	fmt.Println("  addtx     - Add a new transaction (addtx sender recipient amount)")
	fmt.Println("  mine      - Mine all pending transactions into a new block")
	fmt.Println("  balances  - Show all account balances")
	fmt.Println("  validate  - Verify the integrity of the blockchain")
	fmt.Println("  print     - Print the entire JSON blockchain")
}
