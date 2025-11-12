
-- Connect to urls database
\c urls;

-- Create table
CREATE TABLE IF NOT EXISTS short_links (
    code          VARCHAR(16) PRIMARY KEY,
    original_url  TEXT        NOT NULL,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    expires_at    TIMESTAMPTZ NULL
);