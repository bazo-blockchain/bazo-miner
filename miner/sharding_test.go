package miner

import (
	"fmt"
	"github.com/bazo-blockchain/bazo-miner/crypto"
	"github.com/bazo-blockchain/bazo-miner/p2p"
	"github.com/bazo-blockchain/bazo-miner/protocol"
	"github.com/bazo-blockchain/bazo-miner/storage"
	"os"
	"testing"
)

const(
	NodesDirectory 		= "nodes/"
)

var (
	NodeNames			[]string
	TotalNodes			int
)

func TestShardingWith20Nodes(t *testing.T) {
	TotalNodes = 20 //Set number of nodes for this test

	//Generate wallet directories for all nodes, i.e., validators and non-validators
	for i := 1; i <= TotalNodes; i++ {
		strNode := fmt.Sprintf("Node_%d",i)
		if(!stringAlreadyInSlice(NodeNames,strNode)){
			NodeNames = append(NodeNames,strNode)
		}
		if _, err := os.Stat(NodesDirectory+strNode); os.IsNotExist(err) {
			err = os.MkdirAll(NodesDirectory+strNode, 0755)
			if err != nil {
				t.Errorf("Error while creating node directory %v\n",err)
			}
		}
		storage.Init(NodesDirectory+strNode+"/storage.db", TestIpPort)
		_, err := crypto.ExtractECDSAPublicKeyFromFile(NodesDirectory+strNode+"/wallet.key")
		if err != nil {
			return
		}
		_, err = crypto.ExtractRSAKeyFromFile(NodesDirectory+strNode+"/commitment.key")
		if err != nil {
			return
		}
	}

	storage.State = nil
	storage.State = make(map[[64]byte]*protocol.Account)
	//Start the miners
	validatorPubKey1,_ := crypto.ExtractECDSAPublicKeyFromFile(NodesDirectory+"Node_1/wallet.key")
	commPrivKey1,_ := crypto.ExtractRSAKeyFromFile(NodesDirectory+"Node_1/commitment.key")
	go InitFirstStart(validatorPubKey1,commPrivKey1)

	validatorPubKey2,_ := crypto.ExtractECDSAPublicKeyFromFile(NodesDirectory+"Node_2/wallet.key")
	commPrivKey2,_ := crypto.ExtractRSAKeyFromFile(NodesDirectory+"Node_2/commitment.key")
	p2p.Init(TestIpPortNodeB)
	Init(validatorPubKey2,commPrivKey2)
	//
	//validatorPubKey3,_ := crypto.ExtractECDSAPublicKeyFromFile(NodesDirectory+"Node3/wallet.key")
	//commPrivKey3,_ := crypto.ExtractRSAKeyFromFile(NodesDirectory+"Node3/commitment.key")

	t.Log("Done...")
}


func stringAlreadyInSlice(inputSlice []string, str string) bool {
	for _, entry := range inputSlice {
		if entry == str {
			return true
		}
	}
	return false
}