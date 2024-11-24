package service

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
)

// Get all diagrams grouped by endpoint and then by diagram type and sorted by the latest one first
func GetAllDiagramsByType(w http.ResponseWriter, r *http.Request) {
	repoID := mux.Vars(r)["repo_id"]
	rows, err := db.Query(`
		SELECT id, repo_id, endpoint, diagram_type, file_location, file_extension, diagram_code, image
		FROM diagrams
		WHERE repo_id = $1
		ORDER BY endpoint, diagram_type, id DESC`, repoID)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error fetching diagrams: %v", err), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var diagrams []Diagram
	for rows.Next() {
		var diagram Diagram
		if err := rows.Scan(&diagram.ID, &diagram.RepoID, &diagram.Endpoint, &diagram.DiagramType, &diagram.FileLocation, &diagram.FileExtension, &diagram.DiagramCode, &diagram.Image); err != nil {
			http.Error(w, fmt.Sprintf("Error scanning diagram: %v", err), http.StatusInternalServerError)
			return
		}
		diagrams = append(diagrams, diagram)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(diagrams)
}
