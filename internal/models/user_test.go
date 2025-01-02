package models

import (
	"database/sql/driver"
	"encoding/json"
	"testing"
	"time"
)

func TestRoleValue(t *testing.T) {
	tests := []struct {
		name    string
		role    Role
		want    driver.Value
		wantErr bool
	}{
		{
			name:    "technician role",
			role:    RoleTechnician,
			want:    "technician",
			wantErr: false,
		},
		{
			name:    "manager role",
			role:    RoleManager,
			want:    "manager",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.role.Value()
			if (err != nil) != tt.wantErr {
				t.Errorf("Role.Value() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("Role.Value() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRoleScan(t *testing.T) {
	tests := []struct {
		name    string
		input   interface{}
		want    Role
		wantErr bool
	}{
		{
			name:    "valid technician string",
			input:   "technician",
			want:    RoleTechnician,
			wantErr: false,
		},
		{
			name:    "valid manager string",
			input:   "manager",
			want:    RoleManager,
			wantErr: false,
		},
		{
			name:    "valid technician bytes",
			input:   []byte("technician"),
			want:    RoleTechnician,
			wantErr: false,
		},
		{
			name:    "valid manager bytes",
			input:   []byte("manager"),
			want:    RoleManager,
			wantErr: false,
		},
		{
			name:    "invalid role string",
			input:   "invalid",
			want:    "",
			wantErr: true,
		},
		{
			name:    "nil input",
			input:   nil,
			want:    "",
			wantErr: true,
		},
		{
			name:    "invalid type",
			input:   123,
			want:    "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got Role
			err := got.Scan(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("Role.Scan() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && got != tt.want {
				t.Errorf("Role.Scan() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRoleJSON(t *testing.T) {
	tests := []struct {
		name    string
		role    Role
		json    string
		wantErr bool
	}{
		{
			name:    "marshal technician",
			role:    RoleTechnician,
			json:    `"technician"`,
			wantErr: false,
		},
		{
			name:    "marshal manager",
			role:    RoleManager,
			json:    `"manager"`,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test marshaling
			got, err := json.Marshal(tt.role)
			if (err != nil) != tt.wantErr {
				t.Errorf("Role.MarshalJSON() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if string(got) != tt.json {
				t.Errorf("Role.MarshalJSON() = %v, want %v", string(got), tt.json)
			}

			// Test unmarshaling
			var gotRole Role
			err = json.Unmarshal([]byte(tt.json), &gotRole)
			if (err != nil) != tt.wantErr {
				t.Errorf("Role.UnmarshalJSON() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotRole != tt.role {
				t.Errorf("Role.UnmarshalJSON() = %v, want %v", gotRole, tt.role)
			}
		})
	}

	// Test invalid JSON unmarshaling
	invalidTests := []struct {
		name    string
		json    string
		wantErr bool
	}{
		{
			name:    "invalid role value",
			json:    `"invalid"`,
			wantErr: true,
		},
		{
			name:    "invalid json format",
			json:    `{"role": "technician"}`,
			wantErr: true,
		},
		{
			name:    "empty string",
			json:    `""`,
			wantErr: true,
		},
	}

	for _, tt := range invalidTests {
		t.Run(tt.name, func(t *testing.T) {
			var role Role
			err := json.Unmarshal([]byte(tt.json), &role)
			if (err != nil) != tt.wantErr {
				t.Errorf("Role.UnmarshalJSON() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestUser(t *testing.T) {
	now := time.Now()
	user := User{
		ID:        1,
		Username:  "testuser",
		Password:  "password123",
		Role:      RoleTechnician,
		CreatedAt: now,
		UpdatedAt: now,
	}

	// Test JSON marshaling of User
	data, err := json.Marshal(user)
	if err != nil {
		t.Errorf("Failed to marshal User: %v", err)
	}

	// Verify password is not included in JSON
	var unmarshaled map[string]interface{}
	if err := json.Unmarshal(data, &unmarshaled); err != nil {
		t.Errorf("Failed to unmarshal User JSON: %v", err)
	}

	if _, exists := unmarshaled["password"]; exists {
		t.Error("Password field should not be present in JSON")
	}

	// Verify other fields are present
	expectedFields := []string{"id", "username", "role", "created_at", "updated_at"}
	for _, field := range expectedFields {
		if _, exists := unmarshaled[field]; !exists {
			t.Errorf("Expected field %s missing from JSON", field)
		}
	}
}
