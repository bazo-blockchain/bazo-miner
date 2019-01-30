package miner

import (
	"encoding/binary"
	"fmt"
	"github.com/bazo-blockchain/bazo-miner/protocol"
	"github.com/bazo-blockchain/bazo-miner/storage"
	"sort"
)

//The code here is needed if a new block is built. All open (not yet validated) transactions are first fetched
//from the mempool and then sorted. The sorting is important because if transactions are fetched from the mempool
//they're received in random order (because it's implemented as a map). However, if a user wants to issue more fundsTxs
//they need to be sorted according to increasing txCnt, this greatly increases throughput.

type openTxs []protocol.Transaction

func prepareBlock(block *protocol.Block) {
	//Fetch all txs from mempool (opentxs).
	opentxs := storage.ReadAllOpenTxs()

	//This copy is strange, but seems to be necessary to leverage the sort interface.
	//Shouldn't be too bad because no deep copy.
	var tmpCopy openTxs
	tmpCopy = opentxs

	sort.Sort(tmpCopy)

	for i, tx := range opentxs {
		/*When fetching and adding Txs from the MemPool, first check if it belongs to my shard. Only if so, then add tx to the block*/
		txAssignedShard := assignTransactionToShard(tx)

		if int(txAssignedShard) == ValidatorShardMap.ValMapping[validatorAccAddress]{
			//logger.Printf("---- Transaction (%x) in shard: %d\n", tx.Hash(),txAssignedShardAbs)
			FileConnectionsLog.WriteString(fmt.Sprintf("---- Transaction (%x) in shard: %d\n", tx.Hash(),txAssignedShard))
			//Prevent block size to overflow.
			//if block.GetSize()+tx.Size() > activeParameters.Block_size {
			//	break
			//}
			//Prevent block size to overflow. +10 Because of the bloomFilter
			if int(block.GetSize()+10)+(i*int(len(tx.Hash()))) > int(activeParameters.Block_size){
				break
			}

			switch tx.(type) {
			case *protocol.StakeTx:
				//Add StakeTXs only when preparing the last block before the next epoch block
				if (int(lastBlock.Height) == int(lastEpochBlock.Height) + int(activeParameters.epoch_length) - 1) {
					err := addTx(block, tx)
					if err != nil {
						//If the tx is invalid, we remove it completely, prevents starvation in the mempool.
						storage.DeleteOpenTx(tx)
					}
				}
			case *protocol.ContractTx, *protocol.FundsTx, *protocol.ConfigTx:
				err := addTx(block, tx)
				if err != nil {
					//If the tx is invalid, we remove it completely, prevents starvation in the mempool.
					storage.DeleteOpenTx(tx)
				}
			}

		}
	}
}

func assignTransactionToShard(transaction protocol.Transaction) (shardNr int) {
	//Convert Address/Issuer ([64]bytes) included in TX to bigInt for the modulo operation to determine the assigned shard ID.
	switch transaction.(type) {
		case *protocol.ContractTx:
			var byteToConvert [64]byte
			byteToConvert = transaction.(*protocol.ContractTx).Issuer
			var calculatedInt int
			calculatedInt = int(binary.BigEndian.Uint64(byteToConvert[:8]))
			return int((Abs(int32(calculatedInt)) % int32(NumberOfShards)) + 1)
		case *protocol.FundsTx:
			var byteToConvert [64]byte
			byteToConvert = transaction.(*protocol.FundsTx).From
			var calculatedInt int
			calculatedInt = int(binary.BigEndian.Uint64(byteToConvert[:8]))
			return int((Abs(int32(calculatedInt)) % int32(NumberOfShards)) + 1)
		case *protocol.ConfigTx:
			var byteToConvert [64]byte
			byteToConvert = transaction.(*protocol.ConfigTx).Sig
			var calculatedInt int
			calculatedInt = int(binary.BigEndian.Uint64(byteToConvert[:8]))
			return int((Abs(int32(calculatedInt)) % int32(NumberOfShards)) + 1)
		case *protocol.StakeTx:
			var byteToConvert [64]byte
			byteToConvert = transaction.(*protocol.StakeTx).Account
			var calculatedInt int
			calculatedInt = int(binary.BigEndian.Uint64(byteToConvert[:8]))
			return int((Abs(int32(calculatedInt)) % int32(NumberOfShards)) + 1)
		default:
			return 1 // default shard Nr.
		}
}

func Abs(x int32) int32 {
	if x < 0 {
		return -x
	}
	return x
}

//Implement the sort interface
func (f openTxs) Len() int {
	return len(f)
}

func (f openTxs) Swap(i, j int) {
	f[i], f[j] = f[j], f[i]
}

func (f openTxs) Less(i, j int) bool {
	//Comparison only makes sense if both tx are fundsTxs.
	//Why can we only do that with switch, and not e.g., if tx.(type) == ..?
	switch f[i].(type) {
	case *protocol.ContractTx:
		//We only want to sort a subset of all transactions, namely all fundsTxs.
		//However, to successfully do that we have to place all other txs at the beginning.
		//The order between contractTxs and configTxs doesn't matter.
		return true
	case *protocol.ConfigTx:
		return true
	case *protocol.StakeTx:
		return true
	}

	switch f[j].(type) {
	case *protocol.ContractTx:
		return false
	case *protocol.ConfigTx:
		return false
	case *protocol.StakeTx:
		return false
	}

	return f[i].(*protocol.FundsTx).TxCnt < f[j].(*protocol.FundsTx).TxCnt
}
