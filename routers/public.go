package routers

import (
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	promMetrics "github.com/slok/go-http-metrics/metrics/prometheus"
	"github.com/slok/go-http-metrics/middleware"
	mGin "github.com/slok/go-http-metrics/middleware/gin"
	"github.com/zapier/prom-aggregation-gateway/metrics"
)

type ApiRouterConfig struct {
	CorsDomain   string
	Accounts     []string
	authAccounts gin.Accounts
}

func setupAPIRouter(cfg ApiRouterConfig, agg *metrics.Aggregate, promConfig promMetrics.Config) *gin.Engine {
	corsConfig := cors.Config{}
	if cfg.CorsDomain != "*" {
		corsConfig.AllowOrigins = []string{cfg.CorsDomain}
	} else {
		corsConfig.AllowAllOrigins = true
	}
	corsHandler := cors.New(corsConfig)
	cfg.authAccounts = processAuthConfig(cfg.Accounts)

	metricsMiddleware := middleware.New(middleware.Config{
		Recorder: promMetrics.NewRecorder(promConfig),
	})

	r := gin.New()
	r.RedirectTrailingSlash = false

	// add metric middleware for NoRoute handler
	r.NoRoute(mGin.Handler("noRoute", metricsMiddleware))

	neededHandlers := []gin.HandlerFunc{corsHandler}
	if len(cfg.Accounts) > 0 {
		neededHandlers = append(neededHandlers, gin.BasicAuth(cfg.authAccounts))
	}

	r.GET("/metrics",
		mGin.Handler("getMetrics", metricsMiddleware),
		corsHandler,
		agg.HandleRender,
	)

	postHandlers := []gin.HandlerFunc{
		mGin.Handler("postMetrics", metricsMiddleware),
	}
	postHandlers = append(postHandlers, neededHandlers...)
	postHandlers = append(postHandlers, agg.HandleInsert)

	r.POST("/metrics", postHandlers...)
	r.POST("/metrics/*labels", postHandlers...)
	r.PUT("/metrics", postHandlers...)
	r.PUT("/metrics/*labels", postHandlers...)

	return r
}
