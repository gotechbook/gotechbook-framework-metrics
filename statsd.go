package metrics

import (
	"fmt"
	"github.com/DataDog/datadog-go/statsd"
	config "github.com/gotechbook/gotechbook-framework-config"
	logger "github.com/gotechbook/gotechbook-framework-logger"
)

type StatsdReporter struct {
	client      Client
	rate        float64
	serverType  string
	defaultTags []string
}

func NewStatsdReporter(metrics config.Metrics, serverType string, clientOrNil ...Client) (*StatsdReporter, error) {
	sr := &StatsdReporter{
		rate:       metrics.GoTechBookFrameworkMetricsStatsdRate,
		serverType: serverType,
	}
	sr.buildDefaultTags(metrics.GoTechBookFrameworkMetricsConstTags)

	if len(clientOrNil) > 0 {
		sr.client = clientOrNil[0]
	} else {
		c, err := statsd.New(metrics.GoTechBookFrameworkMetricsStatsdHost)
		if err != nil {
			return nil, err
		}
		c.Namespace = metrics.GoTechBookFrameworkMetricsStatsdPrefix
		sr.client = c
	}
	return sr, nil
}

func (s *StatsdReporter) buildDefaultTags(tagsMap map[string]string) {
	defaultTags := make([]string, len(tagsMap)+1)
	defaultTags[0] = fmt.Sprintf("serverType:%s", s.serverType)
	idx := 1
	for k, v := range tagsMap {
		defaultTags[idx] = fmt.Sprintf("%s:%s", k, v)
		idx++
	}
	s.defaultTags = defaultTags
}

func (s *StatsdReporter) ReportCount(metric string, tagsMap map[string]string, count float64) error {
	fullTags := s.defaultTags
	for k, v := range tagsMap {
		fullTags = append(fullTags, fmt.Sprintf("%s:%s", k, v))
	}
	err := s.client.Count(metric, int64(count), fullTags, s.rate)
	if err != nil {
		logger.Log.Errorf("failed to report count: %q", err)
	}
	return err
}

func (s *StatsdReporter) ReportGauge(metric string, tagsMap map[string]string, value float64) error {
	fullTags := s.defaultTags

	for k, v := range tagsMap {
		fullTags = append(fullTags, fmt.Sprintf("%s:%s", k, v))
	}

	err := s.client.Gauge(metric, value, fullTags, s.rate)
	if err != nil {
		logger.Log.Errorf("failed to report gauge: %q", err)
	}

	return err
}

func (s *StatsdReporter) ReportSummary(metric string, tagsMap map[string]string, value float64) error {
	fullTags := s.defaultTags
	for k, v := range tagsMap {
		fullTags = append(fullTags, fmt.Sprintf("%s:%s", k, v))
	}

	err := s.client.TimeInMilliseconds(metric, float64(value), fullTags, s.rate)
	if err != nil {
		logger.Log.Errorf("failed to report summary: %q", err)
	}

	return err
}

func (s *StatsdReporter) ReportHistogram(metric string, tagsMap map[string]string, value float64) error {
	return ErrNotImplemented
}
