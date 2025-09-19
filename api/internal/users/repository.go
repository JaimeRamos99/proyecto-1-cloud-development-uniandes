package users

import (
	"errors"
	"fmt"

	"proyecto1/root/internal/database"
)

type Repository struct {
	db *database.DB
}

func NewRepository(db *database.DB) *Repository {
	return &Repository{db: db}
}

// CreateUser creates a new user in the database
func (r *Repository) CreateUser(user *User) (*User, error) {
	query := `
		INSERT INTO users (first_name, last_name, email, password_hash, city, country)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id`

	var id int
	err := r.db.QueryRow(query, user.FirstName, user.LastName, user.Email,
		user.PasswordHash, user.City, user.Country).Scan(&id)

	if err != nil {
		// Check if it's a duplicate email error
		if err.Error() == `pq: duplicate key value violates unique constraint "users_email_key"` {
			return nil, errors.New("email already exists")
		}
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	user.ID = id
	return user, nil
}

// EmailExists checks if an email already exists in the database
func (r *Repository) EmailExists(email string) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM users WHERE email = $1)`

	var exists bool
	err := r.db.QueryRow(query, email).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check email existence: %w", err)
	}

	return exists, nil
}

// GetUserByEmail retrieves a user by email
func (r *Repository) GetUserByEmail(email string) (*User, error) {
	query := `
		SELECT id, first_name, last_name, email, password_hash, city, country
		FROM users 
		WHERE email = $1`

	var user User
	err := r.db.QueryRow(query, email).Scan(
		&user.ID, &user.FirstName, &user.LastName, &user.Email,
		&user.PasswordHash, &user.City, &user.Country,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to get user by email: %w", err)
	}

	return &user, nil
}

// GetUserByID retrieves a user by their ID
func (r *Repository) GetUserByID(id int) (*User, error) {
	query := `
		SELECT id, first_name, last_name, email, password_hash, city, country
		FROM users
		WHERE id = $1`

	var user User
	err := r.db.QueryRow(query, id).Scan(
		&user.ID, &user.FirstName, &user.LastName, &user.Email,
		&user.PasswordHash, &user.City, &user.Country,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get user by id: %w", err)
	}

	return &user, nil
}
