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

func MPT(){
	Trie, _ := trie.New(common.Hash{}, trie.NewDatabase(ethdb.NewMemDatabase()))
	println(Trie)
}

