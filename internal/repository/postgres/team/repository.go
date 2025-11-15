package team

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"mor80/service-reviewer/internal/model"
)

type TeamRepository struct {
	pool *pgxpool.Pool
}

func New(pool *pgxpool.Pool) *TeamRepository {
	return &TeamRepository{pool: pool}
}

func (r *TeamRepository) Create(ctx context.Context, teamName string) error {
	const query = `
		INSERT INTO teams (team_name)
		VALUES ($1)
	`

	if _, err := r.pool.Exec(ctx, query, teamName); err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return model.ErrTeamExists
		}

		return fmt.Errorf("database error: %w", err)
	}

	return nil
}

func (r *TeamRepository) Exists(ctx context.Context, teamName string) (bool, error) {
	const query = `
		SELECT 1
		FROM teams
		WHERE team_name = $1
	`

	var exists int
	if err := r.pool.QueryRow(ctx, query, teamName).Scan(&exists); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return false, nil
		}

		return false, fmt.Errorf("database error: %w", err)
	}

	return true, nil
}

func (r *TeamRepository) GetByName(ctx context.Context, teamName string) (*model.Team, error) {
	const queryTeam = `
		SELECT team_name
		FROM teams
		WHERE team_name = $1
	`

	var name string
	if err := r.pool.QueryRow(ctx, queryTeam, teamName).Scan(&name); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, model.ErrNotFound
		}

		return nil, fmt.Errorf("database error: %w", err)
	}

	const queryMembers = `
		SELECT user_id, username, is_active
		FROM users
		WHERE team_name = $1
		ORDER BY user_id
	`

	rows, err := r.pool.Query(ctx, queryMembers, teamName)
	if err != nil {
		return nil, fmt.Errorf("database error: %w", err)
	}
	defer rows.Close()

	team := &model.Team{
		Name: name,
	}

	for rows.Next() {
		member, err := scanTeamMember(rows)
		if err != nil {
			return nil, fmt.Errorf("database error: %w", err)
		}

		team.Members = append(team.Members, member)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("database error: %w", err)
	}

	return team, nil
}

type memberScanner interface {
	Scan(dest ...any) error
}

func scanTeamMember(row memberScanner) (model.TeamMember, error) {
	var member model.TeamMember

	if err := row.Scan(
		&member.ID,
		&member.Username,
		&member.IsActive,
	); err != nil {
		return model.TeamMember{}, err
	}

	return member, nil
}
