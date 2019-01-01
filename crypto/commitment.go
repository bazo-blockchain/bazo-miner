package crypto

import (
	"bufio"
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"math/big"
	"os"
)

const (
	// Note that this is the default public exponent set by Golang in rsa.go
	// See https://github.com/golang/go/blob/6269dcdc24d74379d8a609ce886149811020b2cc/src/crypto/rsa/rsa.go#L226
	COMM_PUBLIC_EXPONENT = 65537
	COMM_KEY_BITS        = 2048
	COMM_PROOF_LENGTH    = 256
	COMM_KEY_LENGTH      = 256
	COMM_NOF_PRIMES      = 2
)

func ExtractRSAKeyFromFile(filename string) (privKey *rsa.PrivateKey, err error) {

	if _, err = os.Stat(filename); os.IsNotExist(err) {
		err = CreateRSAKeyFile(filename)
		if err != nil {
			return privKey, err
		}
	}

	filehandle, err := os.Open(filename)
	if err != nil {
		return privKey, errors.New(fmt.Sprintf("%v", err))
	}
	defer filehandle.Close()

	scanner := bufio.NewScanner(filehandle)

	strModulus := nextLine(scanner)
	strPrivExponent := nextLine(scanner)
	strPrimes := make([]string, COMM_NOF_PRIMES)
	for i := 0; i < COMM_NOF_PRIMES; i++ {
		strPrimes[i] = nextLine(scanner)
	}

	if scanErr := scanner.Err(); scanErr != nil || err != nil {
		return privKey, errors.New(fmt.Sprintf("Could not read key from file: %v", err))
	}

	privKey, err = CreateRSAPrivKeyFromBase64(strModulus, strPrivExponent, strPrimes)
	if err != nil {
		return privKey, err
	}

	return privKey, VerifyRSAKey(privKey)
}
func ExtractRSAPubKeyFromFile(filename string) (pubKey *rsa.PublicKey, err error) {
	if _, err = os.Stat(filename); os.IsNotExist(err) {
		err = CreateRSAKeyFile(filename)
		if err != nil {
			return pubKey, err
		}
	}

	filehandle, err := os.Open(filename)
	if err != nil {
		return pubKey, errors.New(fmt.Sprintf("%v", err))
	}
	defer filehandle.Close()

	scanner := bufio.NewScanner(filehandle)

	strModulus := nextLine(scanner)

	if scanErr := scanner.Err(); scanErr != nil || err != nil {
		return pubKey, errors.New(fmt.Sprintf("Could not read key from file: %v", err))
	}

	pubKey, err = CreateRSAPubKeyFromBase64(strModulus)
	if err != nil {
		return pubKey, err
	}

	return pubKey, nil
}

func VerifyRSAKey(privKey *rsa.PrivateKey) error {
	message := "Test"
	cipher, err := SignMessageWithRSAKey(privKey, message)
	if err != nil {
		return errors.New(fmt.Sprintf("Could not sign message. Failed with error: %v", err))
	}

	err = VerifyMessageWithRSAKey(&privKey.PublicKey, message, cipher)
	if err != nil {
		return errors.New(fmt.Sprintf("Could not verify message. Failed with error: %v", err))
	}
	return nil
}

func CreateRSAPubKeyFromBytes(bytModulus [COMM_KEY_LENGTH]byte) (pubKey *rsa.PublicKey, err error) {
	modulus := new(big.Int).SetBytes(bytModulus[:])
	pubKey = new(rsa.PublicKey)
	pubKey.N = modulus
	pubKey.E = COMM_PUBLIC_EXPONENT
	return
}

func CreateRSAPubKeyFromBase64(strModulus string) (pubKey *rsa.PublicKey, err error) {
	modulus, err := fromBase64(strModulus, &err)
	pubKey = new(rsa.PublicKey)
	pubKey.N = modulus
	pubKey.E = COMM_PUBLIC_EXPONENT
	return
}

func CreateRSAPrivKeyFromBase64(strModulus string, strPrivExponent string, strPrimes []string) (privKey *rsa.PrivateKey, err error) {
	modulus, err := fromBase64(strModulus, &err)
	privExponent, err := fromBase64(strPrivExponent, &err)
	primes := make([]*big.Int, COMM_NOF_PRIMES)
	for i := 0; i < COMM_NOF_PRIMES; i++ {
		primes[i], err = fromBase64(strPrimes[i], &err)
	}

	privKey = &rsa.PrivateKey{
		PublicKey: rsa.PublicKey{
			N: modulus,
			E: COMM_PUBLIC_EXPONENT,
		},
		D:      privExponent,
		Primes: primes,
	}
	privKey.Precompute()
	return
}

func SignMessageWithRSAKey(privKey *rsa.PrivateKey, msg string) (fixedSig [COMM_PROOF_LENGTH]byte, err error) {
	hashed := sha256.Sum256([]byte(msg))
	sig, err := rsa.SignPKCS1v15(rand.Reader, privKey, crypto.SHA256, hashed[:])
	if err != nil {
		return fixedSig, err
	}
	copy(fixedSig[:], sig[:])
	return fixedSig, nil
}

func VerifyMessageWithRSAKey(pubKey *rsa.PublicKey, msg string, fixedSig [COMM_PROOF_LENGTH]byte) (err error) {
	hashed := sha256.Sum256([]byte(msg))
	return rsa.VerifyPKCS1v15(pubKey, crypto.SHA256, hashed[:], fixedSig[:])
}

func fromBase64(encoded string, err *error) (*big.Int, error) {
	if *err != nil {
		return nil, *err
	}

	byteArray, encodeErr := base64.StdEncoding.DecodeString(encoded)
	if encodeErr != nil {
		return nil, encodeErr
	}

	return new(big.Int).SetBytes(byteArray), nil
}

func nextLine(scanner *bufio.Scanner) string {
	scanner.Scan()
	return scanner.Text()
}

// Creates an RSA key file with the following lines
// 1 	Public Modulus N
// 2 	Private Exponent D
// 3+	Private Primes (depending on COMM_NOF_PRIMES)
func CreateRSAKeyFile(filename string) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}

	key, err := GenerateRSAKey()
	if err != nil {
		return err
	}

	_, err = file.WriteString(stringifyRSAKey(key))
	return err
}

func GenerateRSAKey() (*rsa.PrivateKey, error) {
	return rsa.GenerateMultiPrimeKey(rand.Reader, COMM_NOF_PRIMES, COMM_KEY_BITS)
}

func stringifyRSAKey(key *rsa.PrivateKey) string {
	keyString :=
		base64.StdEncoding.EncodeToString(key.N.Bytes()) +
		"\n" +
		base64.StdEncoding.EncodeToString(key.D.Bytes())

	for _, prime := range key.Primes {
		keyString += "\n" + base64.StdEncoding.EncodeToString(prime.Bytes())
	}

	return keyString
}

func GetBytesFromRSAPubKey(key *rsa.PublicKey) (commPubKey [COMM_KEY_LENGTH]byte) {
	copy(commPubKey[:], key.N.Bytes())
	return commPubKey
}