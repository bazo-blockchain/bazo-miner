package protocol

import (
	"bytes"
	"encoding/gob"
	"fmt"
)

/*This struct is a convenient way to broadcast the hashes of the transactions a miner has validated and included in a block.
These hashes will be consumed by the miners of the other shards, who then check the MemPool and based on the TXs, re-create the global state*/
type TransactionPayload struct {
	ShardID			int
	ContractTxData  [][32]byte
	FundsTxData  	[][32]byte
	ConfigTxData 	[][32]byte
	StakeTxData  	[][32]byte
}

func NewTransactionPayload(shardID int, contractTx [][32]byte, fundsTx [][32]byte, configTx [][32]byte, stakeTx [][32]byte) *TransactionPayload {
	newPayload := TransactionPayload{
		ShardID:				shardID,
		ContractTxData: 		contractTx,
		FundsTxData: 			fundsTx,
		ConfigTxData: 			configTx,
		StakeTxData: 			stakeTx,
	}

	return &newPayload
}

func (txPayload *TransactionPayload) HashPayload() [32]byte {
	if txPayload == nil {
		return [32]byte{}
	}

	payloadHash := struct {
		shardID					 int
		contractTxData           [][32]byte
		fundsTxData              [][32]byte
		configTxData             [][32]byte
		stakeTxData              [][32]byte
	}{
		txPayload.ShardID,
		txPayload.ContractTxData,
		txPayload.FundsTxData,
		txPayload.ConfigTxData,
		txPayload.StakeTxData,
	}
	return SerializeHashContent(payloadHash)
}

func (txPayload *TransactionPayload) GetPayloadSize() int {
	size :=
		len(txPayload.ContractTxData) + len(txPayload.FundsTxData) + len(txPayload.ConfigTxData) + len(txPayload.StakeTxData)
	return size
}

func (txPayload *TransactionPayload) EncodePayload() []byte {
	if txPayload == nil {
		return nil
	}

	encoded := TransactionPayload{
		ShardID:			   txPayload.ShardID,
		ContractTxData:        txPayload.ContractTxData,
		FundsTxData:           txPayload.FundsTxData,
		ConfigTxData:          txPayload.ConfigTxData,
		StakeTxData:           txPayload.StakeTxData,
	}

	buffer := new(bytes.Buffer)
	gob.NewEncoder(buffer).Encode(encoded)
	return buffer.Bytes()
}

func (txPayload *TransactionPayload) DecodePayload(encoded []byte) (txP *TransactionPayload) {
	if encoded == nil {
		return nil
	}

	var decoded TransactionPayload
	buffer := bytes.NewBuffer(encoded)
	decoder := gob.NewDecoder(buffer)
	decoder.Decode(&decoded)
	return &decoded
}

func (txPayload TransactionPayload) StringPayload() string {
	payloadHash := txPayload.HashPayload()

	return fmt.Sprintf("\nHash: %x\n"+
		"TX Payload Hashes: %v\n",
		payloadHash[0:8],
		txPayload.PayloadToString(),
	)
}

func (txPayload TransactionPayload) PayloadToString() (payload string) {

	if(len(txPayload.ContractTxData) == 0){
		payload += "=== NO Contract Tx ==="
	} else {
		payload += "=== Contract Tx ==="

		for _, tx := range txPayload.ContractTxData {
			payload += fmt.Sprintf("\n%x", tx[:8])
		}
	}

	if(len(txPayload.FundsTxData) == 0){
		payload += "\n=== NO Funds Tx ==="
	} else {
		payload += "\n=== Funds Tx ==="

		for _, tx := range txPayload.FundsTxData {
			payload += fmt.Sprintf("\n%x", tx[:8])
		}
	}

	if(len(txPayload.ConfigTxData) == 0){
		payload += "\n=== NO Config Tx ==="
	} else {
		payload += "\n=== Config Tx ==="

		for _, tx := range txPayload.ConfigTxData {
			payload += fmt.Sprintf("\n%x", tx[:8])
		}
	}

	if(len(txPayload.StakeTxData) == 0){
		payload += "\n=== NO Stake Tx ==="
	} else {
		payload += "\n=== Stake Tx ==="

		for _, tx := range txPayload.StakeTxData {
			payload += fmt.Sprintf("\n%x", tx[:8])
		}
	}


	return payload
}