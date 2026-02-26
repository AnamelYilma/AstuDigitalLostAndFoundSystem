package models

import "time"

type User struct {
	ID       uint   `json:"id"`
	Email    string `json:"email"`
	Password string `json:"-"`
	Role     string `json:"role"` // "admin" or "user"
}

type Item struct {
	ID          uint      `json:"id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Type        string    `json:"type"` // "lost" or "found"
	Status      string    `json:"status"` // "open", "resolved"
	ReporterID  uint      `json:"reporter_id"`
	CreatedAt   time.Time `json:"created_at"`
}
