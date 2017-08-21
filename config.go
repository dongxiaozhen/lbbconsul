package lbbconsul

import (
	"encoding/json"
	"flag"
)

var Ccfg ConsulConfig

func init() {
	flag.StringVar(&Ccfg.Ip, "ip", "127.0.0.1", "server ip")
	flag.IntVar(&Ccfg.Port, "port", 8991, "server port")
	flag.StringVar(&Ccfg.ServerId, "sid", "ppproxy_id_1", "server id")
	flag.StringVar(&Ccfg.ServerName, "sname", "ppproxy", "server name")
	flag.StringVar(&Ccfg.MAddr, "maddr", "127.0.0.1:8891", "monitor addr")
	flag.StringVar(&Ccfg.CAddr, "caddr", "127.0.0.1:8500", "consul addr")
	flag.StringVar(&Ccfg.MInterval, "minterval", "5s", "monitor interval")
	flag.StringVar(&Ccfg.MTimeOut, "mtimeout", "2s", "monitor timeout")
	flag.StringVar(&Ccfg.DeregisterTime, "deregist", "20s", "DeregisterTime")
	flag.StringVar(&Ccfg.MMethod, "mmethod", "http", "monitor method")
}

func GetConsulInfo() []byte {
	sj, err := json.Marshal(Ccfg)
	if err != nil {
		panic(err)
	}
	return sj
}
