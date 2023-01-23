package models

type User struct {
	ID        int    `json:"id" db:"user_id"`
	LastName  string `json:"lastName" db:"last_name"`
	FirstName string `json:"firstName" db:"first_name"`
}
