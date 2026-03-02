package usecase_test

import (
	"context"
	"data-automation-service/configs"
	"data-automation-service/internal/domain"
	"data-automation-service/internal/usecase"
	"data-automation-service/pkg/logger"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"data-automation-service/internal/repository/mocks"

	"github.com/stretchr/testify/assert"
)

func TestReportUsecase_GenerateSalesTrendReport(t *testing.T) {
	mockRepo := mocks.NewReportRepository(t)
	log := logger.NewLogger()
	cfg := &configs.Config{}

	uc := usecase.NewReportUsecase(mockRepo, log, cfg)
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		mockData := []domain.SalesTrendReport{
			{ReportDate: "2024-11-29", TotalRevenue: 1000},
		}

		mockRepo.On("GetSalesTrendReport", ctx, 90).Return(mockData, nil).Once()

		tmpDir := t.TempDir()
		err := uc.GenerateSalesTrendReport(ctx, tmpDir)

		assert.NoError(t, err)
		mockRepo.AssertExpectations(t)

		// Check if file was created
		files, err := os.ReadDir(tmpDir)
		assert.NoError(t, err)
		assert.NotEmpty(t, files)
		assert.Contains(t, files[0].Name(), "sales_trend_")
	})

	t.Run("repo error", func(t *testing.T) {
		mockRepo.On("GetSalesTrendReport", ctx, 90).Return(nil, assert.AnError).Once()

		err := uc.GenerateSalesTrendReport(ctx, t.TempDir())

		assert.Error(t, err)
		mockRepo.AssertExpectations(t)
	})
}

func TestReportUsecase_GenerateProductPerformanceReport(t *testing.T) {
	mockRepo := mocks.NewReportRepository(t)
	log := logger.NewLogger()
	cfg := &configs.Config{}

	uc := usecase.NewReportUsecase(mockRepo, log, cfg)
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		mockData := []domain.ProductPerformanceReport{
			{ProductName: "Phone", TotalRevenue: 5000},
		}

		mockRepo.On("GetProductPerformanceReport", ctx).Return(mockData, nil).Once()

		tmpDir := t.TempDir()
		err := uc.GenerateProductPerformanceReport(ctx, tmpDir)

		assert.NoError(t, err)
		mockRepo.AssertExpectations(t)

		// Check if file was created
		files, err := os.ReadDir(tmpDir)
		assert.NoError(t, err)
		assert.NotEmpty(t, files)
		assert.Contains(t, files[0].Name(), "product_performance_")
	})
}

func TestReportUsecase_PushDailySalesReport(t *testing.T) {
	mockRepo := mocks.NewReportRepository(t)
	log := logger.NewLogger()

	// Setup mock HTTP server
	serverCallCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		serverCallCount++
		// Verify headers
		assert.Equal(t, "Bearer test-token", r.Header.Get("Authorization"))
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))

		// Verify body
		var payload domain.DailySalesPayload
		err := json.NewDecoder(r.Body).Decode(&payload)
		assert.NoError(t, err)
		assert.Equal(t, "daily_sales", payload.ReportType)
		assert.Equal(t, "2024-11-29", payload.Date)
		assert.Equal(t, 125000.5, payload.Data.TotalRevenue)

		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	cfg := &configs.Config{
		URL:   server.URL,
		Token: "test-token",
	}

	uc := usecase.NewReportUsecase(mockRepo, log, cfg)
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		serverCallCount = 0
		mockData := &domain.DailySalesPayloadData{
			TotalRevenue: 125000.5,
		}

		mockRepo.On("GetDailySalesSummary", ctx, "2024-11-29").Return(mockData, nil).Once()

		err := uc.PushDailySalesReport(ctx, "2024-11-29")

		assert.NoError(t, err)
		assert.Equal(t, 1, serverCallCount)
		mockRepo.AssertExpectations(t)
	})

	t.Run("repo error", func(t *testing.T) {
		mockRepo.On("GetDailySalesSummary", ctx, "2024-11-30").Return(nil, assert.AnError).Once()

		err := uc.PushDailySalesReport(ctx, "2024-11-30")

		assert.Error(t, err)
		mockRepo.AssertExpectations(t)
	})

	t.Run("retry logic - fail then success", func(t *testing.T) {
		failCount := 0
		retryServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			failCount++
			if failCount < 3 {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			w.WriteHeader(http.StatusOK)
		}))
		defer retryServer.Close()

		retryCfg := &configs.Config{
			URL:   retryServer.URL,
			Token: "test-token",
		}

		retryUc := usecase.NewReportUsecase(mockRepo, log, retryCfg)

		mockData := &domain.DailySalesPayloadData{
			TotalRevenue: 100,
		}

		// Use short timeout for testing retry loop speed
		// The sleep is hardcoded to `time.Sleep(time.Duration(attempt) * time.Second)` in code,
		// so this test will take ~3 seconds to run successfully.
		mockRepo.On("GetDailySalesSummary", ctx, "2024-12-01").Return(mockData, nil).Once()

		start := time.Now()
		err := retryUc.PushDailySalesReport(ctx, "2024-12-01")

		assert.NoError(t, err)
		assert.Equal(t, 3, failCount)
		assert.True(t, time.Since(start) > 2*time.Second)
		mockRepo.AssertExpectations(t)
	})
}
