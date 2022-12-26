package models

type User struct {
	ID         int    `db:"id"`
	LastName   string `db:"last_name"`
	FirstName  string `db:"first_name"`
	MiddleName string `db:"middle_name"`
}
