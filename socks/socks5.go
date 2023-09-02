package socks

import (
	"encoding/binary"
	"errors"
	"fmt"
	"net"
	"regexp"
	"time"
)

/*
SOCKS5 IMPLENTATION

Implements only:
 - Auth
 - No auth
 - CMD CONNECT
*/

const (
	SOCKS5_Version        byte = 0x05 // SOCKS5 VERSION
	SOCKS5_METHOD_NO_AUTH byte = 0x00 // NO AUTHENTICATION REQUIRED
	SOCKS5_METHOD_AUTH    byte = 0x02 // USERNAME/PASSWORD
	SOCKS5_CMD_CONNECT    byte = 0x01 // CMD CONNECT
	SOCKS5_RESERVED       byte = 0x00 // RESERVED FIELD

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
	SOCKS5_Client struct {
		Timeout time.Duration // READ & WRITE TIMEOUT
	}
	
	SOCKS5_Conn struct {
		*net.TCPConn
	}

	// CONNECT METHOD, see: SOCKS5_AUTH_METHODS
	METHOD byte

	// Auth func, should return user and pass strings
	// nil for no auth
	AuthFunc func() (string, string)

	// Proxy server response
	ServerResponse []byte

	// Can be IPV6, IPV4 or domain name
	Target string
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

// All aorund wrapper
func (c *SOCKS5_Client) Connect(proxy_host string, proxy_port int, authenticate bool, auth AuthFunc, target Target, target_port int) error {
	user, pass := auth(); if authenticate && (user == ""  || pass == ""){
		return errors.New("User/Pass cannot be NILL if authenticate is set to true")
	}

	m := SOCKS5_METHOD_NO_AUTH
	if authenticate {
		m = SOCKS5_METHOD_AUTH
	}

	conn, _, err := c.connectToProxy(proxy_host, proxy_port, METHOD(m))
	if err != nil {
		return err
	}

	if authenticate {
		if err := conn.establishAuth(user,pass); err != nil {
			return err
		}
	}

	atyp := []byte{}
	switch(true){
		case isIPv4(string(target)):
			atyp = append(atyp, SOCKS5_ATYP_IPV4)
			atyp = append(atyp, []byte(target)...)
		case isIPv6(string(target)):
			atyp = append(atyp, SOCKS5_ATYP_IPV6)
			atyp = append(atyp, []byte(target)...)
		case isDomain(string(target)):
			atyp = append(atyp, SOCKS5_ATYP_DOMAIN)
			atyp = append(atyp, byte(len(target)))
			atyp = append(atyp, []byte(target)...)
		default:
			return errors.New("Given target is none of ipv4, ipv6 or domain name")
	}

	header := []byte{
		SOCKS5_Version,
		SOCKS5_CMD_CONNECT,
		SOCKS5_RESERVED,
	}

	fmt.Println(atyp)
	header = append(header, atyp...)

	port := make([]byte,2)
	binary.BigEndian.PutUint16(port, uint16(target_port))

	header = append(header, port...)
	

	if _, err := conn.Write(header); err != nil {
		return err
	}

	rep := make([]byte, 5)
	if _, err := conn.Read(rep); err != nil {
		return err
	}

	fmt.Println(rep)
	return nil
}

func (c *SOCKS5_Client) connectToProxy(proxy_host string, proxy_port int, method METHOD) (*SOCKS5_Conn, ServerResponse, error) {
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

	if rep[0] != SOCKS5_Version || rep[1] != SOCKS5_METHOD_AUTH {
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
		return errors.New("Failed establishing auth")
	}
	return nil
}
// etc etc komt nog

func isIPv4(input string) bool {
	regex := regexp.MustCompile(`^(?:(?:[0-9]{1,3}\.){3}[0-9]{1,3})$`)
	return regex.MatchString(input)
}
  
func isIPv6(input string) bool {
	regex := regexp.MustCompile(`^(?:[0-9a-fA-F]{1,4}:){7}[0-9a-fA-F]{1,4}$`)
	return regex.MatchString(input)
}

func isDomain(input string) bool {
	regex := regexp.MustCompile(`^[a-z0-9]([-a-z0-9]*[a-z0-9])?(\.[a-z0-9]([-a-z0-9]*[a-z0-9])?)*$`)
	return regex.MatchString(input)
}