package metrics

import (
	"fmt"
	"net/http"
	"sync"
	"sync/atomic"
	"time"
)

type Metrics struct {
	SearchRequests    atomic.Int64
	SearchErrors      atomic.Int64
	CacheHits         atomic.Int64
	CacheMisses       atomic.Int64
	NLPCalls          atomic.Int64
	NLPFailures       atomic.Int64
	AvgResponseTimeNs atomic.Int64
	responseSum       atomic.Int64
	responseCount    atomic.Int64
}

var Default = &Metrics{}

func (m *Metrics) RecordSearch(duration time.Duration, err bool) {
	m.SearchRequests.Add(1)
	if err {
		m.SearchErrors.Add(1)
	}
	m.responseSum.Add(duration.Nanoseconds())
	m.responseCount.Add(1)
	avg := m.responseSum.Load() / m.responseCount.Load()
	m.AvgResponseTimeNs.Store(avg)
}

func (m *Metrics) RecordCacheHit()   { m.CacheHits.Add(1) }
func (m *Metrics) RecordCacheMiss()  { m.CacheMisses.Add(1) }
func (m *Metrics) RecordNLPCall()    { m.NLPCalls.Add(1) }
func (m *Metrics) RecordNLPFailure() { m.NLPFailures.Add(1) }

func (m *Metrics) Handler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain; version=0.0.4")
		fprintf := func(format string, args ...interface{}) {
			fmt.Fprintf(w, format, args...)
		}

		fprintf("# HELP geo_search_requests_total Total search requests\n")
		fprintf("# TYPE geo_search_requests_total counter\n")
		fprintf("geo_search_requests_total %d\n", m.SearchRequests.Load())

		fprintf("# HELP geo_search_errors_total Total search errors\n")
		fprintf("# TYPE geo_search_errors_total counter\n")
		fprintf("geo_search_errors_total %d\n", m.SearchErrors.Load())

		fprintf("# HELP geo_search_cache_hits_total Cache hits\n")
		fprintf("# TYPE geo_search_cache_hits_total counter\n")
		fprintf("geo_search_cache_hits_total %d\n", m.CacheHits.Load())

		fprintf("# HELP geo_search_cache_misses_total Cache misses\n")
		fprintf("# TYPE geo_search_cache_misses_total counter\n")
		fprintf("geo_search_cache_misses_total %d\n", m.CacheMisses.Load())

		fprintf("# HELP geo_search_nlp_calls_total NLP service calls\n")
		fprintf("# TYPE geo_search_nlp_calls_total counter\n")
		fprintf("geo_search_nlp_calls_total %d\n", m.NLPCalls.Load())

		fprintf("# HELP geo_search_nlp_failures_total NLP service failures\n")
		fprintf("# TYPE geo_search_nlp_failures_total counter\n")
		fprintf("geo_search_nlp_failures_total %d\n", m.NLPFailures.Load())

		fprintf("# HELP geo_search_avg_response_time_ns Average response time in nanoseconds\n")
		fprintf("# TYPE geo_search_avg_response_time_ns gauge\n")
		fprintf("geo_search_avg_response_time_ns %d\n", m.AvgResponseTimeNs.Load())
	}
}

type responseWriter struct {
	http.ResponseWriter
	statusCode int
	once       sync.Once
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.once.Do(func() { rw.statusCode = code })
	rw.ResponseWriter.WriteHeader(code)
}
