package registry

import (
	"fmt"
	"github.com/google/uuid"
	"httpbin/pkg/options"
	"strings"
)

type ServiceProtocol string

const (
	Http ServiceProtocol = "http"
	Grpc ServiceProtocol = "grpc"
)

type Service struct {
	ID           string
	ServiceName  string
	InstanceName string
	NodeName     string
	Ip           string
	Port         string
	Protocol     ServiceProtocol
	ServiceTags  []string
	ServiceMeta  map[string]string
	CheckPath    string
}

func NewServiceFromOption(option *options.Option) (*Service, error) {
	uuid, _ := uuid.NewUUID()
	service := &Service{
		ID:           uuid.String(),
		ServiceName:  option.ServiceName,
		InstanceName: option.InstanceName,
		NodeName:     option.NodeName,
		Port:         fmt.Sprintf("%d", option.ServerPort),
		Ip:           option.ServerIp,
		ServiceTags:  strings.Split(option.ServiceTags, ","),
		ServiceMeta:  option.ServiceNeta,
		CheckPath:    option.ServiceCheckPath,
		Protocol:     Http,
	}
	return service, nil
}
