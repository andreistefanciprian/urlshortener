-- Simple PostgreSQL setup for URL shortener

-- Create user if not exists
DO $$
BEGIN
    IF NOT EXISTS (SELECT FROM pg_catalog.pg_roles WHERE rolname = 'url_gen_user') THEN
        CREATE USER url_gen_user WITH PASSWORD 'Auth123';
    END IF;
END
$$;

-- Create database if not exists
SELECT 'CREATE DATABASE urls' WHERE NOT EXISTS (SELECT FROM pg_database WHERE datname = 'urls')\gexec

-- Grant database privileges
GRANT ALL PRIVILEGES ON DATABASE urls TO url_gen_user;

-- Connect to urls database
\c urls;

-- Create table
CREATE TABLE IF NOT EXISTS short_links (
    code          VARCHAR(16) PRIMARY KEY,
    original_url  TEXT        NOT NULL,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    expires_at    TIMESTAMPTZ NULL
);

-- Grant table privileges
GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA public TO url_gen_user;
GRANT ALL PRIVILEGES ON ALL SEQUENCES IN SCHEMA public TO url_gen_user;