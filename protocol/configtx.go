package protocol

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/rand"
	"encoding/binary"
	"fmt"
)

const (
	CONFIGTX_SIZE = 83

	BLOCK_SIZE_ID           = 1
	DIFF_INTERVAL_ID        = 2
	FEE_MINIMUM_ID          = 3
	BLOCK_INTERVAL_ID       = 4
	BLOCK_REWARD_ID         = 5
	STAKING_MINIMUM_ID      = 6
	WAITING_MINIMUM_ID      = 7
	ACCEPTANCE_TIME_DIFF_ID = 8
	SLASHING_WINDOW_SIZE_ID = 9
	SLASHING_REWARD_ID      = 10
	CONSOLIDATION_INTERVAL  = 11

	MIN_BLOCK_SIZE = 1000      //1KB
	MAX_BLOCK_SIZE = 100000000 //100MB

	MIN_DIFF_INTERVAL = 3 //amount in seconds
	MAX_DIFF_INTERVAL = 9223372036854775807

	MIN_FEE_MINIMUM = 0
	MAX_FEE_MINIMUM = 9223372036854775807

	MIN_BLOCK_INTERVAL = 3     //30 seconds
	MAX_BLOCK_INTERVAL = 86400 //24 hours

	MIN_BLOCK_REWARD = 0
	MAX_BLOCK_REWARD = 1152921504606846976 //2^60

	MIN_STAKING_MINIMUM = 5                   //minimum number of coins for staking
	MAX_STAKING_MINIMUM = 9223372036854775807 //2^60

	MIN_WAITING_TIME = 0 //number of blocks that must a new validator must wait before it can start validating
	MAX_WAITING_TIME = 100000

	MIN_ACCEPTANCE_TIME_DIFF = 0  //semi-synchronous time difference between local clock of validators
	MAX_ACCEPTANCE_TIME_DIFF = 60 //1min

	MIN_SLASHING_WINDOW_SIZE = 0     //window size where a node has to commit itself to a competing branch in case of a fork
	MAX_SLASHING_WINDOW_SIZE = 10000 //1000 Blocks (totally random)

	MIN_SLASHING_REWARD = 0                   // reward for providing a valid slashing proof
	MAX_SLASHING_REWARD = 1152921504606846976 //2^60

	MIN_CONS_INTERVAL = 0  // 0 to disable consolidation
	MAX_CONS_INTERVAL = 9223372036854775807 //2^60
)

type ConfigTx struct {
	Header  byte
	Id      uint8
	Payload uint64
	Fee     uint64
	TxCnt   uint8
	Sig     [64]byte
}

func ConstrConfigTx(header byte, id uint8, payload uint64, fee uint64, txCnt uint8, rootPrivKey *ecdsa.PrivateKey) (tx *ConfigTx, err error) {

	tx = new(ConfigTx)
	tx.Header = header
	tx.Id = id
	tx.Payload = payload
	tx.Fee = fee
	tx.TxCnt = txCnt

	txHash := tx.Hash()

	r, s, err := ecdsa.Sign(rand.Reader, rootPrivKey, txHash[:])

	if err != nil {
		return nil, err
	}

	copy(tx.Sig[32-len(r.Bytes()):32], r.Bytes())
	copy(tx.Sig[64-len(s.Bytes()):], s.Bytes())

	return tx, nil
}

func (tx *ConfigTx) Hash() (hash [32]byte) {

	if tx == nil {
		return [32]byte{}
	}

	txHash := struct {
		Header  byte
		Id      uint8
		Payload uint64
		Fee     uint64
		TxCnt   uint8
	}{
		tx.Header,
		tx.Id,
		tx.Payload,
		tx.Fee,
		tx.TxCnt,
	}
	return SerializeHashContent(txHash)
}

func (tx *ConfigTx) Encode() (encodedTx []byte) {

	if tx == nil {
		return nil
	}

	var buf bytes.Buffer
	var payloadBuf [8]byte
	var feeBuf [8]byte

	binary.Write(&buf, binary.BigEndian, tx.Payload)
	copy(payloadBuf[:], buf.Bytes())
	buf.Reset()
	binary.Write(&buf, binary.BigEndian, tx.Fee)
	copy(feeBuf[:], buf.Bytes())
	buf.Reset()

	encodedTx = make([]byte, CONFIGTX_SIZE)
	encodedTx[0] = tx.Header
	encodedTx[1] = tx.Id
	copy(encodedTx[2:10], payloadBuf[:])
	copy(encodedTx[10:18], feeBuf[:])
	encodedTx[18] = byte(tx.TxCnt)
	copy(encodedTx[19:83], tx.Sig[:])

	return encodedTx
}

func (*ConfigTx) Decode(encodedTx []byte) (tx *ConfigTx) {

	if len(encodedTx) != CONFIGTX_SIZE {
		return nil
	}

	tx = new(ConfigTx)
	tx.Header = encodedTx[0]
	tx.Id = encodedTx[1]
	tx.Payload = binary.BigEndian.Uint64(encodedTx[2:10])
	tx.Fee = binary.BigEndian.Uint64(encodedTx[10:18])
	tx.TxCnt = uint8(encodedTx[18])
	copy(tx.Sig[:], encodedTx[19:83])

	return tx
}

func (tx *ConfigTx) TxFee() uint64 { return tx.Fee }
func (tx *ConfigTx) Size() uint64  { return CONFIGTX_SIZE }

func (tx ConfigTx) String() string {
	return fmt.Sprintf(
		"\n"+
			"Id: %v\n"+
			"Payload: %v\n"+
			"Fee: %v\n"+
			"TxCnt: %v\n",
		tx.Id,
		tx.Payload,
		tx.Fee,
		tx.TxCnt,
	)
}
