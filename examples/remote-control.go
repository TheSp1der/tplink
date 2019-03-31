package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/TheSp1der/tplink"
)

func main() {
	var (
		host     string
		getState bool
		turnOn   bool
		turnOff  bool
	)

	flag.StringVar(&host, "host", "", "hostname of device")
	flag.BoolVar(&getState, "get-state", false, "get device power state")
	flag.BoolVar(&turnOn, "on", false, "turn the device on")
	flag.BoolVar(&turnOff, "off", false, "turn the device off")
	flag.Parse()

	if len(strings.TrimSpace(host)) == 0 {
		fmt.Println("host must be defined")
		flag.PrintDefaults()
		os.Exit(100)
	}

	if !getState && !turnOn && !turnOff {
		fmt.Println("an action must be defined")
		flag.PrintDefaults()
		os.Exit(100)
	}

	if getState && (turnOn || turnOff) {
		fmt.Println("only one action may be defined")
		flag.PrintDefaults()
		os.Exit(100)
	}

	if turnOn && turnOff {
		fmt.Println("only one action may be defined")
		flag.PrintDefaults()
		os.Exit(100)
	}

	device := tplink.Tplink{Host: host}

	if getState {
		response, err := device.SystemInfo()
		if err != nil {
			fmt.Println("unable to communicate with device")
			os.Exit(100)
		}
		if response.System.GetSysinfo.RelayState == 1 {
			fmt.Println(host + " is ON")
		} else {
			fmt.Println(host + " is OFF")
		}
	}

	if turnOn {
		if err := device.TurnOn(); err != nil {
			fmt.Println("unable to communicate with device")
			os.Exit(100)
		}
	}
	if turnOff {
		if err := device.TurnOff(); err != nil {
			fmt.Println("unable to communicate with device")
			os.Exit(100)
		}
	}
}
