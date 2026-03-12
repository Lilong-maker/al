package inits

import (
	"al/srv/dasic/config"
	"fmt"
	"log"
	"math/rand"
	"sync"
	"time"

	"github.com/hashicorp/consul/api"
)

var (
	consulClient    *api.Client
	serviceID       string
	serviceCache    map[string][]*api.AgentService
	cacheMutex      sync.RWMutex
	cacheExpiration time.Time
	cacheDuration   = 30 * time.Second
)

func ConsulInit() error {
	consulConfig := api.DefaultConfig()
	consulConfig.Address = fmt.Sprintf("%s:%d", config.Gen.ConSul.Host, config.Gen.ConSul.Port)
	var err error
	if consulClient, err = api.NewClient(consulConfig); err != nil {
		return fmt.Errorf("创建Consul客户端失败: %w", err)
	}

	serviceCache = make(map[string][]*api.AgentService)
	serviceID = fmt.Sprintf("%s-%d", config.Gen.ConSul.ServiceName, time.Now().Unix())
	checkID := fmt.Sprintf("%s-health", serviceID)
	registration := &api.AgentServiceRegistration{
		ID:      serviceID,
		Name:    config.Gen.ConSul.ServiceName,
		Address: "localhost",
		Port:    config.Gen.ConSul.ServicePort,
		Checks: []*api.AgentServiceCheck{
			{
				CheckID:                        checkID,
				Name:                           "TTL Health Check",
				TTL:                            fmt.Sprintf("%ds", config.Gen.ConSul.TTL),
				DeregisterCriticalServiceAfter: "1m",
			},
		},
	}
	if err := consulClient.Agent().ServiceRegister(registration); err != nil {
		return fmt.Errorf("注册服务失败: %w", err)
	}
	go func() {
		ticker := time.NewTicker(time.Duration(config.Gen.ConSul.TTL/2) * time.Second)
		defer ticker.Stop()
		for _ = range ticker.C {
			if err := consulClient.Agent().UpdateTTL(checkID, "服务正常", api.HealthPassing); err != nil {
				log.Printf("健康检查更新失败: %v", err)
			}
		}
	}()
	go func() {
		ticker := time.NewTicker(cacheDuration)
		defer ticker.Stop()

		for _ = range ticker.C {
			UpdateServiceCache()
		}
	}()
	return nil
}

func GetServiceWithLoadBalancer(serviceName string) (*api.AgentService, error) {
	healthyServices, err := GetHealthyService(serviceName)
	if err != nil {
		return nil, err
	}
	if len(healthyServices) == 0 {
		return nil, fmt.Errorf("没有可用的健康服务实例")
	}
	randomIndex := rand.Intn(len(healthyServices))
	return healthyServices[randomIndex], nil
}
func GetHealthyService(serviceName string) ([]*api.AgentService, error) {
	healthChecks, _, err := consulClient.Health().Service(serviceName, "", true, nil)
	if err != nil {
		return nil, fmt.Errorf("健康检查失败: %w", err)
	}
	var healthyServices []*api.AgentService
	for _, check := range healthChecks {
		if check.Checks.AggregatedStatus() == api.HealthPassing {
			healthyServices = append(healthyServices, check.Service)
		} else {
			log.Printf("服务 %s 状态异常: %s", check.Service.ID, check.Checks.AggregatedStatus())
		}
	}
	return healthyServices, nil
}

func UpdateServiceCache() {
	services, err := consulClient.Agent().Services()
	if err != nil {
		log.Printf("更新服务缓存失败: %v", err)
		return
	}
	cacheMutex.Lock()
	defer cacheMutex.Unlock()
	serviceMap := make(map[string][]*api.AgentService)
	for _, service := range services {
		serviceMap[service.Service] = append(serviceMap[service.Service], service)
	}
	serviceCache = serviceMap
	cacheExpiration = time.Now().Add(cacheDuration)
}

func ConsulShutdown() error {
	if consulClient == nil {
		return nil
	}
	if err := consulClient.Agent().ServiceDeregister(serviceID); err != nil {
		return fmt.Errorf("注销服务失败: %w", err)
	}
	log.Printf("服务 %s 已从Consul注销", serviceID)
	return nil
}
