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

var cfg lbbconsul.ConsulConfig
var foundServer string

func main() {
	flag.StringVar(&cfg.Ip, "ip", "127.0.0.1", "server ip")
	flag.IntVar(&cfg.Port, "port", 3333, "server port")
	flag.StringVar(&cfg.ServerId, "sid", "client_id_1", "server id")
	flag.StringVar(&cfg.ServerName, "sname", "client_id", "server name")
	flag.StringVar(&cfg.MAddr, "maddr", "127.0.0.1:3331", "monitor addr")
	flag.StringVar(&cfg.CAddr, "caddr", "127.0.0.1:8500", "consul addr")
	flag.StringVar(&foundServer, "fdsvr", "proxy_name", "found server name")
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

	go func() {
		for {

			time.Sleep(3 * time.Second)
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
				sendData(v)
				fmt.Println("use ", k)
				break
			}

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
