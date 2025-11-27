-- File: seeds/000001_permissions.up.sql
-- Seed to insert initial permissions
INSERT INTO "permissions" (code) VALUES
('sale:create'),
('sale:view'),
('sale:delete'),
('sale:update'),
('product:create'),
('product:view'),
('product:delete'),
('product:update'),
('users:create'),
('users:view'),
('users:delete'),
('users:update'),
('self:create'),
('self:view'),
('self:delete'),
('self:update');