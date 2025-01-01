package handlers

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/gorilla/mux"
	"github.com/makcim392/swordhealth-interviewer/internal/middleware"
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

		taskJSON, err := json.Marshal(task)
		assert.NoError(t, err)

		// Create request with context
		req := httptest.NewRequest("POST", "/tasks", bytes.NewBuffer(taskJSON))
		req.Header.Set("Content-Type", "application/json")

		// Add user context
		ctx := context.WithValue(req.Context(), middleware.UserIDContextKey, 1)
		ctx = context.WithValue(ctx, middleware.RoleContextKey, string(models.RoleTechnician))
		req = req.WithContext(ctx)

		rr := httptest.NewRecorder()

		mock.ExpectExec("INSERT INTO tasks").
			WithArgs(sqlmock.AnyArg(), 1, task.Summary, fixedTime).
			WillReturnResult(sqlmock.NewResult(1, 1))

		handler.CreateTask(rr, req)

		assert.Equal(t, http.StatusCreated, rr.Code)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("unauthorized role", func(t *testing.T) {
		task := models.Task{
			Summary:     "Test task",
			PerformedAt: fixedTime,
		}

		taskJSON, err := json.Marshal(task)
		assert.NoError(t, err)

		req := httptest.NewRequest("POST", "/tasks", bytes.NewBuffer(taskJSON))
		req.Header.Set("Content-Type", "application/json")

		// Add manager role context (should be unauthorized)
		ctx := context.WithValue(req.Context(), middleware.UserIDContextKey, 1)
		ctx = context.WithValue(ctx, middleware.RoleContextKey, string(models.RoleManager))
		req = req.WithContext(ctx)

		rr := httptest.NewRecorder()

		handler.CreateTask(rr, req)

		assert.Equal(t, http.StatusForbidden, rr.Code)
		assert.Contains(t, rr.Body.String(), "Unauthorized to create task")
	})

	// Previous test cases remain the same...
	t.Run("summary too long", func(t *testing.T) {
		longSummary := string(make([]byte, 2501))
		task := models.Task{
			Summary:     longSummary,
			PerformedAt: fixedTime,
		}

		taskJSON, err := json.Marshal(task)
		assert.NoError(t, err)

		req := httptest.NewRequest("POST", "/tasks", bytes.NewBuffer(taskJSON))
		req.Header.Set("Content-Type", "application/json")

		ctx := context.WithValue(req.Context(), middleware.UserIDContextKey, 1)
		ctx = context.WithValue(ctx, middleware.RoleContextKey, string(models.RoleTechnician))
		req = req.WithContext(ctx)

		rr := httptest.NewRecorder()

		handler.CreateTask(rr, req)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
		assert.Contains(t, rr.Body.String(), "Summary must not exceed 2500 characters")
	})

	t.Run("missing context values", func(t *testing.T) {
		task := models.Task{
			Summary:     "Test task",
			PerformedAt: fixedTime,
		}

		taskJSON, err := json.Marshal(task)
		assert.NoError(t, err)

		req := httptest.NewRequest("POST", "/tasks", bytes.NewBuffer(taskJSON))
		req.Header.Set("Content-Type", "application/json")
		rr := httptest.NewRecorder()

		handler.CreateTask(rr, req)

		assert.Equal(t, http.StatusInternalServerError, rr.Code)
		assert.Contains(t, rr.Body.String(), "Unable to get user ID from context")
	})
}

func TestUpdateTask(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create mock: %v", err)
	}
	defer db.Close()

	handler := NewTaskHandler(db)
	fixedTime := time.Date(2024, 12, 25, 10, 0, 0, 0, time.UTC)

	t.Run("successful update by technician", func(t *testing.T) {
		task := models.Task{
			Summary:     "Updated task",
			PerformedAt: fixedTime,
		}

		taskJSON, err := json.Marshal(task)
		assert.NoError(t, err)

		req := httptest.NewRequest("PUT", "/tasks/123", bytes.NewBuffer(taskJSON))
		req.Header.Set("Content-Type", "application/json")

		// Add technician context
		ctx := context.WithValue(req.Context(), middleware.UserIDContextKey, 1)
		ctx = context.WithValue(ctx, middleware.RoleContextKey, string(models.RoleTechnician))
		req = req.WithContext(ctx)

		// Add URL parameters
		vars := map[string]string{
			"id": "123",
		}
		req = mux.SetURLVars(req, vars)

		rr := httptest.NewRecorder()

		// Expect the check for existing task
		mock.ExpectQuery("SELECT technician_id FROM tasks WHERE id = ?").
			WithArgs("123").
			WillReturnRows(sqlmock.NewRows([]string{"technician_id"}).AddRow(1))

		// Expect the update
		mock.ExpectExec("UPDATE tasks").
			WithArgs(task.Summary, task.PerformedAt, "123", 1).
			WillReturnResult(sqlmock.NewResult(1, 1))

		handler.UpdateTask(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("task not found", func(t *testing.T) {
		task := models.Task{
			Summary:     "Updated task",
			PerformedAt: fixedTime,
		}

		taskJSON, err := json.Marshal(task)
		assert.NoError(t, err)

		req := httptest.NewRequest("PUT", "/tasks/nonexistent", bytes.NewBuffer(taskJSON))
		req.Header.Set("Content-Type", "application/json")

		ctx := context.WithValue(req.Context(), middleware.UserIDContextKey, 1)
		ctx = context.WithValue(ctx, middleware.RoleContextKey, string(models.RoleTechnician))
		req = req.WithContext(ctx)

		vars := map[string]string{
			"id": "nonexistent",
		}
		req = mux.SetURLVars(req, vars)

		rr := httptest.NewRecorder()

		mock.ExpectQuery("SELECT technician_id FROM tasks WHERE id = ?").
			WithArgs("nonexistent").
			WillReturnError(sql.ErrNoRows)

		handler.UpdateTask(rr, req)

		assert.Equal(t, http.StatusNotFound, rr.Code)
		assert.Contains(t, rr.Body.String(), "Task not found")
	})

	t.Run("unauthorized update attempt", func(t *testing.T) {
		task := models.Task{
			Summary:     "Updated task",
			PerformedAt: fixedTime,
		}

		taskJSON, err := json.Marshal(task)
		assert.NoError(t, err)

		req := httptest.NewRequest("PUT", "/tasks/123", bytes.NewBuffer(taskJSON))
		req.Header.Set("Content-Type", "application/json")

		// Add different technician's context
		ctx := context.WithValue(req.Context(), middleware.UserIDContextKey, 2)
		ctx = context.WithValue(ctx, middleware.RoleContextKey, string(models.RoleTechnician))
		req = req.WithContext(ctx)

		vars := map[string]string{
			"id": "123",
		}
		req = mux.SetURLVars(req, vars)

		rr := httptest.NewRecorder()

		mock.ExpectQuery("SELECT technician_id FROM tasks WHERE id = ?").
			WithArgs("123").
			WillReturnRows(sqlmock.NewRows([]string{"technician_id"}).AddRow(1))

		handler.UpdateTask(rr, req)

		assert.Equal(t, http.StatusForbidden, rr.Code)
		assert.Contains(t, rr.Body.String(), "Unauthorized to modify this task")
	})

	t.Run("database error during update", func(t *testing.T) {
		task := models.Task{
			Summary:     "Updated task",
			PerformedAt: fixedTime,
		}

		taskJSON, err := json.Marshal(task)
		assert.NoError(t, err)

		req := httptest.NewRequest("PUT", "/tasks/123", bytes.NewBuffer(taskJSON))
		req.Header.Set("Content-Type", "application/json")

		ctx := context.WithValue(req.Context(), middleware.UserIDContextKey, 1)
		ctx = context.WithValue(ctx, middleware.RoleContextKey, string(models.RoleTechnician))
		req = req.WithContext(ctx)

		vars := map[string]string{
			"id": "123",
		}
		req = mux.SetURLVars(req, vars)

		rr := httptest.NewRecorder()

		mock.ExpectQuery("SELECT technician_id FROM tasks WHERE id = ?").
			WithArgs("123").
			WillReturnRows(sqlmock.NewRows([]string{"technician_id"}).AddRow(1))

		mock.ExpectExec("UPDATE tasks").
			WithArgs(task.Summary, task.PerformedAt, "123", 1).
			WillReturnError(sql.ErrConnDone)

		handler.UpdateTask(rr, req)

		assert.Equal(t, http.StatusInternalServerError, rr.Code)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}
