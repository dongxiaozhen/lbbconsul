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
	flag.Parse()
	exist := make(chan os.Signal, 1)
	signal.Notify(exist, syscall.SIGTERM)

	err := lbbconsul.GConsulClient.Open(&lbbconsul.Ccfg)
	if err != nil {
		fmt.Println("GConsulClient,open err:", err)
		return
	}

	t := time.NewTicker(1 * time.Second)
	for range t.C {
		alive, err := lbbconsul.GConsulClient.CheckService("user_info", "user_info_1")
		fmt.Println("------------>lbb check ", alive, err)
	}

	<-exist
	lbbconsul.GConsulClient.Close()
	fmt.Println("vim-go")
}
