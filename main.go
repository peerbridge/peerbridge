package main

import (
	"fmt"

	"crypto/sha256"
	"encoding/hex"
)

type Transaction struct {

}

type Block struct {
	Index        uint32
	Timestamp    string
	PrevHash     string
}

func (block Block) Hash() string {
	record := string(block.Index) + block.Timestamp + block.PrevHash
	hash := sha256.New()
	hash.Write([]byte(record))
	hashed := hash.Sum(nil)
	return hex.EncodeToString(hashed)
}

func main() {
	block  := Block{1, "1", "1"}
	fmt.Println(block.Hash())
}
