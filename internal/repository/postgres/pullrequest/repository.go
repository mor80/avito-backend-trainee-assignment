package pullrequest

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"mor80/service-reviewer/internal/model"
)

type PullRequestRepository struct {
	pool *pgxpool.Pool
}

func New(pool *pgxpool.Pool) *PullRequestRepository {
	return &PullRequestRepository{pool: pool}
}

func (r *PullRequestRepository) Create(ctx context.Context, pr model.PullRequestDB, reviewerIDs []string) (*model.PullRequest, error) {
	tx, err := r.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return nil, fmt.Errorf("database error: %w", err)
	}

	prQuery := `
		INSERT INTO pull_requests (pull_request_id, pull_request_name, author_id, status, created_at, merged_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`

	if _, err := tx.Exec(ctx, prQuery, pr.ID, pr.Name, pr.AuthorID, pr.Status, pr.CreatedAt, pr.MergedAt); err != nil {
		_ = tx.Rollback(ctx)

		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return nil, model.ErrPRExists
		}

		return nil, fmt.Errorf("database error: %w", err)
	}

	reviewersQuery := `
		INSERT INTO pull_request_reviewers (pull_request_id, reviewer_id)
		VALUES ($1, $2)
	`

	for _, reviewerID := range reviewerIDs {
		if _, err := tx.Exec(ctx, reviewersQuery, pr.ID, reviewerID); err != nil {
			_ = tx.Rollback(ctx)
			return nil, fmt.Errorf("database error: %w", err)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		_ = tx.Rollback(ctx)
		return nil, fmt.Errorf("database error: %w", err)
	}

	return r.GetByID(ctx, pr.ID)
}

func (r *PullRequestRepository) GetByID(ctx context.Context, prID string) (*model.PullRequest, error) {
	const prQuery = `
		SELECT pull_request_id, pull_request_name, author_id, status, created_at, merged_at
		FROM pull_requests
		WHERE pull_request_id = $1
	`

	pr, err := scanPullRequest(r.pool.QueryRow(ctx, prQuery, prID))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, model.ErrNotFound
		}

		return nil, fmt.Errorf("database error: %w", err)
	}

	reviewers, err := r.getReviewers(ctx, prID)
	if err != nil {
		return nil, fmt.Errorf("database error: %w", err)
	}
	pr.AssignedReviewers = reviewers

	return pr, nil
}

func (r *PullRequestRepository) UpdateStatus(ctx context.Context, prID string, status model.PullRequestStatus, mergedAt *time.Time) (*model.PullRequest, error) {
	if err := status.Validate(); err != nil {
		return nil, err
	}

	const query = `
		UPDATE pull_requests
		SET status = $2, merged_at = $3
		WHERE pull_request_id = $1
		RETURNING pull_request_id, pull_request_name, author_id, status, created_at, merged_at
	`

	pr, err := scanPullRequest(r.pool.QueryRow(ctx, query, prID, status, mergedAt))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, model.ErrNotFound
		}

		return nil, fmt.Errorf("database error: %w", err)
	}

	reviewers, err := r.getReviewers(ctx, prID)
	if err != nil {
		return nil, fmt.Errorf("database error: %w", err)
	}
	pr.AssignedReviewers = reviewers

	return pr, nil
}

func (r *PullRequestRepository) ReplaceReviewer(ctx context.Context, prID, oldReviewerID, newReviewerID string) (*model.PullRequest, error) {
	tx, err := r.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return nil, fmt.Errorf("database error: %w", err)
	}

	const deleteQuery = `
		DELETE FROM pull_request_reviewers
		WHERE pull_request_id = $1 AND reviewer_id = $2
	`

	tag, err := tx.Exec(ctx, deleteQuery, prID, oldReviewerID)
	if err != nil {
		_ = tx.Rollback(ctx)
		return nil, fmt.Errorf("database error: %w", err)
	}

	if tag.RowsAffected() == 0 {
		_ = tx.Rollback(ctx)
		return nil, model.ErrNotAssigned
	}

	const insertQuery = `
		INSERT INTO pull_request_reviewers (pull_request_id, reviewer_id)
		VALUES ($1, $2)
	`

	if _, err := tx.Exec(ctx, insertQuery, prID, newReviewerID); err != nil {
		_ = tx.Rollback(ctx)
		return nil, fmt.Errorf("database error: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		_ = tx.Rollback(ctx)
		return nil, fmt.Errorf("database error: %w", err)
	}

	return r.GetByID(ctx, prID)
}

func (r *PullRequestRepository) ListByReviewer(ctx context.Context, reviewerID string) ([]model.PullRequestShort, error) {
	const query = `
		SELECT pr.pull_request_id, pr.pull_request_name, pr.author_id, pr.status
		FROM pull_requests pr
		JOIN pull_request_reviewers prr ON pr.pull_request_id = prr.pull_request_id
		WHERE prr.reviewer_id = $1
		ORDER BY pr.created_at DESC
	`

	rows, err := r.pool.Query(ctx, query, reviewerID)
	if err != nil {
		return nil, fmt.Errorf("database error: %w", err)
	}
	defer rows.Close()

	var prs []model.PullRequestShort

	for rows.Next() {
		pr, scanErr := scanPullRequestShort(rows)
		if scanErr != nil {
			return nil, fmt.Errorf("database error: %w", err)
		}

		prs = append(prs, pr)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("database error: %w", err)
	}

	return prs, nil
}

func (r *PullRequestRepository) getReviewers(ctx context.Context, prID string) ([]string, error) {
	const query = `
		SELECT reviewer_id
		FROM pull_request_reviewers
		WHERE pull_request_id = $1
		ORDER BY reviewer_id
	`

	rows, err := r.pool.Query(ctx, query, prID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var reviewers []string

	for rows.Next() {
		var reviewerID string
		if err := rows.Scan(&reviewerID); err != nil {
			return nil, err
		}

		reviewers = append(reviewers, reviewerID)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return reviewers, nil
}

type pullRequestScanner interface {
	Scan(dest ...any) error
}

func scanPullRequest(row pullRequestScanner) (*model.PullRequest, error) {
	var pr model.PullRequest

	if err := row.Scan(
		&pr.ID,
		&pr.Name,
		&pr.AuthorID,
		&pr.Status,
		&pr.CreatedAt,
		&pr.MergedAt,
	); err != nil {
		return nil, err
	}

	return &pr, nil
}

func scanPullRequestShort(row pullRequestScanner) (model.PullRequestShort, error) {
	var pr model.PullRequestShort

	if err := row.Scan(
		&pr.ID,
		&pr.Name,
		&pr.AuthorID,
		&pr.Status,
	); err != nil {
		return model.PullRequestShort{}, err
	}

	return pr, nil
}
