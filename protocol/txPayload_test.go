package protocol

import (
	"reflect"
	"testing"
)

func TestTxPayloadCreation(t *testing.T) {
	var contractTxData 		[][32]byte
	var fundsTxData 		[][32]byte
	var configTxData 		[][32]byte
	var stakeTxData		 	[][32]byte

	contract1 		:= [32]byte{'0', '1'}
	contract2 		:= [32]byte{'0', '2'}
	contract3 		:= [32]byte{'0', '3'}
	funds1 			:= [32]byte{'0', '4'}
	funds2 			:= [32]byte{'0', '5'}
	funds3 			:= [32]byte{'0', '6'}
	config1 		:= [32]byte{'0', '7'}
	config2 		:= [32]byte{'0', '8'}
	config3 		:= [32]byte{'0', '9'}
	stake1 			:= [32]byte{'1', '0'}
	stake2 			:= [32]byte{'1', '1'}
	stake3 			:= [32]byte{'1', '2'}


	contractTxData = append(contractTxData,contract1,contract2, contract3)
	fundsTxData = append(fundsTxData,funds1,funds2,funds3)
	configTxData = append(configTxData,config1,config2,config3)
	stakeTxData = append(stakeTxData,stake1,stake2,stake3)

	createdPayload := NewTransactionPayload(3, 5,contractTxData,fundsTxData,configTxData,stakeTxData)

	if !reflect.DeepEqual(createdPayload.ContractTxData, contractTxData) {
		t.Errorf("ContractTxData hash does not match the given one: %x vs. %x", createdPayload.ContractTxData, contractTxData)
	}

	if !reflect.DeepEqual(createdPayload.FundsTxData, fundsTxData) {
		t.Errorf("FundsTxData hash does not match the given one: %x vs. %x", createdPayload.FundsTxData, fundsTxData)
	}

	if !reflect.DeepEqual(createdPayload.ConfigTxData, configTxData) {
		t.Errorf("ConfigTxData hash does not match the given one: %x vs. %x", createdPayload.ConfigTxData, configTxData)
	}

	if !reflect.DeepEqual(createdPayload.StakeTxData, stakeTxData) {
		t.Errorf("StakeTxData hash does not match the given one: %x vs. %x", createdPayload.StakeTxData, stakeTxData)
	}
}

func TestTransactionPayloadSerialization(t *testing.T) {
	var contractTxData 		[][32]byte
	var fundsTxData 		[][32]byte
	var configTxData 		[][32]byte
	var stakeTxData		 	[][32]byte

	contract1 		:= [32]byte{'0', '1'}
	contract2 		:= [32]byte{'0', '2'}
	contract3 		:= [32]byte{'0', '3'}
	funds1 			:= [32]byte{'0', '4'}
	funds2 			:= [32]byte{'0', '5'}
	funds3 			:= [32]byte{'0', '6'}
	config1 		:= [32]byte{'0', '7'}
	config2 		:= [32]byte{'0', '8'}
	config3 		:= [32]byte{'0', '9'}
	stake1 			:= [32]byte{'1', '0'}
	stake2 			:= [32]byte{'1', '1'}
	stake3 			:= [32]byte{'1', '2'}


	contractTxData = append(contractTxData,contract1,contract2, contract3)
	fundsTxData = append(fundsTxData,funds1,funds2,funds3)
	configTxData = append(configTxData,config1,config2,config3)
	stakeTxData = append(stakeTxData,stake1,stake2,stake3)

	createdPayload := NewTransactionPayload(3, 5,contractTxData,fundsTxData,configTxData,stakeTxData)


	var compareTransactionPayload *TransactionPayload
	encodedPayload := createdPayload.EncodePayload()
	compareTransactionPayload = compareTransactionPayload.DecodePayload(encodedPayload)

	if !reflect.DeepEqual(createdPayload, compareTransactionPayload) {
		t.Error("Block encoding/decoding failed!")
	}
}
