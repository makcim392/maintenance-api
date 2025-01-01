package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"

	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	"github.com/makcim392/swordhealth-interviewer/internal/handlers"
	"github.com/makcim392/swordhealth-interviewer/internal/middleware"
)

func main() {
	// Load .env file
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}

	// Get database connection details from environment variables
	dbUser := os.Getenv("DB_USER")
	dbPassword := os.Getenv("DB_PASSWORD")
	dbHost := os.Getenv("DB_HOST")
	dbName := os.Getenv("DB_NAME")

	// Connect to database
	dsn := dbUser + ":" + dbPassword + "@tcp(" + dbHost + ":3306)/" + dbName
	db, err := sql.Open("mysql", dsn)
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
	router.HandleFunc("/tasks/update", middleware.AuthMiddleware(taskHandler.UpdateTask)).Methods("PUT")
	router.HandleFunc("/test", handlers.TestHandler).Methods("GET")

	// Get server port from environment variables
	port := os.Getenv("APP_PORT_HOST")
	if port == "" {
		port = "8080" // Default to 8080 if not set
	}

	// Start server
	log.Printf("Server starting on port %v", port)
	log.Fatal(http.ListenAndServe(":"+port, router))
}
