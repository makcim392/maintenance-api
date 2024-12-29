package handlers

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/makcim392/swordhealth-interviewer/internal/models"
	"github.com/stretchr/testify/assert"
)

func TestCreateTask(t *testing.T) {
	// Create a new mock database connection
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create mock: %v", err)
	}
	defer db.Close()

	handler := NewTaskHandler(db)

	// Define a fixed timestamp for testing
	fixedTime := time.Date(2024, 12, 25, 10, 0, 0, 0, time.UTC)

	t.Run("successful task creation", func(t *testing.T) {
		// Test data
		task := models.Task{
			Summary:     "Test task",
			PerformedAt: fixedTime,
		}

		// Convert task to JSON
		taskJSON, err := json.Marshal(task)
		assert.NoError(t, err)

		// Create request
		req := httptest.NewRequest("POST", "/tasks", bytes.NewBuffer(taskJSON))
		req.Header.Set("Content-Type", "application/json")

		// Create response recorder
		rr := httptest.NewRecorder()

		// Set up expected SQL query
		mock.ExpectExec("INSERT INTO tasks").
			WithArgs(sqlmock.AnyArg(), 1, task.Summary, fixedTime).
			WillReturnResult(sqlmock.NewResult(1, 1))

		// Handle the request
		handler.CreateTask(rr, req)

		// Assert response
		assert.Equal(t, http.StatusCreated, rr.Code)

		// Verify all expected SQL calls were made
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("summary too long", func(t *testing.T) {
		// Create a task with summary > 2500 characters
		longSummary := string(make([]byte, 2501))
		task := models.Task{
			Summary:     longSummary,
			PerformedAt: fixedTime, // Use the same fixed time
		}

		taskJSON, err := json.Marshal(task)
		assert.NoError(t, err)

		req := httptest.NewRequest("POST", "/tasks", bytes.NewBuffer(taskJSON))
		req.Header.Set("Content-Type", "application/json")
		rr := httptest.NewRecorder()

		handler.CreateTask(rr, req)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
		assert.Contains(t, rr.Body.String(), "Summary must not exceed 2500 characters")
	})

	t.Run("invalid JSON payload", func(t *testing.T) {
		// Send invalid JSON
		req := httptest.NewRequest("POST", "/tasks", bytes.NewBufferString("{invalid json}"))
		req.Header.Set("Content-Type", "application/json")
		rr := httptest.NewRecorder()

		handler.CreateTask(rr, req)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})

	t.Run("database error", func(t *testing.T) {
		task := models.Task{
			Summary:     "Test task",
			PerformedAt: fixedTime,
		}

		taskJSON, err := json.Marshal(task)
		assert.NoError(t, err)

		req := httptest.NewRequest("POST", "/tasks", bytes.NewBuffer(taskJSON))
		req.Header.Set("Content-Type", "application/json")
		rr := httptest.NewRecorder()

		// Mock database error
		mock.ExpectExec("INSERT INTO tasks").
			WithArgs(sqlmock.AnyArg(), 1, task.Summary, fixedTime).
			WillReturnError(sql.ErrConnDone)

		handler.CreateTask(rr, req)

		assert.Equal(t, http.StatusInternalServerError, rr.Code)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("missing required field", func(t *testing.T) {
		// Create task with zero time (missing PerformedAt)
		task := models.Task{
			Summary: "Test task",
			// PerformedAt intentionally omitted
		}

		taskJSON, err := json.Marshal(task)
		assert.NoError(t, err)

		req := httptest.NewRequest("POST", "/tasks", bytes.NewBuffer(taskJSON))
		req.Header.Set("Content-Type", "application/json")
		rr := httptest.NewRecorder()

		handler.CreateTask(rr, req)

		// Note: You might need to adjust the expected status code based on your validation logic
		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})
}
