/*
Part 1.1: Schema Analysis

1. Performance Issues (1M+ customers, 10M+ orders, 50M+ order items):
- Missing Foreign Key Indexes: PostgreSQL does not automatically index foreign keys. Joins between customers, orders, and order_items will cause sequential scans, making queries EXTREMELY slow as tables grow to 10M-50M rows.
- Unindexed Filtering/Sorting: Queries filtering or sorting by order_date (e.g., daily sales, cohorts, recent orders) will require full table scans of 10M+ rows.
- Expensive Aggregations: Real-time sum/count aggregations on 50M+ rows for dashboards will block CPU and memory.

2. Index Recommendations:
- CREATE INDEX idx_orders_customer_id ON orders(customer_id); 
  Reason: Crucial for fetching a customer's order history and joining customers with orders.
- CREATE INDEX idx_orders_order_date ON orders(order_date);
  Reason: Most analytical and operational queries filter, group, or sort by date (e.g., sales in the last 90 days).
- CREATE INDEX idx_order_items_order_id ON order_items(order_id);
  Reason: Required for looking up the items of an order, fundamental for joining orders and order_items.
- CREATE INDEX idx_order_items_product_id ON order_items(product_id);
  Reason: Necessary for product-centric aggregations (which products were sold in which orders).
- CREATE INDEX idx_orders_status ON orders(status);
  Reason: Allows quick filtering of 'completed', 'cancelled', or 'pending' orders without reading the entire table.

3. Schema Improvements:
- Partitioning: Partition `orders` and `order_items` tables by `order_date` (e.g., monthly) to improve query speed on recent data and simplify data archiving.
- Denormalization / materialized views: Add a summary table or materialized view for daily/monthly product sales and customer metrics to avoid recalculating heavy metrics (like RFM) directly from 50M rows continuously.
- Audit / Metadata columns: Add `updated_at` timestamps to `orders` and allow tracking of specific status transitions (e.g., when an order became completed).
*/

-- ==============================================================
-- Query 1: Customer Cohort Analysis with Running Totals
-- Approach: Use CTEs to find each customer's first order month. 
-- Group by cohort month to calculate new customers, initial revenue, 
-- and use SUM() OVER() for running totals. 
-- For retention, left join to orders placed in cohort_month + 1.
-- ==============================================================
WITH cohort_data AS (
    SELECT 
        customer_id,
        MIN(DATE_TRUNC('month', order_date)) AS cohort_month
    FROM orders
    WHERE EXTRACT(YEAR FROM order_date) = 2024
    GROUP BY customer_id
),
first_month_revenue AS (
    SELECT 
        cd.cohort_month,
        cd.customer_id,
        SUM(o.total_amount) AS revenue
    FROM cohort_data cd
    JOIN orders o ON cd.customer_id = o.customer_id 
        AND DATE_TRUNC('month', o.order_date) = cd.cohort_month
    GROUP BY cd.cohort_month, cd.customer_id
),
retention_data AS (
    SELECT DISTINCT
        cd.customer_id,
        cd.cohort_month
    FROM cohort_data cd
    JOIN orders o ON cd.customer_id = o.customer_id 
        AND DATE_TRUNC('month', o.order_date) = cd.cohort_month + INTERVAL '1 month'
)
SELECT 
    cd.cohort_month,
    COUNT(DISTINCT cd.customer_id) AS new_customers,
    SUM(fmr.revenue) AS first_month_revenue,
    SUM(COUNT(DISTINCT cd.customer_id)) OVER(ORDER BY cd.cohort_month) AS running_total_customers,
    ROUND(
        COUNT(DISTINCT rd.customer_id)::NUMERIC / NULLIF(COUNT(DISTINCT cd.customer_id), 0) * 100, 
        2
    ) AS retention_rate_percentage
FROM cohort_data cd
JOIN first_month_revenue fmr ON cd.customer_id = fmr.customer_id
LEFT JOIN retention_data rd ON cd.customer_id = rd.customer_id
GROUP BY cd.cohort_month
ORDER BY cd.cohort_month;

-- ==============================================================
-- Query 2: Product Performance with Ranking and Comparison
-- Approach: First calculate each product's lifetime metrics.
-- Then use window functions for ranking and percentage of category.
-- Calculate MoM revenue for the last completed month using LAG().
-- Using PERCENT_RANK() to identify top 20% products.
-- ==============================================================
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
    cm.total_revenue,
    cm.total_units_sold,
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

-- ==============================================================
-- Query 3: Customer Segmentation with RFM Analysis
-- Approach: Calculate base R, F, M metrics per customer.
-- Then use NTILE(5) for multi-dimensional grading (score 1-5).
-- We manually bin recency and frequency as requested, 
-- but NTILE handles monetary grading smoothly.
-- ==============================================================
WITH rfm_base AS (
    SELECT 
        c.id AS customer_id,
        c.name,
        c.email,
        EXTRACT(DAY FROM CURRENT_TIMESTAMP - MAX(o.order_date)) AS recency_days,
        COUNT(o.id) AS frequency_count,
        SUM(o.total_amount) AS monetary_value
    FROM customers c
    JOIN orders o ON c.id = o.customer_id
    WHERE o.status = 'completed'
    GROUP BY c.id, c.name, c.email
),
rfm_scoring AS (
    SELECT 
        *,
        CASE 
            WHEN recency_days <= 30 THEN 5
            WHEN recency_days > 365 THEN 1
            ELSE 3 -- Intermediate default per instructions lacking strict middle bounds
        END AS r_score,
        CASE 
            WHEN frequency_count >= 20 THEN 5
            WHEN frequency_count <= 2 THEN 1
            ELSE 3
        END AS f_score,
        NTILE(5) OVER(ORDER BY monetary_value ASC) AS m_score -- 5 is top 20%
    FROM rfm_base
),
rfm_final AS (
    SELECT 
        *,
        (r_score + f_score + m_score) AS rfm_score
    FROM rfm_scoring
)
SELECT 
    name,
    email,
    recency_days,
    frequency_count,
    monetary_value,
    rfm_score,
    CASE 
        WHEN rfm_score >= 12 THEN 'Champions'
        WHEN rfm_score >= 9 THEN 'Loyal'
        WHEN rfm_score >= 6 THEN 'At Risk'
        ELSE 'Lost'
    END AS segment
FROM rfm_final
ORDER BY rfm_score DESC;

-- ==============================================================
-- Query 4: Advanced Sales Trend Analysis
-- Approach: Generate a date series for the last 90 days to ensure 
-- days with zero orders are included.
-- Use window functions to calculate 7-day moving averages 
-- with preceding and current rows.
-- ==============================================================
WITH date_series AS (
    -- Get last 90 days up to today
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
        AVG(total_revenue) OVER (ORDER BY report_date ROWS BETWEEN 6 PRECEDING AND CURRENT ROW) AS revenue_7d_avg,
        AVG(total_orders) OVER (ORDER BY report_date ROWS BETWEEN 6 PRECEDING AND CURRENT ROW) AS orders_7d_avg,
        TO_CHAR(report_date, 'Day') AS day_of_week
    FROM daily_metrics
)
SELECT 
    report_date,
    total_orders,
    total_revenue,
    ROUND(revenue_7d_avg, 2) AS revenue_7d_avg,
    ROUND(orders_7d_avg, 2) AS orders_7d_avg,
    ROUND(
        CASE WHEN revenue_7d_avg = 0 THEN 0 
        ELSE ((total_revenue - revenue_7d_avg) / NULLIF(revenue_7d_avg, 0)) * 100 
        END, 2
    ) AS pct_diff_from_avg,
    day_of_week,
    CASE 
        WHEN revenue_7d_avg > 0 AND total_revenue > (revenue_7d_avg * 1.3) THEN 'High Anomaly'
        WHEN revenue_7d_avg > 0 AND total_revenue < (revenue_7d_avg * 0.7) THEN 'Low Anomaly'
        ELSE 'Normal'
    END AS revenue_flag
FROM moving_averages
ORDER BY report_date DESC;

-- ==============================================================
-- Query 5: Inventory Turnover and Stock Analysis
-- Approach: Calculate recent sales by joining product to order_items.
-- Calculate daily rate (sales/90) and estimated stockout days.
-- Use CASE expressions for complex bucketing rules.
-- ==============================================================
WITH recent_sales AS (
    SELECT 
        p.id AS product_id,
        p.name AS product_name,
        p.stock_quantity,
        COALESCE(SUM(oi.quantity), 0) AS sold_last_90d,
        MAX(o.order_date) AS last_order_date
    FROM products p
    LEFT JOIN order_items oi ON p.id = oi.product_id
    LEFT JOIN orders o ON oi.order_id = o.id AND o.order_date >= CURRENT_TIMESTAMP - INTERVAL '90 days'
    GROUP BY p.id, p.name, p.stock_quantity
    HAVING p.stock_quantity > 0 OR SUM(oi.quantity) IS NOT NULL
),
stock_metrics AS (
    SELECT 
        *,
        sold_last_90d / 90.0 AS daily_sales_rate,
        CASE WHEN (sold_last_90d / 90.0) > 0 
             THEN stock_quantity / NULLIF((sold_last_90d / 90.0), 0)
             ELSE NULL 
        END AS days_until_stockout
    FROM recent_sales
)
SELECT 
    product_name,
    stock_quantity,
    sold_last_90d,
    ROUND(daily_sales_rate, 2) AS daily_sales_rate,
    ROUND(days_until_stockout, 2) AS days_until_stockout,
    last_order_date,
    CASE 
        WHEN days_until_stockout < 7 THEN 'Critical'
        WHEN days_until_stockout < 30 THEN 'Low'
        WHEN days_until_stockout <= 90 THEN 'Adequate'
        WHEN days_until_stockout > 90 THEN 'Overstocked'
        WHEN sold_last_90d = 0 AND stock_quantity > 0 THEN 'Dead Stock'
        ELSE 'Unknown'
    END AS stock_status,
    GREATEST(0, CEIL((daily_sales_rate * 45) - stock_quantity)) AS reorder_recommendation,
    CASE 
        WHEN days_until_stockout < 7 THEN 1
        WHEN sold_last_90d = 0 AND stock_quantity > 0 THEN 2
        WHEN days_until_stockout < 30 THEN 3
        WHEN days_until_stockout <= 90 THEN 4
        ELSE 5
    END AS priority_score
FROM stock_metrics
ORDER BY priority_score ASC, days_until_stockout ASC NULLS LAST;

-- ==============================================================
-- Query 6: Customer Purchase Pattern Analysis
-- Approach: Using window functions and subqueries to calculate intervals
-- between orders. Calculate increasing/decreasing trends comparing
-- the first 3 vs last 3 orders using filtered averages.
-- ==============================================================
WITH customer_orders AS (
    SELECT 
        c.id AS customer_id,
        c.name,
        c.email,
        o.id AS order_id,
        o.order_date,
        o.total_amount,
        LAG(o.order_date) OVER (PARTITION BY c.id ORDER BY o.order_date) AS prev_order_date,
        ROW_NUMBER() OVER (PARTITION BY c.id ORDER BY o.order_date ASC) AS order_asc,
        ROW_NUMBER() OVER (PARTITION BY c.id ORDER BY o.order_date DESC) AS order_desc
    FROM customers c
    JOIN orders o ON c.id = o.customer_id
    WHERE EXTRACT(YEAR FROM o.order_date) = 2024
),
order_intervals AS (
    SELECT 
        customer_id,
        name,
        email,
        COUNT(order_id) AS total_orders,
        AVG(EXTRACT(EPOCH FROM (order_date - prev_order_date))/86400.0) AS avg_days_between,
        STDDEV(EXTRACT(EPOCH FROM (order_date - prev_order_date))/86400.0) AS stddev_days_between,
        AVG(total_amount) AS avg_order_value,
        MIN(order_date) AS first_order_date,
        MAX(order_date) AS last_order_date,
        AVG(CASE WHEN order_asc <= 3 THEN total_amount END) AS first_3_avg,
        AVG(CASE WHEN order_desc <= 3 THEN total_amount END) AS last_3_avg
    FROM customer_orders
    GROUP BY customer_id, name, email
    HAVING COUNT(order_id) >= 3
),
favorite_category AS (
    SELECT 
        c.id AS customer_id,
        p.category,
        ROW_NUMBER() OVER(PARTITION BY c.id ORDER BY COUNT(oi.id) DESC) as rn
    FROM customers c
    JOIN orders o ON c.id = o.customer_id
    JOIN order_items oi ON o.id = oi.order_id
    JOIN products p ON oi.product_id = p.id
    WHERE EXTRACT(YEAR FROM o.order_date) = 2024
    GROUP BY c.id, p.category
)
SELECT 
    oi.name,
    oi.email,
    oi.total_orders,
    ROUND(oi.avg_days_between::numeric, 2) AS avg_days_between_orders,
    ROUND(oi.stddev_days_between::numeric, 2) AS consistency_metric,
    fc.category AS most_frequent_category,
    ROUND(oi.avg_order_value::numeric, 2) AS avg_order_value,
    CASE 
        WHEN oi.last_3_avg > oi.first_3_avg THEN 'Increasing'
        ELSE 'Decreasing'
    END AS trend_indicator,
    EXTRACT(DAY FROM (oi.last_order_date - oi.first_order_date)) AS lifetime_days
FROM order_intervals oi
JOIN favorite_category fc ON oi.customer_id = fc.customer_id AND fc.rn = 1
ORDER BY oi.total_orders DESC;
