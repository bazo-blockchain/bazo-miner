package protocol

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/rsa"
	"io/ioutil"
	"log"
	"math/big"
	"os"
	"testing"
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
	CommPubB = "n7vb+4YNgTDwjJ1St3/UQP+bXrN/mMmsPgTjKthIMpoMYN7mRhpk6/MGa6Gv0p1Zbw39g6fVsluHSXvyYO6VmsahTQ0gI9MEmxgKt4c6ZQct6M+kWP7E3omXT68NsXXXaZBjBuewfHrJReTz/znbS66HgY8BML55YDRKQBsmDz+cb/H6FWT7/mmPBRXufz7sf6OqvwiMRGXNlRbktbEn3gpumXpndlGhmGL0ZVZj2VklqWSHtgsfBut+rov7uuIN28StPZYZvllnCCvP1DHeImExWHOltWTnZAE0pRUbaX3q3NVAqU4ngL1sbkMSghF8bmz8G26qawM7YNiiDrAmcQ=="
	CommPrivB = "P0og9Hz99tVcSmq/boOQpxxgBFrc0L3/qCcplz1RBfOxueQ3m0kz+aU2QwkycCH2YKFLdJHYgy3u4bfhpnSCBGx1VuE/fdJLfeQ9wtAq3ALHNvqm5Lg1avNbZ7A1nb3SVzplckP00q2X+ECqSNM0x7zkZfoyf4zI7MxrKxFWuC1c1BT7zj7EUT1idG+n/yz3WCx4Xr+4XM3CIt1dTrddhCboLdLlNYCOIh4t5JSTfYysp8YR4FSc96vRVCe+QCVtMOfo7RCR8bcZDoIQjat+u5umnyAsyXLetBerh/MABqHq8wOgC6a8vCqRnyAwhLOT+VQbTbFMQzLO9Lw9T8v9EQ=="
	CommPrim1B = "zzoomDPTH/WxxtqTIApnecinr+BuAhcxJephkDHOhlRWK1IH16yLIal9V6OmC/REGCLgZpJHzUeesATn7QnsTIFnmEDKxIPVk54etYAXJo8G51pB9mylTUJiXqY1hu5O1GSEgtD+EAdHrIRJZ2E7/Pyp/wFrLG5ymXULZ5BFicU="
	CommPrim2B = "xVQiK60JgjOqSdQ6EKEjDxdxfWp1NDGpHbBLElzbqJyAHfd5KCPdwASLIR8V0WHIa12df877xGGL1W+SlXXXOsJaER+FfnlzxzaO5D8a3GqaYMJBYWyUBnf1f0/lgVvnJzh0hKHlKSlvJX2mbObD9mPeuYhEXNO1v7Vo0846sL0="
)

const (
	//P-256
	RootPub1 = "6323cc034597195ae69bcfb628ecdffa5989c7503154c566bab4a87f3e9910ac"
	RootPub2 = "f6115b77a15852764c609c6a5c1739e698ebc6e49bf14617c561b9110039cec7"
	RootPriv = "277ed539f56122c25a6fc115d07d632b47e71416c9aebf1beb54ee704f11842c"
)

var (
	accA, accB, minerAcc 			*Account
	PrivKeyA, PrivKeyB   			ecdsa.PrivateKey
	PubKeyA, PubKeyB     			ecdsa.PublicKey
	RootPrivKey          			ecdsa.PrivateKey
	CommitmentKeyA, CommitmentKeyB 	rsa.PrivateKey
	MinerHash            			[32]byte
	MinerPrivKey         			*ecdsa.PrivateKey
)

func TestMain(m *testing.M) {

	addTestingAccounts()
	addRootAccounts()

	log.SetOutput(ioutil.Discard)
	os.Exit(m.Run())
}

func addTestingAccounts() {

	accA, accB, minerAcc = new(Account), new(Account), new(Account)

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

	CommitmentKeyA, _ = CreateRSAPrivKeyFromBase64(CommPubA, CommPrivA, []string{CommPrim1A, CommPrim2A})

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

	CommitmentKeyB, _ = CreateRSAPrivKeyFromBase64(CommPubB, CommPrivB, []string{CommPrim1B, CommPrim2B})

	copy(accA.Address[0:32], PrivKeyA.PublicKey.X.Bytes())
	copy(accA.Address[32:64], PrivKeyA.PublicKey.Y.Bytes())
	copy(accA.CommitmentKey[:], CommitmentKeyA.N.Bytes())
	accAHash := SerializeHashContent(accA.Address)

	//This one is just for testing purposes
	copy(accB.Address[0:32], PrivKeyB.PublicKey.X.Bytes())
	copy(accB.Address[32:64], PrivKeyB.PublicKey.Y.Bytes())
	copy(accB.CommitmentKey[:], CommitmentKeyB.N.Bytes())
	accBHash := SerializeHashContent(accB.Address)

	//just to bootstrap
	var shortHashA [8]byte
	var shortHashB [8]byte
	copy(shortHashA[:], accAHash[0:8])
	copy(shortHashB[:], accBHash[0:8])

	MinerPrivKey, _ = ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	var pubKey [64]byte
	var shortMiner [8]byte
	copy(pubKey[:32], MinerPrivKey.X.Bytes())
	copy(pubKey[32:], MinerPrivKey.Y.Bytes())
	MinerHash = SerializeHashContent(pubKey)
	copy(shortMiner[:], MinerHash[0:8])
	minerAcc.Address = pubKey
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

	rootHash := SerializeHashContent(pubKey)

	var shortRootHash [8]byte
	copy(shortRootHash[:], rootHash[0:8])
}