package vm

import (
	"log"
	"testing"

	"github.com/bazo-blockchain/bazo-miner/protocol"
)

// Function generates random bytes, if an exception occurs, it is catched and printed out with the random bytes,
// so the specific failing test can be recreated
func Fuzz() {
	code := protocol.RandomBytes()
	vm := NewTestVM([]byte{})
	mc := NewMockContext(code)
	mc.Fee = 10000
	vm.context = mc

	defer func() {
		if err := recover(); err != nil {
			log.Println("Execution failed", err, code)
		}
	}()

	vm.Exec(false)
}

func TestFuzz(t *testing.T) {
	for i := 0; i <= 5000000; i++ {
		Fuzz()
	}
}
