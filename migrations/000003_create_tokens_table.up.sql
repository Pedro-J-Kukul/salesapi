-- File: migrations/000003_create_tokens_table.up.sql
-- Migration to create the tokens table
CREATE TABLE IF NOT EXISTS "tokens" (
    "hash" bytea PRIMARY KEY,
    "user_id" BIGINT NOT NULL REFERENCES "users"("id") ON DELETE CASCADE,
    "scope" TEXT NOT NULL,
    "expires_at" TIMESTAMP NOT NULL
);