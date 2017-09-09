package lbbconsul

import (
	"errors"
	"fmt"
	"net/http"
	"sync"

	log "github.com/donnie4w/go-logger/logger"
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
		Tags      []string
	}
)

var GConsulClient = &ConsulClient{}

type ConsulConfig struct {
	Ip             string `json:"ip"`   // server ip
	Port           int    `json:"port"` // server port
	ServerId       string `json:"id"`   // server id, must unique
	ServerName     string `json:"-"`    // server name,used to functional classification
	MAddr          string `json:"-"`    // monitor addr
	MInterval      string `json:"-"`    // monitor internal
	MTimeOut       string `json:"-"`    // monitor timeout
	MMethod        string `json:"-"`    // monitor method
	CAddr          string `json:"-"`    // consul addr
	DeregisterTime string
	Tags           []string //server desc
}

func NewConsulClient() *ConsulClient {
	c := &ConsulClient{}
	apiConf.Address = cfg.CAddr
	c.Clients = make(map[string]map[string]*ServiceInfo)
	c.ServiceID = cfg.ServerId
	return c
}

func (c *ConsulClient) Open(cfg *ConsulConfig) (err error) {
	apiConf := api.DefaultConfig()
	c.Client, err = api.NewClient(apiConf)
	if err != nil {
		log.Warn("ConsulClient open err", err)
		return err
	}

	err = c.RegistService(cfg)
	return err
}

func (c *ConsulClient) Close() error {
	return c.Deregister()
}

// check handler ,could overwrites
func (c *ConsulClient) StatusHandler(w http.ResponseWriter, r *http.Request) {
	log.Info("check status.")
}

func (c *ConsulClient) Deregister() error {
	return c.Agent().ServiceDeregister(c.ServiceID)
}
func (c *ConsulClient) RegistService(cfg *ConsulConfig) error {
}

func (c *ConsulClient) registService(cfg *ConsulConfig) error {
	service := &api.AgentServiceRegistration{
		ID:      cfg.ServerId,
		Name:    cfg.ServerName,
		Port:    cfg.Port,
		Address: cfg.Ip,
		Tags:    cfg.Tags,
		Check: &api.AgentServiceCheck{
			Interval: cfg.MInterval,
			Timeout:  cfg.MTimeOut,
			DeregisterCriticalServiceAfter: cfg.DeregisterTime,
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
			log.Debug("start listen...")
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
func (c *ConsulClient) DiscoverService(foundService string) error {
	services, err := c.Agent().Services()
	if err != nil {
		log.Warn("DiscoverService get Services err", err)
		return err
	}

	var sers = make(map[string]*ServiceInfo)
	for id, ser := range services {
		if ser.Service == foundService {
			node := &ServiceInfo{}
			node.IP = ser.Address
			node.Port = ser.Port
			node.ServiceID = id
			node.Tags = ser.Tags
			if _, ok := sers[id]; ok {
				log.Warn("DiscoverService repeat serviceID ", id)
				continue
			} else {
				sers[id] = node
			}
			log.Debug("DiscoverService ", *node)
		}
	}

	if sers == nil {
		log.Warn("DiscoverService empty")
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

func (c *ConsulClient) DiscoverServiceV2(foundService string) error {
	var sers = make(map[string]*ServiceInfo)
	servicesData, _, err := c.Health().Service(foundService, "", false, &api.QueryOptions{})
	if err != nil {
		log.Warn("DiscoverAliveServiceV2 err", err)
		return err
	}
	for _, entry := range servicesData {
		if foundService != entry.Service.Service {
			continue
		}

		node := &ServiceInfo{}
		node.IP = entry.Service.Address
		node.Port = entry.Service.Port
		node.ServiceID = entry.Service.ID
		if _, ok := sers[node.ServiceID]; ok {
			log.Warn("repeat serviceID ", node.ServiceID)
			continue
		} else {
			log.Debug("DiscoverServiceV2 ", *node)
			sers[node.ServiceID] = node
		}
	}

	if sers == nil {
		log.Warn("DiscoverAliveServiceV2 empty")
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

func (c *ConsulClient) DiscoverAliveService(foundService string) error {
	var sers = make(map[string]*ServiceInfo)
	servicesData, _, err := c.Health().Service(foundService, "", true, &api.QueryOptions{})
	if err != nil {
		log.Warn("DiscoverAliveService err", err)
		return err
	}
	for _, entry := range servicesData {
		if foundService != entry.Service.Service {
			continue
		}

		node := &ServiceInfo{}
		node.IP = entry.Service.Address
		node.Port = entry.Service.Port
		node.ServiceID = entry.Service.ID
		if _, ok := sers[node.ServiceID]; ok {
			log.Warn("repeat serviceID ", node.ServiceID)
			continue
		} else {
			log.Debug("DiscoverAliveService", *node)
			sers[node.ServiceID] = node
		}
	}

	if sers == nil {
		log.Warn("DiscoverAliveService empty")
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

func (c *ConsulClient) CheckService(serviceName, serviceId string) (alive bool, err error) {
	healthCheck, _, err := c.Health().Checks(serviceName, nil)
	if err != nil {
		log.Warn("heal check err ", err)
		return
	}

	for i := range healthCheck {
		if healthCheck[i].ServiceID == serviceId && healthCheck[i].ServiceName == serviceName {
			if healthCheck[i].Status == api.HealthPassing {
				alive = true
				return
			}
		}
	}
	return
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
		log.Info(err)
		return nil, err
	}
	return kv.Value, nil
}
