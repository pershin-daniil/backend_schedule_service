package models

import (
	"time"
)

type UserRequest struct {
	ID           *int    `json:"id" db:"id"`
	LastName     *string `json:"lastName" db:"last_name"`
	FirstName    *string `json:"firstName" db:"first_name"`
	Phone        *string `json:"phone" db:"phone"`
	Email        *string `json:"email" db:"email"`
	PasswordHash *string `json:"-" db:"password_hash"`
	Role         *string `json:"role" db:"role"`
	Password     *string `json:"password" db:"-"`
}

type User struct {
	ID           int       `json:"id" db:"id"`
	LastName     string    `json:"lastName" db:"last_name"`
	FirstName    string    `json:"firstName" db:"first_name"`
	Phone        string    `json:"phone" db:"phone"`
	Email        string    `json:"email" db:"email"`
	PasswordHash string    `json:"-" db:"password_hash"`
	Role         string    `json:"role" db:"role"`
	Deleted      bool      `json:"deleted" db:"deleted"`
	CreatedAt    time.Time `json:"createdAt" db:"created_at"`
	UpdatedAt    time.Time `json:"updatedAt" db:"updated_at"`
}
