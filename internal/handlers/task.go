package handlers

import (
	"database/sql"
	"encoding/json"
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
