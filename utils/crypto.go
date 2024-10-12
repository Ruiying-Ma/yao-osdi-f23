package utils

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/sha256"
	"math/big"
	"log"

	"github.com/btcsuite/golangcrypto/ripemd160"
)

func RawPK(pk []byte) ecdsa.PublicKey {
	x := big.Int{}
	y := big.Int{}
	x.SetBytes(pk[:len(pk) / 2])
	y.SetBytes(pk[len(pk) / 2:])

	return ecdsa.PublicKey{
		Curve: elliptic.P256(),
		X: &x,
		Y: &y,
	}
}

// Compute the hash of `pk`
// hash(pk) = RIPEMD160(SHA256(pk))
func HashPublicKey(pk []byte) []byte {
	sha256_pk := sha256.Sum256(pk)
	ripemd160_hasher := ripemd160.New()
	_, err := ripemd160_hasher.Write(sha256_pk[:])
	if err != nil {
		log.Panic(err)
	}

	return ripemd160_hasher.Sum(nil)
}

// Compute the address of `pk`
// Layout of address (before Base58Encode):
// version (1B) | hash_pk | checksum (4B)
// version = 0x00
// hash_pk = hash(pk)
// checksum = the first 4 bytes of SHA256(SHA256(version | pk_hash))
func PKToAdress(pk []byte) []byte {
	hash_pk := HashPublicKey(pk)
	version_hashpk := append([]byte{byte(0x00)}, hash_pk...)
	first := sha256.Sum256(version_hashpk)
	second := sha256.Sum256(first[:])
	checksum := second[0:4]
	addr := append(version_hashpk, checksum...)

	return Base58Encode(addr)
}

// Compute the hash_pk from `addr`
func AddressToHashPK(addr []byte) []byte {
	hash_pk := Base58Decode(addr)
	hash_pk = hash_pk[1:len(hash_pk) - 4]

	return hash_pk
}