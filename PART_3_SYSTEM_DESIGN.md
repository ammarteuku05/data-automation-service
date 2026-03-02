# Part 3: Problem-Solving & System Design

## 3.1 Automation Opportunity Identification

**Business context:** The team manually runs 4 requests: Weekly sales (Mon), Low stock alerts (daily), Churn risk analysis, Monthly revenue by category.

### 1. Approach: Automating Recurring Requests
We can automate these tasks by building a **Scheduler Service** that triggers data extraction jobs. The service will use the Go "Data Retrieval Service" CLI, either directly via cron jobs or orchestrated by a centralized job scheduler like Temporal, Apache Airflow, or Argo Workflows.

### 2. Architecture
- **Worker/Executor layer**: A Go application running on a container or cloud function, capable of executing predefined SQL queries and formatting results to JSON/CSV or sending API payloads.
- **Scheduling Layer**: A Cron-based scheduler (or Airflow) invoking the worker.
- **Reporting/Delivery Layer**: Once the file is exported (e.g., pushed to AWS S3), a notification is triggered via webhook/Slack/email, or the data is continuously PUSHED to a REST API.

### 3. Technology Choices
- **Scheduler**: Kubernetes CronJobs or standalone crond for simplicity. If workflows become complex and interdependent, **Temporal** or **Airflow**.
- **Queues**: If generation is heavy, RabbitMQ/SQS to queue the generation requests, processing them asynchronously to protect the database layer from traffic spikes.
- **Notification**: Slack webhook for internal alerts (like low stock), SendGrid for emailing reports.

### 4. Prioritization
I would automate **2. Low stock alerts (every day)** first.
*Why?* Low stock has direct revenue impact. If an item runs out of stock without being noticed, it translates immediately to lost sales. Automating daily reports offers immediate, high-value operations relief compared to monthly analytical views.

### 5. Scalability
Handling 10 concurrent automated reports:
- Implement **Connection Pooling** (e.g., using `pgxpool` in Go or `PgBouncer` proxy) so 10 requests don't choke Postgres limits.
- Run queries asynchronously by offloading work to a queue (like Kafka/RabbitMQ) handled by multiple Go worker containers.
- For large analytical queries, create Read Replicas of the database to route analytical traffic without impacting customer-facing order flow.

---

## 3.2 Query Performance Debugging

### Slow Query
```sql
SELECT c.name, c.email, COUNT(o.id) as order_count, SUM(o.total_amount) as total_revenue
FROM customers c LEFT JOIN orders o ON c.id = o.customer_id
WHERE o.order_date >= '2024-01-01'
GROUP BY c.id, c.name, c.email HAVING COUNT(o.id) > 5 ORDER BY total_revenue DESC;
```

### 1. Root Cause: Why is this query slow?
- **Missing Foreign Key Index**: Without an index on `o.customer_id`, joining 2 million customers against 15 million orders forces a **Sequential Scan** across 15M records.
- **Aggregating Unfiltered Rows**: Grouping by strings (`c.name`, `c.email`) after a huge table hash join uses massive temporary memory buffers.
- **LEFT JOIN logic flaw**: Using `o.order_date >= ...` in the `WHERE` clause fundamentally changes the `LEFT JOIN` into an `INNER JOIN`, misleading the query planner.

### 2. Quick Fixes
- Create indexes:
  ```sql
  CREATE INDEX idx_orders_customer_id_date ON orders(customer_id, order_date);
  CREATE INDEX idx_orders_order_date ON orders(order_date);
  ```

### 3. Long-term Solutions
- Move heavy, repetitive aggregate logic to an incrementally updated **Materialized View** (e.g., `customer_yearly_metrics`) that gets refreshed nightly, so analytical reports take milliseconds regardless of dataset size.

### 4. Rewrite (Optimized Version)
*Optimization: Use an Inner Join, aggregate orders PRE-join rather than post-join to reduce data cardinality, reducing the load on string grouping and sorting.*

```sql
WITH filtered_orders AS (
    SELECT 
        customer_id,
        COUNT(id) AS order_count,
        SUM(total_amount) AS total_revenue
    FROM orders
    WHERE order_date >= '2024-01-01'
    GROUP BY customer_id
    HAVING COUNT(id) > 5
)
SELECT 
    c.name,
    c.email,
    fo.order_count,
    fo.total_revenue
FROM customers c
JOIN filtered_orders fo ON c.id = fo.customer_id
ORDER BY fo.total_revenue DESC;
```

---

## 3.3 REST API Integration Scenario
The company wants to automatically push daily sales reports to their business intelligence platform via REST API.

I have built this functionality into the Go CLI via a new `api-push` command.
It queries the DB for the requested payload structure, marshals it to JSON, and sends an `Authorization: Bearer <token>` POST request to the API.

### Usage
```bash
go run main.go api-push
```

### Scheduling
To fully automate this to push daily sales at 2:00 AM, you would add an entry to the server's cron tab:
1. Run `crontab -e`
2. Add the following line:
```bash
0 2 * * * cd /path/to/data-automation-service && ./data-automation api-push >> /var/log/daily-api-push.log 2>&1
```
