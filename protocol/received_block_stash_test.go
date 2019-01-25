package protocol

import (
	"reflect"
	"testing"
)

func TestStashSetMethod(t *testing.T) {
	var sampleStash = NewBlockStash()
	var blockHash1 = [32]byte{'1'}
	var block1 = NewBlock([32]byte{'1'},1)

	var blockHash2 = [32]byte{'2'}
	var block2 = NewBlock([32]byte{'2'},2)

	var blockHash3 = [32]byte{'3'}
	var block3 = NewBlock([32]byte{'3'},3)

	var blockHash4 = [32]byte{'4'}
	var block4 = NewBlock([32]byte{'4'},4)

	var blockHash5 = [32]byte{'5'}
	var block5 = NewBlock([32]byte{'5'},5)

	var blockHash6 = [32]byte{'6'}
	var block6 = NewBlock([32]byte{'6'},6)

	sampleStash.Set(blockHash1,block1)
	sampleStash.Set(blockHash2,block2)
	sampleStash.Set(blockHash3,block3)
	sampleStash.Set(blockHash4,block4)
	sampleStash.Set(blockHash5,block5)
	sampleStash.Set(blockHash6,block6)

	if !reflect.DeepEqual(6, len(sampleStash.m)) && !reflect.DeepEqual(6, len(sampleStash.keys)){
		t.Errorf("Stash size does not equal 6")
	}

	//same block as block6 included in the stash
	var blockHash6b = [32]byte{'6'}
	var block6b = NewBlock([32]byte{'6'},6)

	sampleStash.Set(blockHash6b,block6b)

	if !reflect.DeepEqual(6, len(sampleStash.m)) && !reflect.DeepEqual(6, len(sampleStash.keys)){
		t.Errorf("Stash included a duplicate block")
	}
}

func TestStashSetWhenSizeOver50Entries(t *testing.T) {
	var sampleStash = NewBlockStash()

	var blockHash = [32]byte{'0'}
	var block = NewBlock([32]byte{'0'},0)

	/*Fill the stash with 50 blocks*/
	for i := 0; i < 50; i++ {
		blockHash = [32]byte{byte(i)}
		block = NewBlock(blockHash,uint32(i))

		sampleStash.Set(blockHash,block)
	}

	if !reflect.DeepEqual(50, len(sampleStash.m)) && !reflect.DeepEqual(50, len(sampleStash.keys)){
		t.Errorf("Error in filling the stash: Length should be: %d - Lenght is actually: %d",50,len(sampleStash.m))
	}

	//Keep track of first entry in the stash, this one should be deleted
	firstHash,firstBlock := ReturnItemForPosition(sampleStash,0)
	if !reflect.DeepEqual(uint32(0), firstBlock.Height) && !reflect.DeepEqual([32]byte{'0'}, firstHash){
		t.Errorf("Error retrieving the first entry of the stash")
	}

	secondHash,secondBlock := ReturnItemForPosition(sampleStash,1)
	if !reflect.DeepEqual(uint32(1), secondBlock.Height) && !reflect.DeepEqual([32]byte{'1'}, secondHash){
		t.Errorf("Error retrieving the second entry of the stash")
	}

	thirdHash,thirdBlock := ReturnItemForPosition(sampleStash,2)
	if !reflect.DeepEqual(uint32(2), thirdBlock.Height) && !reflect.DeepEqual([32]byte{'2'}, thirdHash){
		t.Errorf("Error retrieving the second entry of the stash")
	}

	outofboundHash,outofboundBlock := ReturnItemForPosition(sampleStash,50)
	if(outofboundHash != [32]byte{} && outofboundBlock != nil){
		t.Errorf("Error expected out of bound exception")
	}
}
