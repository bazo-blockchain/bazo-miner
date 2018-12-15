package protocol

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethdb"
	"github.com/ethereum/go-ethereum/trie"
)

type MerklePatriciaTree struct {
	PatriciaRoot *NodePatricia
	merklePatriciaRoot [32] byte
	PatriciaLeafs []*Node
}

type NodePatricia struct {

}

func BuildMPT(state map[[64]byte]*Account) (*trie.Trie, error){
	Trie, _ := trie.New(common.Hash{}, trie.NewDatabase(ethdb.NewMemDatabase()))

	//loop through state of the blockchain and add nodes to the MPT
	for _, acc := range state {
		updateString(Trie,string(acc.Address[:]),string(acc.Balance))
	}

	return Trie, nil
}

func getString(trie *trie.Trie, k string) []byte {
	return trie.Get([]byte(k))
}

func updateString(trie *trie.Trie, k, v string) {
	trie.Update([]byte(k), []byte(v))
}

func deleteString(trie *trie.Trie, k string) {
	trie.Delete([]byte(k))
}

/*
This function creates a MPT Proof for a given MPT and a key
*/
func createProver(trie *trie.Trie, key []byte) (*ethdb.MemDatabase,error) {
	proof := ethdb.NewMemDatabase()
	proofEerror := trie.Prove(key, 0, proof)
	if proofEerror != nil{
		return nil, proofEerror
	}
	return proof, nil
}
