package integration

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"os"
	"testing"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	"github.com/makcim392/swordhealth-interviewer/internal/auth"
	"github.com/makcim392/swordhealth-interviewer/internal/handlers"
	"github.com/makcim392/swordhealth-interviewer/internal/middleware"
)

type TestServer struct {
	DB     *sql.DB
	Router *mux.Router
	Server *http.Server
	// Add cleanup function
	cleanup func()
}

func SetupTestServer(t *testing.T) *TestServer {

	// Load test environment variables
	if err := loadTestEnv(); err != nil {
		t.Fatalf("Failed to load test environment: %v", err)
	}

	// Setup database connection with context and timeout
	db, err := setupTestDB()
	if err != nil {
		t.Fatalf("Failed to setup test database: %v", err)
	}

	// Setup router and handlers
	router := setupRouter(db)

	// Create test server with proper configuration
	server := &http.Server{
		Addr:         ":8081",
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Create cleanup function
	cleanup := func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := server.Shutdown(ctx); err != nil {
			t.Errorf("Failed to shutdown server: %v", err)
		}
		if err := db.Close(); err != nil {
			t.Errorf("Failed to close database connection: %v", err)
		}
	}

	return &TestServer{
		DB:      db,
		Router:  router,
		Server:  server,
		cleanup: cleanup,
	}
}

func loadTestEnv() error {
	paths := []string{
		"../.env.test",    // Try local directory first
		"../../.env.test", // Then parent directory
	}

	var err error
	for _, path := range paths {
		err = godotenv.Load(path)
		if err == nil {
			return nil
		}
	}
	return fmt.Errorf("no valid .env.test file found in paths: %v", paths)
}

func setupTestDB() (*sql.DB, error) {
	dbHost := getEnvWithDefault("TEST_DB_HOST", "127.0.0.1") // Use 127.0.0.1 instead of localhost
	dbPort := getEnvWithDefault("TEST_DB_PORT", "3309")
	dbUser := getEnvWithDefault("TEST_DB_USER", "test_user")
	dbPassword := getEnvWithDefault("TEST_DB_PASSWORD", "test_password")
	dbName := getEnvWithDefault("TEST_DB_NAME", "test_db")

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true&timeout=30s",
		dbUser, dbPassword, dbHost, dbPort, dbName)

	fmt.Printf("Attempting to connect to database with DSN: %s\n",
		fmt.Sprintf("%s:%s@tcp(%s:%s)/%s", dbUser, "****", dbHost, dbPort, dbName))

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open database connection: %w", err)
	}

	// Set connection pool parameters
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(25)
	db.SetConnMaxLifetime(5 * time.Minute)

	// Wait for database with better logging
	deadline := time.Now().Add(30 * time.Second)
	var lastErr error
	for time.Now().Before(deadline) {
		err := db.Ping()
		if err == nil {
			fmt.Println("Successfully connected to database")
			return db, nil
		}
		lastErr = err
		fmt.Printf("Database not ready, retrying... Error: %v\n", err)
		time.Sleep(time.Second)
	}
	return nil, fmt.Errorf("database not ready after 30 seconds, last error: %v", lastErr)
}

func setupRouter(db *sql.DB) *mux.Router {
	router := mux.NewRouter()

	taskHandler := handlers.NewTaskHandler(db)
	authHandler := handlers.NewAuthHandler(db)
	validator := &auth.JWTValidator{}
	authMiddleware := middleware.NewAuthMiddlewareHandler(validator)

	router.HandleFunc("/login", authHandler.Login).Methods("POST")
	router.HandleFunc("/register", authHandler.Register).Methods("POST")
	router.HandleFunc("/tasks", authMiddleware.AuthMiddleware(taskHandler.CreateTask)).Methods("POST")
	router.HandleFunc("/tasks/{id}", authMiddleware.AuthMiddleware(taskHandler.UpdateTask)).Methods("PUT")

	return router
}

// Helper function to get environment variable with default value
func getEnvWithDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// CleanDB now returns error instead of failing test
func (ts *TestServer) CleanDB() error {
	// Reverse the order - truncate tasks first, then users
	tables := []string{"tasks", "users"}
	for _, table := range tables {
		// Temporarily disable foreign key checks before truncating
		if _, err := ts.DB.Exec("SET FOREIGN_KEY_CHECKS = 0"); err != nil {
			return fmt.Errorf("failed to disable foreign key checks: %w", err)
		}

		query := fmt.Sprintf("TRUNCATE TABLE %s", table)
		if _, err := ts.DB.Exec(query); err != nil {
			return fmt.Errorf("failed to truncate table %s: %w", table, err)
		}

		// Re-enable foreign key checks after truncating
		if _, err := ts.DB.Exec("SET FOREIGN_KEY_CHECKS = 1"); err != nil {
			return fmt.Errorf("failed to enable foreign key checks: %w", err)
		}
	}
	return nil
}

// Cleanup method to be called in tests
func (ts *TestServer) Cleanup() {
	if ts.cleanup != nil {
		ts.cleanup()
	}
}

// Helper method to recreate database schema
func (ts *TestServer) ResetSchema(t *testing.T, schemaSQL string) {
	_, err := ts.DB.Exec(schemaSQL)
	if err != nil {
		t.Fatalf("Failed to reset schema: %v", err)
	}
}
