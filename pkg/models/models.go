package models

type User struct {
	ID        int    `db:"user_id"`
	LastName  string `db:"last_name"`
	FirstName string `db:"first_name"`
}
