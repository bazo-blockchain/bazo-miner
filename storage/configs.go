package storage

const (
	//Aduno
	//BOOTSTRAP_SERVER_IP = "104.40.213.93"
	//CSG
	//BOOTSTRAP_SERVER_IP = "192.41.136.199"
	//Local
	BOOTSTRAP_SERVER_IP   = "127.0.0.1"
	BOOTSTRAP_SERVER_PORT = ":8000"

	BOOTSTRAP_SERVER = BOOTSTRAP_SERVER_IP + BOOTSTRAP_SERVER_PORT

	//Root Public Keys at initialization time. This is the only existing account at startup
	//All other accounts are created
	INITROOTPUBKEY1 = "f894ba7a24c1c324bc4b0a833d4b076a0e0f675a380fb7e782672c6568aaab06"
	INITROOTPUBKEY2 = "69ddbc62f79cb521411840d83ff0abf941a8e717d81af3dfc2973f1bac30308a"

	GENESIS_SEED = "FfhHEtSNQO6JyAUcKPlrlWzUqFXo1EoB"

	//File name under which the seed will be stored
	SEED_FILE_NAME = "seed.json"

	DEFAULT_KEY_FILE_NAME = "root"

	INIT_ROOT_SEED = "GfhHEtSNQO6JyAUcKPlrlWzUqFXo1EoB"
)
