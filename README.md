# GoLang TP-Link Module

``` go
package main

import (
    "github.com/TheSp1der/tplink"
    "fmt"
    "log"
)

func main() {
	d := tplink{HostName: "light-office.tplink.example.com"}
	r, err := d.SystemInfo()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(r)
}
```