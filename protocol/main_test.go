package protocol

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
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
	CommA = "-----BEGIN RSA PRIVATE KEY-----\n" +
			"MIIEpAIBAAKCAQEA5CLGs7+KmZq9yNB//old/yxY0xwan3/Rm+pTy4gW1exdR9s6\n" +
			"NBMwVYKfPkeVZEdsn4ZOjjFW61fRsTHcXYdqy58zlHCQIPYam7tsHXQR1X45rkoZ\n" +
			"a5cRkoRwyXA6/+D+b1ZYaMt/YZDujsXMlRpSxS7Xp7dYqe/CKerC1eaiN1nffY/n\n" +
			"QBBuePwuziOaJB5Nu911Uy4X3ZbbUgHks2HcxmETlpYUdaJ6B6ClRLt9SkwXDGqr\n" +
			"MwsJuPMLAQ9skwSl6pHLq7R14zoiggU/Y4XLhKk0oPz/Cwa3rd7eUHM3p+aq1YTQ\n" +
			"X62UmDz5MFUzsMsOerAv1I3j1S2yvG7pW04uUQIDAQABAoIBAFkeIKLc4waoBRUq\n" +
			"JaXIAXkJ2NT2+ItwAfC3M+6hBdKhV7sXL4BiMpJkyVIp4nje0dbrP0qaiYq7roVa\n" +
			"peu/V3+dfCezZQoLOU+2gkBrNABDI8Mq3Q1DYTDsHacC+Xk1ag8SGs0tGWCCnj4V\n" +
			"lJp2QvkWGFZC8BbKOv3m4B9wzdNywirCQgUme1gUPBICkmCBkY6AoiFY5B+UQxNX\n" +
			"+RoItoLQHYCiTNr+BzT0W+Ihwo1V4PHZMFNKKs9g4S159yxfie/jfMS7ElW0cNQt\n" +
			"NnwQv5sVQcG8L7ECBJwdGRnOrqCbT8HJeTg2Zmd+ChFeehLRySkgZ0QwfO9cRk/a\n" +
			"O42GSjUCgYEA/dGYz7pCVoKMolrH1Ox9WOPkpMBDUbbVsdiinpS3WxUIk4ed1d/x\n" +
			"Kp4B0L84ldgp9Cw16PJdtyfOqtUJEVPbu9c+61D1Zn3YPNL3Ag42ePbMkSmVyfLk\n" +
			"iHmkKhzzcTXfycnHAu7dl/Vi9cXVw4COBuYNQ0gcFV9SWqaxM6hmPVsCgYEA5hit\n" +
			"PKrJ4Pqrdv0vyclNuBNHGGlJ2wkujRqqViYq5/k3dVdmGvQFJRq6VKlfnl4jyxoa\n" +
			"PPP8MD4rHq5UDAGlJXqYb+WNcQEjeWv/cYxlNwRUIB8jHaSrw0lzjz5Z0oMWOY+5\n" +
			"VSYE6Julv/HOnFf/jV3DTw8P2VyJI/xWK3To9sMCgYBMrATiMxyY72S2IoAc5LdU\n" +
			"o7rMvbtYMsfIqm0tRDVDEU5+6keWdMhwHDzmJu1b7ml19ejvDk+a5S570lCj6FYH\n" +
			"HxVFljYbGMa6UOwGte5kigDvlMVHtNSuGTiq9AXh2+lXFlnEnA1aOukC3xkcrne4\n" +
			"w8Ob4GuDVUEWWyZKOYNw4wKBgQC+fjWF4VtLIBwuYYRbyYXHXGZipmBXr21TsnzM\n" +
			"38Jr1F5+jgHhVJ6hzlPu3V5lLUjyz8RjLBdgFUf7mZXJbt87fRiQovoLUUb+MlQD\n" +
			"vJjbCIFhKoYW94qgHcJHF/ajGpWeyAdGoDg2Hw4FL/q+YvgWIEcev7h+WmbLXRA1\n" +
			"4A8yowKBgQD5BmFRZG7uQXY/ZGEFhLX0/zGjib/whjBQb5zcFyIOlt0WPCyyB6kE\n" +
			"RR/QIbfTnDfc2Ql1FfQzGamY4le+6bf5S4Eb/1l9paEGGJz4KWEad6XH3KnaBAqL\n" +
			"knaP0QyFXaIYckiMhYXBpgzAHKjzCQOsqUjNP6gS3ROsWpcwII3EVw==\n" +
			"-----END RSA PRIVATE KEY-----"
	PubB1 = "5d7eefd58e3d2f309471928ab4bbd104e52973372c159fa652b8ca6b57ff68b8"
	PubB2 = "ab301a6a77b201c416ddc13a2d33fdf200a5302f6f687e0ea09085debaf8a1d9"
	PrivB = "7a0a9babcc97ea7991ed67ed7f800f70c5e04e99718960ad8efab2ca052f00c7"
)

const (
	//P-256
	RootPub1 = "6323cc034597195ae69bcfb628ecdffa5989c7503154c566bab4a87f3e9910ac"
	RootPub2 = "f6115b77a15852764c609c6a5c1739e698ebc6e49bf14617c561b9110039cec7"
	RootPriv = "277ed539f56122c25a6fc115d07d632b47e71416c9aebf1beb54ee704f11842c"
)

var (
	accA, accB, minerAcc *Account
	PrivKeyA, PrivKeyB   ecdsa.PrivateKey
	PubKeyA, PubKeyB     ecdsa.PublicKey
	RootPrivKey          ecdsa.PrivateKey
	CommKeyA			 rsa.PrivateKey
	MinerHash            [32]byte
	MinerPrivKey         *ecdsa.PrivateKey
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

	block, _ := pem.Decode([]byte(CommA))
	cKeyA, _ := x509.ParsePKCS1PrivateKey(block.Bytes)
	CommKeyA = *cKeyA

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
	copy(accA.CommitmentPubKey[:], CommKeyA.PublicKey.N.Bytes())
	accAHash := SerializeHashContent(accA.Address)

	//This one is just for testing purposes
	copy(accB.Address[0:32], PrivKeyB.PublicKey.X.Bytes())
	copy(accB.Address[32:64], PrivKeyB.PublicKey.Y.Bytes())
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