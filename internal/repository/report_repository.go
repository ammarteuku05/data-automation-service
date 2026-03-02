package repository

import (
	"context"
	"data-automation-service/internal/domain"

	"github.com/jmoiron/sqlx"
)

type ReportRepository interface {
	GetSalesTrendReport(ctx context.Context, days int) ([]domain.SalesTrendReport, error)
	GetProductPerformanceReport(ctx context.Context) ([]domain.ProductPerformanceReport, error)
	GetDailySalesSummary(ctx context.Context, dateStr string) (*domain.DailySalesPayloadData, error)
}

type reportRepository struct {
	db *sqlx.DB
}

func NewReportRepository(db *sqlx.DB) ReportRepository {
	return &reportRepository{db: db}
}

func (r *reportRepository) GetProductPerformanceReport(ctx context.Context) ([]domain.ProductPerformanceReport, error) {
	query := `
		WITH product_metrics AS (
			SELECT 
				p.id AS product_id,
				p.name AS product_name,
				p.category,
				SUM(oi.quantity * oi.unit_price) AS total_revenue,
				SUM(oi.quantity) AS total_units_sold
			FROM products p
			JOIN order_items oi ON p.id = oi.product_id
			JOIN orders o ON o.id = oi.order_id
			GROUP BY p.id, p.name, p.category
		),
		category_metrics AS (
			SELECT 
				pm.*,
				RANK() OVER (PARTITION BY pm.category ORDER BY pm.total_revenue DESC) AS revenue_rank,
				ROUND((pm.total_revenue / NULLIF(SUM(pm.total_revenue) OVER (PARTITION BY pm.category), 0)) * 100, 2) AS pct_category_revenue,
				PERCENT_RANK() OVER (PARTITION BY pm.category ORDER BY pm.total_revenue DESC) AS pct_rank_in_category
			FROM product_metrics pm
		),
		monthly_product_revenue AS (
			SELECT 
				oi.product_id,
				DATE_TRUNC('month', o.order_date) AS order_month,
				SUM(oi.quantity * oi.unit_price) AS monthly_revenue
			FROM order_items oi
			JOIN orders o ON oi.order_id = o.id
			WHERE o.status = 'completed'
			GROUP BY oi.product_id, DATE_TRUNC('month', o.order_date)
		),
		mom_revenue AS (
			SELECT 
				product_id,
				order_month,
				monthly_revenue,
				LAG(monthly_revenue) OVER (PARTITION BY product_id ORDER BY order_month) AS prev_month_revenue,
				ROUND(
					(monthly_revenue - LAG(monthly_revenue) OVER (PARTITION BY product_id ORDER BY order_month)) / 
					NULLIF(LAG(monthly_revenue) OVER (PARTITION BY product_id ORDER BY order_month), 0) * 100,
					2
				) AS mom_change_pct
			FROM monthly_product_revenue
		)
		SELECT 
			cm.product_name,
			cm.category,
			COALESCE(cm.total_revenue, 0) AS total_revenue,
			COALESCE(cm.total_units_sold, 0) AS total_units_sold,
			cm.revenue_rank,
			cm.pct_category_revenue,
			COALESCE(mr.mom_change_pct, 0) AS last_month_mom_change_pct,
			CASE WHEN pct_rank_in_category <= 0.20 THEN true ELSE false END AS is_top_20_percent
		FROM category_metrics cm
		LEFT JOIN (
			SELECT DISTINCT ON (product_id) *
			FROM mom_revenue
			ORDER BY product_id, order_month DESC
		) mr ON cm.product_id = mr.product_id
		ORDER BY cm.category, cm.revenue_rank;
	`
	var results []domain.ProductPerformanceReport
	err := r.db.SelectContext(ctx, &results, query)
	if err != nil {
		return nil, err
	}
	return results, nil
}

func (r *reportRepository) GetSalesTrendReport(ctx context.Context, days int) ([]domain.SalesTrendReport, error) {
	query := `
		WITH date_series AS (
			SELECT generate_series(
				DATE_TRUNC('day', CURRENT_TIMESTAMP - INTERVAL '89 days'),
				DATE_TRUNC('day', CURRENT_TIMESTAMP),
				INTERVAL '1 day'
			)::DATE AS report_date
		),
		daily_metrics AS (
			SELECT 
				ds.report_date,
				COUNT(o.id) AS total_orders,
				COALESCE(SUM(o.total_amount), 0) AS total_revenue
			FROM date_series ds
			LEFT JOIN orders o 
				ON DATE_TRUNC('day', o.order_date)::DATE = ds.report_date 
				AND o.status = 'completed'
			GROUP BY ds.report_date
		),
		moving_averages AS (
			SELECT 
				report_date,
				total_orders,
				total_revenue,
				COALESCE(AVG(total_revenue) OVER (ORDER BY report_date ROWS BETWEEN 6 PRECEDING AND CURRENT ROW), 0) AS revenue_7d_avg,
				COALESCE(AVG(total_orders) OVER (ORDER BY report_date ROWS BETWEEN 6 PRECEDING AND CURRENT ROW), 0) AS orders_7d_avg,
				TO_CHAR(report_date, 'Day') AS day_of_week
			FROM daily_metrics
		)
		SELECT 
			TO_CHAR(report_date, 'YYYY-MM-DD') AS report_date,
			total_orders,
			total_revenue,
			ROUND(revenue_7d_avg::numeric, 2) AS revenue_7d_avg,
			ROUND(orders_7d_avg::numeric, 2) AS orders_7d_avg,
			ROUND(
				CASE WHEN revenue_7d_avg = 0 THEN 0 
				ELSE ((total_revenue - revenue_7d_avg) / NULLIF(revenue_7d_avg, 0)) * 100 
				END::numeric, 2
			) AS pct_diff_from_avg,
			TRIM(day_of_week) AS day_of_week,
			CASE 
				WHEN revenue_7d_avg > 0 AND total_revenue > (revenue_7d_avg * 1.3) THEN 'High Anomaly'
				WHEN revenue_7d_avg > 0 AND total_revenue < (revenue_7d_avg * 0.7) THEN 'Low Anomaly'
				ELSE 'Normal'
			END AS revenue_flag
		FROM moving_averages
		ORDER BY report_date DESC
		LIMIT $1;
	`
	var results []domain.SalesTrendReport
	err := r.db.SelectContext(ctx, &results, query, days)
	if err != nil {
		return nil, err
	}
	return results, nil
}

func (r *reportRepository) GetDailySalesSummary(ctx context.Context, dateStr string) (*domain.DailySalesPayloadData, error) {
	query := `
		WITH daily_orders AS (
			SELECT o.id, o.total_amount
			FROM orders o
			WHERE o.status = 'completed'
			  AND DATE_TRUNC('day', o.order_date)::DATE = $1::DATE
		),
		daily_stats AS (
			SELECT 
				COALESCE(SUM(total_amount), 0) AS total_revenue,
				COUNT(id) AS total_orders,
				COALESCE(AVG(total_amount), 0) AS average_order_value
			FROM daily_orders
		),
		top_category AS (
			SELECT 
				p.category
			FROM orders o
			JOIN order_items oi ON o.id = oi.order_id
			JOIN products p ON oi.product_id = p.id
			WHERE o.status = 'completed'
			  AND DATE_TRUNC('day', o.order_date)::DATE = $1::DATE
			GROUP BY p.category
			ORDER BY SUM(oi.quantity * oi.unit_price) DESC
			LIMIT 1
		)
		SELECT 
			ds.total_revenue,
			ds.total_orders,
			ds.average_order_value,
			COALESCE((SELECT category FROM top_category), 'None') AS top_category
		FROM daily_stats ds;
	`
	var result domain.DailySalesPayloadData
	err := r.db.GetContext(ctx, &result, query, dateStr)
	if err != nil {
		return nil, err
	}
	return &result, nil
}
