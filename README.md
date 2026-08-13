# Toy Blockchain & Ledger Simulator

A completely from-scratch blockchain and ledger simulator written in pure Go. This project implements a deterministic chain of blocks, proof-of-work mining, full-chain validation, and a P2P command-line client.

## Running the P2P Cluster

You can run a local cluster of 3 peer-to-peer nodes using the provided PowerShell script:

```powershell
.\start_cluster.ps1
```

This will spin up three separate node processes listening on ports `8080`, `8081`, and `8082`. 

**Auto-Discovery**: Port `8080` acts as the "Seed Node". When nodes on other ports start up, they automatically register with `8080`, download its address book, broadcast their existence to all other discovered peers, and synchronize their blockchains!

*(Want to run an individual node manually on a different port? You can do so by explicitly passing the `--port` flag to the node binary):*
```bash
go run cmd/node/main.go --port 8083
```

## Using the P2P Client

The project includes a client CLI for generating cryptographic keypairs and interacting with the network. All client commands accept an optional `--port` flag to target a specific node (defaults to `8080`).

### 1. Generate Wallets

Generate a wallet for Alice and Bob:
```bash
go run cmd/client/main.go generate-wallet Alice
go run cmd/client/main.go generate-wallet Bob
```
This automatically creates a `Secret/` folder in your workspace and saves `.key` (private) and `.pub` (public) files for each user.

### 2. Mine a Block (Block Rewards)

To inject money into the economy, you can mine an empty block and collect the Coinbase reward (50 coins). The node will automatically look inside your `Secret/` folder for the public key.
```bash
go run cmd/client/main.go mine --reward Alice.pub
```

### 3. Send a Transaction

Now that Alice has money, she can send a transaction to Bob! The client is smart and automatically resolves filenames to the `Secret/` directory.
```bash
go run cmd/client/main.go send --from Alice.key --to Bob.pub --amount 25
```
*(Note: Transactions are mathematically secured with an embedded timestamp to prevent replay attacks!)*

### 4. Client Commands

You can run the following commands to interact with the blockchain. Append `--port 8081` or `--port 8082` to talk to different nodes in your cluster!

- `help` : Show the help message (and a tip on how to start new nodes)
- `send` : Send a signed transaction to the network (requires --from, --to, --amount)
- `mine` : Mine all pending transactions into a new block. Accepts `--reward <file.pub>` to collect a block reward.
- `balances` : Dynamically calculate and show all account balances. Automatically translates hex keys to friendly names by scanning your `Secret/` folder!
- `validate` : Ask a node to verify the integrity of its blockchain (Hashes, Links, PoW, Timestamps)
- `print` : Print the entire JSON representation of the node's blockchain
- `sync` : Manually trigger a node to ask its peers for the longest valid chain
- `peers` : View the node's address book (list of connected peers)

Example:
```bash
go run cmd/client/main.go mine --reward Bob.pub --port 8082
go run cmd/client/main.go balances --port 8081
```

## Running Tests

To run all unit tests for the core logic:
```bash
go test -v ./...
```

## Design Decisions

1. **Architecture & Separation of Concerns**: The project is split into `cmd` (for the CLI entry points), `internal/core` (for strict domain models like Block, Transaction, Blockchain, and Ledger logic), and `internal/network` and `internal/api` for the P2P networking layer.
2. **Deterministic Hashing**: Hashing uses `fmt.Sprintf` to serialize fields in a strict order. The Genesis block uses a hardcoded timestamp to ensure its hash is absolutely identical on every run.
3. **Dynamic Ledger**: Balances are never statically stored. They are calculated dynamically by replaying the entire chain from the Genesis block to ensure trustless verification.
4. **Pointers for Efficiency**: Blocks are passed around as pointers (`*Block`) to avoid heavy memory copying.

## Known Limitations

1. **No Merkle Trees**: All transactions are serialized as a flat string rather than hashed into a Merkle Root.
