package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/dongxiaozhen/lbbconsul"
)

var foundServer string

func main() {
	flag.StringVar(&foundServer, "fdsvr", "serverNode_1", "found server name")
	flag.Parse()

	exist := make(chan os.Signal, 1)
	signal.Notify(exist, syscall.SIGTERM)

	err := lbbconsul.GConsulClient.Open(&lbbconsul.Ccfg)
	if err != nil {
		fmt.Println("open return", err)
		return
	}

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
						sendData(s)
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

func sendData(service *lbbconsul.ServiceInfo) {
	conn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", service.IP, service.Port))
	log.Println("lbb", service.ServiceID, service.IP, service.Port)

	if err != nil {
		log.Println(err)
		return
	}

	defer conn.Close()

	buf := make([]byte, RECV_BUF_LEN)
	i := 0
	for {
		i++
		msg := fmt.Sprintf("Hello World, %03d", i)
		n, err := conn.Write([]byte(msg))
		if err != nil {
			println("Write Buffer Error:", err.Error())
			break
		}

		n, err = conn.Read(buf)
		if err != nil {
			println("Read Buffer Error:", err.Error())
			break
		}
		log.Println("get:", string(buf[0:n]))

		//等一秒钟
		time.Sleep(time.Second)
	}
}
