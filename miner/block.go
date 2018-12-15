package miner

import (
	"encoding/binary"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/bazo-blockchain/bazo-miner/p2p"
	"github.com/bazo-blockchain/bazo-miner/protocol"
	"github.com/bazo-blockchain/bazo-miner/storage"
	"github.com/bazo-blockchain/bazo-miner/vm"
	"github.com/bazo-blockchain/bazo-miner/crypto"
	"golang.org/x/crypto/sha3"
)

//Datastructure to fetch the payload of all transactions, needed for state validation.
type blockData struct {
	contractTxSlice    []*protocol.ContractTx
	fundsTxSlice  []*protocol.FundsTx
	configTxSlice []*protocol.ConfigTx
	stakeTxSlice  []*protocol.StakeTx
	block         *protocol.Block
}

//This function prepares the block to broadcast into the network. No new txs are added at this point.
func finalizeBlock(block *protocol.Block) error {
	//Check if we have a slashing proof that we can add to the block.
	//The slashingDict is updated when a new block is received and when a slashing proof is provided.
	if len(slashingDict) != 0 {
		//Get the first slashing proof.
		for address, slashingProof := range slashingDict {
			block.SlashedAddress = address
			block.ConflictingBlockHash1 = slashingProof.ConflictingBlockHash1
			block.ConflictingBlockHash2 = slashingProof.ConflictingBlockHash2
			//TODO @simibac Why do you break?
			break
		}
	}

	block.TxBucketData = [][32]byte{}
	for _, bucket := range block.TxBuckets.Sort() {
		block.TxBucketData = append(block.TxBucketData, bucket.Hash())
	}

	block.MerkleRoot = block.BuildMerkleTree().MerkleRoot()
	block.InitBloomFilter(storage.GetTxPubKeys(block))

	validatorAcc, err := storage.ReadAccount(validatorAccAddress)
	if err != nil {
		return err
	}

	copy(block.Beneficiary[:], validatorAcc.Address[:])

	// Cryptographic Sortition for PoS in Bazo
	// The commitment proof stores a signed message of the Height that this block was created at.
	commitmentProof, err := crypto.SignMessageWithRSAKey(commPrivKey, fmt.Sprint(block.Height))
	if err != nil {
		return err
	}
	copy(block.CommitmentProof[0:crypto.COMM_PROOF_LENGTH], commitmentProof[:])

	partialHash := block.HashBlock()
	prevProofs := GetLatestProofs(activeParameters.num_included_prev_proofs, block)

	nonce, err := proofOfStake(getDifficulty(), block.PrevHash, prevProofs, block.Height, validatorAcc.Balance, commitmentProof)
	if err != nil {
		return err
	}

	var nonceBuf [8]byte
	binary.BigEndian.PutUint64(nonceBuf[:], uint64(nonce))
	block.Nonce = nonceBuf
	block.Timestamp = nonce

	//Put pieces together to get the final hash.
	block.Hash = sha3.Sum256(append(nonceBuf[:], partialHash[:]...))

	//This doesn't need to be hashed, because we already have the merkle tree taking care of consistency.
	block.NrContractTx = uint16(len(block.ContractTxData))
	block.NrFundsTx = uint16(len(block.FundsTxData))
	block.NrConfigTx = uint8(len(block.ConfigTxData))
	block.NrStakeTx = uint16(len(block.StakeTxData))
	block.NrTxBucket = uint16(len(block.TxBucketData))

	return nil
}

//Transaction validation operates on a copy of a tiny subset of the state (all accounts involved in transactions).
//We do not operate global state because the work might get interrupted by receiving a block that needs validation
//which is done on the global state.
func addTx(b *protocol.Block, tx protocol.Transaction) error {
	//There is a trade-off what tests can be made now and which have to be delayed (when dynamic state is needed
	//for inspection. The decision made is to check whether contractTx and configTx have been signed with rootAcc. This
	//is a dynamic test because it needs to have access to the rootAcc state. The other option would be to include
	//the address (public key of signature) in the transaction inside the tx -> would resulted in bigger tx size.
	//So the trade-off is effectively clean abstraction vs. tx size. Everything related to fundsTx is postponed because
	//the txs depend on each other.
	if err := verify(tx); err != nil {
		return errors.New(fmt.Sprintf("transaction could not be verified: %v\n", err))
	}

	switch tx.(type) {
	case *protocol.ContractTx:
		err := addContractTx(b, tx.(*protocol.ContractTx))
		if err != nil {
			logger.Printf("Adding contractTx tx failed (%v): %v\n", err, tx.(*protocol.ContractTx))
			return err
		}
	case *protocol.FundsTx:
		err := addFundsTx(b, tx.(*protocol.FundsTx))
		if err != nil {
			logger.Printf("Adding fundsTx tx failed (%v): %v\n", err, tx.(*protocol.FundsTx))
			return err
		}
	case *protocol.ConfigTx:
		err := addConfigTx(b, tx.(*protocol.ConfigTx))
		if err != nil {
			logger.Printf("Adding configTx tx failed (%v): %v\n", err, tx.(*protocol.ConfigTx))
			return err
		}
	case *protocol.StakeTx:
		err := addStakeTx(b, tx.(*protocol.StakeTx))
		if err != nil {
			logger.Printf("Adding stakeTx tx failed (%v): %v\n", err, tx.(*protocol.StakeTx))
			return err
		}
	default:
		return errors.New("Transaction type not recognized.")
	}

	b.TotalFees += tx.TxFee()

	return nil
}

func addContractTx(b *protocol.Block, tx *protocol.ContractTx) error {
	//According to the contractTx specification, we only accept new accounts except if the removal bit is
	//set in the header (2nd bit).
	if tx.Header&0x02 != 0x02 {
		if _, exists := storage.State[tx.PubKey]; exists {
			return errors.New("Account already exists.")
		}
	}

	//Add the tx hash to the block header and write it to open storage (non-validated transactions).
	b.ContractTxData = append(b.ContractTxData, tx.Hash())
	logger.Printf("Added tx to the ContractTxData slice: %v", *tx)
	return nil
}

func addFundsTx(b *protocol.Block, tx *protocol.FundsTx) error {
	//Checking if the sender account is already in the local state copy. If not and account exist, create local copy.
	if _, exists := b.StateCopy[tx.From]; !exists {
		if acc := storage.State[tx.From]; acc != nil {
			if acc.Address == tx.From {
				newAcc := protocol.Account{}
				newAcc = *acc
				b.StateCopy[tx.From] = &newAcc
			}
		} else {
			newFromAcc := protocol.NewAccount(tx.From, [64]byte{}, 0, false, [crypto.COMM_KEY_LENGTH]byte{}, nil, nil)
			b.StateCopy[tx.From] = &newFromAcc
		}
	}

	//Vice versa for receiver account.
	if _, exists := b.StateCopy[tx.To]; !exists {
		if acc := storage.State[tx.To]; acc != nil {
			if acc.Address == tx.To {
				newAcc := protocol.Account{}
				newAcc = *acc
				b.StateCopy[tx.To] = &newAcc
			}
		} else {
			newToAcc := protocol.NewAccount(tx.To, [64]byte{}, 0, false, [crypto.COMM_KEY_LENGTH]byte{}, nil, nil)
			b.StateCopy[tx.To] = &newToAcc
		}
	}

	//Root accounts are exempt from balance requirements. All other accounts need to have (at least)
	//fee + amount to spend as balance available.
	if !storage.IsRootKey(tx.From) {
		if (tx.Amount + tx.Fee) > b.StateCopy[tx.From].Balance {
			return errors.New("Not enough funds to complete the transaction!")
		}
	}

	//Transaction count need to match the state, preventing replay attacks.
	if b.StateCopy[tx.From].TxCnt != tx.TxCnt {
		err := fmt.Sprintf("Sender txCnt does not match: %v (tx.txCnt) vs. %v (state txCnt)", tx.TxCnt, b.StateCopy[tx.From].TxCnt)
		return errors.New(err)
	}

	//Prevent balance overflow in receiver account.
	if b.StateCopy[tx.To].Balance+tx.Amount > MAX_MONEY {
		err := fmt.Sprintf("Transaction amount (%v) leads to overflow at receiver account balance (%v).\n", tx.Amount, b.StateCopy[tx.To].Balance)
		return errors.New(err)
	}

	//Check if transaction has data and the receiver account has a smart contract
	if tx.Data != nil && b.StateCopy[tx.To].Contract != nil {
		context := protocol.NewContext(*b.StateCopy[tx.To], *tx)
		virtualMachine := vm.NewVM(context)

		// Check if vm execution run without error
		if !virtualMachine.Exec(false) {
			return errors.New(virtualMachine.GetErrorMsg())
		}

		//Update changes vm has made to the contract variables
		context.PersistChanges()
	}

	//Update state copy.
	accSender := b.StateCopy[tx.From]
	accSender.TxCnt += 1
	accSender.Balance -= tx.Amount

	accReceiver := b.StateCopy[tx.To]
	accReceiver.Balance += tx.Amount

	b.AddFundsTx(tx)
	logger.Printf("Added tx to the FundsTxData slice: %v", *tx)

	return nil
}

func addConfigTx(b *protocol.Block, tx *protocol.ConfigTx) error {
	//No further checks needed, static checks were already done with verify().
	b.ConfigTxData = append(b.ConfigTxData, tx.Hash())
	logger.Printf("Added tx to the ConfigTxData slice: %v", *tx)
	return nil
}

func addStakeTx(b *protocol.Block, tx *protocol.StakeTx) error {
	//Checking if the sender account is already in the local state copy. If not and account exist, create local copy
	//If account does not exist in state, abort.
	if _, exists := b.StateCopy[tx.Account]; !exists {
		if acc := storage.State[tx.Account]; acc != nil {
			if acc.Address == tx.Account {
				newAcc := protocol.Account{}
				newAcc = *acc
				b.StateCopy[tx.Account] = &newAcc
			}
		} else {
			newAcc := protocol.NewAccount(tx.Account, [64]byte{}, 0, false, [crypto.COMM_KEY_LENGTH]byte{}, nil, nil)
			b.StateCopy[tx.Account] = &newAcc
		}
	}

	//Root accounts are exempt from balance requirements. All other accounts need to have (at least)
	//fee + minimum amount that is required for staking.
	if !storage.IsRootKey(tx.Account) {
		if (tx.Fee + activeParameters.Staking_minimum) >= b.StateCopy[tx.Account].Balance {
			return errors.New("Not enough funds to complete the transaction!")
		}
	}

	//Account has bool already set to the desired value.
	if b.StateCopy[tx.Account].IsStaking == tx.IsStaking {
		return errors.New("Account has bool already set to the desired value.")
	}

	//Update state copy.
	accSender := b.StateCopy[tx.Account]
	accSender.IsStaking = tx.IsStaking
	accSender.CommitmentKey = tx.CommitmentKey

	//No further checks needed, static checks were already done with verify().
	b.StakeTxData = append(b.StakeTxData, tx.Hash())
	logger.Printf("Added tx to the StakeTxData slice: %v", *tx)
	return nil
}

//We use slices (not maps) because order is now important.
func fetchContractTxData(block *protocol.Block, contractTxSlice []*protocol.ContractTx, initialSetup bool, errChan chan error) {
	for cnt, txHash := range block.ContractTxData {
		var tx protocol.Transaction
		var contractTx *protocol.ContractTx

		closedTx := storage.ReadClosedTx(txHash)
		if closedTx != nil {
			if initialSetup {
				contractTx = closedTx.(*protocol.ContractTx)
				contractTxSlice[cnt] = contractTx
				continue
			} else {
				//Reject blocks that have txs which have already been validated.
				errChan <- errors.New("Block validation had contractTx that was already in a previous block.")
				return
			}
		}

		//TODO Optimize code (duplicated)
		//Tx is either in open storage or needs to be fetched from the network.
		tx = storage.ReadOpenTx(txHash)
		if tx != nil {
			contractTx = tx.(*protocol.ContractTx)
		} else {
			err := p2p.TxReq(txHash, p2p.CONTRACTTX_REQ)
			if err != nil {
				errChan <- errors.New(fmt.Sprintf("ContractTx could not be read: %v", err))
				return
			}

			//Blocking Wait
			select {
			case contractTx = <-p2p.ContractTxChan:
				//Limit the waiting time for TXFETCH_TIMEOUT seconds.
			case <-time.After(TXFETCH_TIMEOUT * time.Second):
				errChan <- errors.New("ContractTx fetch timed out.")
			}
			//This check is important. A malicious miner might have sent us a tx whose hash is a different one
			//from what we requested.
			if contractTx.Hash() != txHash {
				errChan <- errors.New("Received txHash did not correspond to our request.")
			}
		}

		contractTxSlice[cnt] = contractTx
	}

	errChan <- nil
}

func fetchFundsTxData(block *protocol.Block, fundsTxSlice []*protocol.FundsTx, initialSetup bool, errChan chan error) {
	for cnt, txHash := range block.FundsTxData {
		var tx protocol.Transaction
		var fundsTx *protocol.FundsTx

		closedTx := storage.ReadClosedTx(txHash)
		if closedTx != nil {
			if initialSetup {
				fundsTx = closedTx.(*protocol.FundsTx)
				fundsTxSlice[cnt] = fundsTx
				continue
			} else {
				errChan <- errors.New("Block validation had fundsTx that was already in a previous block.")
				return
			}
		}

		//TODO Optimize code (duplicated)
		tx = storage.ReadOpenTx(txHash)
		if tx != nil {
			fundsTx = tx.(*protocol.FundsTx)
		} else {
			err := p2p.TxReq(txHash, p2p.FUNDSTX_REQ)
			if err != nil {
				errChan <- errors.New(fmt.Sprintf("FundsTx could not be read: %v", err))
				return
			}

			select {
			case fundsTx = <-p2p.FundsTxChan:
			case <-time.After(TXFETCH_TIMEOUT * time.Second):
				errChan <- errors.New("FundsTx fetch timed out.")
				return
			}
			if fundsTx.Hash() != txHash {
				errChan <- errors.New("Received txHash did not correspond to our request.")
			}
		}

		fundsTxSlice[cnt] = fundsTx
	}

	errChan <- nil
}

func fetchConfigTxData(block *protocol.Block, configTxSlice []*protocol.ConfigTx, initialSetup bool, errChan chan error) {
	for cnt, txHash := range block.ConfigTxData {
		var tx protocol.Transaction
		var configTx *protocol.ConfigTx

		closedTx := storage.ReadClosedTx(txHash)
		if closedTx != nil {
			if initialSetup {
				configTx = closedTx.(*protocol.ConfigTx)
				configTxSlice[cnt] = configTx
				continue
			} else {
				errChan <- errors.New("Block validation had configTx that was already in a previous block.")
				return
			}
		}

		//TODO Optimize code (duplicated)
		tx = storage.ReadOpenTx(txHash)
		if tx != nil {
			configTx = tx.(*protocol.ConfigTx)
		} else {
			err := p2p.TxReq(txHash, p2p.CONFIGTX_REQ)
			if err != nil {
				errChan <- errors.New(fmt.Sprintf("ConfigTx could not be read: %v", err))
				return
			}

			select {
			case configTx = <-p2p.ConfigTxChan:
			case <-time.After(TXFETCH_TIMEOUT * time.Second):
				errChan <- errors.New("ConfigTx fetch timed out.")
				return
			}
			if configTx.Hash() != txHash {
				errChan <- errors.New("Received txHash did not correspond to our request.")
			}
		}

		configTxSlice[cnt] = configTx
	}

	errChan <- nil
}

func fetchStakeTxData(block *protocol.Block, stakeTxSlice []*protocol.StakeTx, initialSetup bool, errChan chan error) {
	for cnt, txHash := range block.StakeTxData {
		var tx protocol.Transaction
		var stakeTx *protocol.StakeTx

		closedTx := storage.ReadClosedTx(txHash)
		if closedTx != nil {
			if initialSetup {
				stakeTx = closedTx.(*protocol.StakeTx)
				stakeTxSlice[cnt] = stakeTx
				continue
			} else {
				errChan <- errors.New("Block validation had stakeTx that was already in a previous block.")
				return
			}
		}

		//TODO Optimize code (duplicated)
		tx = storage.ReadOpenTx(txHash)
		if tx != nil {
			stakeTx = tx.(*protocol.StakeTx)
		} else {
			err := p2p.TxReq(txHash, p2p.STAKETX_REQ)
			if err != nil {
				errChan <- errors.New(fmt.Sprintf("StakeTx could not be read: %v", err))
				return
			}

			select {
			case stakeTx = <-p2p.StakeTxChan:
			case <-time.After(TXFETCH_TIMEOUT * time.Second):
				errChan <- errors.New("StakeTx fetch timed out.")
				return
			}
			if stakeTx.Hash() != txHash {
				errChan <- errors.New("Received txHash did not correspond to our request.")
			}
		}

		stakeTxSlice[cnt] = stakeTx
	}

	errChan <- nil
}

//This function is split into block syntax/PoS check and actual state change
//because there is the case that we might need to go fetch several blocks
// and have to check the blocks first before changing the state in the correct order.
func validate(b *protocol.Block, initialSetup bool) error {
	//TODO Optimize code

	//This mutex is necessary that own-mined blocks and received blocks from the network are not
	//validated concurrently.
	blockValidation.Lock()
	defer blockValidation.Unlock()

	//Prepare datastructure to fill tx payloads.
	blockDataMap := make(map[[32]byte]blockData)

	//Get the right branch, and a list of blocks to rollback (if necessary).
	blocksToRollback, blocksToValidate, err := getBlockSequences(b)
	if err != nil {
		return err
	}

	//Verify block time is dynamic and corresponds to system time at the time of retrieval.
	//If we are syncing or far behind, we cannot do this dynamic check,
	//therefore we include a boolean uptodate. If it's true we consider ourselves uptodate and
	//do dynamic time checking.
	if len(blocksToValidate) > DELAYED_BLOCKS {
		uptodate = false
	} else {
		uptodate = true
	}

	if len(blocksToRollback) > 0 {
		for _, block := range blocksToRollback {
			if err := rollback(block); err != nil {
				return err
			}
			logger.Printf("Rolled back block: %vState:\n%v", block, getState())
		}
	}

	for _, block := range blocksToValidate {
		//Fetching payload data from the txs (if necessary, ask other miners).
		data, err := preValidate(block, initialSetup)

		//Check if the validator that added the block has previously voted on different competing chains (find slashing proof).
		//The proof will be stored in the global slashing dictionary.
		if block.Height > 0 {
			seekSlashingProof(block)
		}

		if err != nil {
			return err
		}

		blockDataMap[block.Hash] = *data
		if err := validateState(blockDataMap[block.Hash]); err != nil {
			return err
		}

		postValidate(blockDataMap[block.Hash], initialSetup)
	}

	deleteZeroBalanceAccounts()

	return nil
}

//Doesn't involve any state changes.
func preValidate(block *protocol.Block, initialSetup bool) (data *blockData, err error) {
	//This dynamic check is only done if we're up-to-date with syncing, otherwise timestamp is not checked.
	//Other miners (which are up-to-date) made sure that this is correct.
	if !initialSetup && uptodate {
		if err := timestampCheck(block.Timestamp); err != nil {
			return nil, err
		}
	}

	//Check block size.
	if block.GetSize() > activeParameters.Block_size {
		return nil, errors.New("Block size too large.")
	}

	//Duplicates are not allowed, use tx hash hashmap to easily check for duplicates.
	duplicates := make(map[[32]byte]bool)
	for _, txHash := range block.ContractTxData {
		if _, exists := duplicates[txHash]; exists {
			return nil, errors.New("Duplicate Account Transaction Hash detected.")
		}
		duplicates[txHash] = true
	}
	for _, txHash := range block.FundsTxData {
		if _, exists := duplicates[txHash]; exists {
			return nil, errors.New("Duplicate Funds Transaction Hash detected.")
		}
		duplicates[txHash] = true
	}
	for _, txHash := range block.ConfigTxData {
		if _, exists := duplicates[txHash]; exists {
			return nil, errors.New("Duplicate Config Transaction Hash detected.")
		}
		duplicates[txHash] = true
	}
	for _, txHash := range block.StakeTxData {
		if _, exists := duplicates[txHash]; exists {
			return nil, errors.New("Duplicate Stake Transaction Hash detected.")
		}
		duplicates[txHash] = true
	}

	//We fetch tx data for each type in parallel -> performance boost.
	errChan := make(chan error, 4)

	//We need to allocate slice space for the underlying array when we pass them as reference.
	contractTxSlice := make([]*protocol.ContractTx, block.NrContractTx)
	fundsTxSlice := make([]*protocol.FundsTx, block.NrFundsTx)
	configTxSlice := make([]*protocol.ConfigTx, block.NrConfigTx)
	stakeTxSlice := make([]*protocol.StakeTx, block.NrStakeTx)

	go fetchContractTxData(block, contractTxSlice, initialSetup, errChan)
	go fetchFundsTxData(block, fundsTxSlice, initialSetup, errChan)
	go fetchConfigTxData(block, configTxSlice, initialSetup, errChan)
	go fetchStakeTxData(block, stakeTxSlice, initialSetup, errChan)

	//Wait for all goroutines to finish.
	for cnt := 0; cnt < 4; cnt++ {
		err = <-errChan
		if err != nil {
			return nil, err
		}
	}

	data = &blockData{contractTxSlice, fundsTxSlice, configTxSlice, stakeTxSlice, block}

	//Check state contains beneficiary.
	acc, err := storage.ReadAccount(block.Beneficiary)
	if err != nil {
		return nil, err
	}

	//Check if node is part of the validator set.
	if !acc.IsStaking {
		return nil, errors.New(fmt.Sprintf("Validator (%x) is not part of the validator set.", acc.Address[0:8]))
	}

	//First, initialize an RSA Public Key instance with the modulus of the proposer of the block (acc)
	//Second, check if the commitment proof of the proposed block can be verified with the public key
	//Invalid if the commitment proof can not be verified with the public key of the proposer
	commitmentPubKey, err := crypto.CreateRSAPubKeyFromBytes(acc.CommitmentKey)
	if err != nil {
		return nil, errors.New("Invalid commitment key in account.")
	}

	err = crypto.VerifyMessageWithRSAKey(commitmentPubKey, fmt.Sprint(block.Height), block.CommitmentProof)
	if err != nil {
		return nil, errors.New("The submitted commitment proof can not be verified.")
	}

	//Invalid if PoS calculation is not correct.
	prevProofs := GetLatestProofs(activeParameters.num_included_prev_proofs, block)

	//PoS validation
	if !validateProofOfStake(getDifficulty(), prevProofs, block.Height, acc.Balance, block.CommitmentProof, block.Timestamp) {
		return nil, errors.New("The nonce is incorrect.")
	}

	//Invalid if PoS is too far in the future.
	now := time.Now()
	if block.Timestamp > now.Unix()+int64(activeParameters.Accepted_time_diff) {
		return nil, errors.New("The timestamp is too far in the future. " + string(block.Timestamp) + " vs " + string(now.Unix()))
	}

	//Check for minimum waiting time.
	if block.Height-acc.StakingBlockHeight < uint32(activeParameters.Waiting_minimum) {
		return nil, errors.New("The miner must wait a minimum amount of blocks before start validating. Block Height:" + fmt.Sprint(block.Height) + " - Height when started validating " + string(acc.StakingBlockHeight) + " MinWaitingTime: " + string(activeParameters.Waiting_minimum))
	}

	//Check if block contains a proof for two conflicting block hashes, else no proof provided.
	if block.SlashedAddress != [64]byte{} {
		if _, err = slashingCheck(block.SlashedAddress, block.ConflictingBlockHash1, block.ConflictingBlockHash2); err != nil {
			return nil, err
		}
	}

	// Total Fees validation
	var totalFees uint64 = 0
	for _, tx := range contractTxSlice {
		totalFees += tx.Fee
	}
	for _, tx := range fundsTxSlice {
		totalFees += tx.Fee
	}
	for _, tx := range configTxSlice {
		totalFees += tx.Fee
	}
	for _, tx := range stakeTxSlice {
		totalFees += tx.Fee
	}
	if totalFees != block.TotalFees {
		return nil, errors.New(fmt.Sprintf("computed total fees do not equal the block's total fees %v vs. %v", totalFees, block.TotalFees))
	}

	/*for _, tx := range fundsTxSlice {
		block.AddFundsTx(tx)
	}*/

	//Merkle Tree validation
	if block.BuildMerkleTree().MerkleRoot() != block.MerkleRoot {
		return nil, errors.New("Merkle Root is incorrect.")
	}

	return data, err
}

//Dynamic state check.
//The sequence of validation matters
func validateState(data blockData) (err error) {
	accStateChange(data.contractTxSlice)

	err = fundsStateChange(data.fundsTxSlice)
	if err != nil {
		accStateChangeRollback(data.contractTxSlice)
		return err
	}

	if err := stakeStateChange(data.stakeTxSlice, data.block.Height); err != nil {
		fundsStateChangeRollback(data.fundsTxSlice)
		accStateChangeRollback(data.contractTxSlice)
		return err
	}

	if err := collectTxFees(data.contractTxSlice, data.fundsTxSlice, data.configTxSlice, data.stakeTxSlice, data.block.Beneficiary); err != nil {
		stakeStateChangeRollback(data.stakeTxSlice)
		fundsStateChangeRollback(data.fundsTxSlice)
		accStateChangeRollback(data.contractTxSlice)
		return err
	}

	if err := collectBlockReward(activeParameters.Block_reward, data.block.Beneficiary); err != nil {
		collectTxFeesRollback(data.contractTxSlice, data.fundsTxSlice, data.configTxSlice, data.stakeTxSlice, data.block.Beneficiary)
		stakeStateChangeRollback(data.stakeTxSlice)
		fundsStateChangeRollback(data.fundsTxSlice)
		accStateChangeRollback(data.contractTxSlice)
		return err
	}

	if err := collectSlashReward(activeParameters.Slash_reward, data.block); err != nil {
		collectBlockRewardRollback(activeParameters.Block_reward, data.block.Beneficiary)
		collectTxFeesRollback(data.contractTxSlice, data.fundsTxSlice, data.configTxSlice, data.stakeTxSlice, data.block.Beneficiary)
		stakeStateChangeRollback(data.stakeTxSlice)
		fundsStateChangeRollback(data.fundsTxSlice)
		accStateChangeRollback(data.contractTxSlice)
		return err
	}

	if err := updateStakingHeight(data.block); err != nil {
		collectSlashRewardRollback(activeParameters.Slash_reward, data.block)
		collectBlockRewardRollback(activeParameters.Block_reward, data.block.Beneficiary)
		collectTxFeesRollback(data.contractTxSlice, data.fundsTxSlice, data.configTxSlice, data.stakeTxSlice, data.block.Beneficiary)
		stakeStateChangeRollback(data.stakeTxSlice)
		fundsStateChangeRollback(data.fundsTxSlice)
		accStateChangeRollback(data.contractTxSlice)
		return err
	}

	return nil
}

func postValidate(data blockData, initialSetup bool) {
	//The new system parameters get active if the block was successfully validated
	//This is done after state validation (in contrast to contractTx/fundsTx).
	//Conversely, if blocks are rolled back, the system parameters are changed first.
	configStateChange(data.configTxSlice, data.block.Hash)
	//Collects meta information about the block (and handled difficulty adaption).
	collectStatistics(data.block)

	if !initialSetup {
		//Write all open transactions to closed/validated storage.
		for _, tx := range data.contractTxSlice {
			storage.WriteClosedTx(tx)
			storage.DeleteOpenTx(tx)
		}

		for _, tx := range data.fundsTxSlice {
			storage.WriteClosedTx(tx)
			storage.DeleteOpenTx(tx)
		}

		for _, tx := range data.configTxSlice {
			storage.WriteClosedTx(tx)
			storage.DeleteOpenTx(tx)
		}

		for _, tx := range data.stakeTxSlice {
			storage.WriteClosedTx(tx)
			storage.DeleteOpenTx(tx)
		}

		if len(data.fundsTxSlice) > 0 {
			broadcastVerifiedTxs(data.fundsTxSlice)
		}

		//It might be that block is not in the openblock storage, but this doesn't matter.
		storage.DeleteOpenBlock(data.block.Hash)
		storage.WriteClosedBlock(data.block)

		// Write last block to db and delete last block's ancestor.
		storage.DeleteAllLastClosedBlock()
		storage.WriteLastClosedBlock(data.block)
	}
}

//Only blocks with timestamp not diverging from system time (past or future) more than one hour are accepted.
func timestampCheck(timestamp int64) error {
	systemTime := p2p.ReadSystemTime()

	if timestamp > systemTime {
		if timestamp-systemTime > int64(time.Hour.Seconds()) {
			return errors.New("Timestamp was too far in the future.System time: " + strconv.FormatInt(systemTime, 10) + " vs. timestamp " + strconv.FormatInt(timestamp, 10) + "\n")
		}
	} else {
		if systemTime-timestamp > int64(time.Hour.Seconds()) {
			return errors.New("Timestamp was too far in the past. System time: " + strconv.FormatInt(systemTime, 10) + " vs. timestamp " + strconv.FormatInt(timestamp, 10) + "\n")
		}
	}

	return nil
}

func slashingCheck(slashedAddress [64]byte, conflictingBlockHash1, conflictingBlockHash2 [32]byte) (bool, error) {
	prefix := "Invalid slashing proof: "

	if conflictingBlockHash1 == [32]byte{} || conflictingBlockHash2 == [32]byte{} {
		return false, errors.New(fmt.Sprintf(prefix + "Invalid conflicting block hashes provided."))
	}

	if conflictingBlockHash1 == conflictingBlockHash2 {
		return false, errors.New(fmt.Sprintf(prefix + "Conflicting block hashes are the same."))
	}

	//Fetch the blocks for the provided block hashes.
	conflictingBlock1 := storage.ReadClosedBlock(conflictingBlockHash1)
	conflictingBlock2 := storage.ReadClosedBlock(conflictingBlockHash2)

	if IsInSameChain(conflictingBlock1, conflictingBlock2) {
		return false, errors.New(fmt.Sprintf(prefix + "Conflicting block hashes are on the same chain."))
	}

	//TODO Optimize code (duplicated)
	//If this block is unknown we need to check if its in the openblock storage or we must request it.
	if conflictingBlock1 == nil {
		conflictingBlock1 = storage.ReadOpenBlock(conflictingBlockHash1)
		if conflictingBlock1 == nil {
			//Fetch the block we apparently missed from the network.
			p2p.BlockReq(conflictingBlockHash1)

			//Blocking wait
			select {
			case encodedBlock := <-p2p.BlockReqChan:
				conflictingBlock1 = conflictingBlock1.Decode(encodedBlock)
				//Limit waiting time to BLOCKFETCH_TIMEOUT seconds before aborting.
			case <-time.After(BLOCKFETCH_TIMEOUT * time.Second):
				return false, errors.New(fmt.Sprintf(prefix + "Could not find a block with the provided conflicting hash (1)."))
			}
		}

		ancestor, _ := getNewChain(conflictingBlock1)
		if ancestor == nil {
			return false, errors.New(fmt.Sprintf(prefix + "Could not find a ancestor for the provided conflicting hash (1)."))
		}
	}

	//TODO Optimize code (duplicated)
	//If this block is unknown we need to check if its in the openblock storage or we must request it.
	if conflictingBlock2 == nil {
		conflictingBlock2 = storage.ReadOpenBlock(conflictingBlockHash2)
		if conflictingBlock2 == nil {
			//Fetch the block we apparently missed from the network.
			p2p.BlockReq(conflictingBlockHash2)

			//Blocking wait
			select {
			case encodedBlock := <-p2p.BlockReqChan:
				conflictingBlock2 = conflictingBlock2.Decode(encodedBlock)
				//Limit waiting time to BLOCKFETCH_TIMEOUT seconds before aborting.
			case <-time.After(BLOCKFETCH_TIMEOUT * time.Second):
				return false, errors.New(fmt.Sprintf(prefix + "Could not find a block with the provided conflicting hash (2)."))
			}
		}

		ancestor, _ := getNewChain(conflictingBlock2)
		if ancestor == nil {
			return false, errors.New(fmt.Sprintf(prefix + "Could not find a ancestor for the provided conflicting hash (2)."))
		}
	}

	// We found the height of the blocks and the height of the blocks can be checked.
	// If the height is not within the active slashing window size, we must throw an error. If not, the proof is valid.
	if !(conflictingBlock1.Height < uint32(activeParameters.Slashing_window_size)+conflictingBlock2.Height) {
		return false, errors.New(fmt.Sprintf(prefix + "Could not find a ancestor for the provided conflicting hash (2)."))
	}

	//Delete the proof from local slashing dictionary. If proof has not existed yet, nothing will be deleted.
	delete(slashingDict, slashedAddress)

	return true, nil
}
