-- Insert mock customers
INSERT INTO customers (email, name, country, created_at)
VALUES 
    ('alice@example.com', 'Alice Smith', 'USA', '2024-01-15 10:00:00'),
    ('bob@example.com', 'Bob Johnson', 'UK', '2024-02-20 14:30:00'),
    ('charlie@example.com', 'Charlie Brown', 'Canada', '2024-03-05 09:15:00')
ON CONFLICT DO NOTHING;

-- Insert mock products
INSERT INTO products (name, category, price, stock_quantity)
VALUES 
    ('Laptop Pro', 'Electronics', 1200.00, 50),
    ('Wireless Mouse', 'Electronics', 25.00, 200),
    ('Mechanical Keyboard', 'Electronics', 150.00, 75),
    ('Coffee Mug', 'Home', 15.00, 300)
ON CONFLICT DO NOTHING;

-- Insert mock orders (Ensuring dates are recent for trend reports)
INSERT INTO orders (customer_id, order_date, status, total_amount)
VALUES 
    (1, CURRENT_TIMESTAMP - INTERVAL '2 days', 'completed', 1225.00),
    (2, CURRENT_TIMESTAMP - INTERVAL '5 days', 'completed', 150.00),
    (3, CURRENT_TIMESTAMP - INTERVAL '1 days', 'completed', 30.00),
    (1, CURRENT_TIMESTAMP - INTERVAL '10 days', 'completed', 25.00)
ON CONFLICT DO NOTHING;

-- Insert mock order items
INSERT INTO order_items (order_id, product_id, quantity, unit_price)
VALUES 
    (1, 1, 1, 1200.00),
    (1, 2, 1, 25.00),
    (2, 3, 1, 150.00),
    (3, 4, 2, 15.00),
    (4, 2, 1, 25.00)
ON CONFLICT DO NOTHING;
