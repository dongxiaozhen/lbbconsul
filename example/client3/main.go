package main

import (
	"flag"
	"fmt"
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
	flag.IntVar(&cfg.Port, "port", 1111, "server port")
	flag.StringVar(&cfg.ServerId, "sid", "test_discover_1", "server id")
	flag.StringVar(&cfg.ServerName, "sname", "test_discover", "server name")
	flag.StringVar(&cfg.MAddr, "maddr", "127.0.0.1:1112", "monitor addr")
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

	go func() {
		for {

			time.Sleep(3 * time.Second)
			err = lbbconsul.GConsulClient.DiscoverServiceV2(foundServer)
			if err != nil {
				fmt.Println("discoveralive service2 server err", foundServer)
			}
			fmt.Println("-------------------------------------------------")
			err = lbbconsul.GConsulClient.DiscoverAliveService(foundServer)
			if err != nil {
				fmt.Println("discoveralive service server err", foundServer)
			}
			fmt.Println("-------------------------------------------------")

			err = lbbconsul.GConsulClient.DiscoverService(foundServer)
			if err != nil {
				fmt.Println("discover server err", foundServer)
			}
			fmt.Println("-------------------------------------------------")
		}
	}()
	<-exist
	lbbconsul.GConsulClient.Close()
}
