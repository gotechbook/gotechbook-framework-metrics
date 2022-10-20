package metrics

import "errors"

var (
	ErrMetricNotKnown = errors.New("the provided metric does not exist")
	ErrNotImplemented = errors.New("method not implemented")
)
