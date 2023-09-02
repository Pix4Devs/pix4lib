package socks

import (
	"encoding/binary"
	"errors"
	"fmt"
	"net"
	"time"
)

/*
SOCKS4 implementation

Implements only:
 - Auto resolve domain A record (ip)
 - CMD CONNECT
*/

const (
	SOCKS4_VERSION byte = 0x04 // SOCKS4 version
	SOCKS4_ID byte = 0x00 // user id null
)

var (
	// SOCKS4 REPLIES
	SOCKS4_REPS = map[byte]string{
		0x5A: "Request granted",
		0x5B: "Request rejected or failed",
		0x5C: "Request failed because client is not running identd (or not reachable from server)",
		0x5D: "Request failed because client's identd could not confirm the user ID in the request",
	}
)

type (
	// SOCKS4 Client
	SOCKS4_Client struct {
		Timeout time.Duration // READ & WRITE TIMEOUT
	}
	
	// TCP Connection
	SOCKS4_Conn struct {
		*net.TCPConn
	}
)

// Initializes an SOCKS4 client
func NewSocks4Client(timeout time.Duration) *SOCKS4_Client {
	return &SOCKS4_Client{timeout}
}

// Connect connects to the proxy server and tunnels the connection from client to the proxy and fromt he proxy to the target.
// You can use the returned TCPConn to do I/O stuff directly to the target through the PROXY server
func (c *SOCKS4_Client) Connect(proxy ProxyCtx, target TargetCtx) (*net.TCPConn, error){
	if isIPv6(proxy.IP) || isIPv6(target.IP) {
		return nil, errors.New("Proxy IP or Target IP is ipv6 but only domain names or ipv4 are supported")
	}

	if !isIPv4(proxy.IP){
		return nil, errors.New("Proxy IP should be ipv4")
	}

	if isDomain(target.IP){
		ips, err := net.LookupIP(target.IP); if err != nil {
			return nil, err
		}
		target.IP = ips[0].String()
	}

	conn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", proxy.IP, proxy.Port)); if err != nil {
		return nil, err
	}

	conn.SetDeadline(time.Now().Add(c.Timeout))

	ip := net.ParseIP(target.IP)
	header := []byte{
		SOCKS4_VERSION,
		CMD_CONNECT,
	}

	port := make([]byte, 2)
	binary.BigEndian.PutUint16(port, uint16(target.Port))
	
	header = append(header, port...)
	header = append(header, ip.To4()...)
	header = append(header, SOCKS4_ID)

	if _, err := conn.Write(header); err != nil {
		return nil, err
	}

	rep := make([]byte, 2)
	if _, err := conn.Read(rep); err != nil {
		fmt.Println("hier")
		return nil, err
	}

	if rep[1] != 0x5A {
		return nil, errors.New(fmt.Sprintf("Request to target from proxy failed with code: %s", SOCKS4_REPS[rep[1]]))
	}
	return conn.(*net.TCPConn), nil
}
