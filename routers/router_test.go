package routers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	promMetrics "github.com/slok/go-http-metrics/metrics/prometheus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/zapier/prom-aggregation-gateway/metrics"
)

func setupTestRouter(cfg ApiRouterConfig) *gin.Engine {
	agg := metrics.NewAggregates(time.Microsecond * 100)
	promConfig := promMetrics.Config{
		Registry: prometheus.NewRegistry(),
	}
	return setupAPIRouter(cfg, agg, promConfig)
}

func TestHealthCheck(t *testing.T) {
	router := gin.New()
	router.GET("/", handleHealthCheck)

	req, err := http.NewRequest("GET", "/", nil)
	require.NoError(t, err)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	responseHeaders := w.Header()
	assert.Equal(t, "application/json", responseHeaders.Get("content-type"))

	responseData, _ := io.ReadAll(w.Body)
	var response HealthResponse
	err = json.Unmarshal(responseData, &response)
	require.NoError(t, err)
	assert.Equal(t, true, response.IsAlive)
}

func TestMultiLabelPosting(t *testing.T) {
	tests := []struct {
		name         string
		path, metric string
		expected     string
	}{
		{
			"multiple labels",
			"/metrics/label1/value1/label2/value2",
			`# TYPE some_counter counter
some_counter 1
`,
			`# TYPE some_counter counter
some_counter{label1="value1",label2="value2"} 1
`},
		{
			"job label",
			"/metrics/job/someJob",
			`# TYPE some_counter counter
some_counter 1
`,
			`# TYPE some_counter counter
some_counter{job="someJob"} 1
`,
		},
		{
			"no labels, no trailing slash",
			"/metrics",
			"# TYPE some_counter counter\nsome_counter 1\n",
			"# TYPE some_counter counter\nsome_counter 1\n",
		},
		{
			"no labels, trailing slash",
			"/metrics/",
			"# TYPE some_counter counter\nsome_counter 1\n",
			"# TYPE some_counter counter\nsome_counter 1\n",
		},
		{
			"duplicate labels",
			"/metrics/testing/one/testing/two/testing/three",
			"# TYPE some_counter counter\n some_counter 1\n",
			"# TYPE some_counter counter\nsome_counter{testing=\"one\",testing=\"two\",testing=\"three\"} 1\n",
		},
	}

	for idx, test := range tests {
		t.Run(fmt.Sprintf("test #%d: %s", idx+1, test.name), func(t *testing.T) {
			// setup router
			router := setupTestRouter(ApiRouterConfig{CorsDomain: "https://cors-domain"})

			// ---- insert metric ----
			// setup request
			buf := bytes.NewBufferString(test.metric)
			req, err := http.NewRequest("PUT", test.path, buf)
			require.NoError(t, err)

			// make request
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, 202, w.Code)

			// ---- retrieve metric ----
			req, err = http.NewRequest("GET", "/metrics", nil)
			require.NoError(t, err)

			time.Sleep(time.Microsecond * 200)

			w = httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, 200, w.Code)
			body := w.Body.String()
			assert.Equal(t, test.expected, body)
		})
	}
}

func TestAuthRouter(t *testing.T) {
	tests := []struct {
		name                   string
		path, metric           string
		accounts               []string
		authName, authPassword string
		statusCode             int
		expected               string
	}{
		{
			"Passing 202 basic auth",
			"/metrics",
			"# TYPE some_counter counter\nsome_counter 1\n",
			[]string{"user=password"},
			"user", "password",
			202,
			"# TYPE some_counter counter\nsome_counter 1\n",
		},
		{
			"Failing 401 basic auth",
			"/metrics",
			"# TYPE some_counter counter\nsome_counter 1\n",
			[]string{"user=password"},
			"user1", "password1",
			401,
			"",
		},
	}

	for idx, test := range tests {
		t.Run(fmt.Sprintf("test #%d: %s", idx+1, test.name), func(t *testing.T) {
			// setup router
			router := setupTestRouter(ApiRouterConfig{CorsDomain: "https://cors-domain", Accounts: test.accounts})

			buf := bytes.NewBufferString(test.metric)
			req, err := http.NewRequest("PUT", test.path, buf)
			require.NoError(t, err)

			req.SetBasicAuth(test.authName, test.authPassword)

			// make request
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, test.statusCode, w.Code)

			// ---- retrieve metric ----
			req, err = http.NewRequest("GET", "/metrics", nil)
			require.NoError(t, err)

			time.Sleep(time.Microsecond * 300)

			w = httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, 200, w.Code)
			body := w.Body.String()
			assert.Equal(t, test.expected, body)

		})
	}
}

func TestCorsRouter(t *testing.T) {
	tests := []struct {
		name           string
		path           string
		method         string
		corsDomain     string
		origin         string
		statusCode     int
		expectedHeader string
	}{
		{
			"GET returning header on all allowed origins",
			"/metrics",
			"GET",
			"*",
			"https://cors-domain",
			200,
			"*",
		},
		{
			"PUT returning header on all allowed origins",
			"/metrics",
			"PUT",
			"*",
			"https://cors-domain",
			202,
			"*",
		},
		{
			"GET returning 403 and not returning header on origin not in cors config",
			"/metrics",
			"GET",
			"https://cors-domain",
			"https://invalid-domain",
			403,
			"",
		},
		{
			"PUT returning 403 and not returning header on origin not in cors config",
			"/metrics",
			"PUT",
			"https://cors-domain",
			"https://invalid-domain",
			403,
			"",
		},
	}

	for idx, test := range tests {
		t.Run(fmt.Sprintf("test #%d: %s", idx+1, test.name), func(t *testing.T) {
			// setup router
			router := setupTestRouter(ApiRouterConfig{CorsDomain: test.corsDomain})

			buf := bytes.NewBufferString("")
			req, err := http.NewRequest(test.method, test.path, buf)
			require.NoError(t, err)

			req.Header.Set("origin", test.origin)

			// make request
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, test.statusCode, w.Code)

			responseHeaders := w.Header()
			assert.Equal(t, test.expectedHeader, responseHeaders.Get("access-control-allow-origin"))
		})
	}
}
