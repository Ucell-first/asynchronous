// storage/postgres/user_repository.go
package postgres

import (
	models "asynchronous/model/db"
	"asynchronous/storage"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type UserRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) storage.IUserStorage {
	return &UserRepository{db: db}
}

func (r *UserRepository) CreateUser(ctx context.Context, user models.User) (string, error) {
	user.ID = uuid.New().String()
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.PasswordHash), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("failed to hash password: %w", err)
	}

	query := `
        INSERT INTO users (id, email, password_hash, created_at, updated_at)
        VALUES ($1, $2, $3, $4, $5)`

	_, err = r.db.ExecContext(ctx, query,
		user.ID,
		user.Email,
		string(hashedPassword),
		time.Now(),
		time.Now(),
	)

	return user.ID, err
}

func (r *UserRepository) GetUserByID(ctx context.Context, id string) (models.User, error) {
	var user models.User
	query := `SELECT id, email, password_hash, created_at, updated_at, deleted_at 
              FROM users WHERE id = $1 AND deleted_at IS NULL`

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&user.ID,
		&user.Email,
		&user.PasswordHash,
		&user.CreatedAt,
		&user.UpdatedAt,
		&user.DeletedAt,
	)

	if errors.Is(err, sql.ErrNoRows) {
		return models.User{}, fmt.Errorf("user not found")
	}
	return user, err
}

func (r *UserRepository) GetUserByEmail(ctx context.Context, email string) (models.User, error) {
	var user models.User
	query := `SELECT id, email, password_hash, created_at, updated_at, deleted_at 
              FROM users WHERE email = $1 AND deleted_at IS NULL`

	err := r.db.QueryRowContext(ctx, query, email).Scan(
		&user.ID,
		&user.Email,
		&user.PasswordHash,
		&user.CreatedAt,
		&user.UpdatedAt,
		&user.DeletedAt,
	)

	if errors.Is(err, sql.ErrNoRows) {
		return models.User{}, fmt.Errorf("user not found")
	}
	return user, err
}

func (r *UserRepository) UpdateUser(ctx context.Context, user models.User) error {
	query := `
        UPDATE users SET
            email = $2,
            password_hash = $3,
            updated_at = $4
        WHERE id = $1 AND deleted_at IS NULL`

	_, err := r.db.ExecContext(ctx, query,
		user.ID,
		user.Email,
		user.PasswordHash,
		time.Now(),
	)
	return err
}

func (r *UserRepository) DeleteUser(ctx context.Context, id string) error {
	query := `UPDATE users SET deleted_at = $1 WHERE id = $2`
	_, err := r.db.ExecContext(ctx, query, time.Now(), id)
	return err
}

func (r *UserRepository) ListUsers(ctx context.Context, limit, offset int) ([]models.User, error) {
	query := `
        SELECT id, email, created_at, updated_at 
        FROM users 
        WHERE deleted_at IS NULL
        LIMIT $1 OFFSET $2`

	rows, err := r.db.QueryContext(ctx, query, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []models.User
	for rows.Next() {
		var user models.User
		if err := rows.Scan(
			&user.ID,
			&user.Email,
			&user.CreatedAt,
			&user.UpdatedAt,
		); err != nil {
			return nil, err
		}
		users = append(users, user)
	}
	return users, nil
}
