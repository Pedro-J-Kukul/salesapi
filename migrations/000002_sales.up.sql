CREATE TABLE sales (
    id SERIAL PRIMARY KEY,
    cashier_name VARCHAR(255) NOT NULL,
    total_amount DECIMAL(10, 2) NOT NULL,
    cash_paid DECIMAL(10, 2) NOT NULL,
    change_due DECIMAL(10, 2) NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW()
);