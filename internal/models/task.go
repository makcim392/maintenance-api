package models

import (
	"time"
)

type Task struct {
	ID           string    `json:"id"`
	TechnicianID int64     `json:"technician_id"`
	Summary      string    `json:"summary"`
	PerformedAt  time.Time `json:"performed_at"`
}

type CreateTaskRequest struct {
	Summary     string    `json:"summary" validate:"required,max=2500"`
	PerformedAt time.Time `json:"performed_at" validate:"required"`
}
