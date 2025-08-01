package models

import "database/sql"

type Staff struct {
	ID                   int            `json:"id,omitempty" db:"id,omitempty"`
	FirstName            string         `json:"first_name,omitempty" db:"first_name,omitempty"`
	LastName             string         `json:"last_name,omitempty" db:"last_name,omitempty"`
	Email                string         `json:"email,omitempty" db:"email,omitempty"`
	Username             string         `json:"username,omitempty" db:"username,omitempty"`
	Password             string         `json:"password,omitempty" db:"password,omitempty"`
	PasswordChangedAt    sql.NullString `json:"password_changed_at" db:"password_changed_at"`
	UserCreatedAt        sql.NullString `json:"user_created_at" db:"user_created_at"`
	PasswordResetToken   sql.NullString `json:"password_reset_token" db:"password_reset_token"`
	PasswordTokenExpires sql.NullString `json:"password_token_expires" db:"password_token_expires"`
	InactiveStatus       bool           `json:"inactive_status,omitempty" db:"inactive_status,omitempty"`
	Role                 string         `json:"role,omitempty" db:"role,omitempty"`
}

type UpdatePasswordRequest struct {
	CurrentPassword string `json:"current_password"`
	NewPassword     string `json:"new_password"`
}
type UpdatePasswordResponse struct {
	Token           string `json:"token"`
	PasswordUpdated string `json:"password_password"`
}
