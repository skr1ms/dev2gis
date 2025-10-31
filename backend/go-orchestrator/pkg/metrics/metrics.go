package metrics

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	httpRequestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total number of HTTP requests",
		},
		[]string{"method", "path", "status"},
	)

	httpRequestDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "HTTP request duration in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "path"},
	)

	processingJobsTotal = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "uav_processing_jobs_total",
			Help: "Total number of UAV processing jobs",
		},
	)

	processingJobsCompletedTotal = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "uav_processing_jobs_completed_total",
			Help: "Total number of completed UAV processing jobs",
		},
	)

	processingJobsFailedTotal = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "uav_processing_jobs_failed_total",
			Help: "Total number of failed UAV processing jobs",
		},
	)

	processingJobDuration = prometheus.NewHistogram(
		prometheus.HistogramOpts{
			Name:    "uav_processing_job_duration_seconds",
			Help:    "Duration of UAV processing jobs in seconds",
			Buckets: []float64{1, 5, 10, 30, 60, 120, 300},
		},
	)
)

func init() {
	prometheus.MustRegister(httpRequestsTotal)
	prometheus.MustRegister(httpRequestDuration)
	prometheus.MustRegister(processingJobsTotal)
	prometheus.MustRegister(processingJobsCompletedTotal)
	prometheus.MustRegister(processingJobsFailedTotal)
	prometheus.MustRegister(processingJobDuration)
}

func PrometheusHandler() gin.HandlerFunc {
	h := promhttp.Handler()
	return func(c *gin.Context) {
		h.ServeHTTP(c.Writer, c.Request)
	}
}

func MetricsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		c.Next()

		duration := time.Since(start).Seconds()
		status := strconv.Itoa(c.Writer.Status())
		path := c.FullPath()
		if path == "" {
			path = c.Request.URL.Path
		}

		httpRequestsTotal.WithLabelValues(c.Request.Method, path, status).Inc()
		httpRequestDuration.WithLabelValues(c.Request.Method, path).Observe(duration)
	}
}

func RecordProcessingJob() {
	processingJobsTotal.Inc()
}

func RecordProcessingJobCompleted() {
	processingJobsCompletedTotal.Inc()
}

func RecordProcessingJobFailed() {
	processingJobsFailedTotal.Inc()
}

func RecordProcessingJobDuration(duration time.Duration) {
	processingJobDuration.Observe(duration.Seconds())
}
