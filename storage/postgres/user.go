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
)

type UserRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) storage.IUserStorage {
	return &UserRepository{db: db}
}

func (r *UserRepository) CreateUser(ctx context.Context, user models.User) (string, error) {
	user.ID = uuid.New().String()
	query := `
        INSERT INTO users (id, email, name, surname, role, password_hash, created_at, updated_at)
        VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`

	_, err := r.db.ExecContext(ctx, query,
		user.ID,
		user.Email,
		user.Name,
		user.Surname,
		user.Role,
		user.PasswordHash,
		time.Now(),
		time.Now(),
	)

	return user.ID, err
}

func (r *UserRepository) GetUserByID(ctx context.Context, id string) (models.User, error) {
	var user models.User
	query := `SELECT 
        id, email, name, surname, role, password_hash, 
        created_at, updated_at, deleted_at 
        FROM users 
        WHERE id = $1 AND deleted_at IS NULL`

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&user.ID,
		&user.Email,
		&user.Name,
		&user.Surname,
		&user.Role,
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
	query := `SELECT 
        id, email, name, surname, role, password_hash, 
        created_at, updated_at, deleted_at 
        FROM users 
        WHERE email = $1 AND deleted_at IS NULL`

	err := r.db.QueryRowContext(ctx, query, email).Scan(
		&user.ID,
		&user.Email,
		&user.Name,
		&user.Surname,
		&user.Role,
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
            name = $3,
            surname = $4,
            role = $5,
            password_hash = COALESCE($6, password_hash),
            updated_at = $7
        WHERE id = $1 AND deleted_at IS NULL`

	_, err := r.db.ExecContext(ctx, query,
		user.ID,
		user.Email,
		user.Name,
		user.Surname,
		user.Role,
		user.PasswordHash, // Agar parol yangilanmagan bo'lsa NULL bo'ladi
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
        SELECT id, email, name, surname, role, created_at, updated_at 
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
			&user.Name,
			&user.Surname,
			&user.Role,
			&user.CreatedAt,
			&user.UpdatedAt,
		); err != nil {
			return nil, err
		}
		users = append(users, user)
	}
	return users, nil
}

// Qo'shimcha metod: Role bo'yicha foydalanuvchilarni qidirish
func (r *UserRepository) ListUsersByRole(ctx context.Context, role models.Role, limit, offset int) ([]models.User, error) {
	query := `
        SELECT id, email, name, surname, role, created_at, updated_at 
        FROM users 
        WHERE deleted_at IS NULL AND role = $1
        LIMIT $2 OFFSET $3`

	rows, err := r.db.QueryContext(ctx, query, role, limit, offset)
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
			&user.Name,
			&user.Surname,
			&user.Role,
			&user.CreatedAt,
			&user.UpdatedAt,
		); err != nil {
			return nil, err
		}
		users = append(users, user)
	}
	return users, nil
}
