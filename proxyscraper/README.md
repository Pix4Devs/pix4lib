# Example
```go
package main

import (
	"fmt"
	"log"
	"time"

	"github.com/Pix4Devs/pix4lib/proxyscraper/proxyscraper"
)

func main(){
	client := proxyscraper.NewClient(time.Duration(time.Second * 5))
	proxies, err := client.Execute(); if err != nil {
		log.Fatal(err)
	}

	for _, proxy := range proxies {
		fmt.Println(proxy)
	}
}
```