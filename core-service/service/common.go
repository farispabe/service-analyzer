package service

import "database/sql"

var (
	db *sql.DB
)

func SetDb(sqlDb *sql.DB) {
	db = sqlDb
}

// Repository struct to represent a repository
type Repository struct {
	ID      int    `json:"id"`
	RepoURL string `json:"repo_url"`
}

// Summary struct to represent a summary of an endpoint
type Summary struct {
	ID            int    `json:"id"`
	RepoID        int    `json:"repo_id"`
	Endpoint      string `json:"endpoint"`
	FileLocation  string `json:"file_location"`
	FileExtension string `json:"file_extension"`
	Content       string `json:"content"`
}

// Diagram struct to represent a diagram for an endpoint
type Diagram struct {
	ID            int    `json:"id"`
	RepoID        int    `json:"repo_id"`
	Endpoint      string `json:"endpoint"`
	DiagramType   string `json:"diagram_type"`
	FileLocation  string `json:"file_location"`
	FileExtension string `json:"file_extension"`
	DiagramCode   string `json:"diagram_code"`
	Image         []byte `json:"image"`
}
