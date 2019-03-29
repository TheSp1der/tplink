# GoLang TP-Link Module

``` go
package main

import (
	"fmt"
	"log"

	"github.com/TheSp1der/tplink"
)

func main() {
	d := tplink.Tplink{HostName: "light-office.tplink.example.com"}
	r, err := d.SystemInfo()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(r)
}
```