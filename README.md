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
* KP200 - Smart Outlet

## Credit

* [tp-link](//www.tp-link.com) - Without the products supplied by this company my home automation would rely on cloud based functions. (Yuck!)
* [sausheong](//github.com/sausheong) - Sausheong's [hs1xxplug](//github.com/sausheong/hs1xxplug) was what the majority of this project was built from, but failed to fully support the HS105 at the time.
* [jaedle](//github.com/jaedle) - The [connector](//github.com/jaedle/golang-tplink-hs100/blob/master/internal/connector/connector.go) written by jaedle was instrumental in helping me understand how [tp-link](//www.tp-link.com) devices communicated.

## Usage 

### Import library

To import the library locally:

``` bash
go get github.com/TheSp1der/tplink
```

To import the library in your code:

``` go
import "github.com/TheSp1der/tplink"
```

### Interacting with the device

#### SystemInfo

To obtain some basic information from the device use the example code below. To return other objects please examine the SysInfo struct.

``` go
d := tplink.Tplink{Host: "light-office.tplink.example.com"}
r, err := d.SystemInfo()
if err != nil {
    log.Fatal(err)
}
fmt.Println("MAC: " + r.System.GetSysinfo.Mac)
fmt.Println("Name: " + r.System.GetSysinfo.Alias)

if r.System.GetSysinfo.RelayState == 1 {
    fmt.Println("Power: On")
} else {
    fmt.Println("Power: Off")
}
```

Will return data:

``` generic
MAC: 50:C7:BF:##:##:##
Name: Office Light
Power: On
```

#### Change power state

##### HSXXX or single switch/outlet device

To turn a device on or off:

``` go
// turn on
d := tplink.Tplink{
    Host: "light-office.tplink.example.com",
}
if err := d.ChangeState(1); err != nil {
	log.Fatal(err)
}

// turn off
d := tplink.Tplink{
    Host: "light-office.tplink.example.com",
}
if err := d.ChangeState(0); err != nil {
	log.Fatal(err)
}
```

##### KPXXX or multi switch/outlet device

The SwitchID is typically number starting at 0 for the first outlet/switch on the hardware.
The following code identifies the first outlet/switch on the device:

``` go
// turn on
d := tplink.Tplink{
    Host: "light-office.tplink.example.com",
    SwitchID: 1,
}
if err := d.ChangeStateMultiSwitch(1); err != nil {
	log.Fatal(err)
}

// turn off
d := tplink.Tplink{
    Host: "light-office.tplink.example.com",
    SwitchID: 1,
}
if err := d.ChangeStateMultiSwitch(0); err != nil {
	log.Fatal(err)
}
```

## Disclaimer

The tp-link company and products are trademarks™ or registered® trademarks of their respective holders. Use of them does not imply any affiliation with or endorsement by them. I claim no ownership or control of the tp-link company, products, name, logos or intellectual property.

## License

BSD 2-Clause License

Copyright (c) 2019, TheSp1der
All rights reserved.

Redistribution and use in source and binary forms, with or without modification, are permitted provided that the following conditions are met:

* Redistributions of source code must retain the above copyright notice, this list of conditions and the following disclaimer.

* Redistributions in binary form must reproduce the above copyright notice, this list of conditions and the following disclaimer in the documentation and/or other materials provided with the distribution.

THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS "AS IS" AND ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT LIMITED TO, THE IMPLIED WARRANTIES OF MERCHANTABILITY AND FITNESS FOR A PARTICULAR PURPOSE ARE DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT HOLDER OR CONTRIBUTORS BE LIABLE FOR ANY DIRECT, INDIRECT, INCIDENTAL, SPECIAL, EXEMPLARY, OR CONSEQUENTIAL DAMAGES (INCLUDING, BUT NOT LIMITED TO, PROCUREMENT OF SUBSTITUTE GOODS OR SERVICES; LOSS OF USE, DATA, OR PROFITS; OR BUSINESS INTERRUPTION) HOWEVER CAUSED AND ON ANY THEORY OF LIABILITY, WHETHER IN CONTRACT, STRICT LIABILITY, OR TORT (INCLUDING NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY OUT OF THE USE OF THIS SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF SUCH DAMAGE.