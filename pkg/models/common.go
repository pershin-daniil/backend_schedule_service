package models

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v4"
)

var ErrInvalidCredentials = errors.New("invalid credentials")

const (
	RoleCoach  = `coach`
	RoleClient = `client`
)

type UserNotify struct {
	UserID    int       `json:"userID" db:"users.id"`
	MeetingID int       `json:"meetingID" db:"id"`
	Notified  bool      `json:"notified" db:"notified"`
	LastName  string    `json:"lastName" db:"last_name"`
	FirstName string    `json:"firstName" db:"first_name"`
	StartAt   time.Time `json:"startAt" db:"start_at"`
}

type Claims struct {
	jwt.RegisteredClaims
	UserID int    `json:"userID"`
	Role   string `json:"role"`
}

type TokenResponse struct {
	Token string `json:"token"`
}
