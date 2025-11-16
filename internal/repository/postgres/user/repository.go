package user

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"mor80/service-reviewer/internal/model"
)

type UserRepository struct {
	pool *pgxpool.Pool
}

func New(pool *pgxpool.Pool) *UserRepository {
	return &UserRepository{pool: pool}
}

func (r *UserRepository) GetByID(ctx context.Context, userID string) (*model.User, error) {
	const query = `
		SELECT user_id, username, team_name, is_active
		FROM users
		WHERE user_id = $1
	`

	row := r.pool.QueryRow(ctx, query, userID)

	user, err := scanUser(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, model.ErrNotFound
		}

		return nil, fmt.Errorf("database error: %w", err)
	}

	return user, nil
}

func (r *UserRepository) ListByTeam(ctx context.Context, teamName string) ([]model.User, error) {
	const query = `
		SELECT user_id, username, team_name, is_active
		FROM users
		WHERE team_name = $1
		ORDER BY user_id
	`

	rows, err := r.pool.Query(ctx, query, teamName)
	if err != nil {
		return nil, fmt.Errorf("database error: %w", err)
	}
	defer rows.Close()

	var users []model.User
	for rows.Next() {
		user, err := scanUser(rows)
		if err != nil {
			return nil, fmt.Errorf("database error: %w", err)
		}

		users = append(users, *user)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("database error: %w", err)
	}

	return users, nil
}

func (r *UserRepository) Upsert(ctx context.Context, users []model.User) error {
	if len(users) == 0 {
		return nil
	}

	const query = `
		INSERT INTO users (user_id, username, team_name, is_active)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (user_id) DO UPDATE
		SET
			username = EXCLUDED.username,
			team_name = EXCLUDED.team_name,
			is_active = EXCLUDED.is_active
	`

	tx, err := r.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return fmt.Errorf("database error: %w", err)
	}

	for _, user := range users {
		if _, err := tx.Exec(ctx, query, user.ID, user.Username, user.TeamName, user.IsActive); err != nil {
			_ = tx.Rollback(ctx)
			return fmt.Errorf("database error: %w", err)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		_ = tx.Rollback(ctx)
		return fmt.Errorf("database error: %w", err)
	}

	return nil
}

func (r *UserRepository) SetIsActive(ctx context.Context, userID string, isActive bool) (*model.User, error) {
	const query = `
		UPDATE users
		SET is_active = $2
		WHERE user_id = $1
		RETURNING user_id, username, team_name, is_active
	`

	row := r.pool.QueryRow(ctx, query, userID, isActive)

	user, err := scanUser(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, model.ErrNotFound
		}

		return nil, fmt.Errorf("database error: %w", err)
	}

	return user, nil
}

func (r *UserRepository) ListByIDs(ctx context.Context, teamName string, userIDs []string) ([]model.User, error) {
	if len(userIDs) == 0 {
		return nil, nil
	}

	const query = `
		SELECT user_id, username, team_name, is_active
		FROM users
		WHERE team_name = $1 AND user_id = ANY($2)
	`

	rows, err := r.pool.Query(ctx, query, teamName, userIDs)
	if err != nil {
		return nil, fmt.Errorf("database error: %w", err)
	}
	defer rows.Close()

	var users []model.User
	for rows.Next() {
		user, scanErr := scanUser(rows)
		if scanErr != nil {
			return nil, fmt.Errorf("database error: %w", scanErr)
		}

		users = append(users, *user)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("database error: %w", err)
	}

	return users, nil
}

func (r *UserRepository) DeactivateUsers(ctx context.Context, teamName string, userIDs []string) ([]string, error) {
	if len(userIDs) == 0 {
		return nil, nil
	}

	const query = `
		UPDATE users
		SET is_active = FALSE
		WHERE team_name = $1 AND user_id = ANY($2)
		RETURNING user_id
	`

	rows, err := r.pool.Query(ctx, query, teamName, userIDs)
	if err != nil {
		return nil, fmt.Errorf("database error: %w", err)
	}
	defer rows.Close()

	var updated []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, fmt.Errorf("database error: %w", err)
		}

		updated = append(updated, id)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("database error: %w", err)
	}

	return updated, nil
}

type scanner interface {
	Scan(dest ...any) error
}

func scanUser(row scanner) (*model.User, error) {
	var user model.User

	if err := row.Scan(
		&user.ID,
		&user.Username,
		&user.TeamName,
		&user.IsActive,
	); err != nil {
		return nil, err
	}

	return &user, nil
}
