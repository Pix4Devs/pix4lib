package main

import (
	"log"
	"time"

	"github.com/Pix4Devs/pix4lib/socks"
)

func main() {
	proxy := socks.ProxyCtx{
		IP: "184.170.245.148",
		Port: 4145,
	}
    // if IP is an domain the IP will be automatically extracted
	target := socks.TargetCtx{
		IP: "google.com",
		Port: 80,
	}

	conn, err := socks.NewSocks4Client(time.Second*5).Connect(proxy,target); if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	// do whatever you want with this connection
	// it is now tunneled through your ip -> proxy -> target
}