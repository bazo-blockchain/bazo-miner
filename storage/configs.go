package storage

import "os"

const (
	//Aduno
	//BOOTSTRAP_SERVER_IP = "104.40.213.93"
	//CSG
	//BOOTSTRAP_SERVER_IP = "192.41.136.199"
	//Local
	BOOTSTRAP_SERVER_IP   = "127.0.0.1"
	BOOTSTRAP_SERVER_PORT = ":8000"

	BOOTSTRAP_SERVER = BOOTSTRAP_SERVER_IP + BOOTSTRAP_SERVER_PORT


	GENESIS_SEED = "FfhHEtSNQO6JyAUcKPlrlWzUqFXo1EoB"

	//File name under which the seed will be stored
	SEED_FILE_NAME = "seed.json"

	DEFAULT_KEY_FILE_NAME = "root"

	INIT_ROOT_SEED = "GfhHEtSNQO6JyAUcKPlrlWzUqFXo1EoB"
)


var (
	//Root Public Keys at initialization time. This is the only existing account at startup
	//All other accounts are created
	INITROOTPUBKEY1 = os.Getenv("INITROOTPUBKEY1")
	INITROOTPUBKEY2 = os.Getenv("INITROOTPUBKEY2")
)
