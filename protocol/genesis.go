package protocol

import (
	"bytes"
	"encoding/gob"
	"github.com/bazo-blockchain/bazo-miner/crypto"
)

type Genesis struct {
	RootAddress		[64]byte
	RootCommitment	[crypto.COMM_KEY_LENGTH]byte
	RootMultisig	[64]byte
}

func NewGenesis(rootAddress [64]byte,
	rootCommitment [crypto.COMM_KEY_LENGTH]byte,
	rootMultisig [64]byte) Genesis {
	return Genesis {
		rootAddress,
		rootCommitment,
		rootMultisig,
	}
}

func (genesis *Genesis) Hash() [32]byte {
	if genesis == nil {
		return [32]byte{}
	}

	input := struct {
		rootAddress		[64]byte
		rootCommitment	[crypto.COMM_KEY_LENGTH]byte
		rootMultisig	[64]byte
	} {
		genesis.RootAddress,
		genesis.RootCommitment,
		genesis.RootMultisig,
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
		RootMultisig:	genesis.RootMultisig,
	}

	buffer := new(bytes.Buffer)
	gob.NewEncoder(buffer).Encode(encoded)
	return buffer.Bytes()
}

func (*Genesis) Decode(encoded []byte) (acc *Genesis) {
	var decoded Genesis
	buffer := bytes.NewBuffer(encoded)
	decoder := gob.NewDecoder(buffer)
	decoder.Decode(&decoded)
	return &decoded
}