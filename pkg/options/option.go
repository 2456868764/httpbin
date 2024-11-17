package options

import (
	"fmt"
	"github.com/spf13/pflag"
	"httpbin/pkg/utils"
)

const (
	DefaultSubSystem = "App"
	DefaultNameSpace = "default"
	DefaultVersion   = "v1"
)

const (
	ServiceRegistryTypeConsul = "consul"
	ServiceRegistryTypeNacos  = "nacos"
	ServiceRegistryTypeNone   = "none"
)

const (
	Skywalking = "skywalking"
	Zipkin     = "zipkin"
)

type Option struct {
	TraceProvider         string
	SkywalkingGrpcAddress string
	ZipkinEndpointURL     string
	ServerPort            uint32
	ServerAddress         string
	ServerIp              string
	SamplingRate          float64
	ServiceTags           string
	ServiceMeta           map[string]string
	ServiceCheckPath      string

	ServiceName  string
	InstanceName string
	SubSystem    string
	NameSpace    string
	Version      string
	NodeName     string

	RegistryType string

	ConsulServerAddress string
	ConsulDatacenter    string
	ConsulAuthToken     string

	NacosServerAddress string
	NacosNamespaceId   string
	NacosGroupName     string
	NacosUsername      string
	NacosPassword      string

	GrpcEnable bool
	GrpcPort   uint32

	HttpsEnable   bool
	HttpsPort     uint32
	CACertFile    string
	TlsCertFile   string
	TlsKeyFile    string
	TlsServerName string
	MTLS          bool
}

func (o *Option) AddFlags(flags *pflag.FlagSet) {
	flags.StringVar(&o.TraceProvider, "trace-provider", "", "Trace provider type")
	flags.StringVar(&o.SkywalkingGrpcAddress, "skywalking-grpc-address", "", "Skywalking grpc address.")
	flags.StringVar(&o.ZipkinEndpointURL, "zipkin-endpoint-url", "", "Zipkin http endpoint url.")
	flags.Uint32Var(&o.ServerPort, "server-port", 80, "The server port binds to.")
	flags.Float64Var(&o.SamplingRate, "sample-rate", 1.0, "Trace sample rate")
	flags.StringVar(&o.ServiceCheckPath, "service-check-path", "/ping", "service check path.")
	flags.StringVar(&o.RegistryType, "registry-type", "none", "Registry type")
	flags.StringVar(&o.ServiceTags, "service-tags", "", "service tags.")
	flags.StringVar(&o.ConsulServerAddress, "consul-server-address", "", "Consul server address.")
	flags.StringVar(&o.ConsulDatacenter, "consul-data-center", "dc1", "Consul data center.")
	flags.StringVar(&o.ConsulAuthToken, "consul-auth-token", "", "Consul server auth token")

	flags.StringVar(&o.NacosServerAddress, "nacos-server-address", "", "nacos server address.")
	flags.StringVar(&o.NacosNamespaceId, "nacos-namespace-id", "", "nacos namespace id")
	flags.StringVar(&o.NacosGroupName, "nacos-group-name", "", "nacos group name")
	flags.StringVar(&o.NacosUsername, "nacos-username", "", "nacos username")
	flags.StringVar(&o.NacosPassword, "nacos-password", "", "nacos password")

	flags.BoolVar(&o.GrpcEnable, "grpc-enable", true, "grpc enable")
	flags.Uint32Var(&o.GrpcPort, "grpc-port", 9091, "grpc demo order port")

	flags.BoolVar(&o.HttpsEnable, "https-enable", false, "https enable")
	flags.Uint32Var(&o.HttpsPort, "https-port", 443, "https port")
	flags.StringVar(&o.CACertFile, "cacert", "", "CA cert file")
	flags.StringVar(&o.TlsCertFile, "cert", "", "tls cert file")
	flags.StringVar(&o.TlsKeyFile, "key", "", "tls key file")
	flags.StringVar(&o.TlsServerName, "server-name", "", "tls server name")
	flags.BoolVar(&o.MTLS, "mtls", false, "mtls enable")

}

func (o *Option) Complete() {
	o.ServerAddress = fmt.Sprintf(":%d", o.ServerPort)
}

func (o *Option) FillEnvs() {
	o.InstanceName = utils.GetHostName()
	o.ServiceName = utils.GetServiceName()
	o.SubSystem = utils.GetSubSystem()
	if len(o.SubSystem) == 0 {
		o.SubSystem = DefaultSubSystem
	}
	o.NameSpace = utils.GetNameSpace()
	if len(o.NameSpace) == 0 {
		o.NameSpace = DefaultNameSpace
	}
	o.Version = utils.GetVersion()
	if len(o.Version) == 0 {
		o.Version = DefaultVersion
	}
	o.ServerIp = utils.GetIP()
	o.NodeName = utils.GetNodeName()
}

func NewOption() *Option {
	return &Option{}
}
