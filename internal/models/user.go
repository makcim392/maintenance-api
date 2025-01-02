package models

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"
)

type Role string

const (
	RoleTechnician Role = "technician"
	RoleManager    Role = "manager"
)

// Value - Implementation for sql driver to save Role type to database
func (r Role) Value() (driver.Value, error) {
	return string(r), nil
}

// Scan - Implementation for sql driver to read Role type from database
func (r *Role) Scan(value interface{}) error {
	if value == nil {
		return errors.New("role cannot be null")
	}

	str, ok := value.(string)
	if !ok {
		bytes, ok := value.([]byte)
		if !ok {
			return errors.New("invalid role format")
		}
		str = string(bytes)
	}

	switch Role(str) {
	case RoleTechnician, RoleManager:
		*r = Role(str)
		return nil
	default:
		return errors.New("invalid role value")
	}
}

// UnmarshalJSON - Custom JSON unmarshaling for Role
func (r *Role) UnmarshalJSON(data []byte) error {
	var str string
	if err := json.Unmarshal(data, &str); err != nil {
		return err
	}

	switch Role(str) {
	case RoleTechnician, RoleManager:
		*r = Role(str)
		return nil
	default:
		return errors.New("invalid role value")
	}
}

// MarshalJSON - Custom JSON marshaling for Role
func (r Role) MarshalJSON() ([]byte, error) {
	return json.Marshal(string(r))
}

type User struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	Username  string    `json:"username" gorm:"unique"`
	Password  string    `json:"-"` // '-' prevents password from being shown in JSON
	Role      Role      `json:"role" gorm:"type:varchar(20)"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
