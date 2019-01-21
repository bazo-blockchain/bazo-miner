package protocol

import (
	"bytes"
	"encoding/gob"
	"fmt"
)

type ValShardMapping struct {
	//Header
	ValMapping        map[[64]byte]int
	EpochHeight		  int
}

func NewMapping() *ValShardMapping {
	newMapping := new(ValShardMapping)
	newMapping.ValMapping = make(map[[64]byte]int)
	newMapping.EpochHeight = 0
	return newMapping
}


func (valMapping *ValShardMapping) HashMapping() [32]byte {
	if valMapping == nil {
		return [32]byte{}
	}

	mappingHash := struct {
		ValMapping				  map[[64]byte]int
		EpochHeight				  int
	}{
		valMapping.ValMapping,
		valMapping.EpochHeight,
	}
	return SerializeHashContent(mappingHash)
}

func (valMapping *ValShardMapping) GetSize() int {
	size := len(valMapping.ValMapping)
	return size
}

func (valMapping *ValShardMapping) Encode() []byte {
	if valMapping == nil {
		return nil
	}

	encoded := ValShardMapping{
		ValMapping:                valMapping.ValMapping,
		EpochHeight:			   valMapping.EpochHeight,
	}

	buffer := new(bytes.Buffer)
	gob.NewEncoder(buffer).Encode(encoded)
	return buffer.Bytes()
}

func (valMapping *ValShardMapping) Decode(encoded []byte) (valMappingDecoded *ValShardMapping) {
	if encoded == nil {
		return nil
	}

	var decoded ValShardMapping
	buffer := bytes.NewBuffer(encoded)
	decoder := gob.NewDecoder(buffer)
	decoder.Decode(&decoded)
	return &decoded
}

func (valMapping ValShardMapping) String() string {
	mappingString := "\n"
	for k, v := range valMapping.ValMapping {
		mappingString += fmt.Sprintf("Entry: %x -> %v\n", k[:8],v)
	}
	mappingString += fmt.Sprintf("Epoch Height: %d\n", valMapping.EpochHeight)
	return mappingString
}
