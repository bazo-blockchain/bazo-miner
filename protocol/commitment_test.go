package protocol

import (
	"os"
	"testing"
)

const (
	COMMITMENT_TEST_FILE = "test_commitment.txt"
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
