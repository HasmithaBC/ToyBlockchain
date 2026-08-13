package core

import (
	"crypto/ed25519"
	"encoding/hex"
	"fmt"
)

// Transaction represents a transfer of value from a sender to a recipient.
type Transaction struct {
	Timestamp int64  `json:"timestamp"`
	Sender    string `json:"sender"`
	Recipient string `json:"recipient"`
	Amount    int64  `json:"amount"`
	PublicKey string `json:"public_key"`
	Signature string `json:"signature"`
}

// Payload stringifies the transaction data for signing/verifying.
func (tx *Transaction) Payload() string {
	return fmt.Sprintf("%d:%s:%s:%d", tx.Timestamp, tx.Sender, tx.Recipient, tx.Amount)
}

// Sign uses the given private key to sign the transaction payload.
func (tx *Transaction) Sign(privateKeyHex string) error {
	privKeyBytes, err := hex.DecodeString(privateKeyHex)
	if err != nil {
		return err
	}
	if len(privKeyBytes) != ed25519.PrivateKeySize {
		return fmt.Errorf("invalid private key size")
	}

	privKey := ed25519.PrivateKey(privKeyBytes)
	
	// Ensure the PublicKey field matches the generated signature
	pubKey := privKey.Public().(ed25519.PublicKey)
	tx.PublicKey = hex.EncodeToString(pubKey)
	
	// The Sender address is commonly just the PublicKey in simple blockchains
	if tx.Sender == "" {
		tx.Sender = tx.PublicKey
	}

	// Now that everything is populated, sign the payload!
	sig := ed25519.Sign(privKey, []byte(tx.Payload()))
	tx.Signature = hex.EncodeToString(sig)
	
	return nil
}

// Verify checks if the signature is valid for the transaction payload.
func (tx *Transaction) Verify() bool {
	// The "System" doesn't need a signature for mining rewards
	if tx.Sender == "System" {
		return true
	}

	pubKeyBytes, err := hex.DecodeString(tx.PublicKey)
	if err != nil || len(pubKeyBytes) != ed25519.PublicKeySize {
		return false
	}

	sigBytes, err := hex.DecodeString(tx.Signature)
	if err != nil {
		return false
	}

	return ed25519.Verify(ed25519.PublicKey(pubKeyBytes), []byte(tx.Payload()), sigBytes)
}
