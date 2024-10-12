package utils

import (
	"bytes"
	"encoding/binary"
	"log"
)

// Copied from https://github.com/Jeiwan/blockchain_go/blob/master/utils.go#L10

// IntToHex converts an int64 to a byte array
func IntToHex(num int64) []byte {
	buff := new(bytes.Buffer)
	err := binary.Write(buff, binary.BigEndian, num)
	if err != nil {
		log.Panic(err)
	}

	return buff.Bytes()
}

func BoolToHex(b bool) []byte {
	if b {
		return IntToHex(1)
	} else {
		return IntToHex(0)
	}
}