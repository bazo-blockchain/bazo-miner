package miner

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/rsa"
	"fmt"
	"github.com/bazo-blockchain/bazo-miner/crypto"
	"math/big"
	"os"
	"strings"
	"testing"

	"github.com/bazo-blockchain/bazo-miner/p2p"
	"github.com/bazo-blockchain/bazo-miner/protocol"
	"github.com/bazo-blockchain/bazo-miner/storage"
)

const (
	TestDBFileName   = "test.db"
	TestIpPort       = "127.0.0.1:8010"
	TestIpPortNodeA       = "127.0.0.1:8001"
	TestIpPortNodeB       = "127.0.0.1:8002"
	TestIpPortNodeC       = "127.0.0.1:8003"
	TestIpPortNodeD       = "127.0.0.1:8004"
	TestKeyFileName  = "test_root"
)

const (
	PubA1 = "c2be9abbeaec39a066c2a09cee23bb9ab2a0b88f2880b1e785b4d317adf0dc7c"
	PubA2 = "8ce020fde838d9c443f6c93345dafe7fd74f091c4d2f30b37e2453679a257ed5"
	PrivA = "ba127fa8f802b008b9cdb58f4e44809d48f1b000cff750dda9cd6b312395c1c5"
	CommPubA = "vsl0yfAd3dqJfDEawAl7Xp2/hOvXGN/u0UXBpRSxWAT+FSKlt5Ha8Ibd59tGkM4D8i/MABx0MMNEVL8Ghe1QkXIITJRnFtoqsidTlcHSL4WL7sc+8LwJjIjMdqM5BYIZJap/j2O2qREcEICEN8i+6LF844iMqFysDOuL8F5MAH22twrh0SMVXAM+IAEqa0Z9TymvX8Op3dt/t5IhrA4ivsS/+QMWzr3xJE9XfQxMrDUNoBwXIszOr656m8/wYa9dOEZn8qlglEySAjievkECJZq9Q3DRat5SUoXjG8M6UJp/AeRUUANsXhrPn6Cg7j4ke5Yw0bk6Lz9foYaZ9rugkw=="
	CommPrivA = "n5Xdlei+4sshA3wDlyyXQF6NS78GTi1KE0zZHJ/BdBHBAqbXnURosZbuWTmmvgtFa7ilWFZ0rjE3n/elmjMWmIKdBImB7bCR1DFnDjZw/QUlNpb9Q9rV1fK7rGT9lmjrZgFG8AcFTEgehIMrlYnafsOv5pdaqJ3T4H7KsEYAJsuNZhHAFmReqNdeiUbdAntPLQjttbs43DqaVQ0D3YnHrKxeu7Ekwcs4ap18tkFt7Lp0mkJ3fjpsvJFPDP2CotrZadLilv7dmOrXe26XDLUQ2aBguExV4Wx85J29puOJwpoM60KiFgiBMtRQFRukzRuValiVkXEBLZKlbh6wYy0vwQ=="
	CommPrim1A = "9UbkVH5chUZCZaehntnZWAfTJ9OYvsKfu19Cb39RrBZ9FDMjDoBlKZslyvRzTez33An84JAgwOBtEbaSTAkVqvPmDin3oZhTYbwwDc9SIBVsYhI6VmbjcPkMAFIeoKbS4KzweXneeKBB9FbozcgvnYrv3lTqofVVWONY/EL9q7M="
	CommPrim2A = "xyC4Jl6ojvL+uF2/iK9kRj3yQh8bV2ngl/fongysmUvxCZrwxaEaOZcBHreTiP6SFPOrWCyk6e9zHjtDPP/LhxrHsaiFapv6AjQejML/gCyFj4GRWMzayFBJlW6prsjZfhNG6FpQbFrEj8FtYdM0vRLyDyzeknrC66PJtwEcR6E="

	PubB1 = "5d7eefd58e3d2f309471928ab4bbd104e52973372c159fa652b8ca6b57ff68b8"
	PubB2 = "ab301a6a77b201c416ddc13a2d33fdf200a5302f6f687e0ea09085debaf8a1d9"
	PrivB = "7a0a9babcc97ea7991ed67ed7f800f70c5e04e99718960ad8efab2ca052f00c7"
	CommPubB = "n7vb+4YNgTDwjJ1St3/UQP+bXrN/mMmsPgTjKthIMpoMYN7mRhpk6/MGa6Gv0p1Zbw39g6fVsluHSXvyYO6VmsahTQ0gI9MEmxgKt4c6ZQct6M+kWP7E3omXT68NsXXXaZBjBuewfHrJReTz/znbS66HgY8BML55YDRKQBsmDz+cb/H6FWT7/mmPBRXufz7sf6OqvwiMRGXNlRbktbEn3gpumXpndlGhmGL0ZVZj2VklqWSHtgsfBut+rov7uuIN28StPZYZvllnCCvP1DHeImExWHOltWTnZAE0pRUbaX3q3NVAqU4ngL1sbkMSghF8bmz8G26qawM7YNiiDrAmcQ=="
	CommPrivB = "P0og9Hz99tVcSmq/boOQpxxgBFrc0L3/qCcplz1RBfOxueQ3m0kz+aU2QwkycCH2YKFLdJHYgy3u4bfhpnSCBGx1VuE/fdJLfeQ9wtAq3ALHNvqm5Lg1avNbZ7A1nb3SVzplckP00q2X+ECqSNM0x7zkZfoyf4zI7MxrKxFWuC1c1BT7zj7EUT1idG+n/yz3WCx4Xr+4XM3CIt1dTrddhCboLdLlNYCOIh4t5JSTfYysp8YR4FSc96vRVCe+QCVtMOfo7RCR8bcZDoIQjat+u5umnyAsyXLetBerh/MABqHq8wOgC6a8vCqRnyAwhLOT+VQbTbFMQzLO9Lw9T8v9EQ=="
	CommPrim1B = "zzoomDPTH/WxxtqTIApnecinr+BuAhcxJephkDHOhlRWK1IH16yLIal9V6OmC/REGCLgZpJHzUeesATn7QnsTIFnmEDKxIPVk54etYAXJo8G51pB9mylTUJiXqY1hu5O1GSEgtD+EAdHrIRJZ2E7/Pyp/wFrLG5ymXULZ5BFicU="
	CommPrim2B = "xVQiK60JgjOqSdQ6EKEjDxdxfWp1NDGpHbBLElzbqJyAHfd5KCPdwASLIR8V0WHIa12df877xGGL1W+SlXXXOsJaER+FfnlzxzaO5D8a3GqaYMJBYWyUBnf1f0/lgVvnJzh0hKHlKSlvJX2mbObD9mPeuYhEXNO1v7Vo0846sL0="

	//Root account for testing
	PubRoot1 = "6323cc034597195ae69bcfb628ecdffa5989c7503154c566bab4a87f3e9910ac"
	PubRoot2 = "f6115b77a15852764c609c6a5c1739e698ebc6e49bf14617c561b9110039cec7"
	PrivRoot = "277ed539f56122c25a6fc115d07d632b47e71416c9aebf1beb54ee704f11842c"
	CommPubRoot = "1e2QBjDop/b9Gk4U1YUtxzTpDrMvFTNb4dFIm2mIxhimeiJtHKnc0xDR1LPqkHN9Ke+tCbg6T3csbONoj8NT+ePIYF97DuUUL9d0ok8QZaSoAOGVIQHLbdCE08zwq8qiwzFWsfJSyKVJe1Bwbjsp9OWaxHenA3f2SWALiK1ZHAA13YV+nxm5Jh2O4uSmmz3PLv7Iz7Lfpo1uhpa0qfWap8Eqsp1XSWj60yms+hfy3X/r57FrbHUjJqeVQUPOqPmRRl3r3j1P+l/b+WQNA0WYu1ArjI8T3BEohqLZW3tZcx4NssyVyiS59SU16Yu3qroAdkLnFP4YPBSgQhXRjVzt8w=="
	CommPrivRoot = "jKphuoBsaw1wDdzrvB6PJF65JE5UFjeoIgswF+jD46YPyV1bq65RooN7xcXr5cHaujl76Vk3FkuBbbP2bBl+3WCWwC/oRboBlRex/IvKd1tWkQXDvmlkrzeeL3qhggSDE6AcpnN1VbPBZpFU7FaA1yQmqSsYKaK20jaSPvPlFRAllP1adSd+m3ZrJY5rPWzPkPDmeyLRhbTPMp2ke3gAVXn2JdX6hYwYBeZJv2ZnDM/ZQfmWezHJpjsaichnbB8mUiHOOqBnGXaHKKomgmveZ+UjLD7QN9x12NfRyhFM7Aih8iAgbK06CNzBMPvj4J3MGrJrZ1sjqpOw7ljLiGccGQ=="
	CommPrimRoot1 = "/dCNZfqFkgE3360DnH+wE9eR1KL0xdjC3XY+0ge2rkg3XxJc2hZsv0MO2JiGuqQBsAfjEtJCmqayJaTemPMHBJABrhJnfLaDL2fHLRzwGzGYvEd2LVTGqOOW5+0qfimEV5dwnCVE7CcZ/uXwH0R2baQzWN2S29DxEq706Bhtpsc="
	CommPrimRoot2 = "18UX/SiRcgUsjXlFsurz9Vjuh05vpREyn5wlNIHLjqmm5Rg/A37/gKiVls7o6CArM344Vhc6e5KpTVcM9FY496Xr/CNBA4ujl71hr93D6oCo50fLXdABoGFBjDP9d8eguK/Bi4L9q1MptTf2UM5JV/IkntVy1X4ff50uh1J5o3U="
)

//Globally accessible values for all other tests, (root)account-related
var (
	accA, accB, validatorAcc, rootAcc         	*protocol.Account
	PrivKeyAccA, PrivKeyAccB, PrivKeyRoot 	*ecdsa.PrivateKey
	CommPrivKeyAccA, CommPrivKeyAccB, CommPrivKeyRoot	   	*rsa.PrivateKey
	initialBlock *protocol.Block
)

//Create some accounts that are used by the tests
func addTestingAccounts() {
	accA, accB, validatorAcc = new(protocol.Account), new(protocol.Account), new(protocol.Account)

	pubAccA1, _ := new(big.Int).SetString(PubA1, 16)
	pubAccA2, _ := new(big.Int).SetString(PubA2, 16)
	privAccA, _ := new(big.Int).SetString(PrivA, 16)
	pubKeyAccA := ecdsa.PublicKey{
		elliptic.P256(),
		pubAccA1,
		pubAccA2,
	}
	PrivKeyAccA = &ecdsa.PrivateKey{
		pubKeyAccA,
		privAccA,
	}

	CommPrivKeyAccA, _ = crypto.CreateRSAPrivKeyFromBase64(CommPubA, CommPrivA, []string{CommPrim1A, CommPrim2A})

	copy(accA.Address[0:32], PrivKeyAccA.PublicKey.X.Bytes())
	copy(accA.Address[32:64], PrivKeyAccA.PublicKey.Y.Bytes())
	copy(accA.CommitmentKey[:], CommPrivKeyAccA.PublicKey.N.Bytes())

	pubAccB1, _ := new(big.Int).SetString(PubB1, 16)
	pubAccB2, _ := new(big.Int).SetString(PubB2, 16)
	privAccB, _ := new(big.Int).SetString(PrivB, 16)
	pubKeyAccB := ecdsa.PublicKey{
		elliptic.P256(),
		pubAccB1,
		pubAccB2,
	}
	PrivKeyAccB = &ecdsa.PrivateKey{
		pubKeyAccB,
		privAccB,
	}

	CommPrivKeyAccB, _ = crypto.CreateRSAPrivKeyFromBase64(CommPubB, CommPrivB, []string{CommPrim1B, CommPrim2B})

	copy(accB.Address[0:32], PrivKeyAccB.PublicKey.X.Bytes())
	copy(accB.Address[32:64], PrivKeyAccB.PublicKey.Y.Bytes())
	copy(accB.CommitmentKey[:], CommPrivKeyAccB.PublicKey.N.Bytes())

	privKeyValidator, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)

	copy(validatorAcc.Address[:32], privKeyValidator.X.Bytes())
	copy(validatorAcc.Address[32:64], privKeyValidator.Y.Bytes())

	//Create and store an initial commitment key for the validator account.
	commPrivKeyValidator, _ := rsa.GenerateMultiPrimeKey(rand.Reader, crypto.COMM_NOF_PRIMES, crypto.COMM_KEY_BITS)
	copy(validatorAcc.CommitmentKey[:], commPrivKeyValidator.PublicKey.N.Bytes()[:])

	validatorAcc.Balance = activeParameters.Staking_minimum
	validatorAcc.IsStaking = true

	//Set the global variable in blockchain.go
	validatorAccAddress = validatorAcc.Address
	commPrivKey = commPrivKeyValidator

	storage.State[accA.Address] = accA
	storage.State[accB.Address] = accB
	storage.State[validatorAcc.Address] = validatorAcc
}

//Create some root accounts that are used by the tests
func addRootAccounts() {
	rootAcc = new(protocol.Account)

	pubRoot1, _ := new(big.Int).SetString(PubRoot1, 16)
	pubRoot2, _ := new(big.Int).SetString(PubRoot2, 16)
	privRoot, _ := new(big.Int).SetString(PrivRoot, 16)
	pubKeyRoot := ecdsa.PublicKey{
		elliptic.P256(),
		pubRoot1,
		pubRoot2,
	}
	PrivKeyRoot = &ecdsa.PrivateKey{
		pubKeyRoot,
		privRoot,
	}

	copy(rootAcc.Address[:32], PrivKeyRoot.X.Bytes())
	copy(rootAcc.Address[32:64], PrivKeyRoot.Y.Bytes())

	//Create root file
	file, _ := os.Create(TestKeyFileName)
	_, _ = file.WriteString(PubRoot1 + "\n")
	_, _ = file.WriteString(PubRoot2 + "\n")
	_, _ = file.WriteString(PrivRoot + "\n")

	CommPrivKeyRoot, _ = crypto.CreateRSAPrivKeyFromBase64(CommPubRoot, CommPrivRoot, []string{CommPrimRoot1, CommPrimRoot2})
	copy(rootAcc.CommitmentKey[:], CommPrivKeyRoot.PublicKey.N.Bytes()[:])

	rootAcc.Balance = activeParameters.Staking_minimum
	rootAcc.IsStaking = true

	storage.State[rootAcc.Address] = rootAcc
	storage.RootKeys[rootAcc.Address] = rootAcc
}

//The state changes (accounts, funds, system parameters etc.) need to be reverted before any new test starts
//So every test has the same view on the blockchain
func cleanAndPrepare() {
	storage.DeleteAll()

	tmpState := make(map[[64]byte]*protocol.Account)
	tmpRootKeys := make(map[[64]byte]*protocol.Account)

	storage.State = tmpState
	storage.RootKeys = tmpRootKeys

	globalBlockCount = -1
	localBlockCount = -1

	//Prepare system parameters
	targetTimes = []timerange{}
	currentTargetTime = new(timerange)
	target = append(target, 8)

	var tmpSlice []Parameters
	tmpSlice = append(tmpSlice, NewDefaultParameters())

	slashingDict = make(map[[64]byte]SlashingProof)

	parameterSlice = tmpSlice
	activeParameters = &tmpSlice[0]

	slashingDict = make(map[[64]byte]SlashingProof)

	//Override some params to ensure tests work correctly.
	activeParameters.num_included_prev_proofs = 0
	activeParameters.Block_reward = 1
	activeParameters.Slash_reward = 1

	addTestingAccounts()
	addRootAccounts()

	genesis := protocol.NewGenesis(
		crypto.GetAddressFromPubKey(&PrivKeyRoot.PublicKey),
		crypto.GetBytesFromRSAPubKey(&CommPrivKeyRoot.PublicKey))

	lastEpochBlock = protocol.NewEpochBlock([][32]byte{genesis.Hash()}, 0)
	storage.WriteClosedEpochBlock(lastEpochBlock)

	storage.DeleteAllLastClosedEpochBlock()
	storage.WriteLastClosedEpochBlock(lastEpochBlock)

	commitmentProof, _ := crypto.SignMessageWithRSAKey(CommPrivKeyRoot, "1")
	initialBlock = newBlock(lastEpochBlock.HashEpochBlock(), commitmentProof, 1)
	storage.WriteClosedBlock(initialBlock)
	lastBlock = initialBlock
	lastBlock.Hash = lastBlock.HashBlock()

	collectStatistics(initialBlock)
	if err := storage.WriteClosedBlock(initialBlock); err != nil {
		fmt.Printf("Error: %v\n", err)
	}
	if err := storage.WriteLastClosedBlock(initialBlock); err != nil {
		fmt.Printf("Error: %v\n", err)
	}

	//Some meaningful balance to simplify testing
	//validatorAcc.Balance = 1000
	accA.Balance = 123232345678
	accB.Balance = 823237654321
	accA.TxCnt = 0
	accB.TxCnt = 0

	NumberOfShards = DetNumberOfShards()

	var validatorShardMapping = protocol.NewMapping()
	validatorShardMapping.ValMapping = AssignValidatorsToShards()
	validatorShardMapping.EpochHeight = int(lastEpochBlock.Height)
	ValidatorShardMap = validatorShardMapping
}

func TestMain(m *testing.M) {
	storage.Init(TestDBFileName, TestIpPort)
	p2p.Init(TestIpPort)
	p2p.InitLogging()

	logger = storage.InitLogger()
	FileLogger = storage.InitFileLogger()
	FileConnectionsLog, _ = os.OpenFile(fmt.Sprintf("hlog-for-%v.txt",strings.Split(p2p.Ipport, ":")[1]), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	FileLogger.SetOutput(FileConnectionsLog)

	cleanAndPrepare()
	addTestingAccounts()
	addRootAccounts()
	////We don't want logging msgs when testing, we have designated messages
	//logger = log.New(nil, "", 0)
	//logger.SetOutput(ioutil.Discard)
	retCode := m.Run()

	//Teardown
	storage.TearDown()
	os.Remove(TestDBFileName)
	os.Remove(TestKeyFileName)
	os.Exit(retCode)
}
