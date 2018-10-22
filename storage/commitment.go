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
	"strconv"
)

const (
	COMM_KEY_BITS_SIZE 	= 2048
	COMM_KEY_LENGTH		= 256
	COMM_NOF_PRIMES    	= 2
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
	pubExponent, err := strconv.Atoi(nextLine(scanner))
	privExponent, err := fromBase10(nextLine(scanner), &err)
	primes := make([]*big.Int, COMM_NOF_PRIMES)
	primes[0], err = fromBase10(nextLine(scanner), &err)
	primes[1], err = fromBase10(nextLine(scanner), &err)

	if scanErr := scanner.Err(); scanErr != nil || err != nil {
		return privKey, errors.New(fmt.Sprintf("Could not read key from file: %v", err))
	}

	fmt.Println(modulus)
	fmt.Println(pubExponent)
	fmt.Println(privExponent)

	privKey = rsa.PrivateKey{
		PublicKey: rsa.PublicKey{
			N: modulus,
			E: pubExponent,
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
// 2 	Public Exponent E
// 3 	Private Exponent D
// 4+	Private Primes
func createRSAKeyFile(filename string) (err error) {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}

	key, err := rsa.GenerateMultiPrimeKey(rand.Reader, COMM_NOF_PRIMES, COMM_KEY_BITS_SIZE)
	if err != nil {
		return err
	}

	commString := key.N.String() + "\n" +
		strconv.Itoa(key.E) + "\n" +
		key.D.String()

	for _, prime := range key.Primes {
		commString += "\n" + prime.String()
	}

	_, err = file.WriteString(commString)
	return err
}

func SignMessageWithRsaKey(privKey *rsa.PrivateKey, msg string) (sig []byte, err error) {
	hashed := sha256.Sum256([]byte(msg))
	return rsa.SignPKCS1v15(rand.Reader, privKey, crypto.SHA256, hashed[:])
}

func VerifyMessageWithRsaKey(pubKey *rsa.PublicKey, msg string, sig []byte) (err error) {
	hashed := sha256.Sum256([]byte(msg))
	return rsa.VerifyPKCS1v15(pubKey, crypto.SHA256, hashed[:], sig)
}
