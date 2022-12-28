package models

type User struct {
	ID          int    `db:"id"`
	LastName    string `db:"last_name"`
	FirstName   string `db:"first_name"`
	PhoneNumber int    `db:"phone_number"`
}
