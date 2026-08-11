package main

import (
	"fmt"
	"os"
)

// This is our Client. It does NO blockchain math.
// It will simply send HTTP requests to the Node and print the response.
func main() {
	fmt.Println("--- ToyBlockchain Client ---")
	
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run cmd/client/main.go <command>")
		return
	}
	
	command := os.Args[1]
	fmt.Printf("You requested to run: %s\n", command)
	fmt.Println("Soon, this will send an HTTP request to the Server!")
}
