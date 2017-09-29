package lbbconsul

import (
	"errors"
	"sort"

	"github.com/dongxiaozhen/lbbutil"
	"github.com/labstack/gommon/log"
)

var (
	ErrEmptyConsulServer = errors.New("consul empty server ")
	ErrConsulDiscover    = errors.New("consul discover error ")
)

func FoundRpcWithBalance(foundServer string, nowServer string) (*ServiceInfo, error) {
	err := GConsulClient.DiscoverAliveService(foundServer)
	if err != nil {
		log.Warn("discover server err", foundServer)
		return nil, ErrConsulDiscover
	}
	services, ok := GConsulClient.GetAllService(foundServer)
	if !ok {
		log.Warn("not find server err", foundServer)
		return nil, ErrEmptyConsulServer
	}
	var serverInfo *ServiceInfo
	if len(services) == 1 {
		for _, v := range services {
			serverInfo = v
			break
		}
		if serverInfo.ServiceID == nowServer {
			return nil, ErrEmptyConsulServer
		}
	} else {
		var balance ConsulServers
		balance = make([]*ServiceInfo, 0, len(services))
		for k, server := range services {
			v, err := GConsulClient.GetKV(k)
			if string(v) == nowServer {
				continue
			}
			if err == nil {
				if a := lbbutil.ToInt64(string(v)); a != -1 {
					server.Load = a
					balance = append(balance, server)
				} else {
					log.Error("conv server [ ", k, "] load error", err, string(v))
				}

			} else {
				log.Error("Get server [ ", k, "] load error", err)
			}
		}

		sort.Sort(balance)
		serverInfo = balance[0]
	}

	return serverInfo, nil
}

func FoundRpc(foundServer string, nowServer string) (*ServiceInfo, error) {
	err := GConsulClient.DiscoverAliveService(foundServer)
	if err != nil {
		log.Warn("discover server err", foundServer)
		return nil, ErrConsulDiscover
	}
	services, ok := GConsulClient.GetAllService(foundServer)
	if !ok {
		log.Warn("not find server err", foundServer)
		return nil, ErrEmptyConsulServer
	}
	if len(services) == 1 {
		nowServer = ""
	}

	var serverInfo *ServiceInfo
	if nowServer == "" {
		for _, v := range services {
			serverInfo = v
			break
		}
	} else {
		for k, v := range services {
			if k != nowServer {
				serverInfo = v
				break
			}
		}
	}
	return serverInfo, nil
}
