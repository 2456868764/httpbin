package middleware

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/SkyAPM/go2sky"
	"github.com/SkyAPM/go2sky/reporter"
	"github.com/gin-gonic/gin"
	"github.com/openzipkin/zipkin-go"
	"github.com/openzipkin/zipkin-go/model"
	"github.com/openzipkin/zipkin-go/propagation/b3"
	reporterhttp "github.com/openzipkin/zipkin-go/reporter/http"
	"httpbin/pkg/logs"
	"httpbin/pkg/options"
	agentv3 "skywalking.apache.org/repo/goapi/collect/language/agent/v3"
)

const (
	componentIDGINHttpServer = 5006
	skipProbPrefix           = "/prob/"
	skipMetricsPrefix        = "/metrics"
)

var (
	ZipkinGlobalTracer *zipkin.Tracer
)

func StartSkywalkingTracer(g *gin.Engine, option *options.Option) {
	if len(option.SkywalkingGrpcAddress) == 0 {
		return
	}

	reporter, err := reporter.NewGRPCReporter(option.SkywalkingGrpcAddress)
	if err != nil {
		logs.Errorf("create gosky reporter failed! error:%v", err)
	}
	tracer, err := go2sky.NewTracer(option.ServiceName, go2sky.WithReporter(reporter),
		go2sky.WithInstance(option.InstanceName),
		go2sky.WithSampler(option.SamplingRate))
	g.Use(middlewareSkywalking(g, tracer))
	go2sky.SetGlobalTracer(tracer)
}

func StartZipkinTracer(g *gin.Engine, option *options.Option) {
	if len(option.ZipkinEndpointURL) == 0 {
		return
	}

	reporter := reporterhttp.NewReporter(option.ZipkinEndpointURL)
	localEndpoint := &model.Endpoint{ServiceName: option.ServiceName, Port: uint16(option.ServerPort)}

	sampler, err := zipkin.NewCountingSampler(option.SamplingRate)
	if err != nil {
		logs.Errorf("create zipkin sampler failed! error:%v", err)
	}

	tracer, err := zipkin.NewTracer(
		reporter,
		zipkin.WithLocalEndpoint(localEndpoint),
		zipkin.WithSampler(sampler),
	)

	if err != nil {
		logs.Errorf("Unable to create zipkin tracer: %v", err)
	}

	g.Use(middlewareZipkin(g, tracer))
	ZipkinGlobalTracer = tracer
}

// Middleware gin middleware return HandlerFunc  with tracing.
func middlewareSkywalking(engine *gin.Engine, tracer *go2sky.Tracer) gin.HandlerFunc {
	if engine == nil || tracer == nil {
		return func(c *gin.Context) {
			c.Next()
		}
	}
	return func(c *gin.Context) {
		if strings.HasPrefix(c.Request.URL.String(), skipProbPrefix) || strings.HasPrefix(c.Request.URL.String(), skipMetricsPrefix) {
			c.Next()
			return
		}
		span, ctx, err := tracer.CreateEntrySpan(c.Request.Context(), getOperationName(c), func(key string) (string, error) {
			return c.Request.Header.Get(key), nil
		})
		if err != nil {
			c.Next()
			return
		}
		span.SetComponent(componentIDGINHttpServer)
		span.Tag(go2sky.TagHTTPMethod, c.Request.Method)
		span.Tag(go2sky.TagURL, c.Request.Host+c.Request.URL.Path)
		span.SetSpanLayer(agentv3.SpanLayer_Http)

		c.Request = c.Request.WithContext(ctx)

		c.Next()

		if len(c.Errors) > 0 {
			span.Error(time.Now(), c.Errors.String())
		}
		span.Tag(go2sky.TagStatusCode, strconv.Itoa(c.Writer.Status()))
		span.End()
	}
}

func middlewareZipkin(engine *gin.Engine, tracer *zipkin.Tracer) gin.HandlerFunc {
	if engine == nil || tracer == nil {
		return func(c *gin.Context) {
			c.Next()
		}
	}

	return func(c *gin.Context) {
		spanContext := tracer.Extract(b3.ExtractHTTP(c.Request))
		span := tracer.StartSpan(c.Request.URL.Path, zipkin.Parent(spanContext))
		zipkin.TagHTTPMethod.Set(span, c.Request.Method)
		zipkin.TagHTTPUrl.Set(span, c.Request.Host+c.Request.URL.Path)

		newCtx := zipkin.NewContext(c.Request.Context(), span)
		c.Request = c.Request.WithContext(newCtx)

		c.Next()

		zipkin.TagHTTPStatusCode.Set(span, strconv.Itoa(c.Writer.Status()))
		span.Finish()
	}
}

func getOperationName(c *gin.Context) string {
	return fmt.Sprintf("/%s%s", c.Request.Method, c.FullPath())
}
