package metrics

import (
	"context"
	gContext "github.com/gotechbook/gotechbook-framework-context"
	errors "github.com/gotechbook/gotechbook-framework-errors"
	"runtime"
	"time"
)

func ReportTimingFromCtx(ctx context.Context, reporters []Reporter, typ string, err error) {
	if ctx == nil {
		return
	}
	code := errors.GetErrorCode(err)
	status := "ok"
	if err != nil {
		status = "failed"
	}
	if len(reporters) > 0 {
		startTime := gContext.GetFromPropagateCtx(ctx, StartTimeKey)
		route := gContext.GetFromPropagateCtx(ctx, RouteKey)
		elapsed := time.Since(time.Unix(0, startTime.(int64)))
		tags := getTags(ctx, map[string]string{
			"route":  route.(string),
			"status": status,
			"type":   typ,
			"code":   code,
		})
		for _, r := range reporters {
			r.ReportSummary(ResponseTime, tags, float64(elapsed.Nanoseconds()))
		}
	}
}

func ReportMessageProcessDelayFromCtx(ctx context.Context, reporters []Reporter, typ string) {
	if len(reporters) > 0 {
		startTime := gContext.GetFromPropagateCtx(ctx, StartTimeKey)
		elapsed := time.Since(time.Unix(0, startTime.(int64)))
		route := gContext.GetFromPropagateCtx(ctx, RouteKey)
		tags := getTags(ctx, map[string]string{
			"route": route.(string),
			"type":  typ,
		})
		for _, r := range reporters {
			r.ReportSummary(ProcessDelay, tags, float64(elapsed.Nanoseconds()))
		}
	}
}

func ReportNumberOfConnectedClients(reporters []Reporter, number int64) {
	for _, r := range reporters {
		r.ReportGauge(ConnectedClients, map[string]string{}, float64(number))
	}
}

func ReportSysMetrics(reporters []Reporter, period time.Duration) {
	for {
		for _, r := range reporters {
			num := runtime.NumGoroutine()
			m := &runtime.MemStats{}
			runtime.ReadMemStats(m)

			r.ReportGauge(Goroutines, map[string]string{}, float64(num))
			r.ReportGauge(HeapSize, map[string]string{}, float64(m.Alloc))
			r.ReportGauge(HeapObjects, map[string]string{}, float64(m.HeapObjects))
		}

		time.Sleep(period)
	}
}

func ReportExceededRateLimiting(reporters []Reporter) {
	for _, r := range reporters {
		r.ReportCount(ExceededRateLimiting, map[string]string{}, 1)
	}
}

func tagsFromContext(ctx context.Context) map[string]string {
	val := gContext.GetFromPropagateCtx(ctx, MetricTagsKey)
	if val == nil {
		return map[string]string{}
	}
	tags, ok := val.(map[string]string)
	if !ok {
		return map[string]string{}
	}
	return tags
}

func getTags(ctx context.Context, tags map[string]string) map[string]string {
	for k, v := range tagsFromContext(ctx) {
		tags[k] = v
	}
	return tags
}
