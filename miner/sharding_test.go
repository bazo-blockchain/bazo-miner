package miner

import (
	"errors"
	"fmt"
	"github.com/bazo-blockchain/bazo-miner/crypto"
	"github.com/bazo-blockchain/bazo-miner/p2p"
	"github.com/bazo-blockchain/bazo-miner/protocol"
	"github.com/bazo-blockchain/bazo-miner/storage"
	"os"
	"testing"
	"time"
)

const(
	NodesDirectory 		= "nodes/"
)

var (
	NodeNames			[]string
	TotalNodes			int
)

func TestShardingWith20Nodes(t *testing.T) {
	/**
	Set Total Number of desired nodes. They will be generated automatically. And for each node, a separate go routine is being created.
	This enables parallel issuance of transactions
	 */
	TotalNodes = 20

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

	transferFundsToWallets()

	//Create a goroutine for each wallet and send TX from corresponding wallet to root account
	for i := 1; i <= TotalNodes; i++ {
		strNode := fmt.Sprintf("Node_%d",i)
		go func() {
			txCount := 0
			for{
				fromPrivKey, err := crypto.ExtractECDSAKeyFromFile(NodesDirectory+strNode+"/wallet.key")
				if err != nil {
					return
				}

				toPubKey, err := crypto.ExtractECDSAPublicKeyFromFile("walletMinerA.key")
				if err != nil {
					return
				}

				fromAddress := crypto.GetAddressFromPubKey(&fromPrivKey.PublicKey)
				//t.Logf("fromAddress: (%x)\n",fromAddress[0:8])
				toAddress := crypto.GetAddressFromPubKey(toPubKey)
				//t.Logf("toAddress: (%x)\n",toAddress[0:8])

				tx, err := protocol.ConstrFundsTx(
					byte(0),
					uint64(1),
					uint64(0),
					uint32(txCount),
					fromAddress,
					toAddress,
					fromPrivKey,
					nil)

				if err := SendTx("127.0.0.1:8000", tx, p2p.FUNDSTX_BRDCST); err != nil {
					return
				}
				txCount += 1
			}
		}()
	}

	time.Sleep(60*time.Second)

	//storage.State = nil
	//storage.State = make(map[[64]byte]*protocol.Account)
	////Start the miners
	//validatorPubKey1,_ := crypto.ExtractECDSAPublicKeyFromFile(NodesDirectory+"Node_1/wallet.key")
	//commPrivKey1,_ := crypto.ExtractRSAKeyFromFile(NodesDirectory+"Node_1/commitment.key")
	//p2p.Init(TestIpPortNodeA)
	//go InitFirstStart(validatorPubKey1,commPrivKey1)
	//
	//validatorPubKey2,_ := crypto.ExtractECDSAPublicKeyFromFile(NodesDirectory+"Node_2/wallet.key")
	//commPrivKey2,_ := crypto.ExtractRSAKeyFromFile(NodesDirectory+"Node_2/commitment.key")
	//p2p.Init(TestIpPortNodeB)
	//Init(validatorPubKey2,commPrivKey2)
	
	//validatorPubKey3,_ := crypto.ExtractECDSAPublicKeyFromFile(NodesDirectory+"Node3/wallet.key")
	//commPrivKey3,_ := crypto.ExtractRSAKeyFromFile(NodesDirectory+"Node3/commitment.key")

	t.Log("Done...")
}
func transferFundsToWallets() {
	//Transfer 1 Mio. funds to all wallets from root account
	txCountRootAccBeginning := 0
	for i := 1; i <= TotalNodes; i++ {
		strNode := fmt.Sprintf("Node_%d",i)
		fromPrivKey, err := crypto.ExtractECDSAKeyFromFile("walletMinerA.key")
		if err != nil {
			return
		}

		toPubKey, err := crypto.ExtractECDSAPublicKeyFromFile(NodesDirectory+strNode+"/wallet.key")
		if err != nil {
			return
		}

		fromAddress := crypto.GetAddressFromPubKey(&fromPrivKey.PublicKey)
		//t.Logf("fromAddress: (%x)\n",fromAddress[0:8])
		toAddress := crypto.GetAddressFromPubKey(toPubKey)
		//t.Logf("toAddress: (%x)\n",toAddress[0:8])

		tx, err := protocol.ConstrFundsTx(
			byte(0),
			uint64(1000000),
			uint64(0),
			uint32(txCountRootAccBeginning),
			fromAddress,
			toAddress,
			fromPrivKey,
			nil)

		if err := SendTx("127.0.0.1:8000", tx, p2p.FUNDSTX_BRDCST); err != nil {
			return
		}
		txCountRootAccBeginning += 1
	}
}

func SendTx(dial string, tx protocol.Transaction, typeID uint8) (err error) {
	if conn := p2p.Connect(dial); conn != nil {
		packet := p2p.BuildPacket(typeID, tx.Encode())
		conn.Write(packet)

		header, payload, err := p2p.RcvData_(conn)
		if err != nil || header.TypeID == p2p.NOT_FOUND {
			err = errors.New(string(payload[:]))
		}
		conn.Close()

		return err
	}

	txHash := tx.Hash()
	return errors.New(fmt.Sprintf("Sending tx %x failed.", txHash[:8]))
}


func stringAlreadyInSlice(inputSlice []string, str string) bool {
	for _, entry := range inputSlice {
		if entry == str {
			return true
		}
	}
	return false
}