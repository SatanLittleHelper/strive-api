package repositories

import (
	"context"
	"fmt"

	"github.com/aleksandr/strive-api/internal/models"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type UserRepository interface {
	Create(ctx context.Context, user *models.User) error
	GetByID(ctx context.Context, id uuid.UUID) (*models.User, error)
	GetByEmail(ctx context.Context, email string) (*models.User, error)
	Update(ctx context.Context, user *models.User) error
	UpdateTheme(ctx context.Context, userID uuid.UUID, theme string) error
	Delete(ctx context.Context, id uuid.UUID) error
}

type userRepository struct {
	pool *pgxpool.Pool
}

func NewUserRepository(pool *pgxpool.Pool) UserRepository {
	return &userRepository{
		pool: pool,
	}
}

func (r *userRepository) Create(ctx context.Context, user *models.User) error {
	query := `
		INSERT INTO users (id, email, password_hash, theme, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`

	_, err := r.pool.Exec(ctx, query, user.ID, user.Email, user.PasswordHash, user.Theme, user.CreatedAt, user.UpdatedAt)
	if err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}

	return nil
}

func (r *userRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.User, error) {
	query := `
		SELECT id, email, password_hash, theme, created_at, updated_at
		FROM users
		WHERE id = $1
	`

	row := r.pool.QueryRow(ctx, query, id)
	return r.scanUser(row, "failed to get user by id")
}

func (r *userRepository) GetByEmail(ctx context.Context, email string) (*models.User, error) {
	query := `
		SELECT id, email, password_hash, theme, created_at, updated_at
		FROM users
		WHERE email = $1
	`

	row := r.pool.QueryRow(ctx, query, email)
	return r.scanUser(row, "failed to get user by email")
}

func (r *userRepository) Update(ctx context.Context, user *models.User) error {
	query := `
		UPDATE users
		SET email = $2, password_hash = $3, theme = $4, updated_at = $5
		WHERE id = $1
	`

	_, err := r.pool.Exec(ctx, query, user.ID, user.Email, user.PasswordHash, user.Theme, user.UpdatedAt)
	if err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}

	return nil
}

func (r *userRepository) UpdateTheme(ctx context.Context, userID uuid.UUID, theme string) error {
	query := `
		UPDATE users
		SET theme = $2, updated_at = NOW()
		WHERE id = $1
	`

	_, err := r.pool.Exec(ctx, query, userID, theme)
	if err != nil {
		return fmt.Errorf("failed to update user theme: %w", err)
	}

	return nil
}

func (r *userRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM users WHERE id = $1`

	_, err := r.pool.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}

	return nil
}

func (r *userRepository) scanUser(row interface{ Scan(...interface{}) error }, errorPrefix string) (*models.User, error) {
	user := &models.User{}
	err := row.Scan(
		&user.ID,
		&user.Email,
		&user.PasswordHash,
		&user.Theme,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", errorPrefix, err)
	}

	return user, nil
}
