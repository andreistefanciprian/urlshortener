CREATE TABLE IF NOT EXISTS short_links (
    code          VARCHAR(16) PRIMARY KEY,
    original_url  TEXT        NOT NULL,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    expires_at    TIMESTAMPTZ NULL
);

INSERT INTO short_links(code, original_url, created_at, expires_at) VALUES
('wRZ32pT', 'https://example.com/long-url-1', now(), now() + INTERVAL '1 day'),
('GD38yB5', 'https://example.com/long-url-2', now(), NULL),
('AalGQ9g', 'https://example.com/long-url-3', now(), now() + INTERVAL '2 days'),
('N9qGJxH', 'https://example.com/long-url-4', now(), now() - INTERVAL '1 hour'); -- expired link for testing