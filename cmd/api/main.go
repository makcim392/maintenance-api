package main

import (
	"database/sql"
	"log"
	"net/http"

	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/mux"
	"github.com/makcim392/swordhealth-interviewer/internal/handlers"
	"github.com/makcim392/swordhealth-interviewer/internal/middleware"
)

func main() {
	// Connect to database
	db, err := sql.Open("mysql", "user:password@tcp(mysql:3306)/tasks_db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// Create router
	router := mux.NewRouter()

	// Initialize handlers
	taskHandler := handlers.NewTaskHandler(db)
	authHandler := handlers.NewAuthHandler(db)

	// Auth routes
	router.HandleFunc("/login", authHandler.Login).Methods("POST")
	router.HandleFunc("/register", authHandler.Register).Methods("POST")

	// Register routes
	router.HandleFunc("/tasks", middleware.AuthMiddleware(taskHandler.CreateTask)).Methods("POST")
	router.HandleFunc("/test", handlers.TestHandler).Methods("GET")

	port := ":8080"

	// Start server
	log.Printf("Server starting on port %v", port)
	log.Fatal(http.ListenAndServe(port, router))
}
