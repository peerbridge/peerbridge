package main

import (
	"fmt"
	"time"

	. "github.com/peerbridge/peerbridge/pkg/block"
)

func main() {
	block := Block{1, time.Now(), "1"}
	fmt.Println(block.Hash())
}
