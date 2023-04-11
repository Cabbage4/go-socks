package main

import (
	"encoding/binary"
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"sort"
	"strings"
)

var (
	port int

	isRequireAuth bool
	username      string
	password      string

	methodHandle = map[byte]func(net.Conn) error{
		0: noAuthHandle,
		2: usernamePasswordAuthHandle,
	}
)

func main() {
	flag.IntVar(&port, "port", 18888, "")
	flag.BoolVar(&isRequireAuth, "requireAuth", false, "")
	flag.StringVar(&username, "username", "admin", "")
	flag.StringVar(&password, "password", "admin", "")
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
	defer cnn.Close()

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
	defer sCnn.Close()

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
	var netProtocol string
	if cmd == 1 {
		netProtocol = "tcp"
	} else if cmd == 3 {
		netProtocol = "udp"
	} else {
		return nil, fmt.Errorf("unsupported cmd:%d", cmd)
	}

	// 预留字段，无用的
	var rsv byte
	if err := binary.Read(cnn, binary.BigEndian, &rsv); err != nil {
		return nil, err
	}

	// 目的地址类型
	var dstAddrType byte
	if err := binary.Read(cnn, binary.BigEndian, &dstAddrType); err != nil {
		return nil, err
	}
	if dstAddrType != 1 && dstAddrType != 3 && dstAddrType != 4 {
		return nil, fmt.Errorf("unsupported addrType:%d", dstAddrType)
	}

	// 目标地址解析
	var dstAddrPrefix string
	if dstAddrType == 1 {
		// ipv4
		var ipByte [4]byte
		if err := binary.Read(cnn, binary.BigEndian, &ipByte); err != nil {
			return nil, err
		}
		dstAddrPrefix = net.IP(ipByte[:]).String()
	} else if dstAddrType == 3 {
		// 域名
		var domainLen byte
		if err := binary.Read(cnn, binary.BigEndian, &domainLen); err != nil {
			return nil, err
		}
		domain := make([]byte, domainLen)
		if err := binary.Read(cnn, binary.BigEndian, &domain); err != nil {
			return nil, err
		}
		dstAddrPrefix = string(domain)
	} else if dstAddrType == 4 {
		// ipv6
		var ipByte [16]byte
		if err := binary.Read(cnn, binary.BigEndian, &ipByte); err != nil {
			return nil, err
		}
		dstAddrPrefix = fmt.Sprintf("[%s]", net.IP(ipByte[:]).String())
	}

	// 目标端口解析
	var dstPort [2]byte
	if err := binary.Read(cnn, binary.BigEndian, &dstPort); err != nil {
		return nil, err
	}
	dstAddr := fmt.Sprintf("%s:%d", dstAddrPrefix, int(dstPort[0])<<8+int(dstPort[1]))

	// 生成目标连接句柄
	dstCnn, err := net.Dial(netProtocol, dstAddr)
	if err != nil {
		return nil, err
	}

	// 通知客户端
	rspByte := []byte{5, 0, 0, 1}
	bindAddr := net.ParseIP(strings.Split(dstCnn.LocalAddr().String(), ":")[0]).To4()
	bindPort := []byte{byte(port / 256), byte(port % 256)}
	rspByte = append(rspByte, bindAddr...)
	rspByte = append(rspByte, bindPort...)
	if _, err := cnn.Write(rspByte); err != nil {
		return nil, err
	}

	return dstCnn, nil
}

func noAuthHandle(cnn net.Conn) error {
	if isRequireAuth {
		return fmt.Errorf("need auth")
	}

	if _, err := cnn.Write([]byte{5, 0}); err != nil {
		return err
	}
	return nil
}

func usernamePasswordAuthHandle(cnn net.Conn) error {
	if _, err := cnn.Write([]byte{5, 2}); err != nil {
		return err
	}

	var version byte
	if err := binary.Read(cnn, binary.BigEndian, &version); err != nil {
		log.Println(err)
		return err
	}
	if version != 1 {
		return fmt.Errorf("version error:%d", version)
	}

	var tUsernameLen byte
	if err := binary.Read(cnn, binary.BigEndian, &tUsernameLen); err != nil {
		log.Println(err)
		return err
	}
	tUsername := make([]byte, tUsernameLen)
	if err := binary.Read(cnn, binary.BigEndian, &tUsername); err != nil {
		log.Println(err)
		return err
	}

	var tPasswordLen byte
	if err := binary.Read(cnn, binary.BigEndian, &tPasswordLen); err != nil {
		log.Println(err)
		return err
	}
	tPassword := make([]byte, tPasswordLen)
	if err := binary.Read(cnn, binary.BigEndian, &tPassword); err != nil {
		log.Println(err)
		return err
	}

	if string(tUsername) != username || string(tPassword) != password {
		return fmt.Errorf("username or password error")
	}

	if _, err := cnn.Write([]byte{1, 0}); err != nil {
		return err
	}

	return nil
}
