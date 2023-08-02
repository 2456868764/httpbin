package registry

import (
	"context"
	"fmt"
	"strconv"
	"time"

	consulapi "github.com/hashicorp/consul/api"
	"httpbin/pkg/logs"
	"httpbin/pkg/options"
)

const (
	DefaultConsulTimeout = time.Second * 5
)

type ConsulServiceRegistry struct {
	client *consulapi.Client
}

func (c *ConsulServiceRegistry) RegisterService(ctx context.Context, service *Service) error {
	err := c.doRegistry(service)
	if err == nil {
		return nil
	}
	logs.Errorf("registry service %+v, error:%v", service, err)
	stop := make(chan struct{})
	go func() {
		ticker := time.NewTicker(time.Duration(DefaultConsulTimeout))
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				err = c.doRegistry(service)
				if err == nil {
					close(stop)
					return
				} else {
					logs.Errorf("registry service %+v, error:%v", service, err)
				}
			case <-ctx.Done():
				logs.Infof("ctx.Done()")
				close(stop)
				return
			}
		}
	}()

	<-stop
	return err
}

func (c *ConsulServiceRegistry) doRegistry(service *Service) error {
	logs.Infof("start registry service:%+v", service)
	port, _ := strconv.Atoi(service.Port)
	registration := &consulapi.AgentServiceRegistration{
		ID:      service.InstanceName,
		Name:    service.ServiceName,
		Address: service.Ip,
		Port:    port,
		Tags:    service.ServiceTags,
		Meta:    service.ServiceMeta,
	}
	check := &consulapi.AgentServiceCheck{
		HTTP:                           fmt.Sprintf("http://%s:%s%s", service.Ip, service.Port, service.CheckPath),
		Timeout:                        "5s",
		Interval:                       "600s",
		DeregisterCriticalServiceAfter: "60s",
		//GRPC:                           fmt.Sprintf("%s:%s%s", service.Ip, service.Port, service.CheckPath),
	}
	registration.Check = check
	err := c.client.Agent().ServiceRegister(registration)
	return err
}

func NewConsulServiceRegistry(option *options.Option) (ServiceRegistry, error) {
	config := consulapi.DefaultConfig()
	config.Token = option.ConsulAuthToken
	config.Address = option.ConsulServerAddress
	client, err := consulapi.NewClient(config)
	if err != nil {
		return nil, err
	}
	registry := &ConsulServiceRegistry{
		client: client,
	}
	return registry, nil
}
