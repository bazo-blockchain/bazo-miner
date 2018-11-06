package crypto

import (
	"os"
	"testing"
)

const (
	COMMITMENT_TEST_FILE = "test_commitment.txt"
)

func TestExtractAndVerifyRSAKeyFromNonExistingFile(t *testing.T) {
	os.Remove(COMMITMENT_TEST_FILE)

	_, err := ExtractRSAKeyFromFile(COMMITMENT_TEST_FILE)
	if err != nil {
		t.Errorf("Could not extract RSA key from file. Failed with error: %v", err)
	}

	os.Remove(COMMITMENT_TEST_FILE)
}

func TestExtractAndVerifyRSAKeyFromExistingFile(t *testing.T) {
	os.Remove(COMMITMENT_TEST_FILE)
	err := CreateRSAKeyFile(COMMITMENT_TEST_FILE)
	if err != nil {
		t.Errorf("Could not create RSA key file. Failed with error: %v", err)
	}

	_, err = ExtractRSAKeyFromFile(COMMITMENT_TEST_FILE)
	if err != nil {
		t.Errorf("Could not extract RSA key from file. Failed with error: %v", err)
	}

	os.Remove(COMMITMENT_TEST_FILE)
}