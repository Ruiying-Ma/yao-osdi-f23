package blockchain

import (
	"log"

	"github.com/boltdb/bolt"
)

// A BlockChainIterator stores:
// - The hash of the current block
// - The db that stores the blockchain
// A BlockChainIterator can:
// - Iterate from the tip of the db to the genesis 

type BlockChainIterator struct {
	Hash	[]byte // the hash of the current block
	DB	*bolt.DB
}

func NewBlockChainIterator(bc *BlockChain) *BlockChainIterator {
	iter := BlockChainIterator{
		Hash: []byte{},
		DB: bc.DB,
	}
	var hash []byte
	err := bc.DB.View(func (tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte("blocks"))
		db_hash := bucket.Get([]byte("l"))
		hash = make([]byte, len(db_hash))
		copy(hash, db_hash)
		return nil
	})
	iter.Hash = hash
	if err != nil {
		log.Panic(err)
	}
	return &iter
}

// Return current block and move to its predecessor
func (iter *BlockChainIterator) Next() *Block {
	var b *Block 
	err := iter.DB.View(func (tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte("blocks"))
		b = Deserialize(bucket.Get(iter.Hash))
		return nil
	})
	if err != nil {
		log.Panic(err)
	}
	iter.Hash = b.PrevHash
	return b
}

