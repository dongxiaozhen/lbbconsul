package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/dongxiaozhen/lbbconsul"
)

var cfg lbbconsul.ConsulConfig
var foundServer string

var clients = []net.Conn{}
var proxys = []net.Conn{}

func main() {
	flag.StringVar(&cfg.Ip, "ip", "127.0.0.1", "server ip")
	flag.IntVar(&cfg.Port, "port", 2222, "server port")
	flag.StringVar(&cfg.ServerId, "sid", "proxy_id_1", "server id")
	flag.StringVar(&cfg.ServerName, "sname", "proxy_name", "server name")
	flag.StringVar(&cfg.MAddr, "maddr", "127.0.0.1:2221", "monitor addr")
	flag.StringVar(&cfg.CAddr, "caddr", "127.0.0.1:8500", "consul addr")
	flag.StringVar(&foundServer, "fdsvr", "serverNode_1", "found server name")
	flag.Parse()
	cfg.MInterval = "5s"
	cfg.MTimeOut = "2s"
	cfg.DeregisterTime = "20s"
	cfg.MMethod = "http"
	exist := make(chan os.Signal, 1)
	signal.Notify(exist, syscall.SIGTERM)

	err := lbbconsul.GConsulClient.Open(&cfg)
	if err != nil {
		fmt.Println("open return", err)
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
			proxys = append(proxys, conn)

			go EchoServer(conn)
		}
	}()

	go func() {
		tick := time.NewTicker(2 * time.Second)
		var oldSer = make(map[string]*lbbconsul.ServiceInfo)

		for range tick.C {
			err := lbbconsul.GConsulClient.DiscoverAliveService(foundServer)
			if err != nil {
				fmt.Println("discover server err", foundServer)
				continue
			}
			services, ok := lbbconsul.GConsulClient.GetAllService(foundServer)
			if !ok {
				fmt.Println("not find server err", foundServer)
				continue
			}
			for k, v := range services {
				if _, ok := oldSer[k]; !ok {
					go func(s *lbbconsul.ServiceInfo) {
						conn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", s.IP, s.Port))
						log.Println("lbb", s.ServiceID, s.IP, s.Port)
						if err != nil {
							log.Println(err)
							return
						}
						clients = append(clients, conn)
						sendData(conn)
					}(v)
				}
			}
			oldSer = services
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
			index := rand.Intn(len(clients))
			tmpCon := clients[index]
			tmpCon.Write(buf[:n])
			if err != nil {
				return
			}
			// n, err = tmpCon.Read(buf)
			// if err != nil {
			// return
			// }
			// n, err = conn.Write(buf[:n])
			// if err != nil {
			// return
			// }
		case io.EOF:
			log.Printf("Warning: End of data: %s\n", err)
			return
		default:
			log.Printf("Error: Reading data: %s\n", err)
			return
		}
	}
}
func sendData(conn net.Conn) {
	defer conn.Close()

	buf := make([]byte, RECV_BUF_LEN)
	for {
		n, err := conn.Read(buf)
		if err != nil {
			println("Read Buffer Error:", err.Error())
			break
		}
		log.Println("get:", string(buf[0:n]))

		if len(proxys) == 0 {
			continue
		}
		index := rand.Intn(len(proxys))

		n, err = proxys[index].Write(buf[:n])
		if err != nil {
			println("Write Buffer Error:", err.Error())
			break
		}

		//等一秒钟
		time.Sleep(time.Second)
	}
}
