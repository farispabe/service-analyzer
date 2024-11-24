CREATE TABLE IF NOT EXISTS summaries (
    id SERIAL PRIMARY KEY,
    repo_id INT REFERENCES repositories(id) ON DELETE CASCADE,
    endpoint TEXT NOT NULL,
    file_location TEXT,
    file_extension TEXT,
    content TEXT,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    version INT DEFAULT 1
);
CREATE INDEX IF NOT EXISTS idx_summaries_version ON summaries(repo_id, endpoint, version DESC);