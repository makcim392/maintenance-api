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

func TestListTasks(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create mock: %v", err)
	}
	defer db.Close()

	handler := NewTaskHandler(db)
	fixedTime := time.Date(2024, 12, 25, 10, 0, 0, 0, time.UTC)
	formattedTime := fixedTime.Format("2006-01-02 15:04:05")

	t.Run("successful list tasks for technician", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/tasks", nil)

		// Add technician context
		ctx := context.WithValue(req.Context(), middleware.UserIDContextKey, 1)
		ctx = context.WithValue(ctx, middleware.RoleContextKey, string(models.RoleTechnician))
		req = req.WithContext(ctx)

		rr := httptest.NewRecorder()

		// Expect query for technician's tasks only
		rows := sqlmock.NewRows([]string{"id", "summary", "performed_at", "technician_id", "username"}).
			AddRow("task1", "Task 1 summary", formattedTime, 1, "tech1").
			AddRow("task2", "Task 2 summary", formattedTime, 1, "tech1")

		mock.ExpectQuery("SELECT t.id, t.summary, DATE_FORMAT.*FROM tasks t.*WHERE t.technician_id = ?.*").
			WithArgs(1).
			WillReturnRows(rows)

		handler.ListTasks(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)

		var tasks []map[string]interface{}
		err := json.Unmarshal(rr.Body.Bytes(), &tasks)
		assert.NoError(t, err)
		assert.Len(t, tasks, 2)
		assert.Equal(t, "task1", tasks[0]["id"])
		assert.Equal(t, float64(1), tasks[0]["technician_id"])
		assert.Equal(t, "tech1", tasks[0]["technician_name"])
	})

	t.Run("successful list tasks for manager", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/tasks", nil)

		// Add manager context
		ctx := context.WithValue(req.Context(), middleware.UserIDContextKey, 2)
		ctx = context.WithValue(ctx, middleware.RoleContextKey, string(models.RoleManager))
		req = req.WithContext(ctx)

		rr := httptest.NewRecorder()

		// Expect query for all tasks
		rows := sqlmock.NewRows([]string{"id", "summary", "performed_at", "technician_id", "username"}).
			AddRow("task1", "Task 1 summary", formattedTime, 1, "tech1").
			AddRow("task2", "Task 2 summary", formattedTime, 3, "tech2")

		mock.ExpectQuery("SELECT t.id, t.summary, DATE_FORMAT.*FROM tasks t.*ORDER BY t.performed_at DESC").
			WillReturnRows(rows)

		handler.ListTasks(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)

		var tasks []map[string]interface{}
		err := json.Unmarshal(rr.Body.Bytes(), &tasks)
		assert.NoError(t, err)
		assert.Len(t, tasks, 2)
		// Verify tasks from different technicians are included
		assert.Equal(t, float64(1), tasks[0]["technician_id"])
		assert.Equal(t, float64(3), tasks[1]["technician_id"])
	})

	t.Run("unauthorized role", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/tasks", nil)

		// Add invalid role context
		ctx := context.WithValue(req.Context(), middleware.UserIDContextKey, 1)
		ctx = context.WithValue(ctx, middleware.RoleContextKey, "invalid_role")
		req = req.WithContext(ctx)

		rr := httptest.NewRecorder()

		handler.ListTasks(rr, req)

		assert.Equal(t, http.StatusForbidden, rr.Code)
		assert.Contains(t, rr.Body.String(), "Unauthorized role")
	})

	t.Run("missing context values", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/tasks", nil)
		rr := httptest.NewRecorder()

		handler.ListTasks(rr, req)

		assert.Equal(t, http.StatusInternalServerError, rr.Code)
		assert.Contains(t, rr.Body.String(), "Unable to get user ID from context")
	})

	t.Run("database error", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/tasks", nil)

		ctx := context.WithValue(req.Context(), middleware.UserIDContextKey, 1)
		ctx = context.WithValue(ctx, middleware.RoleContextKey, string(models.RoleTechnician))
		req = req.WithContext(ctx)

		rr := httptest.NewRecorder()

		mock.ExpectQuery("SELECT t.id, t.summary, DATE_FORMAT.*").
			WithArgs(1).
			WillReturnError(sql.ErrConnDone)

		handler.ListTasks(rr, req)

		assert.Equal(t, http.StatusInternalServerError, rr.Code)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("error parsing date", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/tasks", nil)

		ctx := context.WithValue(req.Context(), middleware.UserIDContextKey, 1)
		ctx = context.WithValue(ctx, middleware.RoleContextKey, string(models.RoleTechnician))
		req = req.WithContext(ctx)

		rr := httptest.NewRecorder()

		// Return an invalid date format
		rows := sqlmock.NewRows([]string{"id", "summary", "performed_at", "technician_id", "username"}).
			AddRow("task1", "Task 1 summary", "invalid-date", 1, "tech1")

		mock.ExpectQuery("SELECT t.id, t.summary, DATE_FORMAT.*").
			WithArgs(1).
			WillReturnRows(rows)

		handler.ListTasks(rr, req)

		assert.Equal(t, http.StatusInternalServerError, rr.Code)
		assert.Contains(t, rr.Body.String(), "Error parsing date")
	})
}

func TestDeleteTask(t *testing.T) {
	// Create a new mock database connection
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create mock: %v", err)
	}
	defer db.Close()

	handler := NewTaskHandler(db)

	t.Run("successful deletion by manager", func(t *testing.T) {
		req := httptest.NewRequest("DELETE", "/tasks/123", nil)

		// Add manager context
		ctx := context.WithValue(req.Context(), middleware.UserIDContextKey, 1)
		ctx = context.WithValue(ctx, middleware.RoleContextKey, string(models.RoleManager))
		req = req.WithContext(ctx)

		// Add URL parameters
		vars := map[string]string{
			"id": "123",
		}
		req = mux.SetURLVars(req, vars)

		rr := httptest.NewRecorder()

		// Expect check for existing task
		mock.ExpectQuery("SELECT EXISTS\\(SELECT 1 FROM tasks WHERE id = \\?\\)").
			WithArgs("123").
			WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(1))

		// Expect the delete operation
		mock.ExpectExec("DELETE FROM tasks WHERE id = ?").
			WithArgs("123").
			WillReturnResult(sqlmock.NewResult(0, 1))

		handler.DeleteTask(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)

		var response map[string]string
		err := json.Unmarshal(rr.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "Task deleted successfully", response["message"])
		assert.Equal(t, "123", response["id"])
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("unauthorized role (technician)", func(t *testing.T) {
		req := httptest.NewRequest("DELETE", "/tasks/123", nil)

		// Add technician context
		ctx := context.WithValue(req.Context(), middleware.UserIDContextKey, 1)
		ctx = context.WithValue(ctx, middleware.RoleContextKey, string(models.RoleTechnician))
		req = req.WithContext(ctx)

		vars := map[string]string{
			"id": "123",
		}
		req = mux.SetURLVars(req, vars)

		rr := httptest.NewRecorder()

		handler.DeleteTask(rr, req)

		assert.Equal(t, http.StatusForbidden, rr.Code)
		assert.Contains(t, rr.Body.String(), "Unauthorized to delete tasks")
	})

	t.Run("task not found", func(t *testing.T) {
		req := httptest.NewRequest("DELETE", "/tasks/nonexistent", nil)

		// Add manager context
		ctx := context.WithValue(req.Context(), middleware.UserIDContextKey, 1)
		ctx = context.WithValue(ctx, middleware.RoleContextKey, string(models.RoleManager))
		req = req.WithContext(ctx)

		vars := map[string]string{
			"id": "nonexistent",
		}
		req = mux.SetURLVars(req, vars)

		rr := httptest.NewRecorder()

		mock.ExpectQuery("SELECT EXISTS\\(SELECT 1 FROM tasks WHERE id = \\?\\)").
			WithArgs("nonexistent").
			WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(0))

		handler.DeleteTask(rr, req)

		assert.Equal(t, http.StatusNotFound, rr.Code)
		assert.Contains(t, rr.Body.String(), "Task not found")
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("database error during check", func(t *testing.T) {
		req := httptest.NewRequest("DELETE", "/tasks/123", nil)

		ctx := context.WithValue(req.Context(), middleware.UserIDContextKey, 1)
		ctx = context.WithValue(ctx, middleware.RoleContextKey, string(models.RoleManager))
		req = req.WithContext(ctx)

		vars := map[string]string{
			"id": "123",
		}
		req = mux.SetURLVars(req, vars)

		rr := httptest.NewRecorder()

		mock.ExpectQuery("SELECT EXISTS\\(SELECT 1 FROM tasks WHERE id = \\?\\)").
			WithArgs("123").
			WillReturnError(sql.ErrConnDone)

		handler.DeleteTask(rr, req)

		assert.Equal(t, http.StatusInternalServerError, rr.Code)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("database error during delete", func(t *testing.T) {
		req := httptest.NewRequest("DELETE", "/tasks/123", nil)

		ctx := context.WithValue(req.Context(), middleware.UserIDContextKey, 1)
		ctx = context.WithValue(ctx, middleware.RoleContextKey, string(models.RoleManager))
		req = req.WithContext(ctx)

		vars := map[string]string{
			"id": "123",
		}
		req = mux.SetURLVars(req, vars)

		rr := httptest.NewRecorder()

		mock.ExpectQuery("SELECT EXISTS\\(SELECT 1 FROM tasks WHERE id = \\?\\)").
			WithArgs("123").
			WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(1))

		mock.ExpectExec("DELETE FROM tasks WHERE id = ?").
			WithArgs("123").
			WillReturnError(sql.ErrConnDone)

		handler.DeleteTask(rr, req)

		assert.Equal(t, http.StatusInternalServerError, rr.Code)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("missing context values", func(t *testing.T) {
		req := httptest.NewRequest("DELETE", "/tasks/123", nil)

		vars := map[string]string{
			"id": "123",
		}
		req = mux.SetURLVars(req, vars)

		rr := httptest.NewRecorder()

		handler.DeleteTask(rr, req)

		assert.Equal(t, http.StatusInternalServerError, rr.Code)
		assert.Contains(t, rr.Body.String(), "Unable to get role from context")
	})
}
