package registry

import (
	"context"
	"errors"
	"httpbin/pkg/logs"
	"httpbin/pkg/options"
	"time"
)

type ServiceRegistry interface {
	RegisterService(ctx context.Context, service *Service) error
}

func ServiceRegistryFactory(option *options.Option) (ServiceRegistry, error) {
	switch option.RegistryType {
	case options.ServiceRegistryTypeConsul:
		serviceRegistry, err := NewConsulServiceRegistry(option)
		return serviceRegistry, err
	default:
		return nil, errors.New("not support registry type")
	}
}

func StartRegistry(ctx context.Context, option *options.Option) {
	if option.RegistryType == options.ServiceRegistryTypeNone {
		return
	}
	if serviceRegistry, err := ServiceRegistryFactory(option); err != nil {
		logs.Fatalf("service registry get error:%v", err)
	} else {
		service, _ := NewServiceFromOption(option)
		newCtx, cancel := context.WithTimeout(ctx, time.Second*30)
		defer cancel()
		err = serviceRegistry.RegisterService(newCtx, service)
		if err != nil {
			logs.Fatalf("service registry error:%v", err)
		}
		return
	}
	return
}
