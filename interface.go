package metrics

type Reporter interface {
	ReportCount(metric string, tags map[string]string, count float64) error
	ReportSummary(metric string, tags map[string]string, value float64) error
	ReportHistogram(metric string, tags map[string]string, value float64) error
	ReportGauge(metric string, tags map[string]string, value float64) error
}

type Client interface {
	Count(name string, value int64, tags []string, rate float64) error
	Gauge(name string, value float64, tags []string, rate float64) error
	TimeInMilliseconds(name string, value float64, tags []string, rate float64) error
}
