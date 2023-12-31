package socks

import (
	"encoding/binary"
	"errors"
	"fmt"
	"net"
	"time"
)

/*
SOCKS5 implementation

Implements only:
 - Auth
 - No auth
 - CMD CONNECT
*/

const (
	SOCKS5_Version        byte = 0x05 // SOCKS5 VERSION
	SOCKS5_METHOD_NO_AUTH byte = 0x00 // NO AUTHENTICATION REQUIRED
	SOCKS5_METHOD_AUTH    byte = 0x02 // USERNAME/PASSWORD

	// FOR CMD CONNECT
	SOCKS5_ATYP_IPV4   byte = 0x01 // IPV4
	SOCKS5_ATYP_IPV6   byte = 0x04 // IPV6
	SOCKS5_ATYP_DOMAIN byte = 0x03 // DOMAIN
)

var (
	// PROXY REPS
	SOCKS5_REPS = map[byte]string{
		0x00: "succeeded",
		0x01: "general SOCKS server failure",
		0x02: "connection not allowed by ruleset",
		0x03: "Network unreachable",
		0x04: "Host unreachable",
		0x05: "Connection refused",
		0x06: "TTL expired",
		0x07: "Command not supported",
		0x08: "Address type not supported",
		0x09: "to X'FF' unassigned",
		0xFF: "unassigned",
	}

	// AUTH METHODS
	SOCKS5_AUTH_METHODS = map[byte]string{
		0x00: "NO AUTHENTICATION REQUIRED",
		0x01: "GSSAPI",
		0x02: "USERNAME/PASSWORD",
		0x03: "to X'7F' IANA ASSIGNED",
		0x80: "to X'FE' RESERVED FOR PRIVATE METHODS",
		0xFF: "NO ACCEPTABLE METHODS",
	}
)

type (
	// Client
	SOCKS5_Client struct {
		Timeout time.Duration // READ & WRITE TIMEOUT
	}
	
	// SOCKS5 Connection
	SOCKS5_Conn struct {
		*net.TCPConn
	}
)

func(c *SOCKS5_Conn) CloseConnection() error {
	if err := c.Close(); err != nil { return err }
	return nil
}

// Create new SOCKS5 client
func NewSocks5Client(timeout time.Duration) *SOCKS5_Client{
	return &SOCKS5_Client{
		timeout,
	}
}

// Does all the low level stuff and returns an active TCP connection from your ip -> proxy -> target server.
// Any I/O you perform from this connection is directly relayed as your ip -> proxy -> target server
func (c *SOCKS5_Client) Connect(
	proxyCtx ProxyCtx,
	authenticate bool, 
	auth Auth, 
	targetCtx TargetCtx,
) (*net.TCPConn, error) {
	if authenticate && (auth.User == ""  || auth.Pass == ""){
		return nil, errors.New("User/Pass cannot be NILL if authenticate is set to true")
	}

	m := SOCKS5_METHOD_NO_AUTH
	if authenticate {
		m = SOCKS5_METHOD_AUTH
	}

	conn, _, err := c.connectToProxy(proxyCtx.IP, proxyCtx.Port, SOCKS5_METHOD(m))
	if err != nil {
		return nil, err
	}

	if authenticate {
		if err := conn.establishAuth(auth.User, auth.Pass); err != nil {
			return nil, err
		}
	}

	atyp := []byte{}
	switch(true){
		case isIPv4(string(targetCtx.IP)):
			ipv4 := net.ParseIP(string(targetCtx.IP))
			atyp = append(atyp, SOCKS5_ATYP_IPV4)
			atyp = append(atyp, ipv4.To4()...)
		case isIPv6(string(targetCtx.IP)):
			ipv6 := net.ParseIP(string(targetCtx.IP))
			atyp = append(atyp, SOCKS5_ATYP_IPV6)
			atyp = append(atyp, ipv6.To16()...)
		case isDomain(string(targetCtx.IP)):
			atyp = append(atyp, SOCKS5_ATYP_DOMAIN)
			atyp = append(atyp, byte(len(targetCtx.IP)))
			atyp = append(atyp, []byte(targetCtx.IP)...)
		default:
			return nil, errors.New("Given target is none of ipv4, ipv6 or domain name")
	}

	header := []byte{
		SOCKS5_Version,
		CMD_CONNECT,
		RESERVED,
	}
	header = append(header, atyp...)

	port := make([]byte,2)
	binary.BigEndian.PutUint16(port, uint16(targetCtx.Port))

	header = append(header, port...)
	
	if _, err := conn.Write(header); err != nil {
		return nil, err
	}

	rep := make([]byte, 5)
	if _, err := conn.Read(rep); err != nil {
		return nil, err
	}

	if rep[1] != 0x00 {
		return nil, errors.New(fmt.Sprintf("Proxy server declined connect request with code: %s", SOCKS5_REPS[rep[1]]))
	}
	return conn.TCPConn, nil
}

func (c *SOCKS5_Client) connectToProxy(proxy_host string, proxy_port int, method SOCKS5_METHOD) (*SOCKS5_Conn, ServerResponse, error) {
	conn, err := net.Dial("tcp",fmt.Sprintf("%s:%d", proxy_host, proxy_port)); if err != nil {
		return nil, nil, err
	}

	conn.SetDeadline(time.Now().Add(c.Timeout))

	header := []byte{
		SOCKS5_Version,
		0x01, // Number of METHODS
	}

	x := byte(method) != SOCKS5_METHOD_AUTH
	if x && byte(method) != SOCKS5_METHOD_NO_AUTH {
		return nil, nil, errors.New("Unsupported SOCKS5 method")
	}

	header = append(header, byte(method))
	if _, err := conn.Write(header); err != nil {
		return nil, nil, err
	}

	rep := make([]byte,2)
	if _, err := conn.Read(rep); err != nil {
		return nil, nil, err
	}

	if rep[0] != SOCKS5_Version || rep[1] != byte(method) {
		return nil, nil, errors.New(
			fmt.Sprintf("Invalid SOCKS version or METHOD selection reply: [ %v: %v ]", 
			rep[0], 
			rep[1]),
		)
	}

	return &SOCKS5_Conn{conn.(*net.TCPConn)}, rep, nil
}

func(c *SOCKS5_Conn) establishAuth(user, pass string) error {
	header := []byte{
		0x01, // Current Auth version for 0x02
		byte(len(user)),
	}

	header = append(header, []byte(user)...)
	header = append(header, byte(len(pass)))
	header = append(header, []byte(pass)...)

	if _, err := c.Write(header); err != nil {
		return err
	}

	rep := make([]byte,2)
	if _, err := c.Read(rep); err != nil {
		return err
	}

	if rep[1] != 0x00 {
		return errors.New("Failed establishing auth, invalid credentials")
	}
	return nil
}