package api

import (
	"io"
	"math/rand"
	"net/http"
	"strings"
	"time"

	"github.com/SkyAPM/go2sky"
	"github.com/gin-gonic/gin"
	"github.com/openzipkin/zipkin-go"
	"github.com/openzipkin/zipkin-go/propagation/b3"
	"httpbin/pkg/logs"
	"httpbin/pkg/middleware"
	"httpbin/pkg/model"
	"httpbin/pkg/options"
	"httpbin/pkg/utils"
	v3 "skywalking.apache.org/repo/goapi/collect/language/agent/v3"
)

var defaultTraceHeaders = []string{
	// All applications should propagate x-request-id. This header is
	// included in access log statements and is used for consistent trace
	// sampling and log sampling decisions in Istio.
	"X-Request-Id",

	// Lightstep tracing header. Propagate this if you use lightstep tracing
	// in Istio (see
	// https://istio.io/latest/docs/tasks/observability/distributed-tracing/lightstep/)
	// Note: this should probably be changed to use B3 or W3C TRACE_CONTEXT.
	// Lightstep recommends using B3 or TRACE_CONTEXT and most application
	// libraries from lightstep do not support x-ot-span-context.
	"X-Ot-Span-Context",

	// Datadog tracing header. Propagate these headers if you use Datadog
	// tracing.
	"x-datadog-trace-id",
	"x-datadog-parent-id",
	"x-datadog-sampling-priority",

	// b3 trace headers. Compatible with Zipkin, OpenCensusAgent, and
	// Stackdriver Istio configurations. Commented out since they are
	// propagated by the OpenTracing tracer above.
	"X-B3-TraceId", "X-B3-SpanId", "X-B3-ParentSpanId", "X-B3-Sampled", "X-B3-Flags",

	// Jager
	"uber-trace-id",

	// Grpc binary trace context. Compatible with OpenCensusAgent nad
	// Stackdriver Istio configurations.
	"grpc-trace-bin",

	// W3C Trace Context. Compatible with OpenCensusAgent and Stackdriver Istio
	// configurations.
	"traceparent",
	"tracestate",

	// Cloud trace context. Compatible with OpenCensusAgent and Stackdriver Istio
	// configurations.
	"x-cloud-trace-context",

	// SkyWalking trace headers.
	"sw8",

	// Context and session specific headers
	"cookie", "jwt", "Authorization",

	// Application-specific headers to forward.
	"end-user",
	"user-agent",

	// httpbin headers
	"X-Httpbin-Trace-Host",
	"X-Httpbin-Trace-Service",
}

func Anything(c *gin.Context) {
	// Simulate business call
	r := rand.Intn(45) + 5
	time.Sleep(time.Duration(r) * time.Millisecond)
	// Return
	response := NewResponseFromContext(c)
	c.JSON(http.StatusOK, response)
}

func Ping(c *gin.Context) {
	c.JSON(http.StatusOK, "pong")
}

func HostName(c *gin.Context) {
	response := utils.GetHostName()
	c.JSON(http.StatusOK, response)
}

func Headers(c *gin.Context) {
	headers := c.Request.Header
	response := make(map[string]string, len(headers))
	for hk, hv := range headers {
		response[hk] = strings.Join(hv, ",")
	}
	c.JSON(http.StatusOK, response)
}

func Healthz(c *gin.Context) {
	c.JSON(http.StatusOK, "healthz")
	return
}

func HealthzFile(c *gin.Context) {
	if utils.FileExisted("./healthz.txt") {
		c.JSON(http.StatusOK, "ok")
		return
	}
	c.JSON(http.StatusNotFound, "not healthz")
}

func Readiness(c *gin.Context) {
	c.JSON(http.StatusOK, "readiness")
	return
}

func ReadinessFile(c *gin.Context) {
	if utils.FileExisted("./readiness.txt") {
		c.JSON(http.StatusOK, "ok")
		return
	}
	c.JSON(http.StatusNotFound, "not readiness")
}

func Startup(c *gin.Context) {
	c.JSON(http.StatusOK, "startup")
	return
}

func StartupFile(c *gin.Context) {
	if utils.FileExisted("./startup.txt") {
		c.JSON(http.StatusOK, "ok")
		return
	}
	c.JSON(http.StatusNotFound, "not startup")
}

func Bool(c *gin.Context) {
	c.JSON(http.StatusCreated, true)
}

func ReponseAnyDto(c *gin.Context) {
	c.JSON(http.StatusOK, model.ResponseAny{Code: 1, Data: model.ConditionRouteDto{}})
}

func ReponseAnyArray(c *gin.Context) {
	c.JSON(http.StatusOK, model.ResponseAny{Code: 1, Data: []model.ConditionRouteDto{{}}})
}

func ReponseAnyString(c *gin.Context) {
	c.JSON(http.StatusOK, model.ResponseAny{Code: 1, Data: "hello"})
}

func Service(c *gin.Context, option *options.Option) {
	nextServices := c.Query("services")
	if len(nextServices) == 0 {
		// Simulate business call
		r := rand.Intn(45) + 5
		time.Sleep(time.Duration(r) * time.Millisecond)
		// Return
		response := NewResponseFromContext(c)
		c.JSON(http.StatusOK, response)
		return
	}
	// Call next service
	// Pass headers
	headers := c.Request.Header
	services := strings.Split(nextServices, ",")
	nextUrl := ""
	if len(services) == 1 {
		nextUrl = "http://" + services[0] + "/"
	} else {
		nextUrl = "http://" + services[0] + "/service?services=" + strings.Join(services[1:], ",")
	}
	logs.Infof("service call nexturl:%s", nextUrl)
	req, err := http.NewRequest(c.Request.Method, nextUrl, c.Request.Body)
	if err != nil {
		logs.Error(err)
	}
	lowerCaseHeader := make(http.Header)
	for key, value := range headers {
		headK := strings.ToLower(key)
		for _, traceHeader := range defaultTraceHeaders {
			if headK == strings.ToLower(traceHeader) {
				lowerCaseHeader[strings.ToLower(key)] = value
			}
		}
	}

	// Add service trace header
	traceHeader, ok := lowerCaseHeader["x-httpbin-trace-host"]
	if !ok {
		lowerCaseHeader["x-httpbin-trace-host"] = []string{utils.GetHostName()}
	} else {
		lowerCaseHeader["x-httpbin-trace-host"] = []string{traceHeader[0] + "/" + utils.GetHostName()}
	}

	traceHeader2, ok2 := lowerCaseHeader["x-httpbin-trace-service"]
	if !ok2 {
		lowerCaseHeader["x-httpbin-trace-service"] = []string{utils.GetServiceName()}
	} else {
		lowerCaseHeader["x-httpbin-trace-service"] = []string{traceHeader2[0] + "/" + utils.GetServiceName()}
	}

	req.Header = lowerCaseHeader
	fn := func(req *http.Request) (*http.Response, error) {
		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			logs.Error(err)
		}
		return resp, err
	}

	var resp *http.Response
	if option.TraceProvider == options.Skywalking {
		resp, err = traceHttpCallSkywalking(c, req, nextUrl, fn)
	}
	if option.TraceProvider == options.Zipkin {
		resp, err = traceHttpCallZipkin(c, req, nextUrl, fn)
	}
	if err != nil {
		logs.Errorf("execute traceHttpCall failed: %v", err)
	}

	var bodyBytes []byte
	bodyBytes, _ = io.ReadAll(resp.Body)
	c.Header("Content-Type", "application/json")
	c.String(http.StatusOK, string(bodyBytes))
}

func traceHttpCallSkywalking(c *gin.Context, req *http.Request, url string, fn func(req *http.Request) (*http.Response, error)) (*http.Response, error) {
	tracer := go2sky.GetGlobalTracer()
	if tracer == nil {
		resp, err := fn(req)
		return resp, err
	}

	reqSpan, err := go2sky.GetGlobalTracer().CreateExitSpan(c.Request.Context(), "invoke", url, func(headerKey, headerValue string) error {
		req.Header.Set(headerKey, headerValue)
		return nil
	})
	if err != nil {
	}
	reqSpan.SetComponent(2)
	reqSpan.SetSpanLayer(v3.SpanLayer_Http) // rpc 调用
	resp, err2 := fn(req)
	reqSpan.Tag(go2sky.TagHTTPMethod, http.MethodGet)
	reqSpan.Tag(go2sky.TagURL, url)
	reqSpan.End()
	return resp, err2
}

func traceHttpCallZipkin(c *gin.Context, req *http.Request, url string, fn func(req *http.Request) (*http.Response, error)) (*http.Response, error) {
	tracer := middleware.ZipkinGlobalTracer
	if tracer == nil {
		resp, err := fn(req)
		return resp, err
	}

	reqSpan, _ := tracer.StartSpanFromContext(c.Request.Context(), "invoke")
	err := b3.InjectHTTP(req)(reqSpan.Context())
	if err != nil {
		return nil, err
	}
	resp, err2 := fn(req)
	zipkin.TagHTTPMethod.Set(reqSpan, c.Request.Method)
	zipkin.TagHTTPUrl.Set(reqSpan, url)
	reqSpan.Finish()
	return resp, err2
}
