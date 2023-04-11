package main

import (
	"log"
	"net"
	"testing"
)

func TestUdp(t *testing.T) {
	// TODO
	cnn, err := net.DialUDP("udp", nil, &net.UDPAddr{
		IP:   net.IPv4(127, 0, 0, 1),
		Port: 8081,
	})

	if err != nil {
		panic(err)
	}

	if _, err := cnn.Write([]byte("hello world")); err != nil {
		panic(err)
	}

	buf := make([]byte, 1024)
	if _, _, err := cnn.ReadFromUDP(buf); err != nil {
		panic(err)
	}
	log.Printf("%s", buf)
}
