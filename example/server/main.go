package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/dongxiaozhen/lbbconsul"
)

func main() {
	flag.Parse()
	exist := make(chan os.Signal, 1)
	signal.Notify(exist, syscall.SIGTERM)

	err := lbbconsul.GConsulClient.Open(&lbbconsul.Ccfg)
	if err != nil {
		fmt.Println("GConsulClient,open err:", err)
		return
	}

	ln, err := net.Listen("tcp", fmt.Sprintf("%s:%d", lbbconsul.Ccfg.Ip, lbbconsul.Ccfg.Port))

	if nil != err {
		panic("Error: " + err.Error())
	}

	go func() {
		for {
			conn, err := ln.Accept()

			if err != nil {
				panic("Error: " + err.Error())
			}

			go EchoServer(conn)
		}
	}()

	<-exist
	lbbconsul.GConsulClient.Close()
}

const RECV_BUF_LEN = 1024

func EchoServer(conn net.Conn) {
	buf := make([]byte, RECV_BUF_LEN)
	defer conn.Close()

	for {
		n, err := conn.Read(buf)
		switch err {
		case nil:
			log.Println("get and echo:", "EchoServer "+string(buf[0:n]))
			conn.Write(append([]byte("EchoServer "), buf[0:n]...))
		case io.EOF:
			log.Printf("Warning: End of data: %s\n", err)
			return
		default:
			log.Printf("Error: Reading data: %s\n", err)
			return
		}
	}
}
