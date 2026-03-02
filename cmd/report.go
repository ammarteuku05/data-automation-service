package cmd

import (
	"context"
	"data-automation-service/configs"
	"data-automation-service/internal/repository"
	"data-automation-service/internal/usecase"
	"data-automation-service/pkg/database"
	"data-automation-service/pkg/logger"

	"github.com/spf13/cobra"
)

var (
	reportType string
	outputDir  string
)

var reportCmd = &cobra.Command{
	Use:   "report",
	Short: "Generate a specific data report to JSON",
	Run: func(cmd *cobra.Command, args []string) {
		log := logger.NewLogger()

		cfg := configs.LoadConfig()

		db, err := database.NewPostgresDB(cfg.DSN(), cfg.MaxOpenConns, cfg.MaxIdleConns)
		if err != nil {
			log.Error("Failed to connect to DB", "error", err)
			return
		}
		defer db.Close()

		repo := repository.NewReportRepository(db)
		uc := usecase.NewReportUsecase(repo, log, cfg)

		ctx := context.Background()

		switch reportType {
		case "sales-trend":
			if err := uc.GenerateSalesTrendReport(ctx, outputDir); err != nil {
				log.Error("Failed generating sales-trend report", "error", err)
			}
		case "product-performance":
			if err := uc.GenerateProductPerformanceReport(ctx, outputDir); err != nil {
				log.Error("Failed generating product-performance report", "error", err)
			}
		default:
			log.Error("Unknown report type. Supported: sales-trend, product-performance", "type", reportType)
		}
	},
}

func init() {
	rootCmd.AddCommand(reportCmd)
	reportCmd.Flags().StringVarP(&reportType, "type", "t", "", "Report type (sales-trend, product-performance) (required)")
	reportCmd.Flags().StringVarP(&outputDir, "output", "o", "./reports", "Output directory for JSON")
	reportCmd.MarkFlagRequired("type")
}
