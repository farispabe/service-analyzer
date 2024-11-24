package service

import (
	"encoding/json"
	"fmt"
	"net/http"
)

// Insert latest summary into the database
func InsertSummary(w http.ResponseWriter, r *http.Request) {
	var summary Summary
	err := json.NewDecoder(r.Body).Decode(&summary)
	if err != nil {
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}

	var summaryID int
	err = db.QueryRow("INSERT INTO summaries (repo_id, endpoint, file_location, file_extension, content) VALUES ($1, $2, $3, $4, $5) RETURNING id", summary.RepoID, summary.Endpoint, summary.FileLocation, summary.FileExtension, summary.Content).Scan(&summaryID)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error inserting summary: %v", err), http.StatusInternalServerError)
		return
	}

	summary.ID = summaryID
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(summary)
}
