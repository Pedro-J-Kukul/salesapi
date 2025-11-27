-- Schema Initialization Script for Sales API
-- This script initializes the database schema by running all migrations

-- Users table
CREATE TABLE IF NOT EXISTS "users" (
    "id" BIGSERIAL PRIMARY KEY,
    "email" TEXT NOT NULL UNIQUE,
    "password_hash" bytea NOT NULL,
    "first_name" TEXT NOT NULL,
    "last_name" TEXT NOT NULL,
    "role" TEXT NOT NULL,
    "created_at" TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    "updated_at" TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    "is_active" BOOLEAN DEFAULT TRUE,
    "version" INT DEFAULT 1
);

-- Permissions table
CREATE TABLE IF NOT EXISTS "permissions" (
    "id" BIGSERIAL PRIMARY KEY,
    "code" TEXT NOT NULL UNIQUE
);

-- Tokens table
CREATE TABLE IF NOT EXISTS "tokens" (
    "id" BIGSERIAL PRIMARY KEY,
    "user_id" BIGINT NOT NULL REFERENCES "users"("id") ON DELETE CASCADE,
    "hash" bytea NOT NULL,
    "scope" TEXT NOT NULL,
    "expiry" TIMESTAMP NOT NULL
);

-- Users permissions table
CREATE TABLE IF NOT EXISTS "users_permissions" (
    "user_id" BIGINT NOT NULL REFERENCES "users"("id") ON DELETE CASCADE,
    "permission_id" BIGINT NOT NULL REFERENCES "permissions"("id") ON DELETE CASCADE,
    PRIMARY KEY ("user_id", "permission_id")
);

-- Products table
CREATE TABLE IF NOT EXISTS "products" (
    "id" BIGSERIAL PRIMARY KEY,
    "name" TEXT NOT NULL,
    "price" NUMERIC(10, 2) NOT NULL,
    "created_at" TIMESTAMP NOT NULL DEFAULT NOW(),
    "updated_at" TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Sales table
CREATE TABLE IF NOT EXISTS "sales" (
    "id" BIGSERIAL PRIMARY KEY,
    "user_id" BIGINT NOT NULL REFERENCES "users"("id") ON DELETE CASCADE,
    "product_id" BIGINT NOT NULL REFERENCES "products"("id") ON DELETE CASCADE,
    "quantity" INT NOT NULL,
    "sold_at" TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Create indexes for performance
CREATE INDEX IF NOT EXISTS idx_tokens_user_id ON tokens(user_id);
CREATE INDEX IF NOT EXISTS idx_tokens_hash ON tokens(hash);
CREATE INDEX IF NOT EXISTS idx_sales_user_id ON sales(user_id);
CREATE INDEX IF NOT EXISTS idx_sales_product_id ON sales(product_id);
CREATE INDEX IF NOT EXISTS idx_sales_sold_at ON sales(sold_at);
