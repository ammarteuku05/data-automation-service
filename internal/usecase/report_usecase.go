package usecase

import (
	"bytes"
	"context"
	"data-automation-service/configs"
	"data-automation-service/internal/domain"
	"data-automation-service/internal/repository"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

type ReportUsecase interface {
	GenerateSalesTrendReport(ctx context.Context, outputPath string) error
	GenerateProductPerformanceReport(ctx context.Context, outputPath string) error
	PushDailySalesReport(ctx context.Context, dateStr string) error
}

type reportUsecase struct {
	repo   repository.ReportRepository
	logger *slog.Logger
	config *configs.Config
}

func NewReportUsecase(repo repository.ReportRepository, logger *slog.Logger, cfg *configs.Config) ReportUsecase {
	return &reportUsecase{
		repo:   repo,
		logger: logger,
		config: cfg,
	}
}

func (u *reportUsecase) GenerateSalesTrendReport(ctx context.Context, outputPath string) error {
	u.logger.Info("Starting GenerateSalesTrendReport")
	start := time.Now()

	data, err := u.repo.GetSalesTrendReport(ctx, 90)
	if err != nil {
		u.logger.Error("Failed to fetch SalesTrendReport from database", "error", err)
		return fmt.Errorf("db error: %w", err)
	}

	err = u.exportJSON(outputPath, "sales_trend", data)
	if err != nil {
		return err
	}

	u.logger.Info("Successfully generated SalesTrendReport",
		"duration", time.Since(start),
		"rows", len(data),
	)
	return nil
}

func (u *reportUsecase) GenerateProductPerformanceReport(ctx context.Context, outputPath string) error {
	u.logger.Info("Starting GenerateProductPerformanceReport")
	start := time.Now()

	data, err := u.repo.GetProductPerformanceReport(ctx)
	if err != nil {
		u.logger.Error("Failed to fetch ProductPerformanceReport from database", "error", err)
		return fmt.Errorf("db error: %w", err)
	}

	err = u.exportJSON(outputPath, "product_performance", data)
	if err != nil {
		return err
	}

	u.logger.Info("Successfully generated ProductPerformanceReport",
		"duration", time.Since(start),
		"rows", len(data),
	)
	return nil
}

func (u *reportUsecase) PushDailySalesReport(ctx context.Context, dateStr string) error {
	u.logger.Info("Starting PushDailySalesReport", "date", dateStr)
	start := time.Now()

	data, err := u.repo.GetDailySalesSummary(ctx, dateStr)
	if err != nil {
		u.logger.Error("Failed to fetch daily sales summary", "error", err)
		return fmt.Errorf("db error: %w", err)
	}

	payload := domain.DailySalesPayload{
		ReportType:  "daily_sales",
		Date:        dateStr,
		Data:        *data,
		GeneratedAt: time.Now().UTC().Format(time.RFC3339),
	}

	fmt.Printf("payload: %+v\n", payload)
	jsonData, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("error marshaling payload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", u.config.URL, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("error creating request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	if u.config.Token != "" {
		req.Header.Set("Authorization", "Bearer "+u.config.Token)
	}

	client := &http.Client{Timeout: 10 * time.Second}

	// Basic retry logic
	var lastErr error
	for attempt := 1; attempt <= 3; attempt++ {
		resp, err := client.Do(req)
		if err == nil {
			defer resp.Body.Close()
			if resp.StatusCode >= 200 && resp.StatusCode < 300 {
				u.logger.Info("Successfully pushed daily sales report",
					"duration", time.Since(start),
					"status", resp.StatusCode,
				)
				return nil
			}
			lastErr = fmt.Errorf("api returned status %d", resp.StatusCode)
		} else {
			lastErr = err
		}

		u.logger.Warn("API push failed, retrying...", "attempt", attempt, "error", lastErr)
		time.Sleep(time.Duration(attempt) * time.Second)
	}

	return fmt.Errorf("failed to push daily sales report after 3 attempts: %w", lastErr)
}

func (u *reportUsecase) exportJSON(outputDir string, reportName string, data interface{}) error {
	if outputDir == "" {
		outputDir = "."
	}

	err := os.MkdirAll(outputDir, 0755)
	if err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	timestamp := time.Now().Format("20060102_150405")
	filename := filepath.Join(outputDir, fmt.Sprintf("%s_%s.json", reportName, timestamp))

	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(data); err != nil {
		return fmt.Errorf("failed to encode json: %w", err)
	}

	u.logger.Info("Report exported successfully", "file", filename)
	return nil
}
