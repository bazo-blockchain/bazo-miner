package protocol

import (
	"bytes"
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
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
	Trie, _ := initTrie()

	updateString(Trie,"11111", "45")

	updateString(Trie,"11222", "100")

	updateString(Trie,"22222", "400")

	updateString(Trie,"2211111", "350")
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

func TestProofMPT(t *testing.T) {
	Trie, _ := initTrie()
	m := make(map[string]string)

	updateString(Trie,"key1","1")
	m["key1"] = "1"
	updateString(Trie,"key2","11")
	m["key2"] = "11"
	updateString(Trie,"olaf3","111")
	m["olaf3"] = "111"
	updateString(Trie,"olaf4","1111")
	m["olaf4"] = "1111"

	root := Trie.Hash()


	for k,v := range m {
		proof, err := CreateProver(Trie,[]byte(k))
		if proof == nil {
			t.Fatalf("prover: missing key %x while constructing proof", k)
		}

		val, _, err := trie.VerifyProof(root, []byte(k), proof)

		if err != nil {
			t.Fatalf("prover: failed to verify proof for key %x: %v\nraw ", k,v)
		}
		if !bytes.Equal(val, []byte(v)) {
			t.Fatalf("prover: verified value mismatch for key %x: have %x, want %x", k, val, v)
		}
	}
}

func TestProofMPTFailValueMismatch(t *testing.T) {
	Trie, _ := initTrie()
	m := make(map[string]string)

	updateString(Trie,"key1","1")
	m["key1"] = "1"
	updateString(Trie,"key2","11")
	m["key2"] = "11"

	root := Trie.Hash()

	// test value mismatch for key2
	proof, err := CreateProver(Trie,[]byte("key2"))
	if proof == nil {
		t.Fatalf("prover: missing key %x while constructing proof", "key2")
	}

	updateString(Trie,"key2","newValue")

	val, _, err := trie.VerifyProof(root, []byte("key2"), proof)

	if err != nil {
		t.Fatalf("prover: failed to verify proof for key %x: %v\nraw ", "key2","11")
	}

	assert.NotEqual(t, val, []byte("newValue"))
}

func TestMPTProofMissingKey(t *testing.T)  {
	Trie, _ := initTrie()

	updateString(Trie,"key1","1")
	updateString(Trie,"key2","11")

	// test missing key
	//proof := createProver(Trie,[]byte("key3"))

	for i, key := range []string{"a", "j", "l", "z"} {
		proof := ethdb.NewMemDatabase()
		Trie.Prove([]byte(key), 0, proof)

		if proof.Len() != 1 {
			t.Errorf("test %d: proof should have one element", i)
		}
		val, _, err := trie.VerifyProof(Trie.Hash(), []byte(key), proof)
		if err != nil {
			t.Fatalf("test %d: failed to verify proof: %v\nraw proof: %x", i, err, proof)
		}
		if val != nil {
			t.Fatalf("test %d: verified value mismatch: have %x, want nil", i, val)
		}

		//expect that the retrieved value for the missing key is nil
		assert.Equal(t, []uint8([]byte(nil)),val)
	}
}

func makeProvers(trie *trie.Trie) []func(key []byte) *ethdb.MemDatabase {
	var provers []func(key []byte) *ethdb.MemDatabase

	// Create a direct trie based Merkle prover
	provers = append(provers, func(key []byte) *ethdb.MemDatabase {
		proof := ethdb.NewMemDatabase()
		trie.Prove(key, 0, proof)
		return proof
	})
	// Create a leaf iterator based Merkle prover
	provers = append(provers, func(key []byte) *ethdb.MemDatabase {
		proof := ethdb.NewMemDatabase()
		if it := trie.NodeIterator(key); it.Next(true) && bytes.Equal(key, it.LeafKey()) {
			for _, p := range it.LeafProof() {
				proof.Put(crypto.Keccak256(p), p)
			}
		}
		return proof
	})
	return provers
}
