package main

import (
)

func main() {
	h := tplink{HostName: "192.168.1.55"}
	r, err := h.SystemInfo()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(r)
}
