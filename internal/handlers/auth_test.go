package handlers

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/makcim392/maintenance-api/internal/models"
	"github.com/stretchr/testify/assert"
	"golang.org/x/crypto/bcrypt"
)

func TestLogin(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create mock DB: %v", err)
	}
	defer db.Close()

	handler := NewAuthHandler(db)

	t.Run("successful login", func(t *testing.T) {
		// Create test password hash
		hashedPass, _ := bcrypt.GenerateFromPassword([]byte("correctpass"), bcrypt.DefaultCost)

		// Set up mock DB response
		rows := sqlmock.NewRows([]string{"id", "password", "role"}).
			AddRow(1, string(hashedPass), models.RoleTechnician)
		mock.ExpectQuery("SELECT id, password, role FROM users WHERE username = ?").
			WithArgs("testuser").
			WillReturnRows(rows)

		// Create request with valid credentials
		reqBody := LoginRequest{
			Username: "testuser",
			Password: "correctpass",
			Role:     models.RoleTechnician,
		}
		body, _ := json.Marshal(reqBody)
		req := httptest.NewRequest(http.MethodPost, "/login", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		handler.Login(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]string
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Contains(t, response, "token")
		assert.NotEmpty(t, response["token"])
	})

	t.Run("invalid credentials - wrong password", func(t *testing.T) {
		hashedPass, _ := bcrypt.GenerateFromPassword([]byte("correctpass"), bcrypt.DefaultCost)

		rows := sqlmock.NewRows([]string{"id", "password", "role"}).
			AddRow(1, string(hashedPass), models.RoleTechnician)
		mock.ExpectQuery("SELECT id, password, role FROM users WHERE username = ?").
			WithArgs("testuser").
			WillReturnRows(rows)

		reqBody := LoginRequest{
			Username: "testuser",
			Password: "wrongpass",
			Role:     models.RoleTechnician,
		}
		body, _ := json.Marshal(reqBody)
		req := httptest.NewRequest(http.MethodPost, "/login", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		handler.Login(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
		assert.Contains(t, w.Body.String(), "Invalid credentials")
	})

	t.Run("user not found", func(t *testing.T) {
		mock.ExpectQuery("SELECT id, password, role FROM users WHERE username = ?").
			WithArgs("nonexistent").
			WillReturnError(sql.ErrNoRows)

		reqBody := LoginRequest{
			Username: "nonexistent",
			Password: "anypass",
			Role:     models.RoleTechnician,
		}
		body, _ := json.Marshal(reqBody)
		req := httptest.NewRequest(http.MethodPost, "/login", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		handler.Login(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
		assert.Contains(t, w.Body.String(), "Invalid credentials")
	})

	t.Run("invalid request body", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/login", bytes.NewBuffer([]byte("invalid json")))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		handler.Login(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "Invalid request body")
	})

	t.Run("database error", func(t *testing.T) {
		mock.ExpectQuery("SELECT id, password, role FROM users WHERE username = ?").
			WithArgs("testuser").
			WillReturnError(sql.ErrConnDone)

		reqBody := LoginRequest{
			Username: "testuser",
			Password: "anypass",
			Role:     models.RoleTechnician,
		}
		body, _ := json.Marshal(reqBody)
		req := httptest.NewRequest(http.MethodPost, "/login", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		handler.Login(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		assert.Contains(t, w.Body.String(), "Database error")
	})
}

func TestRegister(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create mock DB: %v", err)
	}
	defer db.Close()

	handler := NewAuthHandler(db)

	t.Run("successful registration - technician", func(t *testing.T) {
		mock.ExpectExec("INSERT INTO users").
			WithArgs("newuser", sqlmock.AnyArg(), models.RoleTechnician).
			WillReturnResult(sqlmock.NewResult(1, 1))

		reqBody := LoginRequest{
			Username: "newuser",
			Password: "newpass",
			Role:     models.RoleTechnician,
		}
		body, _ := json.Marshal(reqBody)
		req := httptest.NewRequest(http.MethodPost, "/register", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		handler.Register(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)

		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.Equal(t, float64(1), response["id"])
		assert.Equal(t, "newuser", response["username"])
		assert.Equal(t, string(models.RoleTechnician), response["role"])
	})

	t.Run("successful registration - manager", func(t *testing.T) {
		mock.ExpectExec("INSERT INTO users").
			WithArgs("manager", sqlmock.AnyArg(), models.RoleManager).
			WillReturnResult(sqlmock.NewResult(2, 1))

		reqBody := LoginRequest{
			Username: "manager",
			Password: "managerpass",
			Role:     models.RoleManager,
		}
		body, _ := json.Marshal(reqBody)
		req := httptest.NewRequest(http.MethodPost, "/register", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		handler.Register(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)
	})

	t.Run("invalid role", func(t *testing.T) {
		reqBody := LoginRequest{
			Username: "newuser",
			Password: "newpass",
			Role:     "invalid_role",
		}
		body, _ := json.Marshal(reqBody)
		req := httptest.NewRequest(http.MethodPost, "/register", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		handler.Register(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "Invalid request body")
	})

	t.Run("invalid request body", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/register", bytes.NewBuffer([]byte("invalid json")))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		handler.Register(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "Invalid request body")
	})

	t.Run("database error", func(t *testing.T) {
		mock.ExpectExec("INSERT INTO users").
			WithArgs("newuser", sqlmock.AnyArg(), models.RoleTechnician).
			WillReturnError(sql.ErrConnDone)

		reqBody := LoginRequest{
			Username: "newuser",
			Password: "newpass",
			Role:     models.RoleTechnician,
		}
		body, _ := json.Marshal(reqBody)
		req := httptest.NewRequest(http.MethodPost, "/register", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		handler.Register(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		assert.Contains(t, w.Body.String(), "Error creating user")
	})
}
