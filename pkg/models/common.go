package models

import (
	"errors"

	"github.com/golang-jwt/jwt/v4"
)

var ErrInvalidCredentials = errors.New("invalid credentials")

const (
	RoleCoach  = `coach`
	RoleClient = `client`
)

type Claims struct {
	jwt.RegisteredClaims
	UserID int    `json:"userID"`
	Role   string `json:"role"`
}

type TokenResponse struct {
	Token string `json:"token"`
}
