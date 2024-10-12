# Project2

A universal assumption: perfect network. No network failure. All messages will be received eventually.
## References

[An implementation of centralized blockchain](https://github.com/Jeiwan/blockchain_go/tree/master)

[Bolt DB](https://pkg.go.dev/github.com/boltdb/bolt)

## Preliminaries

### Bolt DB
Pay attention to the **caveat**:

    The database uses a read-only, memory-mapped data file to ensure that applications cannot corrupt the database, however, this means that keys and values returned from Bolt cannot be changed. Writing to a read-only byte slice will cause Go to panic.

    Keys and values retrieved from the database are only valid for the life of the transaction. When used outside the transaction, these byte slices can point to different data or can point to invalid memory which will cause a panic.

Particularly, when we want to read a slice from the database, what is **NOT** recommended to do is:

```go
    // read from db
    err := db.View(func (tx *bolt.Tx) error {
        // Get some bucket 
        // ...
        data := bucket.Get([]byte("some key")) // `data` is a slice
        //...
    })
    // Outside the db tx, either
    // revise `data`, or
    data = another_data
    // read `data`
    another_data := data
```

This can cause panic, as stated in the doc (since the underlying array of slice `data` is in the field of the transaction). If we want to use the slice read from the db, what is recommended to do is:

```go
    // read a slice of bytes from db
    var data []byte 
    err := db.View(func (tx *bolt.Tx) error {
        // Get some bucket 
        // ...
        db_data := bucket.Get([]byte("some key")) // `data` is a slice
        data = make([]byte, len(db_data)) // allocate mem for `data`
        copy(data, db_data) // deep copy, the underlying array is copied.
        //...
    })
    // Outside the db tx, we can use `data` (read or revise)
```

In this example, the underlying array of `db_data` is copied to the the underlying array of `data`. Hence, `data` has its own underlying array that is out of the field of the db transaction. That's why we can use `data` outside the db transaction.

## BlockChain Layout

### Wallet
A wallet is equivalent to a user in the blockchain network. A real-world person can create multiple wallets. A wallet has a public key and a secret key. Its public key (and the address of the public) is public to the whole network.

### Transaction
A transaction is first signed and then hashed. We hash the transaction so that the hash serves as an abstract of this transaction. When we want to identify this transaction in the future, we only need to use its hash. Hashing is performed after signing so that different transactions have different hashes.

### Block
Notice that when a miner is going to make a block from a set of txs, it should ensure that a tx initiator can only occur once in this block (to defend double-spent attack). 

### BlockChain
Stored in the db. I implement a 1-confirmation blockchain (i.e., when branch occurs, if a branch is longer by 1 block, then this branch is chosen). 

For a more detailed description, see the comments in the codes.

## Experiments
- Time to mine a block vs. `TBITS` (defined in `Project2/blockchain/block.go`). Need a graph.
    1. Fix `TBITS` (defined in `Project2/blockchain/block.go`), draw the histogram of the time to mine a block. 
    2. Put the histograms under different `TBITS` into one graph.

- Time to verify a block/tx (may be improved after we implement the UTXO set and the Merkle tree)
    1. Draw the distribution (histogram) of the time to verify a block/tx.

    When finishing `make expriment`:
    - Enter `stats` and run `./stats.sh`. It extracts needed data from `debugxxxx.out`
    - Enter `plot` and run `mining_time.py` and `verifying_block_time.py`. The figures `mining_time_16.png` and `verifying_block_time.png` will be generated under this file.
- Others

## Network Layout
Currently, I implement a decentralized network. The network consists of 5 miners. Each miner can
- broadcast a tx (can understand this functionality as a client process)
- broadcast a block
- receive a tx
- receive a block

### Problems
<!-- Sometimes the blockchains of the miners are consistent. However, this is not always true.

Unavoidable consistency issue: there exist cases where the miners cannot reach consensus on their offline blockchain (e.g., at a short time interval, network partition occurs. Although messages will finally be received, the blocks accepted by different miners are different. Thus, one miner may broadcast a tx based on its blockchain, which conflicts with other miners' blockchains and thus fails to be accepted by other miners (but the miner itself, and some other miners that agree with it on the blockchain, will accept this tx).)

If we are not going to implement a consensus protocol, I think we can implement a centralized distributed system (same as the implementation in the github ref). I haven't implemented this. -->

## How to run 

`make experiment` launches the following commands. 
`make stop` stops the miners. 

1. Open 5 machines 8060, 8061, 8062, 8063, 8064.
2. On each machine:
    1. Ensure that `/osdata/osgroup10/` directory exists
    2. Ensure that `/osdata/osgroup10/` is empty.
3. Run `go run main.go -mid=xxxx > debugxxxx.out 2>errorxxxx.out` under `~/Project2` on machine xxxx. Try one's best to run these commands simultaneously. 
4. When finishes, don't forget to run `rm /osdata/osgroup10/*` to clear the directory if you want to rerun.

In `debugxxxx.out`, you can see the runtime messages. In `errorxxxx.out`, you can observe the error messages. In `blockchainxxxx`, you can see the final blockchain.

## Fake Clients and Miners
The requirements are 
1. Demonstrate the case when the blocks get corrupted, miners reject these invalid blocks.
2. Demonstrate the cases where if there is a lying miner (which did not correctly solve the PoW puzzle), the other miners should also reject the block.
3. Demonstrate the case when there is a fork, the longest chain rule is correctly applied. 

Among them 1, 3 can directly be demonstrated without writing new client/miner program:
- For 1, search `Invalid block` in `debug8060.out`
- For 3, search `New block branches the chain: ` in `debug8060.out`.

We only need to write a miner that doesn't honestly solve the PoW (instead, it always generate a block with nonce = 1): Write a function `fake_mine()` in `miner/miner.go`. `fake_mine()` is almost the same as `mine()`, except that it changes a line in `mine()`
```go
    // a line in mine()
    new_block := blockchain.NewBlock(txs, false, m.BC)
```
into 
```go
    new_block := Block {
		Txs: txs,
		TxHashes: , //TODO: similar as NewBlock() in blockchain/block.go
		PrevHash: , //TODO: similar as NewBlock() in blockchain/block.go
		Time: time.Now().UnixNano(),
		Nonce: 1, //TODO: may be a wrong nonce
		IsGenisis: false,
		Hash: ,//TODO: similar as NewBlock() in 
    }
    // Similar as how we generate a new block, except that we do not check the nonce for POW. Pay attention to the order of filling the fields of `Block`.
```

## Future Implementations
We may implement the following features/improvements later:
<!-- - A decentralized network (instead of a centralized one). -->
- Allow a user to create a wallet at any time and use it in the future (i.e., a p2p network such that a user can join/leave dynamically).
<!-- - A command-line implementation (see github ref). -->
- A Merkle tree for transactions in the same block (see github ref).
- A UTXO set (see github ref).
- Dynamically adjust the difficulty of POW
- k-Confirmation with k greater than one