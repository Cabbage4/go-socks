package main

import (
	"log"
	"net"
)

func main() {
	ln, err := net.ListenUDP("udp", &net.UDPAddr{
		IP:   net.IPv4(0, 0, 0, 0),
		Port: 8081,
	})
	if err != nil {
		panic(err)
	}

	buf := make([]byte, 1024)
	for {
		_, addr, err := ln.ReadFromUDP(buf)
		if err != nil {
			log.Println(err)
			continue
		}

		log.Printf("%s\n", buf)

		if _, err := ln.WriteToUDP([]byte("done"), addr); err != nil {
			log.Println(err)
		}
	}
}
