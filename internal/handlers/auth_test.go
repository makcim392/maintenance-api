package handlers

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"golang.org/x/crypto/bcrypt"
)

func TestLogin(t *testing.T) {
	// Create a mock database
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
			AddRow(1, string(hashedPass), "technician")
		mock.ExpectQuery("SELECT id, password, role FROM users WHERE username = ?").
			WithArgs("testuser").
			WillReturnRows(rows)

		// Create request
		reqBody := LoginRequest{
			Username: "testuser",
			Password: "correctpass",
		}
		body, _ := json.Marshal(reqBody)
		req := httptest.NewRequest(http.MethodPost, "/login", bytes.NewBuffer(body))
		w := httptest.NewRecorder()

		// Call handler
		handler.Login(w, req)

		// Assert response
		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]string
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.Contains(t, response, "token")
		assert.NotEmpty(t, response["token"])
	})

	t.Run("invalid credentials - wrong password", func(t *testing.T) {
		hashedPass, _ := bcrypt.GenerateFromPassword([]byte("correctpass"), bcrypt.DefaultCost)

		rows := sqlmock.NewRows([]string{"id", "password", "role"}).
			AddRow(1, string(hashedPass), "technician")
		mock.ExpectQuery("SELECT id, password, role FROM users WHERE username = ?").
			WithArgs("testuser").
			WillReturnRows(rows)

		reqBody := LoginRequest{
			Username: "testuser",
			Password: "wrongpass",
		}
		body, _ := json.Marshal(reqBody)
		req := httptest.NewRequest(http.MethodPost, "/login", bytes.NewBuffer(body))
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
		}
		body, _ := json.Marshal(reqBody)
		req := httptest.NewRequest(http.MethodPost, "/login", bytes.NewBuffer(body))
		w := httptest.NewRecorder()

		handler.Login(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
		assert.Contains(t, w.Body.String(), "Invalid credentials")
	})

	t.Run("invalid request body", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/login", bytes.NewBuffer([]byte("invalid json")))
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
		}
		body, _ := json.Marshal(reqBody)
		req := httptest.NewRequest(http.MethodPost, "/login", bytes.NewBuffer(body))
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

	t.Run("successful registration", func(t *testing.T) {
		mock.ExpectExec("INSERT INTO users").
			WithArgs("newuser", sqlmock.AnyArg(), "technician").
			WillReturnResult(sqlmock.NewResult(1, 1))

		reqBody := LoginRequest{
			Username: "newuser",
			Password: "newpass",
		}
		body, _ := json.Marshal(reqBody)
		req := httptest.NewRequest(http.MethodPost, "/register", bytes.NewBuffer(body))
		w := httptest.NewRecorder()

		handler.Register(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)

		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.Equal(t, float64(1), response["id"])
		assert.Equal(t, "newuser", response["username"])
		assert.Equal(t, "technician", response["role"])
	})

	t.Run("invalid request body", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/register", bytes.NewBuffer([]byte("invalid json")))
		w := httptest.NewRecorder()

		handler.Register(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "Invalid request body")
	})

	t.Run("database error", func(t *testing.T) {
		mock.ExpectExec("INSERT INTO users").
			WithArgs("newuser", sqlmock.AnyArg(), "technician").
			WillReturnError(sql.ErrConnDone)

		reqBody := LoginRequest{
			Username: "newuser",
			Password: "newpass",
		}
		body, _ := json.Marshal(reqBody)
		req := httptest.NewRequest(http.MethodPost, "/register", bytes.NewBuffer(body))
		w := httptest.NewRecorder()

		handler.Register(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		assert.Contains(t, w.Body.String(), "Error creating user")
	})
}
