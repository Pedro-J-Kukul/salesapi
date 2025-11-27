-- File: seeds/000001_permissions.down.sql
-- Description: Revert permissions changes made in 000001_permissions.up.sql
TRUNCATE TABLE permissions RESTART IDENTITY CASCADE;