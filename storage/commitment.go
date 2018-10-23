package storage

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
	COMM_KEY_BITS_SIZE 		= 2048
	COMM_PUBLIC_EXPONENT 	= 65537
	COMM_KEY_LENGTH    		= 256
	COMM_NOF_PRIMES    		= 2
)

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

	modulus, err := fromBase10(nextLine(scanner), &err)
	privExponent, err := fromBase10(nextLine(scanner), &err)
	primes := make([]*big.Int, COMM_NOF_PRIMES)
	for i := 0; i < COMM_NOF_PRIMES; i++ {
		primes[i], err = fromBase10(nextLine(scanner), &err)
	}

	if scanErr := scanner.Err(); scanErr != nil || err != nil {
		return privKey, errors.New(fmt.Sprintf("Could not read key from file: %v", err))
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

	return privKey, nil
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
	key.E = COMM_PUBLIC_EXPONENT // make sure that the right public exponent is used in future version of Golang
	key.Precompute()

	if err != nil {
		return err
	}

	commString := key.N.String() + "\n" + key.D.String()

	for _, prime := range key.Primes {
		commString += "\n" + prime.String()
	}

	_, err = file.WriteString(commString)
	return err
}

func CreateRsaPubKey(modulus [COMM_KEY_LENGTH]byte) (*rsa.PublicKey) {
	modulus2 := make([]byte, 0)
	copy(modulus2[:], modulus[:])
	n := new(big.Int).SetBytes(modulus2)
	return &rsa.PublicKey{
		N: n,
		E: COMM_PUBLIC_EXPONENT,
	}
}

func SignMessageWithRsaKey(privKey *rsa.PrivateKey, msg string) (fixedSig [COMM_KEY_LENGTH]byte, err error) {
	hashed := sha256.Sum256([]byte(msg))
	sig, err := rsa.SignPKCS1v15(rand.Reader, privKey, crypto.SHA256, hashed[:])
	if err != nil {
		return fixedSig, err
	}
	copy(fixedSig[:], sig[:])
	return fixedSig, nil
}

func VerifyMessageWithRsaKey(pubKey *rsa.PublicKey, msg string, fixedSig [COMM_KEY_LENGTH]byte) (err error) {
	hashed := sha256.Sum256([]byte(msg))
	return rsa.VerifyPKCS1v15(pubKey, crypto.SHA256, hashed[:], fixedSig[:])
}
