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
