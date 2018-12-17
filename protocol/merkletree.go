package protocol

import (
	"bytes"
	"errors"
	"fmt"
	"golang.org/x/crypto/sha3"
	"math"
)

//MerkleTree is the container for the tree. It holds a pointer to the root of the tree,
//a list of pointers to the leaf nodes, and the merkle root.
type MerkleTree struct {
	Root       *Node
	merkleRoot [32]byte
	Leafs      []*Node
}

//Node represents a node, root, or leaf in the tree. It stores pointers to its immediate
//relationships, a hash, the content stored if it is a leaf, and other metadata.
type Node struct {
	Parent *Node
	Left   *Node
	Right  *Node
	leaf   bool
	dup    bool
	Hash   [32]byte
}

//NewMerkleTree creates a new Merkle Tree using the content cs.
func NewMerkleTree(txSlices HashArray) (*MerkleTree, error) {
	root, leafs, err := buildWithContent(txSlices)
	if err != nil {
		return nil, err
	}
	t := &MerkleTree{
		Root:       root,
		merkleRoot: root.Hash,
		Leafs:      leafs,
	}
	return t, nil
}

//verifyNode walks down the tree until hitting a leaf, calculating the hash at each level
//and returning the resulting hash of Node n.
func (n *Node) verifyNode() [32]byte {
	if n.leaf {
		return n.Hash
	}
	leftHash := n.Left.verifyNode()
	rightHash := n.Right.verifyNode()
	concatHash := append(leftHash[:], rightHash[:]...)
	return MTHash(concatHash)
}

//buildWithContent is a helper function that for a given set of Contents, generates a
//corresponding tree and returns the root node, a list of leaf nodes, and a possible error.
//Returns an error if cs contains no Contents.
func buildWithContent(txSlices HashArray) (*Node, []*Node, error) {
	if len(txSlices) == 0 {
		return nil, nil, errors.New("Error: cannot construct tree with no content.")
	}
	var leafs []*Node
	for _, tx := range txSlices {
		if tx != [32]byte{} {
			leafs = append(leafs, &Node{
				Hash: tx,
				leaf: true,
			})
		}
	}
	if len(leafs)%2 == 1 {
		leafs = append(leafs, leafs[len(leafs)-1])
		leafs[len(leafs)-1].dup = true
	}
	root := buildIntermediate(leafs)
	return root, leafs, nil
}

//buildIntermediate is a helper function that for a given list of leaf nodes, constructs
//the intermediate and root levels of the tree. Returns the resulting root node of the tree.
func buildIntermediate(nl []*Node) *Node {
	var nodes []*Node
	l := len(nl)
	for !checkPower2(l) {
		l--
	}
	for i := 0; i < l; i += 2 {
		concatHash := append(nl[i].Hash[:], nl[i+1].Hash[:]...)
		n := &Node{
			Left:  nl[i],
			Right: nl[i+1],
			Hash:  MTHash(concatHash),
		}
		nodes = append(nodes, n)
		nl[i].Parent = n
		nl[i+1].Parent = n
		if len(nl) == 2 {
			return n
		}
	}
	if l < len(nl) {
		for i := l; i < len(nl); i++ {
			nodes = append(nodes, nl[i])
		}
	}
	return buildIntermediate(nodes)
}

func checkPower2(i int) bool {
	l := math.Log2(float64(i))
	if l == math.Trunc(l) {
		return true
	} else {
		return false
	}
}

//MerkleRoot returns the unverified Merkle Root (hash of the root node) of the tree.
func (m *MerkleTree) MerkleRoot() [32]byte {
	if m != nil {
		return m.merkleRoot
	} else {
		return [32]byte{}
	}
}

//VerifyTree verify tree validates the hashes at each level of the tree and returns true if the
//resulting hash at the root of the tree matches the resulting root hash; returns false otherwise.
func (m *MerkleTree) VerifyTree() bool {
	calculatedMerkleRoot := m.Root.verifyNode()
	if bytes.Compare(m.merkleRoot[:], calculatedMerkleRoot[:]) == 0 {
		return true
	}
	return false
}

func (m *MerkleTree) GetLeaf(leafHash [32]byte) *Node {
	for _, leaf := range m.Leafs {
		if leafHash == leaf.Hash {
			return leaf
		}
	}
	return nil
}

func MTHash(data []byte) [32]byte {
	return sha3.Sum256(data)
}

func (m *MerkleTree) MerkleProof(leafHash [32]byte) (hashes [][33]byte, err error) {
	leaf := m.GetLeaf(leafHash)
	if leaf == nil {
		return nil, errors.New(fmt.Sprintf("could not find leaf for hash %x", leafHash[0:8]))
	}
	currentNode := leaf
	currentParent := leaf.Parent
	for currentParent != nil {
		left := currentParent.Left
		right := currentParent.Right
		var hash [33]byte
		if currentNode.Hash == left.Hash {
			copy(hash[0:1], []byte{'r'})
			copy(hash[1:33], right.Hash[:])
			hashes = append(hashes, hash)
		} else if currentNode.Hash == right.Hash {
			copy(hash[0:1], []byte{'l'})
			copy(hash[1:33], left.Hash[:])
			hashes = append(hashes, hash)
		} else {
			return nil, errors.New(fmt.Sprintf("could not find intermediate node to verify %x\n", leaf.Hash))
		}
		currentNode = currentParent
		currentParent = currentParent.Parent
	}
	return hashes, nil
}

func (m *MerkleTree) VerifyMerkleProof(leafHash [32]byte, hashes [][33]byte) bool {
	computedRootHash := leafHash
	for i := 0; i < len(hashes); i++ {
		var hash [32]byte
		var leftOrRight [1]byte
		copy(leftOrRight[:], hashes[i][0:1])
		copy(hash[:], hashes[i][1:33])

		if leftOrRight == [1]byte{'l'} {
			computedRootHash = MTHash(append(hash[:], computedRootHash[:]...))
		} else {
			computedRootHash = MTHash(append(computedRootHash[:], hash[:]...))
		}
	}

	return computedRootHash == m.MerkleRoot()
}

//VerifyContent indicates whether a given content is in the tree and the hashes are valid for that content.
//Returns true if the expected Merkle Root is equivalent to the Merkle root calculated on the critical path
//for a given content. Returns true if valid and false otherwise.
func GetIntermediate(leaf *Node) (intermediate []*Node, err error) {
	currentNode := leaf
	currentParent := leaf.Parent
	for currentParent != nil {
		left := currentParent.Left
		right := currentParent.Right
		if currentNode.Hash == left.Hash {
			intermediate = append(intermediate, right)
			intermediate = append(intermediate, currentParent)
		} else if currentNode.Hash == right.Hash {
			intermediate = append(intermediate, left)
			intermediate = append(intermediate, currentParent)
		} else {
			return nil, errors.New(fmt.Sprintf("Could not find helper nodes to verify %x\n", leaf.Hash))
		}
		currentNode = currentParent
		currentParent = currentParent.Parent
	}
	return intermediate, nil
}

//String returns a string representation of the tree. Only leaf nodes are included
//in the output.
func (m *MerkleTree) String() string {
	s := ""
	for _, l := range m.Leafs {
		s += fmt.Sprint(l)
		s += "\n"
	}
	return s
}
