package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/prometheus/common/expfmt"
	metrics "github.com/slok/go-http-metrics/metrics/prometheus"
	"github.com/slok/go-http-metrics/middleware"
	mGin "github.com/slok/go-http-metrics/middleware/gin"
)

func setupAPIRouter(corsDomain string, agg *aggregate, promConfig metrics.Config) *gin.Engine {
	corsConfig := cors.Config{}
	if corsDomain != "*" {
		corsConfig.AllowOrigins = []string{corsDomain}
	} else {
		corsConfig.AllowAllOrigins = true
	}

	metricsMiddleware := middleware.New(middleware.Config{
		Recorder: metrics.NewRecorder(promConfig),
	})

	r := gin.New()

	r.GET("/metrics",
		mGin.Handler("metrics", metricsMiddleware),
		cors.New(corsConfig),
		handleRender(agg))
	r.POST("/metrics/job/:job",
		mGin.Handler("/metrics/job", metricsMiddleware),
		handleInsert(agg))

	return r
}

func handleInsert(a *aggregate) gin.HandlerFunc {
	return func(c *gin.Context) {
		job := c.Param("job")
		// TODO: add logic to verify correct format of job label
		if job == "" {
			err := fmt.Errorf("must send in a valid job name, sent: %s", job)
			log.Println(err)
			http.Error(c.Writer, err.Error(), http.StatusBadRequest)
			return
		}

		if err := a.parseAndMerge(c.Request.Body, job); err != nil {
			log.Println(err)
			http.Error(c.Writer, err.Error(), http.StatusBadRequest)
			return
		}

		MetricPushes.WithLabelValues(job).Inc()
	}
}

func handleRender(a *aggregate) gin.HandlerFunc {
	return func(c *gin.Context) {
		contentType := expfmt.Negotiate(c.Request.Header)
		c.Header("Content-Type", string(contentType))
		enc := expfmt.NewEncoder(c.Writer, contentType)

		a.render(enc)
	}
}
