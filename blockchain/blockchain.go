package blockchain

import (
	"log"
	"strings"
	"bytes"
	"fmt"

	"github.com/boltdb/bolt"
)

// A BlockChain stores:
// - DB: the offline place where the blockchain is stored
// A BlockChain can:
// - Append a block to the chain:
//		1. Verify legal block
//		2. If legal, update the blockchain
//		3. If not legal, yell and do nothing

const DBDIR = "/osdata/osgroup10/blockchain-"

type BlockChain struct {
	DB	*bolt.DB
}

func NewBlockChain(machine_id string) *BlockChain {
	filename := DBDIR + machine_id + ".db"
	db, err := bolt.Open(filename, 0644, nil)
	if err != nil {
		log.Panic(err)
	}
	err = db.Update(func (tx *bolt.Tx) error {
		_, err := tx.CreateBucket([]byte("blocks"))
		if err != nil {
			log.Panic(err)
		}
		return nil
	})
	//fmt.Printf("New block chain created\n")///////////////////////////////////////////////////
	if err != nil {
		log.Panic(err)
	}
	return &BlockChain {
		DB: db,
	}
}

func (bc *BlockChain) AppendBlock(b *Block) {
	//fmt.Printf("Append block %#v\n", *b)//////////////////////////////////////////////////////
	if b.Verify(bc) == false {
		fmt.Printf("Invalid block\n")/////////////////////////////////////////////////
		return
	}
	data := b.Serialize()

	err := bc.DB.Update(func (tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte("blocks"))
		err := bucket.Put(b.Hash, data)
		if err != nil {
			log.Panic(err)
		}
		if b.IsGenisis{
			err = bucket.Put([]byte("l"), b.Hash)
			if err != nil {
				log.Panic(err)
			}
			return nil
		}
		last_hash := bucket.Get([]byte("l"))
		last_data := bucket.Get(last_hash)
		last_block := Deserialize(last_data)
		// Nakamoto
		if last_block.Height < b.Height {
			err = bucket.Put([]byte("l"), b.Hash)
			fmt.Printf("New block extends the chain\n")//////////////////////////////
		} else if last_block.Height == b.Height {
			fmt.Printf("New block branches the chain: \n")//////////////////////////////////
			if last_block.Time > b.Time {
				fmt.Printf("\tChoose new block (created earlier)\n")////////////////////////////////
				err = bucket.Put([]byte("l"), b.Hash)
			} else if last_block.Time == b.Time {
				if bytes.Compare(last_block.Hash, b.Hash) == -1 {
					fmt.Print("\tChoose new block (larger hash)\n")//////////////////////////////////////
					err = bucket.Put([]byte("l"), b.Hash)
				}
			}
		}
		// err = bucket.Put([]byte("l"), b.Hash)
		if err != nil {
			log.Panic(err)
		}
		return nil
	})
	if err != nil {
		log.Panic(err)
	}
}

func (bc *BlockChain) PrintBlockChain() string {
	iter := NewBlockChainIterator(bc)
	var string_bc []string
	for {
		cur_block := iter.Next()
		string_bc = append(string_bc, cur_block.PrintBlock())
		if cur_block.IsGenisis {
			break
		}
	}
	return strings.Join(string_bc, "\n")
}