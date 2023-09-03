# SOCKS Client

> Our SOCKS client does currently not support BIND only CONNECT (relay)

SOCKS version 4 and 5 are supported!

With this client you can create an relay within ease:

```
your ip -> proxy -> target

!! direct I/O !!
```

**SOCKS5 support**

- Auth
- No auth
- IPV4, Domain, IPV6

**SOCKS4 support**

- IPV4, Domain
- FULL RFC

### Example SOCKS5 client

```go
package main

import (
	"log"
	"time"

	"github.com/Pix4Devs/pix4lib/socks"
)

func main() {
	proxy := socks.ProxyCtx{
		IP: "2.56.119.93",
		Port: 5074,
	}
	target := socks.TargetCtx{
		IP: "google.com",
		Port: 443,
	}

	// Supports authenticate and no auth, see docs for info
	conn, err := socks.NewSocks5Client(time.Second*5).Connect(proxy, true, socks.Auth{
		User: "rvugmczm",
		Pass: "3j296ogi3b86",
	} ,target); if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	// do whatever you want with this connection
	// it is now tunneled through your ip -> proxy -> target
	if _, err := conn.Write([]byte("GET / HTTP/1.1\r\nConnection: keep-alive\r\n")); err != nil {
		log.Fatal(err)
	}

	var buffer []byte
	for {
		chunk := make([]byte, 1042)
		n, err := conn.Read(chunk)
		if  err == io.EOF {
			break
		}

		if n != 0 {
			buffer = chunk[:n]
		}
	}

	fmt.Println(string(buffer))
}
```

### Example SOCKS4 client

```go
package main

import (
	"fmt"
	"io"
	"log"
	"time"

	"github.com/Pix4Devs/pix4lib/socks"
)

func main() {
	proxy := socks.ProxyCtx{
		IP: "192.111.130.5",
		Port: 17002,
	}
    // if IP is an domain the IP will be automatically extracted
	target := socks.TargetCtx{
		IP: "google.com",
		Port: 443,
	}

	conn, err := socks.NewSocks4Client(time.Second*5).Connect(proxy,target); if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	// do whatever you want with this connection
	// it is now tunneled through your ip -> proxy -> target
	if _, err := conn.Write([]byte("GET / HTTP/1.1\r\nConnection: keep-alive\r\n")); err != nil {
		log.Fatal(err)
	}

	var buffer []byte
	for {
		chunk := make([]byte, 1042)
		n, err := conn.Read(chunk)
		if  err == io.EOF {
			break
		}

		if n != 0 {
			buffer = chunk[:n]
		}
	}

	fmt.Println(string(buffer))
}
```
