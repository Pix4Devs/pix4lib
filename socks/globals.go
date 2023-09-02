package socks

const (
	CMD_CONNECT byte = 0x01 // CMD CONNECT
	RESERVED    byte = 0x00 // RESERVED FIELD
)

type (
	// CONNECT METHOD, see: SOCKS5_AUTH_METHODS
	SOCKS5_METHOD byte

	// Auth func, should return user and pass strings
	// nil for no auth
	Auth struct {
		User string
		Pass string
	}

	// Proxy server response
	ServerResponse []byte

	// Can be IPV6, IPV4 or domain name
	Target string

	// Proxy ip:port
	ProxyCtx struct {
		IP   string
		Port int
	}

	// Target ip:port, for SOCKS5 client only; the IP can be an domain!
	TargetCtx struct {
		IP   string
		Port int
	}
)