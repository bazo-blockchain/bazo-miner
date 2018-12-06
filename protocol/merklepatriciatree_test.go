package protocol

import (
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethdb"
	"github.com/ethereum/go-ethereum/trie"
	"testing"
)

const (
	TestDBFileName   = "test.db"
	TestIpPort       = "127.0.0.1:8000"
	TestKeyFileName  = "test_root"
)

func TestMPT(t *testing.T){
	//storage.DeleteAll()
	fmt.Printf("Starting testrun...")
}

func TestEthereumMPTInsertNodes(t *testing.T){
	Trie, _ := trie.New(common.Hash{}, trie.NewDatabase(ethdb.NewMemDatabase()))
	key1 := []byte("11111")
	value1 := []byte("45")
	Trie.Update(key1,value1)

	key2 := []byte("11222")
	value2 := []byte("100")
	Trie.Update(key2,value2)

	key3 := []byte("22222")
	value3 := []byte("400")
	Trie.Update(key3,value3)

	key4 := []byte("2211111")
	value4 := []byte("350")
	Trie.Update(key4,value4)

	println(Trie)
}

func TestIncludeStateInMPT(t *testing.T){
	db := ethdb.NewMemDatabase()
	db.Put([]byte("key1"),[]byte("val1"))

	fmt.Printf("\n")
}

func TestGetValuesMPT(t *testing.T){
	Trie, _ := trie.New(common.Hash{}, trie.NewDatabase(ethdb.NewMemDatabase()))

	key1 := []byte("a1111111")
	value1 := []byte("45")
	Trie.Update(key1,value1)

	testVal := Trie.Get(key1)

	fmt.Printf("First insert: %v",string(testVal))
	fmt.Printf("\n")

	Trie.Update(key1,[]byte("90"))

	testVal = Trie.Get(key1)
	fmt.Printf("Second insert: %v",string(testVal))
	fmt.Printf("\n")

}

func keybytesToHex(str []byte) []byte {
	l := len(str)*2 + 1
	var nibbles = make([]byte, l)
	for i, b := range str {
		nibbles[i*2] = b / 16
		nibbles[i*2+1] = b % 16
	}
	nibbles[l-1] = 16
	return nibbles
}
