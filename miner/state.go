package miner

import (
	"errors"
	"fmt"
	"github.com/sfontanach/bazo-miner/protocol"
	"github.com/sfontanach/bazo-miner/storage"
	"strconv"
)

func accStateChange(txSlice []*protocol.AccTx) error {
	for _, tx := range txSlice {
		if tx.Header == 0 || tx.Header == 1 || tx.Header == 2 {
			newAcc := protocol.NewAccount(tx.PubKey, 0, false, [32]byte{})
			newAccHash := newAcc.Hash()
			if acc := storage.GetAccount(newAccHash); acc != nil {
				//Shouldn't happen, because this should have been prevented when adding an accTx to the block
				errors.New("Address already exists in the state.")
			}
			//If acc does not exist, write to state
			storage.State[newAccHash] = &newAcc

			switch tx.Header {
			case 1:
				//First bit set, given account will be a new root account
				//It might be cleaner to move this to the storage package (e.g., storage.Delete(...))
				//leave it here for now (not fully convinced yet)
				storage.RootKeys[newAccHash] = &newAcc
			case 2:
				//Second bit set, delete account from root account
				delete(storage.RootKeys, newAccHash)
			}
		}
	}

	return nil
}

func fundsStateChange(txSlice []*protocol.FundsTx) (err error) {
	for index, tx := range txSlice {
		//Check if we have to issue new coins (in case a root account signed the tx)
		if rootAcc := storage.RootKeys[tx.From]; rootAcc != nil {
			if rootAcc.Balance+tx.Amount+tx.Fee > MAX_MONEY {
				err = errors.New("Root account has max amount of coins reached.")
			}
			rootAcc.Balance += tx.Amount
			rootAcc.Balance += tx.Fee
		}

		accSender, accReceiver := storage.GetAccount(tx.From), storage.GetAccount(tx.To)
		if accSender == nil {
			err = errors.New("Sender does not exist in the State.")
		}

		if accReceiver == nil {
			err = errors.New("Receiver does not exist in the State.")
		}

		//Check transaction counter
		if tx.TxCnt != accSender.TxCnt {
			err = errors.New(fmt.Sprintf("Sender txCnt does not match: %v (tx.txCnt) vs. %v (state txCnt).", tx.TxCnt, accSender.TxCnt))
		}

		//Check sender balance
		if (tx.Amount + tx.Fee) > accSender.Balance {
			err = errors.New(fmt.Sprintf("Sender does not have enough funds for the transaction: Balance = %v, Amount = %v, Fee = %v.", accSender.Balance, tx.Amount, tx.Fee))
		}

		//TODO Please check & verify if statement:
		//if accSender.IsStaking && !((tx.Fee + protocol.MIN_STAKING_MINIMUM) > accSender.Balance) to
		//if accSender.IsStaking && ((tx.Fee + protocol.MIN_STAKING_MINIMUM) > accSender.Balance)
		//After Tx fees, account must still have more than the minimum staking amount
		if accSender.IsStaking && ((tx.Fee + protocol.MIN_STAKING_MINIMUM + tx.Amount) > accSender.Balance) {
			err = errors.New("Sender is staking and does not have enough funds in order to fulfill the required staking minimum.")
		}

		//Overflow protection
		if tx.Amount+accReceiver.Balance > MAX_MONEY {
			err = errors.New("Transaction amount would lead to balance overflow at the receiver account.")
		}

		if err != nil {
			//If it was the first tx in the block, no rollback needed
			if index == 0 {
				return err
			}
			fundsStateChangeRollback(txSlice[0 : index-1])
			storage.DeleteOpenTx(tx)
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

func stakeStateChange(txSlice []*protocol.StakeTx, height uint32) error {
	for index, tx := range txSlice {
		var err error

		//Check if we have to issue new coins (in case a root account signed the tx)
		for hash, rootAcc := range storage.RootKeys {
			if hash == tx.Account {
				if rootAcc.Balance+tx.Fee > MAX_MONEY {
					err = errors.New("Account balance is too high.")
				}
				//fmt.Println("Account Balance:", rootAcc.Balance, "TxFee:", tx.Fee, "minStake:", activeParameters.Staking_minimum)

				if rootAcc.Balance < activeParameters.Staking_minimum+tx.Fee {
					rootAcc.Balance = activeParameters.Staking_minimum + tx.Fee
					fmt.Println("Root balance increased to:", rootAcc.Balance)
				}
			}
		}

		accSender := storage.GetAccount(tx.Account)
		if accSender == nil {
			logger.Printf("CRITICAL: Sender does not exist in the State: %x\n", tx.Account[0:8])
			err = errors.New("Sender does not exist in the State.")
		}

		//Check staking state
		if tx.IsStaking == accSender.IsStaking {
			logger.Printf("IsStaking state is already set to " + strconv.FormatBool(accSender.IsStaking) + ".")
			err = errors.New("IsStaking state is already set to " + strconv.FormatBool(accSender.IsStaking) + ".")
		}

		//Check minimum amount
		if tx.IsStaking && accSender.Balance < tx.Fee+activeParameters.Staking_minimum {
			logger.Print("Sender wants to stake but does not have enough funds in order to fulfill the required staking minimum: %x\n", accSender.Balance)
			err = errors.New("Sender wants to stake but does not have enough funds in order to fulfill the required staking minimum.")
		}

		//Check sender balance
		if tx.Fee > accSender.Balance {
			logger.Printf("Sender does not have enough balance: %x\n", accSender.Balance)
			err = errors.New("Sender does not have enough funds for the transaction.")
		}

		if err != nil {
			//If it was the first tx in the block, no rollback needed
			if index == 0 {
				return err
			}
			stakeStateChangeRollback(txSlice[0 : index-1])
			return err
		}

		//We're manipulating pointer, no need to write back

		accSender.IsStaking = tx.IsStaking
		accSender.HashedSeed = tx.HashedSeed
		accSender.StakingBlockHeight = height
	}

	return nil
}

func collectTxFees(accTxSlice []*protocol.AccTx, fundsTxSlice []*protocol.FundsTx, configTxSlice []*protocol.ConfigTx, stakeTxSlice []*protocol.StakeTx, minerHash [32]byte) error {
	var tmpAccTx []*protocol.AccTx
	var tmpFundsTx []*protocol.FundsTx
	var tmpConfigTx []*protocol.ConfigTx
	var tmpStakeTx []*protocol.StakeTx

	minerAcc := storage.GetAccount(minerHash)

	for _, tx := range accTxSlice {
		if minerAcc.Balance+tx.Fee > MAX_MONEY {
			//Rollback of all perviously transferred transaction fees to the protocol's account
			collectTxFeesRollback(tmpAccTx, tmpFundsTx, tmpConfigTx, tmpStakeTx, minerHash)
			logger.Printf("Miner balance (%v) overflows with transaction fee (%v).\n", minerAcc.Balance, tx.Fee)
			return errors.New("Miner balance overflows with transaction fee.\n")
		}

		//Money gets created from thin air, no need to subtract money from root key
		minerAcc.Balance += tx.Fee
		tmpAccTx = append(tmpAccTx, tx)
	}

	//subtract fees from sender (check if that is allowed has already been done in the block validation)
	for _, tx := range fundsTxSlice {
		//Prevent protocol account from overflowing
		if minerAcc.Balance+tx.Fee > MAX_MONEY {
			//Rollback of all perviously transferred transaction fees to the protocol's account
			collectTxFeesRollback(tmpAccTx, tmpFundsTx, tmpConfigTx, tmpStakeTx, minerHash)
			return errors.New("Miner balance overflows with transaction fee.\n")
		}
		minerAcc.Balance += tx.Fee

		senderAcc := storage.GetAccount(tx.From)
		senderAcc.Balance -= tx.Fee

		tmpFundsTx = append(tmpFundsTx, tx)
	}

	for _, tx := range configTxSlice {
		if minerAcc.Balance+tx.Fee > MAX_MONEY {
			//Rollback of all perviously transferred transaction fees to the protocol's account
			collectTxFeesRollback(tmpAccTx, tmpFundsTx, tmpConfigTx, tmpStakeTx, minerHash)
			logger.Printf("Miner balance (%v) overflows with transaction fee (%v).\n", minerAcc.Balance, tx.Fee)
			return errors.New("Miner balance overflows with transaction fee.\n")
		}

		//No need to subtract money because signed by root account
		minerAcc.Balance += tx.Fee
		tmpConfigTx = append(tmpConfigTx, tx)
	}

	for _, tx := range stakeTxSlice {
		if minerAcc.Balance+tx.Fee > MAX_MONEY {
			//Rollback of all previously transferred transaction fees to the protocol's account
			collectTxFeesRollback(tmpAccTx, tmpFundsTx, tmpConfigTx, tmpStakeTx, minerHash)
			logger.Printf("Miner balance (%v) overflows with transaction fee (%v).\n", minerAcc.Balance, tx.Fee)
			return errors.New("Miner balance overflows with transaction fee.\n")
		}

		senderAcc := storage.GetAccount(tx.Account)
		senderAcc.Balance -= tx.Fee

		minerAcc.Balance += tx.Fee
		tmpStakeTx = append(tmpStakeTx, tx)
	}

	return nil
}

func collectBlockReward(reward uint64, minerHash [32]byte) error {
	miner := storage.GetAccount(minerHash)

	if miner == nil {
		return errors.New("Miner doesn't exist in the state!")
	}

	if miner.Balance+reward > MAX_MONEY {
		logger.Printf("Miner balance (%v) overflows with block reward (%v).\n", miner.Balance, reward)
		return errors.New("Miner balance overflows with transaction fee.\n")
	}
	miner.Balance += reward
	return nil
}

func collectSlashReward(reward uint64, block *protocol.Block) error {
	miner := storage.GetAccount(block.Beneficiary)
	if miner == nil {
		return errors.New("Miner doesn't exist in the state!")
	}

	if miner.Balance+reward > MAX_MONEY {
		logger.Printf("Miner balance (%v) overflows with block reward (%v).\n", miner.Balance, reward)
		return errors.New("Miner balance overflows with transaction fee.\n")
	}

	//check if proof is provided. if proof was incorrect, the prevalidation step would already have failed.
	if block.SlashedAddress != [32]byte{} || block.ConflictingBlockHash1 != [32]byte{} || block.ConflictingBlockHash2 != [32]byte{} {

		//validator is rewarded with slashing reward for providing a valid slashing proof
		miner.Balance += reward

		slashedAccount := storage.GetAccount(block.SlashedAddress)
		//slashed account looses the minimum staking amount
		slashedAccount.Balance -= activeParameters.Staking_minimum
		//slashed account is being removed from the validator set
		slashedAccount.IsStaking = false
	}

	return nil
}

//For logging purposes
func getState() (state string) {
	for _, acc := range storage.State {
		state += fmt.Sprintf("Is root: %v, %v\n", storage.IsRootKey(acc.Hash()), acc)
	}
	return state
}

func SetUpInitialState(hashedSeed [32]byte) (block *protocol.Block, err error) {
	var seed [32]byte
	copy(seed[:], storage.GENESIS_SEED)

	initialBlock := newBlock([32]byte{}, seed, hashedSeed, 0)

	allClosedBlocks := storage.ReadAllClosedBlocks()

	//Switch order in array to validate genesis block first
	storage.AllClosedBlocksAsc = InvertBlockArray(allClosedBlocks)

	//Set the last closed block as the initial block
	if len(storage.AllClosedBlocksAsc) > 0 {
		initialBlock = storage.AllClosedBlocksAsc[len(storage.AllClosedBlocksAsc)-1]
	}

	collectStatistics(initialBlock)

	if initialBlock.Hash == [32]byte{} {
		storage.WriteClosedBlock(initialBlock)
	}

	blockDataMap := make(map[[32]byte]blockData)

	//Validate all closed blocks and update state
	for _, blockToValidate := range storage.AllClosedBlocksAsc {
		//Do not validate the genesis block
		if blockToValidate.Hash != [32]byte{} {
			accTxs := make([]*protocol.AccTx, blockToValidate.NrAccTx)
			fundsTxs := make([]*protocol.FundsTx, blockToValidate.NrFundsTx)
			configTxs := make([]*protocol.ConfigTx, blockToValidate.NrConfigTx)
			stakeTxs := make([]*protocol.StakeTx, blockToValidate.NrStakeTx)

			for cnt, txHash := range blockToValidate.AccTxData {
				var tx protocol.Transaction
				var accTx *protocol.AccTx
				tx = storage.ReadClosedTx(txHash)
				if tx != nil {
					accTx = tx.(*protocol.AccTx)
				}

				accTxs[cnt] = accTx
			}

			for cnt, txHash := range blockToValidate.FundsTxData {
				var tx protocol.Transaction
				var fundsTx *protocol.FundsTx
				tx = storage.ReadClosedTx(txHash)
				if tx != nil {
					fundsTx = tx.(*protocol.FundsTx)
				}

				fundsTxs[cnt] = fundsTx
			}

			for cnt, txHash := range blockToValidate.ConfigTxData {
				var tx protocol.Transaction
				var configTx *protocol.ConfigTx
				tx = storage.ReadClosedTx(txHash)
				if tx != nil {
					configTx = tx.(*protocol.ConfigTx)
				}

				configTxs[cnt] = configTx
			}

			for cnt, txHash := range blockToValidate.StakeTxData {
				var tx protocol.Transaction
				var stakeTx *protocol.StakeTx
				tx = storage.ReadClosedTx(txHash)
				if tx != nil {
					stakeTx = tx.(*protocol.StakeTx)
				}

				stakeTxs[cnt] = stakeTx
			}

			blockDataMap[blockToValidate.Hash] = blockData{accTxs, fundsTxs, configTxs, stakeTxs, blockToValidate}
			if err := stateValidation(blockDataMap[blockToValidate.Hash]); err != nil {
				return nil, errors.New(fmt.Sprintf("Received block (%x) could not be validated: %v", blockToValidate.Hash[0:8], err))
			}

			configStateChange(blockDataMap[blockToValidate.Hash].configTxSlice, blockDataMap[blockToValidate.Hash].block.Hash)
		}
	}

	logger.Printf("%v blocks up to this date are validated. Chain good to go.", len(storage.AllClosedBlocksAsc))

	return initialBlock, nil
}

func updateStakingHeight(beneficiary [32]byte, height uint32) (err error) {
	acc := storage.GetAccount(beneficiary)
	if err != nil{
		return err
	}
	acc.StakingBlockHeight = height

	return err
}
