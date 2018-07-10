package storage

const (
	//Aduno
	//BOOTSTRAP_SERVER_IP = "104.40.213.93"
	//CSG
	BOOTSTRAP_SERVER_IP = "192.41.136.199"
	//Local
	//BOOTSTRAP_SERVER_IP   = "127.0.0.1"
	BOOTSTRAP_SERVER_PORT = ":8000"

	BOOTSTRAP_SERVER = BOOTSTRAP_SERVER_IP + BOOTSTRAP_SERVER_PORT

	//Root Public Keys at initialization time. This is the only existing account at startup
	//All other accounts are created
	INITROOTPUBKEY1 = "f64180705b5d2cf85f8f3ac046ce633c1d00f23a9e7483598d6d6267c40cc811"
	INITROOTPUBKEY2 = "18f9a1c5accb44c4529b0d3c51f622997dc11881f44771aacf3c2359a67b04ed"

	GENESIS_SEED = "FfhHEtSNQO6JyAUcKPlrlWzUqFXo1EoB"

	//File name under which the seed will be stored
	SEED_FILE_NAME = "seed.json"

	DEFAULT_KEY_FILE_NAME = "root"

	INIT_ROOT_SEED = "GfhHEtSNQO6JyAUcKPlrlWzUqFXo1EoB"
)
