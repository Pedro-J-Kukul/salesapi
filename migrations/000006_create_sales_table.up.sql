-- File: migrations/000006_create_sales_table.up.sql
-- Migration to create the sales table
CREATE TABLE IF NOT EXISTS "sales" (
    "id" BIGSERIAL PRIMARY KEY,
    "user_id" BIGINT NOT NULL REFERENCES "users"("id") ON DELETE CASCADE,
    "product_id" BIGINT NOT NULL REFERENCES "products"("id") ON DELETE CASCADE,
    "quantity" INT NOT NULL,
    "sold_at" TIMESTAMP NOT NULL DEFAULT NOW()
);