package storage

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rsa"
	"github.com/bazo-blockchain/bazo-miner/crypto"
	"github.com/bazo-blockchain/bazo-miner/protocol"
	"io/ioutil"
	"log"
	"math/big"
	"os"
	"testing"
)

const (
	TestIpPort     = "127.0.0.1:8000"
	TestDBFileName = "test.db"
)

const (
	PubA1 = "c2be9abbeaec39a066c2a09cee23bb9ab2a0b88f2880b1e785b4d317adf0dc7c"
	PubA2 = "8ce020fde838d9c443f6c93345dafe7fd74f091c4d2f30b37e2453679a257ed5"
	PrivA = "ba127fa8f802b008b9cdb58f4e44809d48f1b000cff750dda9cd6b312395c1c5"
	CommPubA = "vsl0yfAd3dqJfDEawAl7Xp2/hOvXGN/u0UXBpRSxWAT+FSKlt5Ha8Ibd59tGkM4D8i/MABx0MMNEVL8Ghe1QkXIITJRnFtoqsidTlcHSL4WL7sc+8LwJjIjMdqM5BYIZJap/j2O2qREcEICEN8i+6LF844iMqFysDOuL8F5MAH22twrh0SMVXAM+IAEqa0Z9TymvX8Op3dt/t5IhrA4ivsS/+QMWzr3xJE9XfQxMrDUNoBwXIszOr656m8/wYa9dOEZn8qlglEySAjievkECJZq9Q3DRat5SUoXjG8M6UJp/AeRUUANsXhrPn6Cg7j4ke5Yw0bk6Lz9foYaZ9rugkw=="
	CommPrivA = "n5Xdlei+4sshA3wDlyyXQF6NS78GTi1KE0zZHJ/BdBHBAqbXnURosZbuWTmmvgtFa7ilWFZ0rjE3n/elmjMWmIKdBImB7bCR1DFnDjZw/QUlNpb9Q9rV1fK7rGT9lmjrZgFG8AcFTEgehIMrlYnafsOv5pdaqJ3T4H7KsEYAJsuNZhHAFmReqNdeiUbdAntPLQjttbs43DqaVQ0D3YnHrKxeu7Ekwcs4ap18tkFt7Lp0mkJ3fjpsvJFPDP2CotrZadLilv7dmOrXe26XDLUQ2aBguExV4Wx85J29puOJwpoM60KiFgiBMtRQFRukzRuValiVkXEBLZKlbh6wYy0vwQ=="
	CommPrim1A = "9UbkVH5chUZCZaehntnZWAfTJ9OYvsKfu19Cb39RrBZ9FDMjDoBlKZslyvRzTez33An84JAgwOBtEbaSTAkVqvPmDin3oZhTYbwwDc9SIBVsYhI6VmbjcPkMAFIeoKbS4KzweXneeKBB9FbozcgvnYrv3lTqofVVWONY/EL9q7M="
	CommPrim2A = "xyC4Jl6ojvL+uF2/iK9kRj3yQh8bV2ngl/fongysmUvxCZrwxaEaOZcBHreTiP6SFPOrWCyk6e9zHjtDPP/LhxrHsaiFapv6AjQejML/gCyFj4GRWMzayFBJlW6prsjZfhNG6FpQbFrEj8FtYdM0vRLyDyzeknrC66PJtwEcR6E="
	PubB1 = "5d7eefd58e3d2f309471928ab4bbd104e52973372c159fa652b8ca6b57ff68b8"
	PubB2 = "ab301a6a77b201c416ddc13a2d33fdf200a5302f6f687e0ea09085debaf8a1d9"
	PrivB = "7a0a9babcc97ea7991ed67ed7f800f70c5e04e99718960ad8efab2ca052f00c7"
)

//Root account for testing
const (
	RootPub1 = "6323cc034597195ae69bcfb628ecdffa5989c7503154c566bab4a87f3e9910ac"
	RootPub2 = "f6115b77a15852764c609c6a5c1739e698ebc6e49bf14617c561b9110039cec7"
	RootPriv = "277ed539f56122c25a6fc115d07d632b47e71416c9aebf1beb54ee704f11842c"
)

var (
	accA, accB, minerAcc, rootAcc *protocol.Account
	PrivKeyA, PrivKeyB ecdsa.PrivateKey
	PubKeyA, PubKeyB ecdsa.PublicKey
	CommitmentKeyA 	*rsa.PrivateKey
	RootPrivKey ecdsa.PrivateKey
)

func TestMain(m *testing.M) {

	Init(TestDBFileName, TestIpPort)

	DeleteAll()
	addTestingAccounts()
	addRootAccounts()
	//we don't want logging msgs when testing, designated messages
	log.SetOutput(ioutil.Discard)
	os.Exit(m.Run())

	TearDown()
}

func addTestingAccounts() {

	accA, accB, minerAcc = new(protocol.Account), new(protocol.Account), new(protocol.Account)

	puba1, _ := new(big.Int).SetString(PubA1, 16)
	puba2, _ := new(big.Int).SetString(PubA2, 16)
	priva, _ := new(big.Int).SetString(PrivA, 16)
	PubKeyA = ecdsa.PublicKey{
		elliptic.P256(),
		puba1,
		puba2,
	}
	PrivKeyA = ecdsa.PrivateKey{
		PubKeyA,
		priva,
	}

	CommitmentKeyA, _ = crypto.CreateRSAPrivKeyFromBase64(CommPubA, CommPrivA, []string{CommPrim1A, CommPrim2A})

	pubb1, _ := new(big.Int).SetString(PubB1, 16)
	pubb2, _ := new(big.Int).SetString(PubB2, 16)
	privb, _ := new(big.Int).SetString(PrivB, 16)
	PubKeyB = ecdsa.PublicKey{
		elliptic.P256(),
		pubb1,
		pubb2,
	}
	PrivKeyB = ecdsa.PrivateKey{
		PubKeyB,
		privb,
	}

	copy(accA.Address[0:32], PrivKeyA.PublicKey.X.Bytes())
	copy(accA.Address[32:64], PrivKeyA.PublicKey.Y.Bytes())
	accAHash := protocol.SerializeHashContent(accA.Address)

	//This one is just for testing purposes
	copy(accB.Address[0:32], PrivKeyB.PublicKey.X.Bytes())
	copy(accB.Address[32:64], PrivKeyB.PublicKey.Y.Bytes())
	accBHash := protocol.SerializeHashContent(accB.Address)

	State[accAHash] = accA
	State[accBHash] = accB
}

func addRootAccounts() {

	var pubKey [64]byte

	pub1, _ := new(big.Int).SetString(RootPub1, 16)
	pub2, _ := new(big.Int).SetString(RootPub2, 16)
	priv, _ := new(big.Int).SetString(RootPriv, 16)
	PubKeyA = ecdsa.PublicKey{
		elliptic.P256(),
		pub1,
		pub2,
	}
	RootPrivKey = ecdsa.PrivateKey{
		PubKeyA,
		priv,
	}

	copy(pubKey[32-len(pub1.Bytes()):32], pub1.Bytes())
	copy(pubKey[64-len(pub2.Bytes()):], pub2.Bytes())

	rootHash := protocol.SerializeHashContent(pubKey)

	rootAcc = new(protocol.Account)
	rootAcc.Address = pubKey

	State[rootHash] = rootAcc
	RootKeys[rootHash] = rootAcc
}
