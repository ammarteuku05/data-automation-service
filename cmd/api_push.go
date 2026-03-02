package cmd

import (
	"context"
	"data-automation-service/configs"
	"data-automation-service/internal/repository"
	"data-automation-service/internal/usecase"
	"data-automation-service/pkg/database"
	"data-automation-service/pkg/logger"
	"time"

	"github.com/spf13/cobra"
)

var pushDate string

var apiPushCmd = &cobra.Command{
	Use:   "api-push",
	Short: "Push daily sales summary to the BI REST API",
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

		if pushDate == "" {
			// default to yesterday
			pushDate = time.Now().Add(-24 * time.Hour).Format("2006-01-02")
		}

		if err := uc.PushDailySalesReport(context.Background(), pushDate); err != nil {
			log.Error("Failed to push report", "error", err)
		}
	},
}

func init() {
	rootCmd.AddCommand(apiPushCmd)
	apiPushCmd.Flags().StringVarP(&pushDate, "date", "d", "", "Date for report (YYYY-MM-DD), defaults to yesterday")
}
