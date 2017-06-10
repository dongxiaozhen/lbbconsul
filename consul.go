package lbbconsul

import (
	"errors"
	"fmt"
	"net/http"
	"sync"

	"github.com/hashicorp/consul/api"
)

type (
	ConsulClient struct {
		*api.Client
		ServiceID string
		Clients   map[string]map[string]*ServiceInfo
		sync.RWMutex
	}
	ServiceInfo struct {
		ServiceID string
		IP        string
		Port      int
	}
)

var GConsulClient = &ConsulClient{}

type ConsulConfig struct {
	Ip         string   // server ip
	Port       int      // server port
	ServerId   string   // server id, must unique
	ServerName string   // server name,used to functional classification
	MAddr      string   // monitor addr
	MInterval  string   // monitor internal
	MTimeOut   string   // monitor timeout
	MMethod    string   // monitor method
	CAddr      string   // consul addr
	Tags       []string //server desc
}

func (c *ConsulClient) Open(cfg *ConsulConfig) (err error) {
	apiConf := api.DefaultConfig()
	apiConf.Address = cfg.CAddr

	c.Client, err = api.NewClient(apiConf)
	if err != nil {
		fmt.Println("ConsulClient open err", err)
		return err
	}

	c.Clients = make(map[string]map[string]*ServiceInfo)
	c.ServiceID = cfg.ServerId
	err = c.RegistService(cfg)
	return err
}

func (c *ConsulClient) Close() error {
	return c.Deregister()
}

// check handler ,could overwrites
func (c *ConsulClient) StatusHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("check status.")
}

func (c *ConsulClient) Deregister() error {
	return c.Agent().ServiceDeregister(c.ServiceID)
}

func (c *ConsulClient) RegistService(cfg *ConsulConfig) error {
	service := &api.AgentServiceRegistration{
		ID:      cfg.ServerId,
		Name:    cfg.ServerName,
		Port:    cfg.Port,
		Address: cfg.Ip,
		Tags:    cfg.Tags,
		Check: &api.AgentServiceCheck{
			Interval: cfg.MInterval,
			Timeout:  cfg.MTimeOut,
		},
	}
	switch cfg.MMethod {
	case "tcp":
		service.Check.TCP = cfg.MAddr
		/* go func() {
			listen, err := net.Listen("tcp", cfg.MAddr)
			if err != nil {
				panic(err)
				return
			}
			for {
				con, err := listen.Accept()
				if err != nil {
					fmt.Println(err)
					return
				}
				c.Debug(con)

			}
		}() */
	case "http":
		service.Check.HTTP = "http://" + cfg.MAddr + "/status"
		go func() {
			http.HandleFunc("/status", c.StatusHandler)
			fmt.Println("start listen...")
			if err := http.ListenAndServe(cfg.MAddr, nil); err != nil {
				fmt.Println(err)
				panic(err)
			}
		}()
	default:
		return errors.New("This method is not implemented")
	}

	err := c.Agent().ServiceRegister(service)
	return err
}

// 服务发现
// serviceID 要监测的服务名称
func (c *ConsulClient) DiscoverService(healthyOnly bool, foundService string) error {
	services, _, err := c.Client.Catalog().Services(&api.QueryOptions{})
	if err != nil {
		fmt.Println("DiscoverService Services err", err)
		return err
	}
	var sers = make(map[string]*ServiceInfo)
	for name := range services {
		servicesData, _, err := c.Client.Health().Service(name, "", healthyOnly, &api.QueryOptions{})
		if err != nil {
			fmt.Println("DiscoverService,Service err", err)
			return err
		}
		for _, entry := range servicesData {
			if foundService != entry.Service.Service {
				continue
			}
			for _, health := range entry.Checks {
				if foundService != health.ServiceName {
					continue
				}

				node := &ServiceInfo{}
				node.IP = entry.Service.Address
				node.Port = entry.Service.Port
				node.ServiceID = health.ServiceID
				if _, ok := sers[health.ServiceID]; ok {
					fmt.Println("repeat serviceID ", health.ServiceID)
					continue
				} else {
					sers[health.ServiceID] = node
				}
			}
		}
	}

	if sers == nil {
		fmt.Println("DiscoverService empty")
		c.Lock()
		delete(c.Clients, foundService)
		c.Unlock()
		return nil
	}
	c.Lock()
	c.Clients[foundService] = sers
	c.Unlock()
	return nil
}

func (c *ConsulClient) GetAllService(serviceName string) (map[string]*ServiceInfo, bool) {
	services, ok := c.Clients[serviceName]
	return services, ok
}

func (c *ConsulClient) PutKV(key string, value []byte) error {
	kv := &api.KVPair{Key: key, Value: value, Flags: 0}
	_, err := c.KV().Put(kv, nil)
	return err
}

func (c *ConsulClient) GetKV(key string) (value []byte, err error) {
	kv, _, err := c.KV().Get(key, nil)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	return kv.Value, nil
}
