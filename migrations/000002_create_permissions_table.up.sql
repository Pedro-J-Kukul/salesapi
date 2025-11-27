-- Migration to create the permissions table
CREATE TABLE IF NOT EXISTS "permissions" (
    "id" BIGSERIAL PRIMARY KEY,
    "code" TEXT NOT NULL UNIQUE
);