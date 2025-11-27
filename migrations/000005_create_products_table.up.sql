-- File: migrations/000005_create_products_table.up.sql
-- Migration to create the products table
CREATE TABLE IF NOT EXISTS "products" (
    "id" BIGSERIAL PRIMARY KEY,
    "name" TEXT NOT NULL,
    "price" NUMERIC(10, 2) NOT NULL,
    "created_at" TIMESTAMP NOT NULL DEFAULT NOW(),
    "updated_at" TIMESTAMP NOT NULL DEFAULT NOW()
);