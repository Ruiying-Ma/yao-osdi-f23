package wallet

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"log"
	"encoding/gob"
	"bytes"
	"io/ioutil"

	"Project2/utils"
)

// A Wallet stores a pair of (sk, pk) and the address of the pk
// A Wallet can:
// - Generate a pair of (sk, pk) and store it to file
// - Read a specific wallet from the file

const DIR = "/osdata/osgroup10/wallet-"

type Wallet struct {
	SK []byte
	PK []byte
	Address	[]byte
}

func NewWallet(machine_id string) []byte {
	// Generate key pair
	curve := elliptic.P256()
	sk, err := ecdsa.GenerateKey(curve, rand.Reader) // `sk` here is of type *ecdsa.PrivateKey
	if err != nil {
		log.Panic(err)
	}
	pk := append(sk.PublicKey.X.Bytes(), sk.PublicKey.Y.Bytes()...)
	serialized_sk, err := x509.MarshalECPrivateKey(sk)
	if err != nil {
		log.Panic(err)
	}
	new_wallet := &Wallet{
		SK: serialized_sk,
		PK: pk,
		Address: utils.PKToAdress(pk),
	}
	// Save the wallet to disk
	var data bytes.Buffer
	encoder := gob.NewEncoder(&data)
	err = encoder.Encode(*new_wallet)
	if err != nil {
		log.Panic(err)
	}
	filename := DIR + machine_id + "-" + string(new_wallet.Address)
	err = ioutil.WriteFile(filename, data.Bytes(), 0600)
	if err != nil {
		log.Panic(err)
	}
	return new_wallet.Address
}

func ReadWallet(machine_id string, address string) *Wallet {
	filename := DIR + machine_id + "-" + address
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		log.Panic(err)
	}
	var wallet Wallet
	gob.Register(elliptic.P256())
	decoder := gob.NewDecoder(bytes.NewReader(data))
	err = decoder.Decode(&wallet)
	if err != nil {
		log.Panic(err)
	}
	return &wallet
}
