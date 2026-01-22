-- CRM Kilang Desa Murni Batik - PostgreSQL Initialization Script
-- ==============================================================

-- Create databases for each service
CREATE DATABASE crm_iam;
CREATE DATABASE crm_sales;
CREATE DATABASE crm_notification;

-- Connect to IAM database and setup
\c crm_iam;

-- Enable UUID extension
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- Create schema
CREATE SCHEMA IF NOT EXISTS iam;

-- Set default search path
ALTER DATABASE crm_iam SET search_path TO iam, public;

-- Grant permissions
GRANT ALL PRIVILEGES ON DATABASE crm_iam TO crm_admin;
GRANT ALL PRIVILEGES ON SCHEMA iam TO crm_admin;

-- Connect to Sales database and setup
\c crm_sales;

-- Enable UUID extension
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- Create schema
CREATE SCHEMA IF NOT EXISTS sales;

-- Set default search path
ALTER DATABASE crm_sales SET search_path TO sales, public;

-- Grant permissions
GRANT ALL PRIVILEGES ON DATABASE crm_sales TO crm_admin;
GRANT ALL PRIVILEGES ON SCHEMA sales TO crm_admin;

-- Connect to Notification database and setup
\c crm_notification;

-- Enable UUID extension
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- Create schema
CREATE SCHEMA IF NOT EXISTS notification;

-- Set default search path
ALTER DATABASE crm_notification SET search_path TO notification, public;

-- Grant permissions
GRANT ALL PRIVILEGES ON DATABASE crm_notification TO crm_admin;
GRANT ALL PRIVILEGES ON SCHEMA notification TO crm_admin;

-- Log completion
\echo 'PostgreSQL initialization completed successfully!'
