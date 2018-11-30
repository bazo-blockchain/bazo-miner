package protocol

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"github.com/bazo-blockchain/bazo-miner/crypto"
)

type Genesis struct {
	RootAddress		[64]byte
	RootCommitment	[crypto.COMM_KEY_LENGTH]byte
}

func NewGenesis(rootAddress [64]byte,
	rootCommitment [crypto.COMM_KEY_LENGTH]byte) Genesis {
	return Genesis {
		rootAddress,
		rootCommitment,
	}
}

func (genesis *Genesis) Hash() [32]byte {
	if genesis == nil {
		return [32]byte{}
	}

	input := struct {
		rootAddress		[64]byte
		rootCommitment	[crypto.COMM_KEY_LENGTH]byte
	} {
		genesis.RootAddress,
		genesis.RootCommitment,
	}

	return SerializeHashContent(input)
}

func (genesis *Genesis) Encode() []byte {
	if genesis == nil {
		return nil
	}

	encoded := Genesis{
		RootAddress:    genesis.RootAddress,
		RootCommitment:	genesis.RootCommitment,
	}

	buffer := new(bytes.Buffer)
	gob.NewEncoder(buffer).Encode(encoded)
	return buffer.Bytes()
}

func (*Genesis) Decode(encoded []byte) (acc *Genesis) {
	if encoded == nil {
		return nil
	}

	var decoded Genesis
	buffer := bytes.NewBuffer(encoded)
	decoder := gob.NewDecoder(buffer)
	decoder.Decode(&decoded)
	return &decoded
}

func (genesis *Genesis) String() string {
	return fmt.Sprintf(
		"\n"+
			"RootAddress: %x\n"+
			"RootCommitment: %x\n",
		genesis.RootAddress[0:8],
		genesis.RootCommitment[0:8],
	)
}