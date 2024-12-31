package models

import "time"

type Role string

const (
	RoleTechnician Role = "technician"
	RoleManager    Role = "manager"
)

type User struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	Username  string    `json:"username" gorm:"unique"`
	Password  string    `json:"-"` // '-' prevents password from being shown in JSON
	Role      Role      `json:"role"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
