package storage

import (
	"encoding/json"
	"os"
	"toyblockchain/internal/core"
)

// SaveToFile saves the blockchain to a JSON file.
func SaveToFile(chain *core.Blockchain, filename string) error {
	// MarshalIndent converts our Go struct into a beautiful, readable JSON string
	data, err := json.MarshalIndent(chain, "", "  ")
	if err != nil {
		return err
	}

	// Write the JSON string to a file on the hard drive
	// 0644 means "read/write for owner, read-only for others"
	return os.WriteFile(filename, data, 0644)
}

// LoadFromFile reads the JSON file and converts it back into a Go struct.
func LoadFromFile(filename string) (*core.Blockchain, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, err // File might not exist yet
	}

	var chain core.Blockchain
	// Unmarshal converts the JSON string back into our empty Go struct
	err = json.Unmarshal(data, &chain)
	if err != nil {
		return nil, err
	}

	return &chain, nil
}
