package models

type User struct {
	ID        int    `json:"id" db:"id"`
	LastName  string `json:"lastName" db:"last_name"`
	FirstName string `json:"firstName" db:"first_name"`
}

type Meeting struct {
	ID        int `json:"id" db:"id"`
	Manager   int `json:"manager" db:"manager"`
	StartTime int `json:"startTime" db:"start_at"`
	EndTime   int `json:"endTime" db:"end_at"`
	Client    int `json:"client" db:"client"`
}
