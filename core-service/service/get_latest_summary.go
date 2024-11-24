package service

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
)

// Get latest summary for each endpoint of a repository
func GetLatestSummaries(w http.ResponseWriter, r *http.Request) {
	repoID := mux.Vars(r)["repo_id"]
	rows, err := db.Query(`
		SELECT id, repo_id, endpoint, file_location, file_extension, content
		FROM summaries
		WHERE repo_id = $1
		ORDER BY id DESC`, repoID)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error fetching summaries: %v", err), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var summaries []Summary
	for rows.Next() {
		var summary Summary
		if err := rows.Scan(&summary.ID, &summary.RepoID, &summary.Endpoint, &summary.FileLocation, &summary.FileExtension, &summary.Content); err != nil {
			http.Error(w, fmt.Sprintf("Error scanning summary: %v", err), http.StatusInternalServerError)
			return
		}
		summaries = append(summaries, summary)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(summaries)
}
