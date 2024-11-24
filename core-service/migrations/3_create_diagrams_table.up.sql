CREATE TABLE IF NOT EXISTS diagrams (
    id SERIAL PRIMARY KEY,
    repo_id INT REFERENCES repositories(id) ON DELETE CASCADE,
    endpoint TEXT NOT NULL,
    diagram_type TEXT NOT NULL,  -- Type of diagram (e.g., 'sequence', 'flow', 'class', etc.)
    file_location TEXT,
    file_extension TEXT,
    diagram_code TEXT,  -- The diagram syntax code (e.g., for PlantUML or similar)
    image BYTEA,  -- This will store the binary image file (for visual diagrams)
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    version INT DEFAULT 1
);
CREATE INDEX IF NOT EXISTS idx_diagrams_type_version ON diagrams(repo_id, endpoint, diagram_type, version DESC);