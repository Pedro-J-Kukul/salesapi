-- Database Setup Script for Sales API
-- This script creates the database and user for the Sales API application

-- NOTE: This script uses a simple password for testing/development purposes.
-- For production environments, use a strong password and consider using 
-- environment variables or a secrets management system.

-- Create the sales database
CREATE DATABASE sales;

-- Create the sales user with password
-- WARNING: Change this password for production use
CREATE USER sales WITH PASSWORD 'sales';

-- Grant all privileges on the sales database to the sales user
GRANT ALL PRIVILEGES ON DATABASE sales TO sales;

-- Connect to the sales database to set additional permissions
\c sales

-- Grant schema permissions to sales user
GRANT ALL ON SCHEMA public TO sales;

-- Grant permissions on all tables in public schema
GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA public TO sales;

-- Grant permissions on all sequences in public schema
GRANT ALL PRIVILEGES ON ALL SEQUENCES IN SCHEMA public TO sales;

-- Set default privileges for future tables
ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT ALL ON TABLES TO sales;
ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT ALL ON SEQUENCES TO sales;

-- Change ownership of the sales database to sales user
ALTER DATABASE sales OWNER TO sales;
