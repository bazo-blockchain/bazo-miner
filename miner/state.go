package miner

import (
	"errors"
	"fmt"
	"github.com/bazo-blockchain/bazo-miner/crypto"
	"github.com/bazo-blockchain/bazo-miner/p2p"
	"github.com/bazo-blockchain/bazo-miner/protocol"
	"github.com/bazo-blockchain/bazo-miner/storage"
	"strconv"
	"time"
)

//Separate function to reuse mechanism in client implementation
func CheckAndChangeParameters(parameters *Parameters, configTxSlice *[]*protocol.ConfigTx) (change bool) {
	for _, tx := range *configTxSlice {
		switch tx.Id {
		case protocol.FEE_MINIMUM_ID:
			if parameterBoundsChecking(protocol.FEE_MINIMUM_ID, tx.Payload) {
				parameters.Fee_minimum = tx.Payload
				change = true
			}
		case protocol.BLOCK_SIZE_ID:
			if parameterBoundsChecking(protocol.BLOCK_SIZE_ID, tx.Payload) {
				parameters.Block_size = tx.Payload
				change = true
			}
		case protocol.BLOCK_REWARD_ID:
			if parameterBoundsChecking(protocol.BLOCK_REWARD_ID, tx.Payload) {
				parameters.Block_reward = tx.Payload
				change = true
			}
		case protocol.DIFF_INTERVAL_ID:
			if parameterBoundsChecking(protocol.DIFF_INTERVAL_ID, tx.Payload) {
				parameters.Diff_interval = tx.Payload
				change = true
			}
		case protocol.BLOCK_INTERVAL_ID:
			if parameterBoundsChecking(protocol.BLOCK_INTERVAL_ID, tx.Payload) {
				parameters.Block_interval = tx.Payload
				change = true
			}
		case protocol.STAKING_MINIMUM_ID:
			if parameterBoundsChecking(protocol.STAKING_MINIMUM_ID, tx.Payload) {
				parameters.Staking_minimum = tx.Payload
				change = true
				//Go through all accounts and remove all validators from the validator sett that no longer fulfill the minimum staking amount
				for _, account := range storage.State {
					if account.IsStaking && account.Balance < 0+tx.Payload {
						account.IsStaking = false
					}
				}
			}
		case protocol.WAITING_MINIMUM_ID:
			if parameterBoundsChecking(protocol.WAITING_MINIMUM_ID, tx.Payload) {
				parameters.Waiting_minimum = tx.Payload
				change = true
			}
		case protocol.ACCEPTANCE_TIME_DIFF_ID:
			if parameterBoundsChecking(protocol.ACCEPTANCE_TIME_DIFF_ID, tx.Payload) {
				parameters.Accepted_time_diff = tx.Payload
				change = true
			}
		case protocol.SLASHING_WINDOW_SIZE_ID:
			if parameterBoundsChecking(protocol.SLASHING_WINDOW_SIZE_ID, tx.Payload) {
				parameters.Slashing_window_size = tx.Payload
				change = true
			}
		case protocol.SLASHING_REWARD_ID:
			if parameterBoundsChecking(protocol.SLASHING_REWARD_ID, tx.Payload) {
				parameters.Slash_reward = tx.Payload
				change = true
			}
		}
	}

	return change
}

//For logging purposes
func getState() (state string) {
	for _, acc := range storage.State {
		state += fmt.Sprintf("Is root: %v, %v\n", storage.IsRootKey(acc.Address), acc)
	}
	return state
}

func GetState() (state map[[64]byte]*protocol.Account){

	stateMap := make(map[[64]byte]*protocol.Account)

	for _, acc := range storage.State {
		stateMap[acc.Address] = acc
	}
	return stateMap
}

func GetValidatorsCount() (validatorsCount int) {
	var returnValCounts int
	returnValCounts = 0
	for _, acc := range storage.State {
		if acc.IsStaking {
			returnValCounts += 1
		}
	}
	return returnValCounts
}

func GetTotalAccountsCount() (accountsCount int) {
	return len(storage.State)
}

func initState() (initialBlock *protocol.Block, err error) {
	genesis, err := initGenesis()
	if err != nil {
		return nil, err
	}

	initialEpochBlock, err := initEpochBlock()

	FileConnections.WriteString(fmt.Sprintf("'GENESIS: %x' -> 'EPOCH BLOCK: %x'\n",[32]byte{},initialEpochBlock.Hash[0:15]))

	if err != nil {
		return nil, err
	}

	initRootAccounts(genesis)

	err = initClosedBlocks(initialEpochBlock)
	if err != nil {
		return nil, err
	}

	initialBlock, err = getInitialBlock(initialEpochBlock)
	if err != nil {
		return nil, err
	}

	err = validateClosedBlocks()
	if err != nil {
		return nil, err
	}

	return initialBlock, nil
}

func initGenesis() (genesis *protocol.Genesis, err error) {
	if genesis, err = storage.ReadGenesis(); err != nil {
		return nil, err
	}

	if genesis == nil {
		p2p.GenesisReq()

		// TODO: @rmnblm parallelize this
		// blocking wait
		select {
		case encodedGenesis := <-p2p.GenesisReqChan:
			genesis = genesis.Decode(encodedGenesis)
			logger.Printf("Received genesis: %v", genesis.String())
		case <-time.After(GENESISFETCH_TIMEOUT * time.Second):
			return nil, errors.New("genesis fetch timeout")
		}

		storage.WriteGenesis(genesis)
	}
	return genesis, nil
}

func initEpochBlock() (initialEpochBlock *protocol.EpochBlock, err error) {
	if initialEpochBlock, err = storage.ReadFirstEpochBlock(); err != nil {
		return nil, err
	}

	if initialEpochBlock == nil {
		p2p.FirstEpochBlockReq()

		// TODO: @rmnblm parallelize this
		// blocking wait
		select {
		case encodedFirstEpochBlock := <-p2p.FirstEpochBlockReqChan:
			initialEpochBlock = initialEpochBlock.Decode(encodedFirstEpochBlock)
			logger.Printf("Received first Epoch Block: %v", initialEpochBlock.String())
		case <-time.After(EPOCHBLOCKFETCH_TIMEOUT* time.Second):
			return nil, errors.New("epoch block fetch timeout")
		}

		storage.WriteClosedEpochBlock(initialEpochBlock)
	}
	return initialEpochBlock, nil
}

func initRootAccounts(genesis *protocol.Genesis) {
	rootAcc := protocol.NewAccount(genesis.RootAddress, [64]byte{}, activeParameters.Staking_minimum, true, genesis.RootCommitment, nil, nil)
	storage.State[genesis.RootAddress] = &rootAcc
	storage.RootKeys[genesis.RootAddress] = &rootAcc
}

func initClosedBlocks(firstEpochBlock *protocol.EpochBlock) error {
	var allClosedBlocks []*protocol.Block
	if p2p.IsBootstrap() {
		allClosedBlocks = storage.ReadAllClosedBlocks()
	} else {
		p2p.LastBlockReq()

		var lastBlock *protocol.Block
		//Blocking wait
		select {
		case encodedBlock := <-p2p.BlockReqChan:
			lastBlock = lastBlock.Decode(encodedBlock)
			//Limit waiting time to BLOCKFETCH_TIMEOUT seconds before aborting.
		case <-time.After(BLOCKFETCH_TIMEOUT * time.Second):
			return errors.New("block fetch timeout")
		}

		storage.WriteClosedBlock(lastBlock)
		storage.WriteLastClosedBlock(lastBlock)
		if len(allClosedBlocks) > 0 && allClosedBlocks[len(allClosedBlocks)-1].Hash == lastBlock.Hash {
			fmt.Printf("Block with height %v already exists", lastBlock.Height)
		} else {
			allClosedBlocks = append(allClosedBlocks, lastBlock)
		}

		for {
			p2p.BlockReq(lastBlock.PrevHash)
			select {
			case encodedBlock := <-p2p.BlockReqChan:
				lastBlock = lastBlock.Decode(encodedBlock)
				//Limit waiting time to BLOCKFETCH_TIMEOUT seconds before aborting.
			case <-time.After(BLOCKFETCH_TIMEOUT * time.Second):
				logger.Println("Timed out")
			}

			storage.WriteClosedBlock(lastBlock)
			if len(allClosedBlocks) > 0 && allClosedBlocks[len(allClosedBlocks)-1].Hash == lastBlock.Hash {
				fmt.Printf("Block with height %v already exists", lastBlock.Height)
			} else {
				allClosedBlocks = append(allClosedBlocks, lastBlock)
			}
			fmt.Println("Last block: ", lastBlock.Height)
			if lastBlock.Height == 0 {
				if lastBlock.PrevHash != firstEpochBlock.HashEpochBlock() {
					return errors.New("invalid first epoch block")
				}
				break
			}
		}
	}

	//Switch array order to validate genesis block first
	storage.AllClosedBlocksAsc = InvertBlockArray(allClosedBlocks)

	return nil
}

func getInitialBlock(firstEpochBlock *protocol.EpochBlock) (initialBlock *protocol.Block, err error) {
	if len(storage.AllClosedBlocksAsc) > 0 {
		//Set the last closed block as the initial block
		initialBlock = storage.AllClosedBlocksAsc[len(storage.AllClosedBlocksAsc)-1]
	} else {
		initialBlock = newBlock(firstEpochBlock.Hash, [crypto.COMM_PROOF_LENGTH]byte{}, 1) // since first epoch block is at height 0
		commitmentProof, err := crypto.SignMessageWithRSAKey(commPrivKey, fmt.Sprint(initialBlock.Height))
		if err != nil {
			return nil, err
		}
		copy(initialBlock.CommitmentProof[:], commitmentProof[:])

		//Append genesis block to the map and save in storage
		storage.AllClosedBlocksAsc = append(storage.AllClosedBlocksAsc, initialBlock)

		storage.WriteLastClosedBlock(initialBlock)
		storage.WriteClosedBlock(initialBlock)
	}

	return initialBlock, nil
}

func validateClosedBlocks() error {

	/*Initialize file which which stores the hash-prevhash connections of the blockchain. This is just for visualization purposes*/
	/*file,err := os.OpenFile("hash-prevhash.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}*/

	//Validate all closed blocks and update state
	for _, blockToValidate := range storage.AllClosedBlocksAsc {
		//Prepare datastructure to fill tx payloads
		blockDataMap := make(map[[32]byte]blockData)

		//Do not validate the genesis block, since a lot of properties are set to nil
		if blockToValidate.Hash != [32]byte{} {
			//Fetching payload data from the txs (if necessary, ask other miners)
			contractTxs, fundsTxs, configTxs, stakeTxs, err := preValidate(blockToValidate, true)
			if err != nil {
				return errors.New(fmt.Sprintf("Block (%x) could not be prevalidated: %v\n", blockToValidate.Hash[0:8], err))
			}

			blockDataMap[blockToValidate.Hash] = blockData{contractTxs, fundsTxs, configTxs, stakeTxs, blockToValidate}

			err = validateState(blockDataMap[blockToValidate.Hash])
			if err != nil {
				return errors.New(fmt.Sprintf("Block (%x) could not be statevalidated: %v\n", blockToValidate.Hash[0:8], err))
			}

			postValidate(blockDataMap[blockToValidate.Hash], true)
		} else {
			blockDataMap[blockToValidate.Hash] = blockData{nil, nil, nil, nil, blockToValidate}

			postValidate(blockDataMap[blockToValidate.Hash], true)
		}

		logger.Printf("Validated block with height %v\n", blockToValidate.Height)

		FileConnections.WriteString(fmt.Sprintf("'%x' -> '%x'\n",blockToValidate.PrevHash[0:15],blockToValidate.Hash[0:15]))

		//Set the last validated block as the lastBlock
		lastBlock = blockToValidate
	}

	logger.Printf("%v block(s) validated. Chain good to go.", len(storage.AllClosedBlocksAsc))
	//file.Close()
	return nil
}

func accStateChange(txSlice []*protocol.ContractTx) {
	for _, tx := range txSlice {
		acc, _ := storage.ReadAccount(tx.PubKey)
		if acc == nil {
			newAcc := protocol.NewAccount(tx.PubKey, tx.Issuer, 0, false, [crypto.COMM_KEY_LENGTH]byte{}, tx.Contract, tx.ContractVariables)
			storage.WriteAccount(&newAcc)
		}
	}
}

func fundsStateChange(txSlice []*protocol.FundsTx) (err error) {
	for _, tx := range txSlice {
		var rootAcc *protocol.Account
		//Check if we have to issue new coins (in case a root account signed the tx)
		if rootAcc, err = storage.ReadRootAccount(tx.From); err != nil {
			return err
		}

		if rootAcc != nil && rootAcc.Balance+tx.Amount+tx.Fee > MAX_MONEY {
			return errors.New("transaction amount would lead to balance overflow at the receiver (root) account")
		}

		//Will not be reached if errors occured
		if rootAcc != nil {
			rootAcc.Balance += tx.Amount
			rootAcc.Balance += tx.Fee
		}

		accSender, _ := storage.ReadAccount(tx.From)
		if accSender == nil {
			newFromAcc := protocol.NewAccount(tx.From, [64]byte{}, 0, false, [crypto.COMM_KEY_LENGTH]byte{}, nil, nil)
			accSender = &newFromAcc
			storage.WriteAccount(accSender)
		}

		accReceiver, _ := storage.ReadAccount(tx.To)
		if accReceiver == nil {
			newToAcc := protocol.NewAccount(tx.To, [64]byte{}, 0, false, [crypto.COMM_KEY_LENGTH]byte{}, nil, nil)
			accReceiver = &newToAcc
			storage.WriteAccount(accReceiver)
		}

		//Check transaction counter
		if tx.TxCnt != accSender.TxCnt {
			err = errors.New(fmt.Sprintf("sender txCnt does not match: %v (tx.txCnt) vs. %v (state txCnt)", tx.TxCnt, accSender.TxCnt))
		}

		//TODO @Kürsat: Replace this check with the included MPT proof of the transaction
		//Check sender balance
		if (tx.Amount + tx.Fee) > accSender.Balance {
			err = errors.New(fmt.Sprintf("sender does not have enough funds for the transaction: Balance = %v, Amount = %v, Fee = %v", accSender.Balance, tx.Amount, tx.Fee))
		}

		//After Tx fees, account must still have more than the minimum staking amount
		if accSender.IsStaking && ((tx.Fee + protocol.MIN_STAKING_MINIMUM + tx.Amount) > accSender.Balance) {
			err = errors.New("sender is staking and does not have enough funds in order to fulfill the required staking minimum")
		}

		//Overflow protection
		if tx.Amount+accReceiver.Balance > MAX_MONEY {
			err = errors.New("transaction amount would lead to balance overflow at the receiver account")
		}

		if err != nil {
			if rootAcc != nil {
				//Rollback root's credits if error occurs
				rootAcc.Balance -= tx.Amount
				rootAcc.Balance -= tx.Fee
			}

			return err
		}

		//We're manipulating pointer, no need to write back
		accSender.TxCnt += 1
		accSender.Balance -= tx.Amount
		accReceiver.Balance += tx.Amount
	}

	return nil
}

//We accept config slices with unknown id, but don't act on the payload. This is in case we have not updated to a new
//software with corresponding code to act on the configTx id/payload
func configStateChange(configTxSlice []*protocol.ConfigTx, blockHash [32]byte) {
	var newParameters Parameters
	//Initialize it to state right now (before validating config txs)
	newParameters = *activeParameters

	if len(configTxSlice) == 0 {
		return
	}

	//Only add a new parameter struct if a relevant system parameter changed
	if CheckAndChangeParameters(&newParameters, &configTxSlice) {
		newParameters.BlockHash = blockHash
		parameterSlice = append(parameterSlice, newParameters)
		activeParameters = &parameterSlice[len(parameterSlice)-1]
		logger.Printf("Config parameters changed. New configuration: %v", *activeParameters)
	}
}

func stakeStateChange(txSlice []*protocol.StakeTx, height uint32) (err error) {
	for _, tx := range txSlice {
		var accSender *protocol.Account
		accSender, err = storage.ReadAccount(tx.Account)

		//Check staking state
		if tx.IsStaking == accSender.IsStaking {
			err = errors.New("IsStaking state is already set to " + strconv.FormatBool(accSender.IsStaking) + ".")
		}

		//Check minimum amount
		if tx.IsStaking && accSender.Balance < tx.Fee+activeParameters.Staking_minimum {
			err = errors.New(fmt.Sprintf("Sender wants to stake but does not have enough funds (%v) in order to fulfill the required staking minimum (%v).", accSender.Balance, STAKING_MINIMUM))
		}

		//Check sender balance
		if tx.Fee > accSender.Balance {
			err = errors.New(fmt.Sprintf("Sender does not have enough funds for the transaction: Balance = %v, Amount = %v, Fee = %v.", accSender.Balance, 0, tx.Fee))
		}

		if err != nil {
			return err
		}

		//We're manipulating pointer, no need to write back
		accSender.IsStaking = tx.IsStaking
		accSender.CommitmentKey = tx.CommitmentKey
		accSender.StakingBlockHeight = height
	}

	return nil
}

func collectTxFees(contractTxSlice []*protocol.ContractTx, fundsTxSlice []*protocol.FundsTx, configTxSlice []*protocol.ConfigTx, stakeTxSlice []*protocol.StakeTx, minerAddress [64]byte) (err error) {
	var tmpContractTx []*protocol.ContractTx
	var tmpFundsTx []*protocol.FundsTx
	var tmpConfigTx []*protocol.ConfigTx
	var tmpStakeTx []*protocol.StakeTx

	minerAcc, err := storage.ReadAccount(minerAddress)
	if err != nil {
		return err
	}

	var senderAcc *protocol.Account

	for _, tx := range contractTxSlice {
		if minerAcc.Balance+tx.Fee > MAX_MONEY {
			err = errors.New("Fee amount would lead to balance overflow at the miner account.")
		}

		if err != nil {
			//Rollback of all perviously transferred transaction fees to the protocol's account
			collectTxFeesRollback(tmpContractTx, tmpFundsTx, tmpConfigTx, tmpStakeTx, minerAddress)
			return err
		}

		//Money gets created from thin air, no need to subtract money from root key
		minerAcc.Balance += tx.Fee
		tmpContractTx = append(tmpContractTx, tx)
	}

	//subtract fees from sender (check if that is allowed has already been done in the block validation)
	for _, tx := range fundsTxSlice {
		//Prevent protocol account from overflowing
		if minerAcc.Balance+tx.Fee > MAX_MONEY {
			err = errors.New("Fee amount would lead to balance overflow at the miner account.")
		}

		senderAcc, err = storage.ReadAccount(tx.From)

		if err != nil {
			//Rollback of all perviously transferred transaction fees to the protocol's account
			collectTxFeesRollback(tmpContractTx, tmpFundsTx, tmpConfigTx, tmpStakeTx, minerAddress)
			return err
		}

		minerAcc.Balance += tx.Fee
		senderAcc.Balance -= tx.Fee
		tmpFundsTx = append(tmpFundsTx, tx)
	}

	for _, tx := range configTxSlice {
		if minerAcc.Balance+tx.Fee > MAX_MONEY {
			err = errors.New("Fee amount would lead to balance overflow at the miner account.")
		}

		if err != nil {
			//Rollback of all perviously transferred transaction fees to the protocol's account
			collectTxFeesRollback(tmpContractTx, tmpFundsTx, tmpConfigTx, tmpStakeTx, minerAddress)
			return err
		}

		//No need to subtract money because signed by root account
		minerAcc.Balance += tx.Fee
		tmpConfigTx = append(tmpConfigTx, tx)
	}

	for _, tx := range stakeTxSlice {
		if minerAcc.Balance+tx.Fee > MAX_MONEY {
			err = errors.New("Fee amount would lead to balance overflow at the miner account.")
		}

		senderAcc, err = storage.ReadAccount(tx.Account)

		if err != nil {
			//Rollback of all perviously transferred transaction fees to the protocol's account
			collectTxFeesRollback(tmpContractTx, tmpFundsTx, tmpConfigTx, tmpStakeTx, minerAddress)
			return err
		}

		senderAcc.Balance -= tx.Fee
		minerAcc.Balance += tx.Fee
		tmpStakeTx = append(tmpStakeTx, tx)
	}

	return nil
}

func collectBlockReward(reward uint64, minerAddress [64]byte) (err error) {
	var miner *protocol.Account
	miner, err = storage.ReadAccount(minerAddress)

	if miner.Balance+reward > MAX_MONEY {
		err = errors.New("Block reward would lead to balance overflow at the miner account.")
	}

	if err != nil {
		return err
	}

	miner.Balance += reward

	return nil
}

func collectSlashReward(reward uint64, block *protocol.Block) (err error) {
	//Check if proof is provided. If proof was incorrect, prevalidation would already have failed.
	if block.SlashedAddress != [64]byte{} || block.ConflictingBlockHash1 != [32]byte{} || block.ConflictingBlockHash2 != [32]byte{} {
		var minerAcc, slashedAcc *protocol.Account
		minerAcc, err = storage.ReadAccount(block.Beneficiary)
		slashedAcc, err = storage.ReadAccount(block.SlashedAddress)

		if minerAcc.Balance+reward > MAX_MONEY {
			err = errors.New("Slash reward would lead to balance overflow at the miner account.")
		}

		if err != nil {
			return err
		}

		//Validator is rewarded with slashing reward for providing a valid slashing proof
		minerAcc.Balance += reward
		//Slashed account looses the minimum staking amount
		slashedAcc.Balance -= activeParameters.Staking_minimum
		//Slashed account is being removed from the validator set
		slashedAcc.IsStaking = false
	}

	return nil
}

//No rollback method exists
func updateStakingHeight(block *protocol.Block) error {
	acc, err := storage.ReadAccount(block.Beneficiary)
	if err != nil {
		return err
	}

	acc.StakingBlockHeight = block.Height

	return nil
}

func deleteZeroBalanceAccounts() error {
	for _, acc := range storage.State {
		if acc.Balance > 0 || acc.Contract != nil {
			continue
		}

		storage.DeleteAccount(acc.Address)
	}

	return nil
}
