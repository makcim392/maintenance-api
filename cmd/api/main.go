package main

import (
	"database/sql"
	"log"
	"net/http"

	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/mux"
	"github.com/makcim392/swordhealth-interviewer/internal/handlers"
)

func main() {
	// Connect to database
	db, err := sql.Open("mysql", "user:password@tcp(localhost:3307)/tasks_db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// Create router
	router := mux.NewRouter()

	// Initialize handler
	taskHandler := handlers.NewTaskHandler(db)

	// Register routes
	router.HandleFunc("/tasks", taskHandler.CreateTask).Methods("POST")

	// Dev server
	devServer := ":8081"

	// Start server
	log.Printf("Server starting on %v", devServer)
	log.Fatal(http.ListenAndServe(devServer, router))
}
