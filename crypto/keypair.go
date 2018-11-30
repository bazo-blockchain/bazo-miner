package crypto

import (
	"bufio"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"errors"
	"fmt"
	"log"
	"math/big"
	"os"
	"strings"
)

func ExtractECDSAKeyFromFile(filename string) (privKey *ecdsa.PrivateKey, err error) {
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		err = CreateECDSAKeyFile(filename)
		if err != nil {
			return nil, err
		}
	}

	filehandle, err := os.Open(filename)
	if err != nil {
		return privKey, errors.New(fmt.Sprintf("%v", err))
	}
	defer filehandle.Close()

	reader := bufio.NewReader(filehandle)
	privKey, err = readECDSAPrivateKey(reader)

	if err != nil {
		return privKey, errors.New(fmt.Sprintf("%v", err))
	}

	return privKey, VerifyECDSAKey(privKey)
}

func ExtractECDSAPublicKeyFromFile(filename string) (pubKey *ecdsa.PublicKey, err error) {
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		err = CreateECDSAKeyFile(filename)
		if err != nil {
			return nil, err
		}
	}

	filehandle, err := os.Open(filename)
	if err != nil {
		return pubKey, errors.New(fmt.Sprintf("%v", err))
	}
	defer filehandle.Close()

	reader := bufio.NewReader(filehandle)

	return readECDSAPublicKey(reader)
}

func readECDSAPrivateKey(reader *bufio.Reader) (privKey *ecdsa.PrivateKey, err error) {
	pubKey, err := readECDSAPublicKey(reader)
	priv, err2 := reader.ReadString('\n')
	if err != nil || err2 != nil {
		return privKey, errors.New(fmt.Sprintf("Could not read key from file: %v", err))
	}

	if err2 == nil {
		privInt, b := new(big.Int).SetString(strings.Split(priv, "\n")[0], 16)
		if !b {
			return privKey, errors.New("failed to convert the key strings to big.Int")
		}

		privKey = &ecdsa.PrivateKey{
			*pubKey,
			privInt,
		}
	}

	return privKey, nil
}

func readECDSAPublicKey(reader *bufio.Reader) (pubKey *ecdsa.PublicKey, err error) {
	//Public Key
	pub1, err := reader.ReadString('\n')
	pub2, err := reader.ReadString('\n')

	if err != nil {
		return pubKey, errors.New(fmt.Sprintf("Could not read key from file: %v", err))
	}

	return GetPubKeyFromString(strings.Split(pub1, "\n")[0], strings.Split(pub2, "\n")[0])
}

func VerifyECDSAKey(privKey *ecdsa.PrivateKey) error {
	//Make sure the key being used is a valid one, that can sign and verify hashes/transactions
	hashed := []byte("testing")
	r, s, err := ecdsa.Sign(rand.Reader, privKey, hashed)
	if err != nil {
		return errors.New("the ecdsa key you provided is invalid and cannot sign hashes")
	}

	if !ecdsa.Verify(&privKey.PublicKey, hashed, r, s) {
		return errors.New("the ecdsa key you provided is invalid and cannot verify hashes")
	}
	return nil
}

func ReadFile(filename string) (lines []string) {
	file, err := os.Open(filename)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}

	return lines
}

func GetAddressFromPubKey(pubKey *ecdsa.PublicKey) (address [64]byte) {
	copy(address[:32], pubKey.X.Bytes())
	copy(address[32:], pubKey.Y.Bytes())

	return address
}

func GetPubKeyFromAddress(address [64]byte) (pubKey *ecdsa.PublicKey) {
	pubKey1Sig, pubKey2Sig := new(big.Int), new(big.Int)
	pubKey1Sig.SetBytes(address[:32])
	pubKey2Sig.SetBytes(address[32:])
	return &ecdsa.PublicKey {
		Curve: elliptic.P256(),
		X:     pubKey1Sig,
		Y:     pubKey2Sig,
	}
}

func GetPubKeyFromString(pub1, pub2 string) (pubKey *ecdsa.PublicKey, err error) {
	pub1Int, b := new(big.Int).SetString(pub1, 16)
	pub2Int, b := new(big.Int).SetString(pub2, 16)
	if !b {
		return pubKey, errors.New("failed to convert the key strings to big.Int")
	}

	pubKey = &ecdsa.PublicKey{
		elliptic.P256(),
		pub1Int,
		pub2Int,
	}

	return pubKey, nil
}

func CreateECDSAKeyFile(filename string) (err error) {
	newKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)

	//Write the public key to the given textfile
	if _, err = os.Stat(filename); !os.IsNotExist(err) {
		return err
	}

	file, err := os.Create(filename)
	if err != nil {
		return err
	}

	var pubKey [64]byte

	_, err1 := file.WriteString(string(newKey.X.Text(16)) + "\n")
	_, err2 := file.WriteString(string(newKey.Y.Text(16)) + "\n")
	_, err3 := file.WriteString(string(newKey.D.Text(16)) + "\n")

	newAccPub1, newAccPub2 := newKey.PublicKey.X.Bytes(), newKey.PublicKey.Y.Bytes()
	copy(pubKey[0:32], newAccPub1)
	copy(pubKey[32:64], newAccPub2)

	if err1 != nil || err2 != nil || err3 != nil {
		return errors.New("failed to write key to file")
	}

	return nil
}
