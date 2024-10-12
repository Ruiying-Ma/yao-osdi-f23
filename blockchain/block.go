package blockchain

import (
	"log"
	"bytes"
	"crypto/sha256"
	"math/big"
	"math/rand"
	"time"
	"encoding/gob"
	"strings"
	"fmt"

	"github.com/boltdb/bolt"

	"Project2/utils"
)

const TBITS = 16 // threshold of pow = 1 << (256 - TBITS). Set 16 when demo.

// A Block stores:
// - A set of transactions (TODO: use Merkle Tree to store txs)
// - A hash of all the txs (TODO: the root hash of the Merkle Tree)
// - A hash of a block that is already in the blockchain
// - A hash of itself
// - Time of creation
// - Nonce (for pow)
// - Height: the distance from the genisis to this block
// - Genisis: whether this block is genisis
// A Block can:
// - Serialize / Deserialize: to get stored on disk
// - Print its information
// - *Blindly* mine a block from given txs: 
// 		1. Find the hash of its previous block
//		2. Run POW to find Nonce
//		3. Compute the hash of the final block
// - Verify legal block:
//		0. Whether the block's TxHashes is correct
//		1. Whether there is at most one reward
//		2. Whether the block's prevhash is correct (no need for genisis)
//		3. Whether the block's height is correct
//		4. Whether the block's txs are legal
//		5. Whether the block's nonce is correct
//		6. Whether the block's hash is correct

type Block struct {
	Txs	[]*Transaction
	TxHashes	[]byte
	PrevHash 	[]byte
	Time 	int64 
	Nonce 	int 
	Height 	int 
	IsGenisis	bool
	Hash 	[]byte
}

func NewBlock(txs []*Transaction, genisis bool, bc *BlockChain) *Block {
	new_block := Block {
		Txs: []*Transaction{},
		TxHashes: []byte{},
		PrevHash: []byte{},
		Time: time.Now().UnixNano(),
		Nonce: 0,
		IsGenisis: genisis,
		Hash: []byte{},
	}
	// find prevhash and height
	var prev_hash []byte
	if !genisis {
		err := bc.DB.View(func (tx *bolt.Tx) error {
			bucket := tx.Bucket([]byte("blocks"))
			// TODO: Caution: here is a concurrency issue: when getting the prevhash and the tip block, some new block may be added to the blockchain after getting prevhash but before getting tip block
			// Don't know whether bolt db will ensure r/w exclusion in this case
			db_prev_hash := bucket.Get([]byte("l"))
			prev_hash = make([]byte, len(db_prev_hash))
			copy(prev_hash, db_prev_hash)
			tip_block := Deserialize(bucket.Get(prev_hash))
			new_block.Height = tip_block.Height + 1
			return nil
		})
		if err != nil {
			log.Panic(err)
		}
		new_block.PrevHash = prev_hash
	}
	// hash the txs
	for _, tx := range txs {
		new_block.Txs = append(new_block.Txs, tx)
	}
	new_block.HashTxs()
	// find nonce and hash
	start := time.Now()
	for {
		new_block.Nonce = rand.Int()
		hash := new_block.mid_hash()
		if is_acceptable_hash(hash) == true {
			new_block.Hash = hash
			elapsed := time.Since(start)
			fmt.Printf("Mining time = %d ns\n", elapsed.Nanoseconds())
			return &new_block
		}
	}
}

func (b *Block) Verify(bc *BlockChain) bool {
	start := time.Now()
	res := b.verify_reward() && b.verify_txhashes() && b.verify_prevhash_and_height(bc) && b.verify_txs(bc) && b.verify_nonce_and_hash()
	elapsed := time.Since(start)
	fmt.Printf("Verifying block time = %d ns\n", elapsed.Nanoseconds())
	return res
}

func (b *Block) HashTxs() {
	var tx_hashes [][]byte
	for _, tx := range b.Txs {
		tx_hashes = append(tx_hashes, tx.Hash)
	}
	hash := sha256.Sum256(bytes.Join(tx_hashes, []byte{}))
	b.TxHashes = hash[:]
}

func (b *Block) Serialize() []byte {
	var data bytes.Buffer
	encoder := gob.NewEncoder(&data)
	err := encoder.Encode(*b)
	if err != nil {
		log.Panic(err)
	}
	return data.Bytes()
}

func Deserialize(data []byte) *Block {
	var block Block 
	decoder := gob.NewDecoder(bytes.NewReader(data))
	err := decoder.Decode(&block)
	if err != nil {
		log.Panic(err)
	}
	return &block
}

func (b *Block) PrintBlock() string {
	var string_block []string
	string_block = append(string_block, fmt.Sprintf("--- Block Hash: %x", b.Hash))
	string_block = append(string_block, fmt.Sprintf("\tPrevHash: %x", b.PrevHash))
	string_block = append(string_block, fmt.Sprintf("\tTime: %d", b.Time))
	string_block = append(string_block, fmt.Sprintf("\tNonce: %d", b.Nonce))
	string_block = append(string_block, fmt.Sprintf("\tHeight: %d", b.Height))
	if b.IsGenisis {
		string_block = append(string_block, fmt.Sprintf("\tIsGenisis: True"))
	} else {
		string_block = append(string_block, fmt.Sprintf("\tIsGenisis: False"))
	}
	string_block = append(string_block, fmt.Sprintf("\tTxHashes: %x", b.TxHashes))
	for _, tx := range b.Txs {
		string_block = append(string_block, tx.PrintTx())
	}
	return strings.Join(string_block, "\n")
}

// Find whether `hash_tx` tx is in block `b`
// TODO: Can be improved from O(n) to O(log n) using Merkle Tree
func (b *Block) find_tx(hash_tx []byte) *Transaction {
	for _, tx := range b.Txs {
		if bytes.Compare(tx.Hash, hash_tx) == 0 {
			return tx
		}
	}
	return nil
}

func (b *Block) verify_reward() bool {
	num_reward := 0
	for _, tx := range b.Txs {
		if tx.IsReward {
			num_reward += 1
		}
	}
	if num_reward > 1 {
		fmt.Printf("verify_reward: more than one reward in a block\n")
		return false
	}
	return true
}


func (b *Block) verify_txhashes() bool {
	tmp_b := *b
	tmp_b.HashTxs()
	if bytes.Compare(tmp_b.TxHashes, b.TxHashes) != 0 {
		fmt.Printf("verify_txhashes: wrong TxHashes\n")
		return false
	}
	return true
}

func (b *Block) verify_prevhash_and_height(bc *BlockChain) bool {
	if b.IsGenisis {
		if b.Height == 0 {
			return true
		} else {
			fmt.Printf("verify_prevhash_and_height: genisis height not 0\n")
			return false
		}
	}
	iter := NewBlockChainIterator(bc)
	for {
		cur_block := iter.Next()
		if bytes.Compare(cur_block.Hash, b.PrevHash) == 0 {
			if cur_block.Height + 1 == b.Height {
				return true
			} else {
				fmt.Printf("verify_prevhash_and_height: wrong height\n")
				return false
			}
		}
		if cur_block.IsGenisis {
			break
		}
	}
	fmt.Printf("verify_prevhash_and_height: prevhash doesn't exist\n")
	return false
}


func (b *Block) verify_txs(bc *BlockChain) bool {
	if b.IsGenisis {
		if len(b.Txs) != 1 {
			fmt.Printf("verify_txs: wrong number of txs in genisis: should be one (reward tx)\n")
			return false
		}
	}
	for _, tx := range b.Txs {
		if tx.Verify(bc, b.PrevHash) == false {
			fmt.Print("verify_txs: wrong tx\n")
			return false
		}
	}
	return true
}

func (b *Block) verify_nonce_and_hash() bool {
	hash := b.mid_hash()
	if is_acceptable_hash(hash) == false {
		fmt.Printf("verify_nonce_and_hash: wrong nonce\n")
		return false
	}
	if bytes.Compare(hash, b.Hash) != 0 {
		fmt.Printf("verify_nonce_and_hash: wrong hash\n")
		return false
	}
	return true
}

func (b *Block) mid_hash() []byte {
	data := bytes.Join(
		[][]byte{
			b.TxHashes,
			b.PrevHash,
			utils.IntToHex(b.Time),
			utils.IntToHex(int64(b.Nonce)),
			utils.IntToHex(int64(b.Height)),
			utils.BoolToHex(b.IsGenisis),
			utils.IntToHex(TBITS),
		},
		[]byte{},
	)
	hash := sha256.Sum256(data)
	return hash[:]
}
func is_acceptable_hash(hash []byte) bool {
	var hashInt big.Int
	hashInt.SetBytes(hash)
	threshold := big.NewInt(1)
	threshold.Lsh(threshold, uint(256 - TBITS))
	if hashInt.Cmp(threshold) == -1 {
		return true
	}
	return false
}

