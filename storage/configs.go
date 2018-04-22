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
	INITROOTPUBKEY1 = "d5a0c62eeaf699eeba121f92e08becd38577f57b83eba981dc057e92fde1ad22"
	INITROOTPUBKEY2 = "a480e4ee6ff8b4edbf9470631ec27d3b1eb27f210d5a994a7cbcffa3bfce958e"

	GENESIS_SEED = "FfhHEtSNQO6JyAUcKPlrlWzUqFXo1EoB"

	//File name under which the seed will be stored
	SEED_FILE_NAME = "seed.json"

	DEFAULT_KEY_FILE_NAME = "root"

	INIT_ROOT_SEED = "GfhHEtSNQO6JyAUcKPlrlWzUqFXo1EoB"
)
