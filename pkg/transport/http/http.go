package http

import (
	"net"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rs/zerolog"
	"source.rad.af/libs/go-lib/pkg/health"
	"source.rad.af/libs/go-lib/pkg/version"
)

type Server interface {
	Listen(net.Listener) error
	ApiRoutes() gin.IRouter
	ServeHTTP(http.ResponseWriter, *http.Request)
}

type server struct {
	engine    *gin.Engine
	apiRoutes *gin.RouterGroup
	options   *options
}

func (s *server) Listen(l net.Listener) error {
	return s.engine.RunListener(l)
}

func (s *server) ApiRoutes() gin.IRouter {
	return s.apiRoutes
}

func (s *server) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	s.engine.ServeHTTP(w, req)
}

func NewServer(config *Configuration, opts ...Option) Server {
	o := &options{
		logger: zerolog.Nop(),
	}
	for _, applyOpt := range opts {
		applyOpt(o)
	}
	if !config.ServerDebug {
		gin.SetMode(gin.ReleaseMode)
	}
	engine := gin.New()
	if o.healthChecker != nil {
		engine.GET("/health", healthHandler(o.healthChecker))
	}
	engine.GET("/ping", pingHandler)
	engine.GET("/version", versionHandler)
	engine.GET("/metrics", gin.WrapH(promhttp.Handler()))
	apiRoutes := engine.Group("/api", o.middleware...)
	return &server{
		engine:    engine,
		apiRoutes: apiRoutes,
		options:   o,
	}
}

func healthHandler(checker health.Checker) gin.HandlerFunc {
	return func(c *gin.Context) {
		body := ""
		code := http.StatusOK
		if !checker.Passed() {
			code = http.StatusServiceUnavailable
		}
		if _, ok := c.GetQuery("verbose"); ok {
			body = checker.String()
		}
		c.String(code, body)
	}
}

func versionHandler(c *gin.Context) {
	c.JSON(http.StatusOK, version.GetVersion())
}

func pingHandler(c *gin.Context) {
	c.String(http.StatusOK, "pong")
}
