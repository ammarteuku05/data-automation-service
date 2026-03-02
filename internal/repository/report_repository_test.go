package repository_test

import (
	"context"
	"data-automation-service/internal/repository"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
)

func prepareSqlmock(t *testing.T) (*sqlx.DB, sqlmock.Sqlmock, func()) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}

	sqlxDB := sqlx.NewDb(db, "postgres")
	return sqlxDB, mock, func() {
		db.Close()
	}
}

func TestReportRepository_GetSalesTrendReport(t *testing.T) {
	db, mock, cleanup := prepareSqlmock(t)
	defer cleanup()

	repo := repository.NewReportRepository(db)
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{
			"report_date", "total_orders", "total_revenue",
			"revenue_7d_avg", "orders_7d_avg", "pct_diff_from_avg",
			"day_of_week", "revenue_flag",
		}).
			AddRow("2024-11-29", 10, 1000.5, 950.0, 9.5, 5.3, "Friday", "Normal")

		mock.ExpectQuery(`(?i)WITH date_series AS \(.*`).
			WithArgs(90).
			WillReturnRows(rows)

		result, err := repo.GetSalesTrendReport(ctx, 90)

		assert.NoError(t, err)
		assert.Len(t, result, 1)
		assert.Equal(t, "2024-11-29", result[0].ReportDate)
		assert.Equal(t, 10, result[0].TotalOrders)
	})

	t.Run("db error", func(t *testing.T) {
		mock.ExpectQuery(`(?i)WITH date_series AS \(.*`).
			WithArgs(90).
			WillReturnError(assert.AnError)

		result, err := repo.GetSalesTrendReport(ctx, 90)

		assert.Error(t, err)
		assert.Nil(t, result)
	})
}

func TestReportRepository_GetProductPerformanceReport(t *testing.T) {
	db, mock, cleanup := prepareSqlmock(t)
	defer cleanup()

	repo := repository.NewReportRepository(db)
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{
			"product_name", "category", "total_revenue",
			"total_units_sold", "revenue_rank", "pct_category_revenue",
			"last_month_mom_change_pct", "is_top_20_percent",
		}).
			AddRow("Phone", "Electronics", 5000.0, 10, 1, 50.0, 10.5, true)

		mock.ExpectQuery(`(?i)WITH product_metrics AS \(.*`).
			WillReturnRows(rows)

		result, err := repo.GetProductPerformanceReport(ctx)

		assert.NoError(t, err)
		assert.Len(t, result, 1)
		assert.Equal(t, "Phone", result[0].ProductName)
		assert.Equal(t, "Electronics", result[0].Category)
	})
}

func TestReportRepository_GetDailySalesSummary(t *testing.T) {
	db, mock, cleanup := prepareSqlmock(t)
	defer cleanup()

	repo := repository.NewReportRepository(db)
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{
			"total_revenue", "total_orders", "average_order_value", "top_category",
		}).
			AddRow(125000.50, 450, 277.78, "Electronics")

		mock.ExpectQuery(`(?i)WITH daily_orders AS \(.*`).
			WithArgs("2024-11-29").
			WillReturnRows(rows)

		result, err := repo.GetDailySalesSummary(ctx, "2024-11-29")

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, 125000.50, result.TotalRevenue)
		assert.Equal(t, 450, result.TotalOrders)
		assert.Equal(t, "Electronics", result.TopCategory)
	})
}
