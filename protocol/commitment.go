package protocol

import (
	"bufio"
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"errors"
	"fmt"
	"math/big"
	"os"
)

const (
	// Note that this is the default public exponent set by Golang in rsa.go
	// See https://github.com/golang/go/blob/6269dcdc24d74379d8a609ce886149811020b2cc/src/crypto/rsa/rsa.go#L226
	COMM_PUBLIC_EXPONENT 	= 65537
	// When changing COMM_KEY_BITS_SIZE, remember to change COMM_KEY_LENGTH
	COMM_KEY_BITS_SIZE 		= 2048
	// When changing COMM_KEY_LENGTH, remember to change ACC_SIZE, STAKETX_SIZE, ...
	COMM_KEY_LENGTH    		= 256
	COMM_NOF_PRIMES    		= 2
)

func ExtractRSAKeyFromFile(filename string) (privKey rsa.PrivateKey, err error) {
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		err = createRSAKeyFile(filename)
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

	return CreateRSAPrivKeyFromBase10(strModulus, strPrivExponent, strPrimes)
}

func CreateRSAPubKeyFromModulus(modulus [COMM_KEY_LENGTH]byte) (*rsa.PublicKey) {
	n := new(big.Int).SetBytes(modulus[:])
	return &rsa.PublicKey{
		N: n,
		E: COMM_PUBLIC_EXPONENT,
	}
}

func CreateRSAPrivKeyFromBase10(strModulus string, strPrivExponent string, strPrimes []string) (privKey rsa.PrivateKey, err error) {
	modulus, err := fromBase10(strModulus, &err)
	privExponent, err := fromBase10(strPrivExponent, &err)
	primes := make([]*big.Int, COMM_NOF_PRIMES)
	for i := 0; i < COMM_NOF_PRIMES; i++ {
		primes[i], err = fromBase10(strPrimes[i], &err)
	}

	privKey = rsa.PrivateKey{
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

func SignMessageWithRSAKey(privKey *rsa.PrivateKey, msg string) (fixedSig [COMM_KEY_LENGTH]byte, err error) {
	hashed := sha256.Sum256([]byte(msg))
	sig, err := rsa.SignPKCS1v15(rand.Reader, privKey, crypto.SHA256, hashed[:])
	if err != nil {
		return fixedSig, err
	}
	copy(fixedSig[:], sig[:])
	return fixedSig, nil
}

func VerifyMessageWithRSAKey(pubKey *rsa.PublicKey, msg string, fixedSig [COMM_KEY_LENGTH]byte) (err error) {
	hashed := sha256.Sum256([]byte(msg))
	return rsa.VerifyPKCS1v15(pubKey, crypto.SHA256, hashed[:], fixedSig[:])
}

func fromBase10(base10 string, err *error) (*big.Int, error) {
	if *err != nil {
		return nil, *err
	}

	i, ok := new(big.Int).SetString(base10, 10)
	if !ok {
		return nil, errors.New("Could not convert to Base10 integer")
	}
	return i, nil
}

func nextLine(scanner *bufio.Scanner) string {
	scanner.Scan()
	return scanner.Text()
}

// Creates an RSA key file with the following lines
// 1 	Public Modulus N
// 2 	Private Exponent D
// 3+	Private Primes (depending on COMM_NOF_PRIMES)
func createRSAKeyFile(filename string) (err error) {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}

	key, err := rsa.GenerateMultiPrimeKey(rand.Reader, COMM_NOF_PRIMES, COMM_KEY_BITS_SIZE)
	if err != nil {
		return err
	}

	_, err = file.WriteString(stringifyRSAKey(key))
	return
}

func stringifyRSAKey(key *rsa.PrivateKey) (keyString string) {
	keyString = key.N.String() + "\n" + key.D.String()
	for _, prime := range key.Primes {
		keyString += "\n" + prime.String()
	}
	return
}
