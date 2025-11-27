-- File: migrations/000004_create_users_permissions_table.up.sql
-- Migration to create the users_permissions table
CREATE TABLE IF NOT EXISTS "users_permissions" (
    "user_id" BIGINT NOT NULL REFERENCES "users"("id") ON DELETE CASCADE,
    "permission_id" BIGINT NOT NULL REFERENCES "permissions"("id") ON DELETE CASCADE,
    PRIMARY KEY ("user_id", "permission_id")
);