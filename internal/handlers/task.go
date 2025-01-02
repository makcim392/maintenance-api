package handlers

import (
	"database/sql"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/makcim392/swordhealth-interviewer/internal/middleware"

	"github.com/google/uuid"
	"github.com/makcim392/swordhealth-interviewer/internal/models"
)

type TaskHandler struct {
	db *sql.DB
}

func NewTaskHandler(db *sql.DB) *TaskHandler {
	return &TaskHandler{
		db: db,
	}
}

func (h *TaskHandler) CreateTask(w http.ResponseWriter, r *http.Request) {
	var task models.Task
	if err := json.NewDecoder(r.Body).Decode(&task); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Validate summary length
	if len(task.Summary) > 2500 {
		http.Error(w, "Summary must not exceed 2500 characters", http.StatusBadRequest)
		return
	}

	// Validate PerformedAt is not zero
	if task.PerformedAt.IsZero() {
		http.Error(w, "PerformedAt is required", http.StatusBadRequest)
		return
	}

	task.ID = uuid.New().String()

	// Get user information from context using your existing context keys
	userID, ok := r.Context().Value(middleware.UserIDContextKey).(int)
	if !ok {
		http.Error(w, "Unable to get user ID from context", http.StatusInternalServerError)
		return
	}

	role, ok := r.Context().Value(middleware.RoleContextKey).(string)
	if !ok {
		http.Error(w, "Unable to get role from context", http.StatusInternalServerError)
		return
	}

	if role != string(models.RoleTechnician) {
		http.Error(w, "Unauthorized to create task", http.StatusForbidden)
		return
	}

	task.TechnicianID = int64(userID)

	// Insert into database
	query := `
        INSERT INTO tasks (id, technician_id, summary, performed_at)
        VALUES (?, ?, ?, ?)
    `
	_, err := h.db.Exec(query, task.ID, task.TechnicianID, task.Summary, task.PerformedAt)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	err = json.NewEncoder(w).Encode(task)
	if err != nil {
		log.Printf("Error encoding task: %v", err)
	}
}

func (h *TaskHandler) UpdateTask(w http.ResponseWriter, r *http.Request) {
	value := r.Context().Value(middleware.UserIDContextKey)
	log.Printf("Value in context: %v, Type: %T", value, value)

	// Get user information from context using your existing context keys
	userID, ok := r.Context().Value(middleware.UserIDContextKey).(int)
	if !ok {
		http.Error(w, "Unable to get user ID from context", http.StatusInternalServerError)
		return
	}

	role, ok := r.Context().Value(middleware.RoleContextKey).(string)
	if !ok {
		http.Error(w, "Unable to get role from context", http.StatusInternalServerError)
		return
	}

	vars := mux.Vars(r)
	taskID := vars["id"]

	// First, check if task exists and get current technician ID
	var currentTechID int
	err := h.db.QueryRow("SELECT technician_id FROM tasks WHERE id = ?", taskID).Scan(&currentTechID)
	if errors.Is(err, sql.ErrNoRows) {
		http.Error(w, "Task not found", http.StatusNotFound)
		return
	} else if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if userID != currentTechID {
		http.Error(w, "Unauthorized to modify this task", http.StatusForbidden)
		return
	}

	// Check authorization using your existing role constants
	if role != string(models.RoleManager) && userID != currentTechID {
		http.Error(w, "Unauthorized to modify this task", http.StatusForbidden)
		return
	}

	var task models.Task
	if err := json.NewDecoder(r.Body).Decode(&task); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if len(task.Summary) > 2500 {
		http.Error(w, "Summary must not exceed 2500 characters", http.StatusBadRequest)
		return
	}

	query := `
        UPDATE tasks SET summary = ?, performed_at = ?
		WHERE 
		id = ? AND technician_id = ?
    `
	result, err := h.db.Exec(query, task.Summary, task.PerformedAt, taskID, userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if rowsAffected == 0 {
		http.Error(w, "Task not found or unauthorized", http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Task updated successfully",
		"id":      taskID,
	})
}

func (h *TaskHandler) ListTasks(w http.ResponseWriter, r *http.Request) {
	// Get user information from context using your existing context keys
	userID, ok := r.Context().Value(middleware.UserIDContextKey).(int)
	if !ok {
		http.Error(w, "Unable to get user ID from context", http.StatusInternalServerError)
		return
	}

	role, ok := r.Context().Value(middleware.RoleContextKey).(string)
	if !ok {
		http.Error(w, "Unable to get role from context", http.StatusInternalServerError)
		return
	}

	var query string
	var args []interface{}

	// Build query based on user role with DATE_FORMAT
	if role == string(models.RoleTechnician) {
		query = `
            SELECT t.id, t.summary, 
            DATE_FORMAT(t.performed_at, '%Y-%m-%d %H:%i:%s') as performed_at, 
            t.technician_id, u.username
            FROM tasks t
            JOIN users u ON t.technician_id = u.id
            WHERE t.technician_id = ?
            ORDER BY t.performed_at DESC`
		args = append(args, userID)
	} else if role == string(models.RoleManager) {
		query = `
            SELECT t.id, t.summary, 
            DATE_FORMAT(t.performed_at, '%Y-%m-%d %H:%i:%s') as performed_at, 
            t.technician_id, u.username
            FROM tasks t
            JOIN users u ON t.technician_id = u.id
            ORDER BY t.performed_at DESC`
	} else {
		http.Error(w, "Unauthorized role", http.StatusForbidden)
		return
	}

	// Execute query
	rows, err := h.db.Query(query, args...)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	type TaskResponse struct {
		ID           string    `json:"id"`
		Summary      string    `json:"summary"`
		PerformedAt  time.Time `json:"performed_at"`
		TechnicianID int64     `json:"technician_id"`
		Username     string    `json:"technician_name"`
	}

	var tasks []TaskResponse

	// Iterate through results
	for rows.Next() {
		var task TaskResponse
		var performedAtStr string // Change to string to receive the formatted date

		err := rows.Scan(
			&task.ID,
			&task.Summary,
			&performedAtStr,
			&task.TechnicianID,
			&task.Username,
		)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// Parse the formatted date string
		parsedTime, err := time.Parse("2006-01-02 15:04:05", performedAtStr)
		if err != nil {
			http.Error(w, "Error parsing date", http.StatusInternalServerError)
			return
		}
		task.PerformedAt = parsedTime

		tasks = append(tasks, task)
	}

	if err = rows.Err(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Return response
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(tasks); err != nil {
		log.Printf("Error encoding tasks: %v", err)
	}
}
