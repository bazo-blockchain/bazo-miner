package protocol

import (
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethdb"
	"github.com/ethereum/go-ethereum/trie"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
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

func TestGetValuesMPT(t *testing.T){
	Trie, _ := trie.New(common.Hash{}, trie.NewDatabase(ethdb.NewMemDatabase()))

	updateString(Trie,"a1111111", "45")

	testVal := Trie.Get([]byte("a1111111"))

	fmt.Printf("First insert: %v",string(testVal))
	fmt.Printf("\n")

	if(string(testVal) != "45"){
		t.Errorf("Retrieved value does not match with inserted value for key: %v", "a1111111")
	}

	updateString(Trie,"a1111111", "90")

	testVal2 := Trie.Get([]byte("a1111111"))
	fmt.Printf("Second insert: %v",string(testVal2))
	fmt.Printf("\n")

	if(string(testVal2) != string([]byte("90"))){
		t.Errorf("Retrieved value does not match with inserted value for key: %v", "a1111111")
	}

	assert.Equal(t,testVal2,getString(Trie,"a1111111"),"The two values should match")
}

/*
This test asserts that the root hash of the MPT does not change due to insertion order of the nodes.
This is in contrast to regular Merkle Trees where insertion order of the leafs is crucial for getting the same root hash
*/
func TestStaleRootInsertOrder(t *testing.T)  {
	Trie, _ := initTrie()

	updateString(Trie, "address111", "testValForAddress111")
	updateString(Trie, "address222", "testValForAddress222")
	updateString(Trie, "address333", "testValForAddress333")
	updateString(Trie, "address444", "testValForAddress444")

	rootHashOne := Trie.Hash()

	deleteString(Trie,"address111")
	deleteString(Trie,"address222")
	deleteString(Trie,"address333")
	deleteString(Trie,"address444")

	updateString(Trie, "address333", "testValForAddress333")
	updateString(Trie, "address444", "testValForAddress444")
	updateString(Trie, "address111", "testValForAddress111")
	updateString(Trie, "address222", "testValForAddress222")

	rootHashTwo := Trie.Hash()

	assert.Equal(t, rootHashOne,rootHashTwo, "The two root hashes should match")
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

func initTrie() (*trie.Trie, error)  {
	Trie, err := trie.New(common.Hash{}, trie.NewDatabase(ethdb.NewMemDatabase()))

	if err != nil {
		return nil, errors.New("error while initializing new trie")
	}

	return Trie, nil
}
