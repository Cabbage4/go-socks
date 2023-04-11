package main

import (
	"encoding/binary"
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"sort"
)

var (
	port int

	methodHandle = map[byte]func(net.Conn) error{
		0: noAuthHandle,
	}
)

func main() {
	flag.IntVar(&port, "port", 18888, "")
	flag.Parse()

	ln, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		panic(err)
	}

	for {
		cnn, err := ln.Accept()
		if err != nil {
			log.Println(err)
			continue
		}

		go worker(cnn)
	}
}

func worker(cnn net.Conn) {
	// shake hands
	if err := shakeHands(cnn); err != nil {
		log.Println(err)
		return
	}
	// parseCnn
	sCnn, err := parseCnn(cnn)
	if err != nil {
		log.Println(err)
		return
	}

	// io copy
	ch := make(chan bool, 2)

	go func() {
		if _, err := io.Copy(cnn, sCnn); err != nil {
			log.Println(err)
		}
		ch <- true
	}()

	go func() {
		if _, err := io.Copy(sCnn, cnn); err != nil {
			log.Println(err)
		}
		ch <- true
	}()

	<-ch
	cnn.Close()
	sCnn.Close()
}

func noAuthHandle(cnn net.Conn) error {
	if _, err := cnn.Write([]byte{5, 0}); err != nil {
		return err
	}
	return nil
}

func shakeHands(cnn net.Conn) error {
	var version byte
	if err := binary.Read(cnn, binary.BigEndian, &version); err != nil {
		log.Println(err)
		return err
	}
	if version != 5 {
		return fmt.Errorf("version error:%d", version)
	}
	var methodCount byte
	if err := binary.Read(cnn, binary.BigEndian, &methodCount); err != nil {
		return err
	}
	if methodCount == 0 {
		return fmt.Errorf("methodCount == 0")
	}

	methodList := make([]byte, methodCount)
	if err := binary.Read(cnn, binary.BigEndian, &methodList); err != nil {
		return err
	}
	sort.Slice(methodList, func(i, j int) bool {
		return methodList[i] > methodList[j]
	})

	var authError error
	for _, v := range methodList {
		handle, ok := methodHandle[v]
		if ok {
			authError = handle(cnn)
			break
		}
	}
	if authError != nil {
		return authError
	}
	return nil
}

func parseCnn(cnn net.Conn) (net.Conn, error) {
	var version byte
	if err := binary.Read(cnn, binary.BigEndian, &version); err != nil {
		return nil, err
	}
	if version != 5 {
		return nil, fmt.Errorf("version error:%d", version)
	}

	var cmd byte
	if err := binary.Read(cnn, binary.BigEndian, &cmd); err != nil {
		return nil, err
	}

	var rsv byte
	if err := binary.Read(cnn, binary.BigEndian, &rsv); err != nil {
		return nil, err
	}

	var addrType byte
	if err := binary.Read(cnn, binary.BigEndian, &addrType); err != nil {
		return nil, err
	}

	var ip string
	if addrType == 1 {
		var ipByte [4]byte
		if err := binary.Read(cnn, binary.BigEndian, &ipByte); err != nil {
			return nil, err
		}
		ip = net.IPv4(ipByte[0], ipByte[1], ipByte[2], ipByte[3]).String()
	} else if addrType == 4 {
		var ipByte [16]byte
		if err := binary.Read(cnn, binary.BigEndian, &ipByte); err != nil {
			return nil, err
		}
		ip = string(ipByte[:])
	} else {
		return nil, fmt.Errorf("unsupported cmd")
	}

	var dstPort [2]byte
	if err := binary.Read(cnn, binary.BigEndian, &dstPort); err != nil {
		return nil, err
	}

	var r net.Conn
	var err error
	if cmd == 1 {
		r, err = net.Dial("tcp", fmt.Sprintf("%s:%s", ip, dstPort))
		if err != nil {
			return nil, err
		}
	} else if cmd == 3 {
		r, err = net.Dial("udp", fmt.Sprintf("%s:%s", ip, dstPort))
		if err != nil {
			return nil, err
		}
	} else {
		return nil, fmt.Errorf("unsupported cmd")
	}

	if _, err := cnn.Write([]byte{5, 0, 0}); err != nil {
		return nil, err
	}

	return r, nil
}
