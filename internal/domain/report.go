package domain

type SalesTrendReport struct {
	ReportDate     string  `json:"report_date" db:"report_date"`
	TotalOrders    int     `json:"total_orders" db:"total_orders"`
	TotalRevenue   float64 `json:"total_revenue" db:"total_revenue"`
	Revenue7dAvg   float64 `json:"revenue_7d_avg" db:"revenue_7d_avg"`
	Orders7dAvg    float64 `json:"orders_7d_avg" db:"orders_7d_avg"`
	PctDiffFromAvg float64 `json:"pct_diff_from_avg" db:"pct_diff_from_avg"`
	DayOfWeek      string  `json:"day_of_week" db:"day_of_week"`
	RevenueFlag    string  `json:"revenue_flag" db:"revenue_flag"`
}

type ProductPerformanceReport struct {
	ProductName        string  `json:"product_name" db:"product_name"`
	Category           string  `json:"category" db:"category"`
	TotalRevenue       float64 `json:"total_revenue" db:"total_revenue"`
	TotalUnitsSold     int     `json:"total_units_sold" db:"total_units_sold"`
	RevenueRank        int     `json:"revenue_rank" db:"revenue_rank"`
	PctCategoryRevenue float64 `json:"pct_category_revenue" db:"pct_category_revenue"`
	LastMonthMomChange float64 `json:"last_month_mom_change_pct" db:"last_month_mom_change_pct"`
	IsTop20Percent     bool    `json:"is_top_20_percent" db:"is_top_20_percent"`
}

type DailySalesPayload struct {
	ReportType  string                `json:"report_type"`
	Date        string                `json:"date"`
	Data        DailySalesPayloadData `json:"data"`
	GeneratedAt string                `json:"generated_at"`
}

type DailySalesPayloadData struct {
	TotalRevenue      float64 `json:"total_revenue" db:"total_revenue"`
	TotalOrders       int     `json:"total_orders" db:"total_orders"`
	AverageOrderValue float64 `json:"average_order_value" db:"average_order_value"`
	TopCategory       string  `json:"top_category" db:"top_category"`
}
