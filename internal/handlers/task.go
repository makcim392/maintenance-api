package handlers

import (
	"database/sql"
	"encoding/json"
	"github.com/google/uuid"
	"github.com/makcim392/swordhealth-interviewer/internal/models"
	"net/http"
	"strconv"
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

	// For now, let's hardcode technician ID (you'll get this from auth later)
	task.TechnicianID = 1

	taskID := uuid.New().String()

	// Insert into database
	query := `
        INSERT INTO tasks (id, technician_id, summary, performed_at)
        VALUES (?, ?, ?, ?)
    `
	result, err := h.db.Exec(query, taskID, task.TechnicianID, task.Summary, task.PerformedAt)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	id, err := result.LastInsertId()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	task.ID = strconv.FormatInt(id, 10)
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(task)
}
