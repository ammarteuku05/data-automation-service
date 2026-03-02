# Part 1: Complex SQL Queries & Database Analysis

## 1.1 Schema Analysis

### Performance Issues
With a scale of 1M+ customers, 10M+ orders, and 50M+ order items, the current schema will face severe performance bottlenecks:
1. **Missing Foreign Key Indexes**: PostgreSQL does not automatically index foreign keys (`customer_id`, `order_id`, `product_id`). Joins between customers, orders, and order_items will cause sequential scans, making queries extremely slow as tables grow.
2. **Unindexed Filtering/Sorting**: Queries filtering or sorting by `order_date` (e.g., daily sales, cohorts, recent orders) will require full table scans of 10M+ rows.
3. **Expensive Aggregations**: Real-time `SUM()` or `COUNT()` aggregations on 50M+ rows for dashboards will block CPU and memory.

### Index Recommendations
To support common query patterns and improve performance, the following 5 indexes are recommended (and included in `scripts/schema.sql`):
1. `CREATE INDEX idx_orders_customer_id ON orders(customer_id);`
   **Reason**: Crucial for fetching a customer's order history and joining customers with orders.
2. `CREATE INDEX idx_orders_order_date ON orders(order_date);`
   **Reason**: Most analytical and operational queries filter, group, or sort by date (e.g., sales in the last 90 days).
3. `CREATE INDEX idx_order_items_order_id ON order_items(order_id);`
   **Reason**: Required for looking up the items of an order, fundamental for joining orders and order_items.
4. `CREATE INDEX idx_order_items_product_id ON order_items(product_id);`
   **Reason**: Necessary for product-centric aggregations (which products were sold in which orders).
5. `CREATE INDEX idx_orders_status ON orders(status);`
   **Reason**: Allows partition-like quick filtering of 'completed', 'cancelled', or 'pending' orders without reading the entire table.

### Schema Improvements
1. **Partitioning**: Partition `orders` and `order_items` tables by `order_date` (e.g., monthly) to improve query speed on recent data and simplify data archiving.
2. **Denormalization / Materialized Views**: Add a summary table or materialized view for daily/monthly product sales and customer metrics to avoid recalculating heavy metrics (like RFM) directly from 50M rows continuously.
3. **Audit / Metadata columns**: Add `updated_at` timestamps to `orders` and allow tracking of specific status transitions (e.g., when an order became completed).

---

## 1.2 Write Complex SQL Queries
All 6 complex SQL queries (Cohort Analysis, Product Performance, RFM Segmentation, Sales Trend Analysis, Inventory Turnover, and Purchase Patterns) have been written and are extensively documented in the **`queries.sql`** file at the root of the project.
