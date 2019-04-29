package protocol

import (
	"bytes"
	"testing"
)

func TestVMContext_GetContractVariable_EncapsulationBreach(t *testing.T) {
	c := Context{}
	c.ContractVariables = [][]byte{[]byte{0x00, 0x00, 0x00}}

	slice1, _ := c.GetContractVariable(0)

	for i := range slice1 {
		slice1[i] = 1
	}

	expected := []byte{0x00, 0x00, 0x00}
	actual, _ := c.GetContractVariable(0)
	if !bytes.Equal(expected, actual) {
		t.Errorf("Expected result to be '%v' but was '%v'", expected, actual)
	}

	expected = []byte{0x01, 0x01, 0x01}
	actual = slice1
	if !bytes.Equal(expected, actual) {
		t.Errorf("Expected result to be '%v' but was '%v'", expected, actual)
	}
}

func TestVMContext_GetContractVariable_RespectChanges(t *testing.T) {
	c := Context{}
	c.ContractVariables = [][]byte{[]byte{0x00, 0x00, 0x00}}
	c.changes = []Change{{
		index: 0,
		value: []byte{0x00, 0x00, 0x01},
	}}

	expected := []byte{0x00, 0x00, 0x01}
	actual, _ := c.GetContractVariable(0)
	if !bytes.Equal(expected, actual) {
		t.Errorf("Expected result to be '%v' but was '%v'", expected, actual)
	}
}

func TestVMContext_SetContractVariable_EncapsulationBreach(t *testing.T) {
	c := Context{}
	c.ContractVariables = [][]byte{[]byte{0x00, 0x00, 0x00}}

	slice1, _ := c.GetContractVariable(0)

	for i := range slice1 {
		slice1[i] = 1
	}

	c.SetContractVariable(0, slice1)
	c.PersistChanges()

	for i := range slice1 {
		slice1[i] = 2
	}

	expected := []byte{0x01, 0x01, 0x01}
	actual, _ := c.GetContractVariable(0)
	if !bytes.Equal(expected, actual) {
		t.Errorf("Expected result to be '%v' but was '%v'", expected, actual)
	}
}

func TestVMContext_SetContractVariable_Simple(t *testing.T) {
	c := Context{}
	c.ContractVariables = [][]byte{[]byte{0x00, 0x00, 0x00}}

	slice1, _ := c.GetContractVariable(0)

	// Change values for the first time
	for i := range slice1 {
		slice1[i] = 1
	}
	c.SetContractVariable(0, slice1)

	c.PersistChanges()

	if len(c.changes) != 1 {
		t.Errorf("Only 1 change with index 0 expected but got %v", len(c.changes))
	}

	expected := []byte{0x01, 0x01, 0x01}
	actual, _ := c.GetContractVariable(0)
	if !bytes.Equal(expected, actual) {
		t.Errorf("Expected result to be '%v' but was '%v'", expected, actual)
	}
}

func TestVMContext_SetContractVariable_ReplaceChange(t *testing.T) {
	c := Context{}
	c.ContractVariables = [][]byte{[]byte{0x00, 0x00, 0x00}}

	slice1, _ := c.GetContractVariable(0)

	// Change values for the first time
	for i := range slice1 {
		slice1[i] = 1
	}
	c.SetContractVariable(0, slice1)

	// Change values for the second time
	for i := range slice1 {
		slice1[i] = 2
	}
	c.SetContractVariable(0, slice1)
	c.PersistChanges()

	if len(c.changes) != 1 {
		t.Errorf("Only 1 change with index 0 expected but got %v", len(c.changes))
	}

	expected := []byte{0x02, 0x02, 0x02}
	actual, _ := c.GetContractVariable(0)
	if !bytes.Equal(expected, actual) {
		t.Errorf("Expected result to be '%v' but was '%v'", expected, actual)
	}
}