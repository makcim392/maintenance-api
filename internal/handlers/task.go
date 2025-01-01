package handlers

import (
	"database/sql"
	"encoding/json"
	"github.com/gorilla/mux"
	"github.com/makcim392/swordhealth-interviewer/internal/middleware"
	"log"
	"net/http"

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

	// For now, let's hardcode technician ID (you'll get this from auth later)
	task.TechnicianID = 1

	task.ID = uuid.New().String()

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
	if err == sql.ErrNoRows {
		http.Error(w, "Task not found", http.StatusNotFound)
		return
	} else if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
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
        UPDATE tasks 
        SET summary = ?, performed_at = ?
        WHERE id = ? AND (? = ? OR technician_id = ?)
    `
	result, err := h.db.Exec(query, task.Summary, task.PerformedAt, taskID, role, models.RoleManager, userID)
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
