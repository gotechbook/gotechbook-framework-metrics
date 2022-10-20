package metrics

import (
	"context"
	"errors"
	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	gContext "github.com/gotechbook/gotechbook-framework-context"
	e "github.com/gotechbook/gotechbook-framework-errors"
	"github.com/gotechbook/gotechbook-framework-metrics/mocks"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestReportTimingFromCtx(t *testing.T) {
	t.Run("test-duration", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		mockMetricsReporter := mocks.NewMockReporter(ctrl)

		originalTs := time.Now().UnixNano()
		expectedRoute := uuid.New().String()
		expectedType := uuid.New().String()
		expectedErr := errors.New(uuid.New().String())
		ctx := gContext.AddToPropagateCtx(context.Background(), StartTimeKey, originalTs)
		ctx = gContext.AddToPropagateCtx(ctx, RouteKey, expectedRoute)

		time.Sleep(200 * time.Millisecond) // to test duration report
		mockMetricsReporter.EXPECT().ReportSummary(ResponseTime, gomock.Any(), gomock.Any()).Do(
			func(metric string, tags map[string]string, duration float64) {
				assert.InDelta(t, duration, time.Now().UnixNano()-originalTs, 10e6)
			},
		)

		ReportTimingFromCtx(ctx, []Reporter{mockMetricsReporter}, expectedType, expectedErr)
	})

	t.Run("test-tags", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		mockMetricsReporter := mocks.NewMockReporter(ctrl)

		originalTs := time.Now().UnixNano()
		expectedRoute := uuid.New().String()
		expectedType := uuid.New().String()
		var expectedErr error
		ctx := gContext.AddToPropagateCtx(context.Background(), StartTimeKey, originalTs)
		ctx = gContext.AddToPropagateCtx(ctx, RouteKey, expectedRoute)
		ctx = gContext.AddToPropagateCtx(ctx, MetricTagsKey, map[string]string{
			"key": "value",
		})

		expectedTags := map[string]string{
			"route":  expectedRoute,
			"status": "ok",
			"type":   expectedType,
			"key":    "value",
			"code":   "",
		}

		mockMetricsReporter.EXPECT().ReportSummary(ResponseTime, expectedTags, gomock.Any())

		ReportTimingFromCtx(ctx, []Reporter{mockMetricsReporter}, expectedType, expectedErr)
	})

	t.Run("test-tags-not-correct-type", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		mockMetricsReporter := mocks.NewMockReporter(ctrl)

		originalTs := time.Now().UnixNano()
		expectedRoute := uuid.New().String()
		expectedType := uuid.New().String()
		var expectedErr error
		ctx := gContext.AddToPropagateCtx(context.Background(), StartTimeKey, originalTs)
		ctx = gContext.AddToPropagateCtx(ctx, RouteKey, expectedRoute)
		ctx = gContext.AddToPropagateCtx(ctx, MetricTagsKey, "not-map")

		expectedTags := map[string]string{
			"route":  expectedRoute,
			"status": "ok",
			"type":   expectedType,
			"code":   "",
		}

		mockMetricsReporter.EXPECT().ReportSummary(ResponseTime, expectedTags, gomock.Any())

		ReportTimingFromCtx(ctx, []Reporter{mockMetricsReporter}, expectedType, expectedErr)
	})

	t.Run("test-failed-route-with-error", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		mockMetricsReporter := mocks.NewMockReporter(ctrl)

		originalTs := time.Now().UnixNano()
		expectedRoute := uuid.New().String()
		expectedType := uuid.New().String()
		code := "GAME-404"
		expectedErr := e.New(errors.New("error"), code)
		ctx := gContext.AddToPropagateCtx(context.Background(), StartTimeKey, originalTs)
		ctx = gContext.AddToPropagateCtx(ctx, RouteKey, expectedRoute)

		mockMetricsReporter.EXPECT().ReportSummary(ResponseTime, map[string]string{
			"route":  expectedRoute,
			"status": "failed",
			"type":   expectedType,
			"code":   code,
		}, gomock.Any())

		ReportTimingFromCtx(ctx, []Reporter{mockMetricsReporter}, expectedType, expectedErr)
	})

	t.Run("test-failed-route", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		mockMetricsReporter := mocks.NewMockReporter(ctrl)

		originalTs := time.Now().UnixNano()
		expectedRoute := uuid.New().String()
		expectedType := uuid.New().String()
		expectedErr := errors.New("error")
		ctx := gContext.AddToPropagateCtx(context.Background(), StartTimeKey, originalTs)
		ctx = gContext.AddToPropagateCtx(ctx, RouteKey, expectedRoute)

		mockMetricsReporter.EXPECT().ReportSummary(ResponseTime, map[string]string{
			"route":  expectedRoute,
			"status": "failed",
			"type":   expectedType,
			"code":   e.ErrUnknownCode,
		}, gomock.Any())

		ReportTimingFromCtx(ctx, []Reporter{mockMetricsReporter}, expectedType, expectedErr)
	})
}

func TestReportMessageProcessDelayFromCtx(t *testing.T) {
	t.Run("test-tags", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		mockMetricsReporter := mocks.NewMockReporter(ctrl)

		originalTs := time.Now().UnixNano()
		expectedRoute := uuid.New().String()
		expectedType := uuid.New().String()
		ctx := gContext.AddToPropagateCtx(context.Background(), StartTimeKey, originalTs)
		ctx = gContext.AddToPropagateCtx(ctx, RouteKey, expectedRoute)
		ctx = gContext.AddToPropagateCtx(ctx, MetricTagsKey, map[string]string{
			"key": "value",
		})
		expectedTags := map[string]string{
			"route": expectedRoute,
			"type":  expectedType,
			"key":   "value",
		}
		mockMetricsReporter.EXPECT().ReportSummary(ProcessDelay, expectedTags, gomock.Any())
		ReportMessageProcessDelayFromCtx(ctx, []Reporter{mockMetricsReporter}, expectedType)
	})
}
