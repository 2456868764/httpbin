package registry

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/nacos-group/nacos-sdk-go/v2/clients"
	"github.com/nacos-group/nacos-sdk-go/v2/clients/naming_client"
	"github.com/nacos-group/nacos-sdk-go/v2/common/constant"
	"github.com/nacos-group/nacos-sdk-go/v2/vo"
	"httpbin/pkg/logs"
	"httpbin/pkg/options"
)

const (
	DefaultInitTimeout          = time.Second * 10
	DefaultNacosTimeout         = 5000
	DefaultNacosLogLevel        = "warn"
	DefaultNacosLogDir          = "/var/log/nacos/log/"
	DefaultNacosCacheDir        = "/var/log/nacos/cache/"
	DefaultNacosNotLoadCache    = true
	DefaultNacosLogRotateTime   = "24h"
	DefaultNacosLogMaxAge       = 3
	DefaultUpdateCacheWhenEmpty = true
	DefaultRefreshInterval      = time.Second * 30
	DefaultRefreshIntervalLimit = time.Second * 10
	DefaultFetchPageSize        = 50
	DefaultJoiner               = "@@"
)

type NacosServiceRegistry struct {
	namingClient      naming_client.INamingClient
	nacosClietConfig  *constant.ClientConfig
	nacosNamespaceId  string
	nacosGroupName    string
	nacosServerDomain string
	nacosServerPort   int
	nacosUsername     string
	nacosPassword     string
}

func (c *NacosServiceRegistry) RegisterService(ctx context.Context, service *Service) error {
	err := c.doRegistry(service)
	if err == nil {
		return nil
	}
	logs.Errorf("registry service %+v, error:%v", service, err)
	stop := make(chan struct{})
	go func() {
		ticker := time.NewTicker(time.Duration(DefaultNacosTimeout))
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

func (c *NacosServiceRegistry) doRegistry(service *Service) error {
	logs.Infof("start registry service:%+v", service)
	port, _ := strconv.Atoi(service.Port)
	registration := vo.RegisterInstanceParam{
		Ip:          service.Ip,
		Port:        uint64(port),
		ServiceName: service.ServiceName,
		GroupName:   c.nacosGroupName,
		Weight:      10,
		Enable:      true,
		Healthy:     true,
		Ephemeral:   true,
		Metadata:    service.ServiceMeta,
	}
	success, err := c.namingClient.RegisterInstance(registration)
	if !success || err != nil {
		return errors.New(fmt.Sprintf("registry error:%v", err))
	}
	return nil
}

func NewNacosServiceRegistry(option *options.Option) (ServiceRegistry, error) {

	nacosClietConfig := constant.NewClientConfig(
		constant.WithTimeoutMs(DefaultNacosTimeout),
		constant.WithLogLevel(DefaultNacosLogLevel),
		constant.WithLogDir(DefaultNacosLogDir),
		constant.WithCacheDir(DefaultNacosCacheDir),
		constant.WithNotLoadCacheAtStart(DefaultNacosNotLoadCache),
		constant.WithLogRollingConfig(&constant.ClientLogRollingConfig{
			MaxAge: DefaultNacosLogMaxAge,
		}),
		constant.WithUpdateCacheWhenEmpty(false),
		constant.WithNamespaceId(option.NacosNamespaceId),
		constant.WithUsername(option.NacosUsername),
		constant.WithPassword(option.NacosPassword),
	)

	addresses := strings.Split(option.NacosServerAddress, ":")
	nacosServerDomain := addresses[0]
	nacosServerPort := 8848
	if len(addresses) == 2 {
		nacosServerPort, _ = strconv.Atoi(addresses[1])
	}
	sc := []constant.ServerConfig{
		*constant.NewServerConfig(nacosServerDomain, uint64(nacosServerPort)),
	}

	namingClient, err := clients.NewNamingClient(vo.NacosClientParam{
		ClientConfig:  nacosClietConfig,
		ServerConfigs: sc,
	})

	if err != nil {
		return nil, err
	}

	registry := &NacosServiceRegistry{
		nacosClietConfig:  nacosClietConfig,
		namingClient:      namingClient,
		nacosNamespaceId:  option.NacosNamespaceId,
		nacosGroupName:    option.NacosGroupName,
		nacosServerDomain: nacosServerDomain,
		nacosServerPort:   nacosServerPort,
		nacosUsername:     option.NacosUsername,
		nacosPassword:     option.NacosPassword,
	}
	return registry, nil
}
