package metrics

import (
	"fmt"
	config "github.com/gotechbook/gotechbook-framework-config"
	logger "github.com/gotechbook/gotechbook-framework-logger"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"net/http"
	"sync"
)

var (
	prometheusReporter *PrometheusReporter
	once               sync.Once
)

type PrometheusReporter struct {
	serverType            string
	game                  string
	countReportersMap     map[string]*prometheus.CounterVec
	summaryReportersMap   map[string]*prometheus.SummaryVec
	histogramReportersMap map[string]*prometheus.HistogramVec
	gaugeReportersMap     map[string]*prometheus.GaugeVec
	additionalLabels      map[string]string
}

func GetPrometheusReporter(serverType string, metrics config.Metrics, spec *config.CustomMetricsSpec) (*PrometheusReporter, error) {
	once.Do(func() {
		prometheusReporter = &PrometheusReporter{
			serverType:            serverType,
			game:                  "",
			countReportersMap:     make(map[string]*prometheus.CounterVec),
			summaryReportersMap:   make(map[string]*prometheus.SummaryVec),
			histogramReportersMap: make(map[string]*prometheus.HistogramVec),
			gaugeReportersMap:     make(map[string]*prometheus.GaugeVec),
		}
		prometheusReporter.registerMetrics(metrics.GoTechBookFrameworkMetricsConstTags, metrics.GoTechBookFrameworkMetricsPrometheusAdditionalTags, spec)
		http.Handle("/metrics", promhttp.Handler())
		go (func() {
			err := http.ListenAndServe(fmt.Sprintf(":%d", metrics.GoTechBookFrameworkMetricsPrometheusPort), nil)
			if err != nil {
				logger.Log.Error("prometheus reporter serve start failed, err: ", err)
			}
		})()
	})
	return prometheusReporter, nil
}
func (p *PrometheusReporter) registerCustomMetrics(constLabels map[string]string, additionalLabelsKeys []string, spec *config.CustomMetricsSpec) {
	for _, summary := range spec.Summaries {
		p.summaryReportersMap[summary.Name] = prometheus.NewSummaryVec(
			prometheus.SummaryOpts{
				Namespace:   config.PREFIX,
				Subsystem:   summary.Subsystem,
				Name:        summary.Name,
				Help:        summary.Help,
				Objectives:  summary.Objectives,
				ConstLabels: constLabels,
			},
			append(additionalLabelsKeys, summary.Labels...),
		)
	}

	for _, histogram := range spec.Histograms {
		p.histogramReportersMap[histogram.Name] = prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Namespace:   config.PREFIX,
				Subsystem:   histogram.Subsystem,
				Name:        histogram.Name,
				Help:        histogram.Help,
				Buckets:     histogram.Buckets,
				ConstLabels: constLabels,
			},
			append(additionalLabelsKeys, histogram.Labels...),
		)
	}
	for _, gauge := range spec.Gauges {
		p.gaugeReportersMap[gauge.Name] = prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace:   config.PREFIX,
				Subsystem:   gauge.Subsystem,
				Name:        gauge.Name,
				Help:        gauge.Help,
				ConstLabels: constLabels,
			},
			append(additionalLabelsKeys, gauge.Labels...),
		)
	}
	for _, counter := range spec.Counters {
		p.countReportersMap[counter.Name] = prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Namespace:   config.PREFIX,
				Subsystem:   counter.Subsystem,
				Name:        counter.Name,
				Help:        counter.Help,
				ConstLabels: constLabels,
			},
			append(additionalLabelsKeys, counter.Labels...),
		)
	}
}
func (p *PrometheusReporter) registerMetrics(constLabels, additionalLabels map[string]string, spec *config.CustomMetricsSpec) {
	constLabels["game"] = p.game
	constLabels["serverType"] = p.serverType

	p.additionalLabels = additionalLabels
	additionalLabelsKeys := make([]string, 0, len(additionalLabels))
	for key := range additionalLabels {
		additionalLabelsKeys = append(additionalLabelsKeys, key)
	}
	p.registerCustomMetrics(constLabels, additionalLabelsKeys, spec)

	p.summaryReportersMap[ResponseTime] = prometheus.NewSummaryVec(
		prometheus.SummaryOpts{
			Namespace:   config.PREFIX,
			Subsystem:   "handler",
			Name:        ResponseTime,
			Help:        "the time to process a msg in nanoseconds",
			Objectives:  map[float64]float64{0.7: 0.02, 0.95: 0.005, 0.99: 0.001},
			ConstLabels: constLabels,
		},
		append([]string{"route", "status", "type", "code"}, additionalLabelsKeys...),
	)

	p.histogramReportersMap[ResponseTime] = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace:   config.PREFIX,
			Subsystem:   "handler",
			Name:        ResponseTime,
			Help:        "the time to process a msg in nanoseconds",
			Buckets:     []float64{1, 5, 10, 50, 100, 300, 500, 1000, 5000, 10000},
			ConstLabels: constLabels,
		},
		append([]string{"route", "status", "type", "code"}, additionalLabelsKeys...),
	)
	p.summaryReportersMap[ProcessDelay] = prometheus.NewSummaryVec(
		prometheus.SummaryOpts{
			Namespace:   config.PREFIX,
			Subsystem:   "handler",
			Name:        ProcessDelay,
			Help:        "the delay to start processing a msg in nanoseconds",
			Objectives:  map[float64]float64{0.7: 0.02, 0.95: 0.005, 0.99: 0.001},
			ConstLabels: constLabels,
		},
		append([]string{"route", "type"}, additionalLabelsKeys...),
	)
	p.gaugeReportersMap[ConnectedClients] = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace:   config.PREFIX,
			Subsystem:   "acceptor",
			Name:        ConnectedClients,
			Help:        "the number of clients connected right now",
			ConstLabels: constLabels,
		},
		additionalLabelsKeys,
	)
	p.gaugeReportersMap[CountServers] = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace:   config.PREFIX,
			Subsystem:   "service_discovery",
			Name:        CountServers,
			Help:        "the number of discovered servers by service discovery",
			ConstLabels: constLabels,
		},
		append([]string{"type"}, additionalLabelsKeys...),
	)
	p.gaugeReportersMap[ChannelCapacity] = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace:   config.PREFIX,
			Subsystem:   "channel",
			Name:        ChannelCapacity,
			Help:        "the available capacity of the channel",
			ConstLabels: constLabels,
		},
		append([]string{"channel"}, additionalLabelsKeys...),
	)

	p.gaugeReportersMap[DroppedMessages] = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace:   config.PREFIX,
			Subsystem:   "rpc_server",
			Name:        DroppedMessages,
			Help:        "the number of rpc server dropped messages (messages that are not handled)",
			ConstLabels: constLabels,
		},
		additionalLabelsKeys,
	)

	p.gaugeReportersMap[Goroutines] = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace:   config.PREFIX,
			Subsystem:   "sys",
			Name:        Goroutines,
			Help:        "the current number of goroutines",
			ConstLabels: constLabels,
		},
		additionalLabelsKeys,
	)
	p.gaugeReportersMap[HeapSize] = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace:   config.PREFIX,
			Subsystem:   "sys",
			Name:        HeapSize,
			Help:        "the current heap size",
			ConstLabels: constLabels,
		},
		additionalLabelsKeys,
	)

	p.gaugeReportersMap[HeapObjects] = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace:   config.PREFIX,
			Subsystem:   "sys",
			Name:        HeapObjects,
			Help:        "the current number of allocated heap objects",
			ConstLabels: constLabels,
		},
		additionalLabelsKeys,
	)

	p.gaugeReportersMap[WorkerJobsRetry] = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace:   config.PREFIX,
			Subsystem:   "worker",
			Name:        WorkerJobsRetry,
			Help:        "the current number of job retries",
			ConstLabels: constLabels,
		},
		additionalLabelsKeys,
	)

	p.gaugeReportersMap[WorkerQueueSize] = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace:   config.PREFIX,
			Subsystem:   "worker",
			Name:        WorkerQueueSize,
			Help:        "the current queue size",
			ConstLabels: constLabels,
		},
		append([]string{"queue"}, additionalLabelsKeys...),
	)

	p.gaugeReportersMap[WorkerJobsTotal] = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace:   config.PREFIX,
			Subsystem:   "worker",
			Name:        WorkerJobsTotal,
			Help:        "the total executed jobs",
			ConstLabels: constLabels,
		},
		append([]string{"status"}, additionalLabelsKeys...),
	)

	p.countReportersMap[ExceededRateLimiting] = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace:   config.PREFIX,
			Subsystem:   "acceptor",
			Name:        ExceededRateLimiting,
			Help:        "the number of blocked requests by exceeded rate limiting",
			ConstLabels: constLabels,
		},
		additionalLabelsKeys,
	)

	toRegister := make([]prometheus.Collector, 0)
	for _, c := range p.countReportersMap {
		toRegister = append(toRegister, c)
	}

	for _, c := range p.gaugeReportersMap {
		toRegister = append(toRegister, c)
	}

	for _, c := range p.summaryReportersMap {
		toRegister = append(toRegister, c)
	}

	prometheus.MustRegister(toRegister...)
}
func (p *PrometheusReporter) ensureLabels(labels map[string]string) map[string]string {
	for key, defaultVal := range p.additionalLabels {
		if _, ok := labels[key]; !ok {
			labels[key] = defaultVal
		}
	}
	return labels
}
func (p *PrometheusReporter) ReportSummary(metric string, labels map[string]string, value float64) error {
	sum := p.summaryReportersMap[metric]
	if sum != nil {
		labels = p.ensureLabels(labels)
		sum.With(labels).Observe(value)
		return nil
	}
	return ErrMetricNotKnown
}
func (p *PrometheusReporter) ReportHistogram(metric string, labels map[string]string, value float64) error {
	hist := p.histogramReportersMap[metric]
	if hist != nil {
		labels = p.ensureLabels(labels)
		hist.With(labels).Observe(value)
		return nil
	}
	return ErrMetricNotKnown
}
func (p *PrometheusReporter) ReportCount(metric string, labels map[string]string, count float64) error {
	cnt := p.countReportersMap[metric]
	if cnt != nil {
		labels = p.ensureLabels(labels)
		cnt.With(labels).Add(count)
		return nil
	}
	return ErrMetricNotKnown
}
func (p *PrometheusReporter) ReportGauge(metric string, labels map[string]string, value float64) error {
	g := p.gaugeReportersMap[metric]
	if g != nil {
		labels = p.ensureLabels(labels)
		g.With(labels).Set(value)
		return nil
	}
	return ErrMetricNotKnown
}
