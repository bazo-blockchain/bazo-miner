package storage

const (
	//Root Public Keys at initialization time. This is the only existing account at startup
	//All other accounts are created
	INITROOTKEY1 = "f894ba7a24c1c324bc4b0a833d4b076a0e0f675a380fb7e782672c6568aaab06"
	INITROOTKEY2 = "69ddbc62f79cb521411840d83ff0abf941a8e717d81af3dfc2973f1bac30308a"
	INITPRIVKEY = "4e8cf4d82d1a376484659b632c0c506affe3c394d18266379c1c6b86eb5ba0fb"

	//Sha3-256 Hash of the MINER'S ECC public key (needs to be part of the state)
	BENEFICIARY = "1c90d27e539d035512d27d072f7b514753157fa1591ff5c5a8a9ef642449d291"

	GENESIS_SEED = "FfhHEtSNQO6JyAUcKPlrlWzUqFXo1EoB"

	//File name under which the seed will be stored
	SEED_FILE_NAME     = "seed.json"

	DEFAULT_KEY_FILE_NAME = "root"
)
