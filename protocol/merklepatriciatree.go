package protocol

import (
	"bytes"
	"encoding/gob"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethdb"
	"github.com/ethereum/go-ethereum/trie"
)

type MPT_Proof struct {
	Proofs map[string][]byte
}

func (proof *MPT_Proof) Hash() (hash [32]byte) {
	if proof == nil {
		return [32]byte{}
	}

	return SerializeHashContent(proof)
}

func (proof *MPT_Proof) Encode() (encodedTx []byte) {
	encodeData := MPT_Proof{}
	buffer := new(bytes.Buffer)
	gob.NewEncoder(buffer).Encode(encodeData)
	return buffer.Bytes()
}

func (proof *MPT_Proof) Decode(encoded []byte) *MPT_Proof {
	var decoded MPT_Proof
	buffer := bytes.NewBuffer(encoded)
	decoder := gob.NewDecoder(buffer)
	decoder.Decode(&decoded)
	return &decoded
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
func CreateProver(trie *trie.Trie, key []byte) (*ethdb.MemDatabase,error) {
	proof := ethdb.NewMemDatabase()
	proofEerror := trie.Prove(key, 0, proof)
	if proofEerror != nil{
		return nil, proofEerror
	}
	return proof, nil
}