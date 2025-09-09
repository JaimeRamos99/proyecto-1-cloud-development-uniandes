package users

// User represents the user model based on the database schema
type User struct {
	ID           int    `json:"id" db:"id"`
	FirstName    string `json:"first_name" db:"first_name"`
	LastName     string `json:"last_name" db:"last_name"`
	Email        string `json:"email" db:"email"`
	PasswordHash string `json:"-" db:"password_hash"`
	City         string `json:"city" db:"city"`
	Country      string `json:"country" db:"country"`
}