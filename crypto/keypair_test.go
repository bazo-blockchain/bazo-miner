package crypto

import (
	"os"
	"testing"
)

const (
	KEY_TEST_FILE = "test_key.txt"
)

func TestExtractAndVerifyEDSAKeyFromNonExistingFile(t *testing.T) {
	os.Remove(KEY_TEST_FILE)

	_, err := ExtractECDSAKeyFromFile(KEY_TEST_FILE)
	if err != nil {
		t.Errorf("Could not extract RSA key from file. Failed with error: %v", err)
	}

	os.Remove(KEY_TEST_FILE)
}

func TestExtractAndVerifyEDSAKeyFromExistingFile(t *testing.T) {
	os.Remove(KEY_TEST_FILE)
	err := CreateECDSAKeyFile(KEY_TEST_FILE)

	_, err = ExtractECDSAKeyFromFile(KEY_TEST_FILE)
	if err != nil {
		t.Errorf("Could not extract RSA key from file. Failed with error: %v", err)
	}

	os.Remove(KEY_TEST_FILE)
}
