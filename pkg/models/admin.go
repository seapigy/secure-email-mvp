package models

import (
	"time"
)

// AdminUser represents an admin user in the system
type AdminUser struct {
	ID        int64     `json:"id"`
	Email     string    `json:"email"`
	Password  string    `json:"-"` // hashed
	Role      string    `json:"role"` // must equal "admin"
	CreatedAt time.Time `json:"created_at"`
}






