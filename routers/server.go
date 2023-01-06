package routers

import (
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	promMetrics "github.com/slok/go-http-metrics/metrics/prometheus"
	"github.com/zapier/prom-aggregation-gateway/config"
	"github.com/zapier/prom-aggregation-gateway/metrics"
)

func RunServers(cfg config.Server) {
	sigChannel := make(chan os.Signal, 1)
	signal.Notify(sigChannel, syscall.SIGTERM, syscall.SIGINT)

	apiCfg := ApiRouterConfig{
		authAccounts: processAuthConfig(cfg.AuthUsers),
		CorsDomain:   cfg.CorsDomain,
	}

	agg := metrics.NewAggregates(time.Duration(cfg.MetricBatchInterval))

	promMetricsConfig := promMetrics.Config{
		Registry: metrics.PromRegistry,
	}

	apiRouter := setupAPIRouter(apiCfg, agg, promMetricsConfig)
	go runServer("api", apiRouter, cfg.ApiListen)

	lifecycleRouter := setupLifecycleRouter(metrics.PromRegistry)
	go runServer("lifecycle", lifecycleRouter, cfg.LifecycleListen)

	// Block until an interrupt or term signal is sent
	<-sigChannel
}

func runServer(label string, r *gin.Engine, listen string) {
	log.Printf("%s server listening at %s", label, listen)
	if err := r.Run(listen); err != nil {
		log.Panicf("error while serving %s: %v", label, err)
	}
}
