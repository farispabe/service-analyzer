package service

import (
	"encoding/json"
	"fmt"
	"net/http"
)

// Insert latest diagram into the database
func InsertDiagram(w http.ResponseWriter, r *http.Request) {
	var diagram Diagram
	err := json.NewDecoder(r.Body).Decode(&diagram)
	if err != nil {
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}

	var diagramID int
	err = db.QueryRow("INSERT INTO diagrams (repo_id, endpoint, diagram_type, file_location, file_extension, diagram_code, image) VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING id", diagram.RepoID, diagram.Endpoint, diagram.DiagramType, diagram.FileLocation, diagram.FileExtension, diagram.DiagramCode, diagram.Image).Scan(&diagramID)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error inserting diagram: %v", err), http.StatusInternalServerError)
		return
	}

	diagram.ID = diagramID
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(diagram)
}
