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

func TestGenerateNodes(t *testing.T) {
	t.Skip("Skipping TestGenerateNodes...")
	TotalNodes = 10

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
}

func TestShardingWith20Nodes(t *testing.T) {
	t.Skip("Skipping TestShardingWith20Nodes...")
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
	//transferFundsToWallets()
	//
	//time.Sleep(15*time.Second)
	//var wg sync.WaitGroup

	//Create a goroutine for each wallet and send TX from corresponding wallet to root account
	for i := 1; i <= TotalNodes; i++ {
		strNode := fmt.Sprintf("Node_%d",i)
		//wg.Add(1)
		go func() {
			txCount := 0
			//for i := 1; i <= 500; i++ {
			for {
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
				if err := SendTx("127.0.0.1:8001", tx, p2p.FUNDSTX_BRDCST); err != nil {
					return
				}
				txCount += 1
			}
			//wg.Done()
		}()
	}

	time.Sleep(60*time.Second)

	//wg.Wait()
	t.Log("Done...")
}

func TestSendingFundsTo20Nodes(t *testing.T) {
	t.Skip("Skipping TestSendingFundsTo20Nodes...")
	TotalNodes = 20
	transferFundsToWallets()
}

func transferFundsToWallets() {
	//Transfer 1 Mio. funds to all wallets from root account
	txCountRootAccBeginning := 1
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
		toAddress := crypto.GetAddressFromPubKey(toPubKey)

		tx, err := protocol.ConstrFundsTx(
			byte(0),
			uint64(10000),
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