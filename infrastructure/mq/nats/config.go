package nats

type Config struct {
	Addr               string // nats://127.0.0.1:4222
	MaxReconnects      int    // 5
	ReconnectWait      int    // 2s
	Token              string
	Username           string
	Password           string
	EmbeddedServerPort int
}