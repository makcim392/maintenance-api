package handlers

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/makcim392/maintenance-api/internal/auth"
	"github.com/makcim392/maintenance-api/internal/models"
	"golang.org/x/crypto/bcrypt"
)

type AuthHandler struct {
	db *sql.DB
}

func NewAuthHandler(db *sql.DB) *AuthHandler {
	return &AuthHandler{
		db: db,
	}
}

type LoginRequest struct {
	Username string      `json:"username"`
	Password string      `json:"password"`
	Role     models.Role `json:"role"`
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Query user from database
	var user models.User
	query := `SELECT id, password, role FROM users WHERE username = ?`
	err := h.db.QueryRow(query, req.Username).Scan(&user.ID, &user.Password, &user.Role)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			http.Error(w, "Invalid credentials", http.StatusUnauthorized)
			return
		}
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	// Verify password
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	// Generate token
	token, err := auth.GenerateToken(user.ID, string(user.Role))
	if err != nil {
		http.Error(w, "Error generating token", http.StatusInternalServerError)
		return
	}

	// Return token
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"token": token,
	})
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate role
	if req.Role != models.RoleTechnician && req.Role != models.RoleManager {
		http.Error(w, "Invalid role. Must be either 'technician' or 'manager'", http.StatusBadRequest)
		return
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		http.Error(w, "Error processing request", http.StatusInternalServerError)
		return
	}

	// Insert user into database
	query := `
        INSERT INTO users (username, password, role)
        VALUES (?, ?, ?)
    `
	result, err := h.db.Exec(query, req.Username, hashedPassword, req.Role)
	if err != nil {
		http.Error(w, "Error creating user", http.StatusInternalServerError)
		return
	}

	id, _ := result.LastInsertId()
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"id":       id,
		"username": req.Username,
		"role":     req.Role,
	})
}
