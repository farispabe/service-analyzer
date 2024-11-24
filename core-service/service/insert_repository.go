package service

import (
	"encoding/json"
	"fmt"
	"net/http"
)

// Insert new repository into the database
func InsertRepository(w http.ResponseWriter, r *http.Request) {
	var repo Repository
	err := json.NewDecoder(r.Body).Decode(&repo)
	if err != nil {
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}

	var repoID int
	err = db.QueryRow("INSERT INTO repositories (repo_url) VALUES ($1) RETURNING id", repo.RepoURL).Scan(&repoID)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error inserting repository: %v", err), http.StatusInternalServerError)
		return
	}

	repo.ID = repoID
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(repo)
}
