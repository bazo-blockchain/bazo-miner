package protocol

import (
	"math/big"
	"os"
	"reflect"
	"testing"
)

const (
	COMMITMENT_TEST_FILE = "test_commitment.txt"
	COMMITMENT_TEST_PUB_KEY = "18548212161594241704562086057105745474725101902253588258556723550630094870988762961560443068881642741657030368708672414829169963663387916382212493645579999921532548551718896960388598422072111998614042524402714882944809342207653938586247432016091165966941453061367573179570610572390465350003274179139174972419154383984809658470079788130967692120601258597645956151785354457176279523123870577986200270725135160744968043298436307626602608775682438224953565643149184132971403302816516228897647081726145971323873493321063058710948023038704239467348321824015654365987281556817284724506489623275566344874728267587394489105959"
)

func TestRSASigningAndVerification(t *testing.T) {
	os.Remove(COMMITMENT_TEST_FILE)

	privKey, err := ExtractRSAKeyFromFile(COMMITMENT_TEST_FILE)
	if err != nil {
		t.Errorf("Could not extract RSA key from file. Failed with error: %v", err)
	}

	message := "Test"
	cipher, err := SignMessageWithRSAKey(&privKey, message)
	if err != nil {
		t.Errorf("Could not sign message. Failed with error: %v", err)
	}

	err = VerifyMessageWithRSAKey(&privKey.PublicKey, message, cipher)
	if err != nil {
		t.Errorf("Could not verify message. Failed with error: %v", err)
	}

	os.Remove(COMMITMENT_TEST_FILE)
}

func TestCreateRSAPubKeyFromModulus(t *testing.T) {
	intModulus, _ := new(big.Int).SetString(COMMITMENT_TEST_PUB_KEY, 10)
	var byteModulus [COMM_KEY_LENGTH]byte
	copy(byteModulus[:], intModulus.Bytes())
	pubKey := CreateRSAPubKeyFromModulus(byteModulus)

	if !reflect.DeepEqual(pubKey.N, intModulus) {
		t.Errorf("Creating RSA PubKey from modulus failed, (%v) vs. (%v)\n", pubKey.N, intModulus)
	}
}