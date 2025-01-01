package integration

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

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
