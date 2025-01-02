package integration

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/makcim392/swordhealth-interviewer/internal/models"
	"github.com/stretchr/testify/assert"
)

func TestTaskCreation(t *testing.T) {
	server := SetupTestServer(t)
	defer server.DB.Close()

	if err := server.CleanDB(); err != nil {
		t.Fatalf("Failed to clean database: %v", err)
	}

	// First register a test user
	user := models.User{
		Username: "test@example.com",
		Password: "password123",
		Role:     models.RoleTechnician,
	}

	// Register user and get token
	token := registerAndLogin(t, server, user)

	t.Run("create task successfully", func(t *testing.T) {
		// Get the technician ID first
		var technicianID int
		err := server.DB.QueryRow("SELECT id FROM users WHERE username = ?", user.Username).Scan(&technicianID)
		assert.NoError(t, err)

		task := models.Task{
			Summary:      "Test task",
			PerformedAt:  time.Now(),
			TechnicianID: int64(technicianID),
		}

		taskJSON, _ := json.Marshal(task)
		req := httptest.NewRequest("POST", "/tasks", bytes.NewBuffer(taskJSON))
		req.Header.Set("Authorization", "Bearer "+token)
		req.Header.Set("Content-Type", "application/json")

		rr := httptest.NewRecorder()
		server.Router.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusCreated, rr.Code)

		// Verify task was created in database
		var count int
		err = server.DB.QueryRow("SELECT COUNT(*) FROM tasks").Scan(&count)
		assert.NoError(t, err)
		assert.Equal(t, 1, count)
	})
}

// Helper function to register and login a user
func registerAndLogin(t *testing.T, server *TestServer, user models.User) string {
	// Register
	userJSON, _ := json.Marshal(user)
	req := httptest.NewRequest("POST", "/register", bytes.NewBuffer(userJSON))
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	server.Router.ServeHTTP(rr, req)
	assert.Equal(t, http.StatusCreated, rr.Code)

	// Login
	loginReq := httptest.NewRequest("POST", "/login", bytes.NewBuffer(userJSON))
	loginReq.Header.Set("Content-Type", "application/json")

	loginRR := httptest.NewRecorder()
	server.Router.ServeHTTP(loginRR, loginReq)

	var response map[string]string
	json.NewDecoder(loginRR.Body).Decode(&response)

	return response["token"]
}

func TestListTasks(t *testing.T) {
	server := SetupTestServer(t)
	defer server.Cleanup()

	// Clean existing data
	if err := server.CleanDB(); err != nil {
		t.Fatalf("Failed to clean database: %v", err)
	}

	// Create test users
	technician := models.User{
		Username: "tech@example.com",
		Password: "password123",
		Role:     models.RoleTechnician,
	}

	manager := models.User{
		Username: "manager@example.com",
		Password: "password123",
		Role:     models.RoleManager,
	}

	// Register and login users
	techToken := registerAndLogin(t, server, technician)
	managerToken := registerAndLogin(t, server, manager)

	// Get the technician ID
	var technicianID int
	err := server.DB.QueryRow("SELECT id FROM users WHERE username = ?", technician.Username).Scan(&technicianID)
	assert.NoError(t, err)
	assert.NotEqual(t, 0, technicianID, "Technician ID should not be 0")

	// Create test tasks
	tasks := []models.Task{
		{
			Summary:      "Task 1 for testing",
			PerformedAt:  time.Now().Add(-1 * time.Hour),
			TechnicianID: int64(technicianID),
		},
		{
			Summary:      "Task 2 for testing",
			PerformedAt:  time.Now(),
			TechnicianID: int64(technicianID),
		},
	}

	// Insert tasks with UUID as ID
	for _, task := range tasks {
		taskID := uuid.New().String()
		_, err = server.DB.Exec(
			"INSERT INTO tasks (id, summary, performed_at, technician_id) VALUES (?, ?, ?, ?)",
			taskID,
			task.Summary,
			task.PerformedAt.Format("2006-01-02 15:04:05"),
			task.TechnicianID,
		)
		assert.NoError(t, err)
	}

	// Verify tasks were inserted
	var count int
	err = server.DB.QueryRow("SELECT COUNT(*) FROM tasks").Scan(&count)
	assert.NoError(t, err)
	assert.Equal(t, 2, count, "Expected 2 tasks to be created")

	t.Run("technician can only see their own tasks", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/tasks", nil)
		req.Header.Set("Authorization", "Bearer "+techToken)

		rr := httptest.NewRecorder()
		server.Router.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)

		var response []struct {
			ID           string    `json:"id"`
			Summary      string    `json:"summary"`
			PerformedAt  time.Time `json:"performed_at"`
			TechnicianID int       `json:"technician_id"`
			Username     string    `json:"technician_name"`
		}
		err := json.NewDecoder(rr.Body).Decode(&response)
		assert.NoError(t, err)

		assert.Len(t, response, 2, "Technician should see 2 tasks")
		for _, task := range response {
			assert.Equal(t, technicianID, task.TechnicianID)
			assert.Equal(t, technician.Username, task.Username)
		}
	})

	t.Run("manager can see all tasks", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/tasks", nil)
		req.Header.Set("Authorization", "Bearer "+managerToken)

		rr := httptest.NewRecorder()
		server.Router.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)

		var response []struct {
			ID           string    `json:"id"`
			Summary      string    `json:"summary"`
			PerformedAt  time.Time `json:"performed_at"`
			TechnicianID int       `json:"technician_id"`
			Username     string    `json:"technician_name"`
		}
		err := json.NewDecoder(rr.Body).Decode(&response)
		assert.NoError(t, err)

		assert.Len(t, response, 2, "Manager should see all 2 tasks")
	})

	t.Run("unauthorized access returns 401", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/tasks", nil)
		rr := httptest.NewRecorder()
		server.Router.ServeHTTP(rr, req)
		assert.Equal(t, http.StatusUnauthorized, rr.Code)
	})

	t.Run("invalid token returns 401", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/tasks", nil)
		req.Header.Set("Authorization", "Bearer invalid-token")
		rr := httptest.NewRecorder()
		server.Router.ServeHTTP(rr, req)
		assert.Equal(t, http.StatusUnauthorized, rr.Code)
	})
}

func TestUpdateTask(t *testing.T) {
	server := SetupTestServer(t)
	defer server.Cleanup()

	if err := server.CleanDB(); err != nil {
		t.Fatalf("Failed to clean database: %v", err)
	}

	// Create test users
	technician1 := models.User{
		Username: "tech1@example.com",
		Password: "password123",
		Role:     models.RoleTechnician,
	}

	technician2 := models.User{
		Username: "tech2@example.com",
		Password: "password123",
		Role:     models.RoleTechnician,
	}

	manager := models.User{
		Username: "manager@example.com",
		Password: "password123",
		Role:     models.RoleManager,
	}

	// Register and login users
	tech1Token := registerAndLogin(t, server, technician1)
	tech2Token := registerAndLogin(t, server, technician2)
	managerToken := registerAndLogin(t, server, manager)

	// Get the technician IDs
	var tech1ID, tech2ID int
	err := server.DB.QueryRow("SELECT id FROM users WHERE username = ?", technician1.Username).Scan(&tech1ID)
	assert.NoError(t, err)
	err = server.DB.QueryRow("SELECT id FROM users WHERE username = ?", technician2.Username).Scan(&tech2ID)
	assert.NoError(t, err)

	// Create a test task
	taskID := uuid.New().String()
	originalPerformedAt := time.Now().Add(-1 * time.Hour)
	_, err = server.DB.Exec(
		"INSERT INTO tasks (id, summary, performed_at, technician_id) VALUES (?, ?, ?, ?)",
		taskID,
		"Original task summary",
		originalPerformedAt.Format("2006-01-02 15:04:05"),
		tech1ID,
	)
	assert.NoError(t, err)

	t.Run("technician can update their own task", func(t *testing.T) {
		updatedTask := models.Task{
			Summary:     "Updated summary",
			PerformedAt: time.Now(),
		}

		taskJSON, _ := json.Marshal(updatedTask)
		req := httptest.NewRequest(http.MethodPut, "/tasks/"+taskID, bytes.NewBuffer(taskJSON))
		req.Header.Set("Authorization", "Bearer "+tech1Token)
		req.Header.Set("Content-Type", "application/json")

		rr := httptest.NewRecorder()
		server.Router.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)

		// Verify the update in database
		var dbTask struct {
			Summary     string
			PerformedAt time.Time
		}
		err := server.DB.QueryRow(
			"SELECT summary, performed_at FROM tasks WHERE id = ?",
			taskID,
		).Scan(&dbTask.Summary, &dbTask.PerformedAt)

		assert.NoError(t, err)
		assert.Equal(t, updatedTask.Summary, dbTask.Summary)
	})

	t.Run("technician cannot update another technician's task", func(t *testing.T) {
		updatedTask := models.Task{
			Summary:     "Unauthorized update attempt",
			PerformedAt: time.Now(),
		}

		taskJSON, _ := json.Marshal(updatedTask)
		req := httptest.NewRequest(http.MethodPut, "/tasks/"+taskID, bytes.NewBuffer(taskJSON))
		req.Header.Set("Authorization", "Bearer "+tech2Token)
		req.Header.Set("Content-Type", "application/json")

		rr := httptest.NewRecorder()
		server.Router.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusForbidden, rr.Code)
	})

	t.Run("rejects update with too long summary", func(t *testing.T) {
		// Create a summary that exceeds 2500 characters
		longSummary := make([]byte, 2501)
		for i := range longSummary {
			longSummary[i] = 'a'
		}

		updatedTask := models.Task{
			Summary:     string(longSummary),
			PerformedAt: time.Now(),
		}

		taskJSON, _ := json.Marshal(updatedTask)
		req := httptest.NewRequest(http.MethodPut, "/tasks/"+taskID, bytes.NewBuffer(taskJSON))
		req.Header.Set("Authorization", "Bearer "+tech1Token)
		req.Header.Set("Content-Type", "application/json")

		rr := httptest.NewRecorder()
		server.Router.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})

	t.Run("returns not found for non-existent task", func(t *testing.T) {
		nonExistentTaskID := uuid.New().String()
		updatedTask := models.Task{
			Summary:     "Update to non-existent task",
			PerformedAt: time.Now(),
		}

		taskJSON, _ := json.Marshal(updatedTask)
		req := httptest.NewRequest(http.MethodPut, "/tasks/"+nonExistentTaskID, bytes.NewBuffer(taskJSON))
		req.Header.Set("Authorization", "Bearer "+tech1Token)
		req.Header.Set("Content-Type", "application/json")

		rr := httptest.NewRecorder()
		server.Router.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusNotFound, rr.Code)
	})

	t.Run("rejects invalid JSON", func(t *testing.T) {
		invalidJSON := []byte(`{"summary": "Invalid JSON`)
		req := httptest.NewRequest(http.MethodPut, "/tasks/"+taskID, bytes.NewBuffer(invalidJSON))
		req.Header.Set("Authorization", "Bearer "+tech1Token)
		req.Header.Set("Content-Type", "application/json")

		rr := httptest.NewRecorder()
		server.Router.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})

	t.Run("unauthorized access returns 401", func(t *testing.T) {
		updatedTask := models.Task{
			Summary:     "Unauthorized update",
			PerformedAt: time.Now(),
		}

		taskJSON, _ := json.Marshal(updatedTask)
		req := httptest.NewRequest(http.MethodPut, "/tasks/"+taskID, bytes.NewBuffer(taskJSON))
		// No authorization header

		rr := httptest.NewRecorder()
		server.Router.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusUnauthorized, rr.Code)
	})

	t.Run("manager should not be able to update technician's task", func(t *testing.T) {
		updatedTask := models.Task{
			Summary:     "Manager update attempt",
			PerformedAt: time.Now(),
		}

		taskJSON, _ := json.Marshal(updatedTask)
		req := httptest.NewRequest(http.MethodPut, "/tasks/"+taskID, bytes.NewBuffer(taskJSON))
		req.Header.Set("Authorization", "Bearer "+managerToken)
		req.Header.Set("Content-Type", "application/json")

		rr := httptest.NewRecorder()
		server.Router.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusForbidden, rr.Code)

		// Verify the task wasn't updated
		var dbTask struct {
			Summary string
		}
		err := server.DB.QueryRow(
			"SELECT summary FROM tasks WHERE id = ?",
			taskID,
		).Scan(&dbTask.Summary)

		assert.NoError(t, err)
		assert.NotEqual(t, updatedTask.Summary, dbTask.Summary)
	})

	t.Run("invalid token returns 401", func(t *testing.T) {
		updatedTask := models.Task{
			Summary:     "Invalid token update",
			PerformedAt: time.Now(),
		}

		taskJSON, _ := json.Marshal(updatedTask)
		req := httptest.NewRequest(http.MethodPut, "/tasks/"+taskID, bytes.NewBuffer(taskJSON))
		req.Header.Set("Authorization", "Bearer invalid-token")
		req.Header.Set("Content-Type", "application/json")

		rr := httptest.NewRecorder()
		server.Router.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusUnauthorized, rr.Code)
	})
}
