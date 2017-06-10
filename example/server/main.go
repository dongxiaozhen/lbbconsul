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

var cfg lbbconsul.ConsulConfig

func main() {
	flag.StringVar(&cfg.Ip, "ip", "127.0.0.1", "server ip")
	flag.IntVar(&cfg.Port, "port", 9527, "server port")
	flag.StringVar(&cfg.ServerId, "sid", "serverNode_1_1", "server id")
	flag.StringVar(&cfg.ServerName, "sname", "serverNode_1", "server name")
	flag.StringVar(&cfg.MAddr, "maddr", "127.0.0.1:8889", "monitor addr")
	flag.StringVar(&cfg.CAddr, "caddr", "127.0.0.1:8500", "consul addr")
	flag.Parse()
	cfg.MInterval = "5s"
	cfg.MTimeOut = "2s"
	cfg.MMethod = "http"
	exist := make(chan os.Signal, 1)
	signal.Notify(exist, syscall.SIGTERM)

	err := lbbconsul.GConsulClient.Open(&cfg)
	if err != nil {
		fmt.Println("GConsulClient,open err:", err)
		return
	}

	ln, err := net.Listen("tcp", fmt.Sprintf("%s:%d", cfg.Ip, cfg.Port))

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
