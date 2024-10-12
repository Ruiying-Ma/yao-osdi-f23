package blockchain

import (
	"bytes"
	"encoding/gob"
	"log"
	"crypto/ecdsa"
	"crypto/rand"
	"crypto/sha256"
	"crypto/x509"
	"math/big"
	"encoding/hex"
	"fmt"
	"strings"

	"Project2/utils"
	"Project2/wallet"
)

// A Transaction stores:
// - An address of a wallet: the initiator of this tx
// - A set of In that represents the incomes of this tx
// - A set of Out that represents the payments of this tx
// - A signature by the initiator of this tx
// - An abstract (hash) of that tx
// - Whether this tx is a reward
// A Transaction can:
// - Sign: sign the tx
// - Hash: hash the tx after signing
// - Verify legal tx:
//		1. Whether the tx's incomes are valid (unspent, belongs to initiator) (no need for reward)
//		2. Whether the tx's payments are valid (payments <= incomes)
//		3. Whether the tx's signature is valid
//		4. Whether the tx's hash is valid

const REWARD = 100 // the reward tx

type In struct {
	HashTx	[]byte
	Idx	int
	Amount	int
}

type Out struct {
	Amount	int
	Recipient	[]byte // wallet address
}

type Transaction struct {
	Initiator	[]byte // pk
	Incomes		[]In 
	Payments	[]Out 
	IsReward 	bool
	Hash		[]byte
	Signature	[]byte
}

// `w`: Initiator's wallet
// `r`: Recipient's address
// `a`: amount
// `is_reward`: whether this tx is a reward
func NewTransaction(w *wallet.Wallet, r []byte, a int, is_reward bool, bc *BlockChain) *Transaction {
	sk, err := x509.ParseECPrivateKey(w.SK)
	if err != nil {
		log.Panic(err)
	}
	if is_reward {
		tx := &Transaction{
			Initiator: w.PK,
			Incomes: []In{},
			Payments: []Out{Out{
				Amount: REWARD,
				Recipient: w.Address,
			}},
			IsReward: true,
			Hash: []byte{},
			Signature: []byte{},
		}
		tx.Sign(*sk)
		tx.HashTx()
		return tx
	}
	// Accummulate incomes
	acc, acc_payments := acc_incomes(w.PK, a, bc)
	if acc < a {
		log.Panic(fmt.Sprintf("ERROR: %s cannot pay %d money: not enough money", string(w.Address), a))
	}
	tx := &Transaction {
		Initiator: w.PK,
		Incomes: acc_payments,
		Payments: []Out{},
		IsReward: false,
		Hash: []byte{},
		Signature: []byte{},
	}
	tx.Payments = append(tx.Payments, Out{
		Amount: a,
		Recipient: r,
	})
	if a < acc {
		tx.Payments = append(tx.Payments, Out{
			Amount: acc - a,
			Recipient: w.Address,
		})
	}
	tx.Sign(*sk)
	tx.HashTx()
	return tx
}

func (tx *Transaction) HashTx() {
	if len(tx.Hash) != 0 {
		log.Panic("Tx already hashed")
	}
	if len(tx.Signature) == 0 {
		log.Panic("Tx haven't been signed")
	}
	hash := sha256.Sum256(tx.serialize())
	tx.Hash = hash[:]
}

func (tx *Transaction) Sign(sk ecdsa.PrivateKey) {
	if len(tx.Signature) != 0 {
		log.Panic("Tx already signed")
	}
	if len(tx.Hash) != 0 {
		log.Panic("Tx hashed before signed")
	}
	r, s, err := ecdsa.Sign(rand.Reader, &sk, tx.serialize())
	if err != nil {
		log.Panic(err)
	}
	tx.Signature = append(r.Bytes(), s.Bytes()...)
}

func (tx *Transaction) Verify(bc *BlockChain, prev_hash []byte) bool {
	return tx.verify_incomes(bc, prev_hash) && tx.verify_payments() && tx.verify_hash() && tx.verify_signature()
}

func (tx *Transaction) PrintTx() string {
	var string_tx []string
	string_tx = append(string_tx, fmt.Sprintf("\t--- Transaction Hash: %x", tx.Hash))
	string_tx = append(string_tx, fmt.Sprintf("\t\tInitiator: %x", tx.Initiator))
	string_tx = append(string_tx, fmt.Sprintf("\t\tSignature: %x", tx.Signature))
	if tx.IsReward {
		string_tx = append(string_tx, fmt.Sprintf("\t\tIsReward: True"))
	} else {
		string_tx = append(string_tx, fmt.Sprintf("\t\tIsReward: False"))
		for iid, in := range tx.Incomes {
			string_tx = append(string_tx, fmt.Sprintf("\t\t\tIncome: %d", iid))
			string_tx = append(string_tx, fmt.Sprintf("\t\t\t\tIncome HashTx: %x", in.HashTx))
			string_tx = append(string_tx, fmt.Sprintf("\t\t\t\tIncome Idx: %d", in.Idx))
			string_tx = append(string_tx, fmt.Sprintf("\t\t\t\tIncome Amount: %d", in.Amount))
		}
	}
	for oid, out := range tx.Payments {
		string_tx = append(string_tx, fmt.Sprintf("\t\t\tPayment: %d", oid))
		string_tx = append(string_tx, fmt.Sprintf("\t\t\t\tPayment Recipient: %x", out.Recipient))
		string_tx = append(string_tx, fmt.Sprintf("\t\t\t\tPayment Amount: %d", out.Amount))
	}
	return strings.Join(string_tx, "\n")
}


// Accumulate `a` incomes for initiator `i` (pk) 
// return accumulations
// return a set of incomes
func acc_incomes(i []byte, a int, bc *BlockChain) (int, []In) {
	acc := 0
	acc_payments := []In{}
	used_payments := make(map[string][]int)
	initiator_addr := utils.PKToAdress(i)
	iter := NewBlockChainIterator(bc)
	for {
		cur_block := iter.Next()
		for _, tx := range cur_block.Txs {
			key := hex.EncodeToString(tx.Hash)
			tx_used_payments, _ := used_payments[key]
			for oid, out := range tx.Payments {
				if bytes.Compare(out.Recipient, initiator_addr) == 0 {
					used := false
					if tx_used_payments != nil {
						for _, id := range tx_used_payments {
							if id == oid {
								used = true
								break
							}
						}
					}
					if used == false {
						acc += out.Amount
						acc_payments = append(acc_payments, In{
							HashTx: tx.Hash,
							Idx: oid,
							Amount: out.Amount,
						})
						if acc >= a {
							return acc, acc_payments
						}
					}
				}
				
			}

			if bytes.Compare(tx.Initiator, i) == 0 && !tx.IsReward {
				for _, in := range tx.Incomes {
					paid_tx_key := hex.EncodeToString(in.HashTx)
					used_payments[paid_tx_key] = append(used_payments[paid_tx_key], in.Idx)
				}
			}
		}
		if cur_block.IsGenisis {
			break
		}
	}
	return acc, acc_payments
}

func (tx *Transaction) verify_incomes(bc *BlockChain, prev_hash []byte) bool {
	if tx.IsReward {
		return true
	}
	for _, in := range tx.Incomes {
		if is_existed(in.HashTx, in.Idx, bc, prev_hash) == false || is_used(in.HashTx, in.Idx, bc, prev_hash) == true {
			return false
		}
	}
	return true
}

func (tx *Transaction) verify_payments() bool {
	in_amount := 0
	if tx.IsReward {
		in_amount = REWARD
	} else {
		for _, in := range tx.Incomes {
			in_amount += in.Amount
		}
	}
	out_amount := 0
	for _, out := range tx.Payments {
		out_amount += out.Amount
	}
	if out_amount > in_amount {
		fmt.Printf("verify_payments: out_amount > in_amount\n")
		return false
	}
	return true
}

func (tx *Transaction) verify_hash() bool {
	tmp_tx := *tx
	tmp_tx.Hash = []byte{}
	//tmp_tx.Signature = []byte{}
	//fmt.Printf("verify the hash of tx %s\n", tmp_tx.PrintTx())///////////////////////////////////////////////////////////
	hash := sha256.Sum256(tmp_tx.serialize())
	if bytes.Compare(hash[:], tx.Hash) != 0 {
		fmt.Printf("verify_hash: wrong tx hash\n")
		return false
	}
	return true
}

func (tx *Transaction) verify_signature() bool {
	if tx.IsReward {
		return true
	}
	tmp_tx := *tx
	tmp_tx.Signature = []byte{}
	tmp_tx.Hash = []byte{}
	
	r := big.Int{}
	s := big.Int{}
	sig_len := len(tx.Signature)
	r.SetBytes(tx.Signature[:sig_len/2])
	s.SetBytes(tx.Signature[sig_len/2:])

	raw_pk := utils.RawPK(tx.Initiator)

	if !ecdsa.Verify(&raw_pk, tmp_tx.serialize(), &r, &s) {
		fmt.Printf("verify_signature: wrong tx signature\n")
		return false
	}
	return true
}

// Check whether the `oid`-th payment of `hash_tx` tx exists
func is_existed(hash_tx []byte, oid int, bc *BlockChain, prev_hash []byte) bool {
	iter := NewBlockChainIterator(bc)
	start := false
	if len(prev_hash) == 0 {
		start = true
	}
	for {
		cur_block := iter.Next()
		if bytes.Compare(cur_block.Hash, prev_hash) == 0 {
			start = true
		}
		if start == true {
			tx := cur_block.find_tx(hash_tx)
			if tx != nil {
				return tx.has_payment(oid)
			}
			if cur_block.IsGenisis {
				break
			}
		}
	}
	fmt.Printf("is_existed: the tx doesn't exist\n")
	return false;
}

// Check whether the `oid`-th payment of `hash_tx` tx exists and has been used
func is_used(hash_tx []byte, oid int, bc *BlockChain, prev_hash []byte) bool {
	iter := NewBlockChainIterator(bc)
	start := false
	if len(prev_hash) == 0 {
		start = true
	}
	for {
		cur_block := iter.Next()
		if bytes.Compare(cur_block.Hash, prev_hash) == 0 {
			start = true
		}
		if start == true {
			for _, tx := range cur_block.Txs {
				if !tx.IsReward {
					for in_id, in := range tx.Incomes {
						if bytes.Compare(in.HashTx, hash_tx) == 0 && in.Idx == oid {
							fmt.Printf("is_used: the %d-th payment of the tx has been used\n", in_id)
							return true
						}
					}
				}
			}
			if cur_block.IsGenisis {
				break
			}
		}
		
	}
	return false
}

func (tx *Transaction) has_payment(oid int) bool {
	if oid < 0 || oid >= len(tx.Payments) {
		fmt.Printf("has_payment: the payment doesn't exist\n")
		return false
	}
	return true
}


func (tx *Transaction) serialize() []byte {
	var data bytes.Buffer
	encoder := gob.NewEncoder(&data)
	err := encoder.Encode(*tx)
	if err != nil {
		log.Panic(err)
	}
	return data.Bytes()
}

func deserialize(data []byte) *Transaction {
	var tx Transaction
	decoder := gob.NewDecoder(bytes.NewReader(data))
	err := decoder.Decode(&tx)
	if err != nil {
		log.Panic(err)
	}
	return &tx
}