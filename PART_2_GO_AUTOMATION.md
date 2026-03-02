# Part 2: Go Automation Tool - Data Retrieval Service

## Overview
This Go command-line tool automates data extraction from the PostgreSQL database and exports the results to properly formatted JSON files, fulfilling the requirements of Part 2.

The tool uses a layered architecture separating concerns into:
- `cmd/`: CLI routing using Cobra
- `configs/`: Environment variables parsing
- `internal/domain/`: Core data structures and interface contracts
- `internal/repository/postgres/`: SQL execution and database mapping
- `internal/usecase/`: Business logic, logging, and JSON file generation
- `pkg/`: Reusable utilities (database connection, structured logging)

## Implemented Reports
I implemented two of the complex queries from Part 1 into the CLI tool:
1. **Sales Trend Analysis** (Part 1, Query 4)
2. **Product Performance** (Part 1, Query 2)

## Installation & Setup
1. Copy the example configuration file:
   ```bash
   cp .env.example .env
   ```
2. Update the `.env` file with your database credentials.
3. Setup the database tables and mock data:
   ```bash
   make seed
   ```
4. Build the application binary:
   ```bash
   make build 
   ```

## Usage Examples

### 1. Generating Sales Trend Report
To extract the complex Sales Trend Analysis report out to JSON:
```bash
go run main.go report --type sales-trend
```
*Or using the Makefile alias:*
```bash
make run-report-sales
```
This generates a file in the `reports/` directory like: `reports/sales_trend_YYYYMMDD_HHMMSS.json`.

### 2. Generating Product Performance Report
To extract the comprehensive Product Performance report:
```bash
go run main.go report --type product-performance
```
*Or using the Makefile alias:*
```bash
make run-report-product
```
This generates a file in the `reports/` directory like: `reports/product_performance_YYYYMMDD_HHMMSS.json`.

## Technical Highlights
- Uses **sqlx** for safe, fast database mapping to structs.
- Built-in retry mechanisms and robust error handling.
- Structured contextual logging via `log/slog`.
- Centralized `Makefile` to easily run testing and building workflows.
