# GoLang TP-Link Module

## Table of Contents

1. [Information](#information)
1. [Credit](#credit)
1. [Usage](#usage)
1. [Disclaimer](#disclaimer)
1. [License](#license)

## Information

This is a support package for interacting and obtaining information from [tp-link](//www.tp-link.com) Smart Wi-Fi devices.

* HS105 - Smart Plug
* HS200 - Smart Switch

## Credit

* [tp-link](//www.tp-link.com) - Without the products supplied by this company my home automation would rely on cloud based functions. (Yuck!)
* [sausheong](//github.com/sausheong) - Sausheong's [hs1xxplug](//github.com/sausheong/hs1xxplug) was what the majority of this project was built from, but failed to fully support the HS105 at the time.
* [jaedle](//github.com/jaedle) - The [connector](//github.com/jaedle/golang-tplink-hs100/blob/master/internal/connector/connector.go) written by jaedle was instrumental in helping me understand how [tp-link](//www.tp-link.com) devices communicated.

## Usage 

Import this library:

``` bash
go get github.com/TheSp1der/tplink
```

Get information from this device:

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

## Disclaimer

The tp-link company and products are trademarks™ or registered® trademarks of their respective holders. Use of them does not imply any affiliation with or endorsement by them. I claim no ownership or control of the tp-link company, products, name, logos or intellectual property.

## License