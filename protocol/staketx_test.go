package protocol

import (
	"io/ioutil"
	"log"
	"math/rand"
	"os"
	"reflect"
	"testing"
	"time"
)

func TestStakeTxSerialization(t *testing.T) {
	rand := rand.New(rand.NewSource(time.Now().Unix()))
	accAHash := SerializeHashContent(accA.Address)
	loopMax := int(rand.Uint32() % 10000)
	for i := 0; i < loopMax; i++ {
		fee := rand.Uint64()%10 + 1
		isStaking := rand.Intn(2) != 0
		seed := createRandomSeed()
		hashedSeed := SerializeHashContent(seed)

		tx, _ := ConstrStakeTx(0x01, fee, isStaking, hashedSeed, accAHash, &PrivKeyA)
		data := tx.Encode()
		var decodedTx *StakeTx
		decodedTx = decodedTx.Decode(data)

		//this is done by verify() which is outside protocol package, we're just testing serialization here
		decodedTx.Fee = fee
		decodedTx.Account = accAHash
		decodedTx.IsStaking = isStaking
		decodedTx.HashedSeed = hashedSeed

		if !reflect.DeepEqual(tx, decodedTx) {
			t.Errorf("StakeTx Serialization failed (%v) vs. (%v)\n", tx, decodedTx)
		}
	}
}

func TestSeedCreation(t *testing.T) {
	//generate random seed and store it in a file called as args[3]
	seed := createRandomSeed()
	file, _ := os.Create("testseed")

	_, _ = file.WriteString(string(seed[:]) + "\n")

	seedFromFile, _ := readSeed("testseed")

	if !reflect.DeepEqual(seed, seedFromFile) {
		t.Errorf("Seed creation and loading failed (%v) vs. (%v)", seed, seedFromFile)
	}
}

func createRandomSeed() [32]byte {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	const chars = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	var seed [32]byte
	for i := range seed {
		seed[i] = chars[r.Intn(len(chars))]
	}
	return seed
}

func readSeed(fileName string) ([32]byte, error) {
	var (
		seedByte []byte
		seed     [32]byte
		err      error
	)

	seedByte, err = ioutil.ReadFile(fileName)

	copy(seed[0:32], seedByte[0:32])

	if err != nil {
		log.Fatal(err)
		return seed, err
	}

	return seed, nil
}
