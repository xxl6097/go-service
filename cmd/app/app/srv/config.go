package srv

type Config struct {
	//ukey.KeyBuffer
	AppTesting string `json:"appTesting"`
	ServerPort int    `json:"serverPort"`
}
