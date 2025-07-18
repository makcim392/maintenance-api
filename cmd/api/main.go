package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/makcim392/maintenance-api/internal/auth"
	"github.com/makcim392/maintenance-api/internal/health"
	"github.com/makcim392/maintenance-api/internal/logger"
	"github.com/makcim392/maintenance-api/internal/metrics"
	"github.com/makcim392/maintenance-api/internal/server"

	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	"github.com/makcim392/maintenance-api/internal/handlers"
	"github.com/makcim392/maintenance-api/internal/middleware"
)

func main() {
	// Load default.env for base configuration
	err := godotenv.Load("default.env")
	if err != nil {
		log.Fatalf("Error loading default.env file: %v", err)
	}

	// Load .env to override default.env
	err = godotenv.Overload(".env")
	if err != nil {
		log.Printf("No .env file found or failed to load it: %v", err)
	}

	// Get application environment
	appEnv := os.Getenv("APP_ENV")
	if appEnv == "" {
		appEnv = "prod" // Default to production if not set
	}

	// Get database connection details from environment variables
	dbHost := os.Getenv("DB_HOST")

	dbPort := "3306" // Default to container port
	if appEnv == "dev" {
		dbHost = os.Getenv("DEV_DB_HOST")
		dbPort = os.Getenv("DEV_DB_PORT")
	}

	dbUser := os.Getenv("DB_USER")
	dbPassword := os.Getenv("DB_PASSWORD")
	dbName := os.Getenv("DB_NAME")

	// Default to local port for debugging if not set
	if dbPort == "" {
		dbPort = "3307"
	}

	// Construct DSN and connect to the database
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s", dbUser, dbPassword, dbHost, dbPort, dbName)
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		log.Fatalf("Error connecting to database: %v", err)
	}
	defer db.Close()

	// Test database connection
	if err := db.Ping(); err != nil {
		log.Fatalf("Error pinging database: %v", err)
	}

	// Initialize logger
	appLogger := logger.New()

	// Create router
	router := mux.NewRouter()

	// Add middleware
	router.Use(middleware.LoggingMiddleware(appLogger))
	router.Use(metrics.MetricsMiddleware)

	// Initialize handlers
	taskHandler := handlers.NewTaskHandler(db)
	authHandler := handlers.NewAuthHandler(db)
	healthChecker := health.New(db, appLogger)

	validator := &auth.JWTValidator{}
	authMiddleware := middleware.NewAuthMiddlewareHandler(validator)

	// Auth routes
	router.HandleFunc("/login", authHandler.Login).Methods("POST")
	router.HandleFunc("/register", authHandler.Register).Methods("POST")

	// Task routes
	router.HandleFunc("/tasks", authMiddleware.AuthMiddleware(taskHandler.CreateTask)).Methods("POST")
	router.HandleFunc("/tasks/{id}", authMiddleware.AuthMiddleware(taskHandler.UpdateTask)).Methods("PUT")
	router.HandleFunc("/tasks", authMiddleware.AuthMiddleware(taskHandler.ListTasks)).Methods("GET")
	router.HandleFunc("/tasks/{id}", authMiddleware.AuthMiddleware(taskHandler.DeleteTask)).Methods("DELETE")

	router.HandleFunc("/test", handlers.TestHandler).Methods("GET")
	
	// Health check endpoints
	router.HandleFunc("/health", healthChecker.HealthHandler).Methods("GET")
	router.HandleFunc("/health/ready", healthChecker.ReadinessHandler).Methods("GET")
	router.HandleFunc("/health/live", healthChecker.LivenessHandler).Methods("GET")
	
	// Add metrics endpoint
	router.Handle("/metrics", metrics.MetricsHandler()).Methods("GET")

	// Get server port from environment variables
	port := os.Getenv("APP_PORT_HOST")
	if port == "" {
		port = "8080" // Default to 8080 if not set
	}

	// Create and start server with graceful shutdown
	srv := server.New(":"+port, router, appLogger, healthChecker)
	appLogger.LogError(srv.Start(), "Server failed to start")
}
