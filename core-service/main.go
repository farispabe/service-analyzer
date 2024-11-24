package main

import (
	"database/sql"
	"log"
	"net/http"

	"github.com/farispabe/service-analyzer/core-service/service"
	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
)

var (
	db *sql.DB
)

// Connect to the database
func initDB() {
	var err error
	connStr := "postgres://user:password@postgres:5432/mydb?sslmode=disable"
	db, err = sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal("Error connecting to the database: ", err)
	}
	service.SetDb(db)
}

func main() {
	initDB()
	defer db.Close()

	r := mux.NewRouter()

	r.HandleFunc("/repository", service.InsertRepository).Methods("POST")
	r.HandleFunc("/summary", service.InsertSummary).Methods("POST")
	r.HandleFunc("/diagram", service.InsertDiagram).Methods("POST")
	r.HandleFunc("/repository/{repo_id}/summaries", service.GetLatestSummaries).Methods("GET")
	r.HandleFunc("/repository/{repo_id}/diagrams", service.GetLatestDiagrams).Methods("GET")
	r.HandleFunc("/repository/{repo_id}/summaries/all", service.GetAllSummaries).Methods("GET")
	r.HandleFunc("/repository/{repo_id}/diagrams/all", service.GetAllDiagramsByType).Methods("GET")

	http.Handle("/", r)
	log.Fatal(http.ListenAndServe(":7012", nil))
}
