# Toy Blockchain & Ledger Simulator

A completely from-scratch blockchain and ledger simulator written in pure Go. This project implements a deterministic chain of blocks, proof-of-work mining, full-chain validation, and a command-line interface.

## Build and Run Instructions

To build the executable:
```bash
go build -o toychain cmd/toychain/main.go
```

To run the program directly:
```bash
go run cmd/toychain/main.go [command]
```

To run all tests:
```bash
go test -v ./...
```

## Command-Line Commands

- `help` : Show the help message
- `addtx <sender> <recipient> <amount>` : Add a new transaction to the pending pool
- `mine` : Mine all pending transactions into a new block on the chain (Proof of Work)
- `balances` : Dynamically calculate and show all account balances from the chain
- `validate` : Verify the integrity of the blockchain (Hashes, Links, PoW, Timestamps)
- `print` : Print the entire JSON representation of the blockchain

## Design Decisions

1. **Architecture & Separation of Concerns**: The project is split into `cmd` (for the CLI entry point), `internal/core` (for strict domain models like Block, Transaction, Blockchain, and Ledger logic), and `internal/storage` (for JSON persistence). This avoids bloated files and circular dependencies.
2. **Deterministic Hashing**: Hashing uses `fmt.Sprintf` to serialize fields in a strict order. The Genesis block uses a hardcoded timestamp to ensure its hash is absolutely identical on every run.
3. **Dynamic Ledger**: Balances are never statically stored. They are calculated dynamically by replaying the entire chain from the Genesis block to ensure trustless verification.
4. **Pointers for Efficiency**: Blocks are passed around as pointers (`*Block`) to avoid heavy memory copying.

## Known Limitations

1. **Single Node (No Networking)**: This is a local simulation. There is no P2P networking or consensus mechanism to resolve forks.
2. **No Digital Signatures**: Transactions do not require cryptographic signatures (e.g., ECDSA). Anyone can currently pretend to be the sender.
3. **No Merkle Trees**: All transactions are serialized as a flat string rather than hashed into a Merkle Root. 
