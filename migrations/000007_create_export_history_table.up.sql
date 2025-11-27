-- File: migrations/000007_create_export_history_table.up.sql
-- Migration to create the export_history table
CREATE TABLE IF NOT EXISTS "export_history" (
    "id" BIGSERIAL PRIMARY KEY,
    "user_id" BIGINT NOT NULL REFERENCES "users"("id") ON DELETE CASCADE,
    "export_type" TEXT NOT NULL,
    "spreadsheet_id" TEXT NOT NULL,
    "sheet_name" TEXT NOT NULL,
    "row_count" INT NOT NULL DEFAULT 0,
    "start_date" TIMESTAMP,
    "end_date" TIMESTAMP,
    "status" TEXT NOT NULL DEFAULT 'completed',
    "error_message" TEXT,
    "created_at" TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Create index on user_id for faster lookups
CREATE INDEX IF NOT EXISTS idx_export_history_user_id ON export_history(user_id);

-- Create index on created_at for date-based queries
CREATE INDEX IF NOT EXISTS idx_export_history_created_at ON export_history(created_at);
