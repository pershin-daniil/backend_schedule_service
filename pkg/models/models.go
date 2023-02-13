package models

import (
	"time"
)

type UserRequest struct {
	ID        *int    `json:"id" db:"id"`
	LastName  *string `json:"lastName" db:"last_name"`
	FirstName *string `json:"firstName" db:"first_name"`
	Phone     *string `json:"phone" db:"phone"`
	Email     *string `json:"email" db:"email"`
}

type User struct {
	ID        int       `json:"id" db:"id"`
	LastName  string    `json:"lastName" db:"last_name"`
	FirstName string    `json:"firstName" db:"first_name"`
	Phone     string    `json:"phone" db:"phone"`
	Email     string    `json:"email" db:"email"`
	Deleted   bool      `json:"deleted" db:"deleted"`
	CreatedAt time.Time `json:"createdAt" db:"created_at"`
	UpdatedAt time.Time `json:"updatedAt" db:"updated_at"`
}
type MeetingRequest struct {
	ID        *int       `json:"id" db:"id"`
	Manager   *int       `json:"manager" db:"manager"`
	StartTime *time.Time `json:"startTime" db:"start_at"`
	EndTime   *time.Time `json:"endTime" db:"end_at"`
	Client    *int       `json:"client" db:"client"`
}
type Meeting struct {
	ID        int       `json:"id" db:"id"`
	Manager   int       `json:"manager" db:"manager"`
	StartTime time.Time `json:"startTime" db:"start_at"`
	EndTime   time.Time `json:"endTime" db:"end_at"`
	Client    int       `json:"client" db:"client"`
	CreatedAt time.Time `json:"createdAt" db:"created_at"`
	UpdatedAt time.Time `json:"updatedAt" db:"updated_at"`
}
