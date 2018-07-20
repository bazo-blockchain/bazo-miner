package integration

import (
	"testing"
	"github.com/bazo-blockchain/bazo-miner/storage"
	"os"
	"crypto/rand"
	"time"
	"crypto/ecdsa"
	"crypto/elliptic"
	"os/exec"
	"fmt"
	"bytes"
)



const (
	DbName          = "integration_test.db"
	RootKey         = "root_key_test"
	MultisigKey     = "multisig_key_test"
	MinerKey     	= "miner_key_test"
	BootstrapIpPort = "127.0.0.1:8000"
	SeedFileName    = "seed_test.json"
	InitRootBalance = 100000
)


func createBootstrapMiner() {
	pubKey, _, _ := storage.ExtractKeyFromFile(RootKey)
	cmd := exec.Command("go", "run", "../main.go", DbName, BootstrapIpPort, RootKey, SeedFileName, MultisigKey)
	cmd.Env = append(os.Environ(),
		fmt.Sprintf("INITROOTPUBKEY1=%s", pubKey.X.Text(16)),
		fmt.Sprintf("INITROOTPUBKEY2=%s", pubKey.Y.Text(16)),
		fmt.Sprintf("INITROOTBALANCE=%d", InitRootBalance),
	)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Start()
	if err != nil {
		os.Stderr.WriteString(fmt.Sprintf("==> Error: %s\n", err.Error()))
	}
	ticker := time.NewTicker(1*time.Second)
	go func(ticker *time.Ticker) {
		for _ = range ticker.C {
			outStr, errStr := string(stdout.Bytes()), string(stderr.Bytes())
			fmt.Printf("out:\n%s\nerr:\n%s\n", outStr, errStr)
		}
	}(ticker)

}

func generateKeypair(name string) {
	newKey, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	file, _ := os.Create(name)
	file.WriteString(string(newKey.X.Text(16)) + "\n")
	file.WriteString(string(newKey.Y.Text(16)) + "\n")
	file.WriteString(string(newKey.D.Text(16)) + "\n")
}

func TestMain(m *testing.M) {

	// Clean
	os.Remove(DbName)
	os.Remove(RootKey)
	os.Remove(MultisigKey)
	os.Remove(MinerKey)
	os.Remove(SeedFileName)

	// Create keys
	generateKeypair(RootKey)
	generateKeypair(MultisigKey)
	time.Sleep(time.Second)

	createBootstrapMiner()
	time.Sleep(5*time.Second)
	retCode := m.Run()
	TearDown()
	os.Exit(retCode)
}

func TearDown() {
	os.Remove(DbName)
}
