# E-commerce Data Automation & Retrieval Service

This repository provides solutions to the Data Engineer Screening Test.

## Part 1 & 3: SQL and System Design
- `queries.sql`: Contains schema analysis, 5 index recommendations, and 6 complex SQL queries (Cohort, RFM, Product Performance, Sales Trend, Inventory, Customer Purchase Patterns).
- `AUTOMATION_DESIGN.md`: Explains the scheduler architecture, query debugging execution plans, and REST API push system design.

## Part 2 & 3: Go Data Automation Service
A Go CLI tool built with Clean Architecture, standard library, `sqlx` and `cobra`.

### Setup
Ensure you have set up a PostgreSQL database and populated it with the e-commerce schema (customers, products, orders, order_items).

1. Clone or download this directory.
2. Initialize dependencies: `go mod tidy`
3. Configure the environment by copying `.env.example` to `.env` and adjust the DB URL.

### Usage Example: Generating Reports
```bash
# General report usage
go run main.go report --type=product-performance --output=./reports

# Using sales trend report
go run main.go report --type=sales-trend
```

### Usage Example: Pushing to BI API (Part 3.3)
```bash
# Push yesterday's sales to API
go run main.go api-push

# Push specific date
go run main.go api-push --date=2024-11-29
```

Logs are output in JSON format using Go 1.21's `slog`.
