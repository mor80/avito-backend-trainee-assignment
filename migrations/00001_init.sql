-- +goose Up
-- +goose StatementBegin
-- Teams
CREATE TABLE teams (
    team_name VARCHAR(255) PRIMARY KEY
);

-- Users
CREATE TABLE users (
    user_id   VARCHAR(255) PRIMARY KEY,
    username  VARCHAR(255) NOT NULL,
    is_active BOOLEAN      NOT NULL DEFAULT TRUE,
    team_name VARCHAR(255) NOT NULL,
    CONSTRAINT fk_users_team
        FOREIGN KEY (team_name)
        REFERENCES teams(team_name)
        ON UPDATE CASCADE
        ON DELETE RESTRICT
);

CREATE INDEX idx_users_team_name ON users(team_name);

-- Pull Requests
CREATE TABLE pull_requests (
    pull_request_id   VARCHAR(255) PRIMARY KEY,
    pull_request_name TEXT        NOT NULL,
    author_id         VARCHAR(255) NOT NULL,
    status            VARCHAR(16)  NOT NULL,
    created_at        TIMESTAMPTZ  NULL DEFAULT NOW(),
    merged_at         TIMESTAMPTZ  NULL,
    CONSTRAINT fk_pr_author
        FOREIGN KEY (author_id)
        REFERENCES users(user_id)
        ON UPDATE CASCADE
        ON DELETE RESTRICT,
    CONSTRAINT chk_pr_status
        CHECK (status IN ('OPEN', 'MERGED'))
);

CREATE INDEX idx_pr_author_id ON pull_requests(author_id);
CREATE INDEX idx_pr_status    ON pull_requests(status);

-- Pull Request Reviewers
CREATE TABLE pull_request_reviewers (
    pull_request_id VARCHAR(255) NOT NULL,
    reviewer_id     VARCHAR(255) NOT NULL,
    CONSTRAINT fk_prr_pr
        FOREIGN KEY (pull_request_id)
        REFERENCES pull_requests(pull_request_id)
        ON UPDATE CASCADE
        ON DELETE CASCADE,
    CONSTRAINT fk_prr_reviewer
        FOREIGN KEY (reviewer_id)
        REFERENCES users(user_id)
        ON UPDATE CASCADE
        ON DELETE RESTRICT,
    PRIMARY KEY (pull_request_id, reviewer_id)
);

CREATE INDEX idx_pr_reviewers_reviewer_id
    ON pull_request_reviewers(reviewer_id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_pr_reviewers_reviewer_id;
DROP TABLE IF EXISTS pull_request_reviewers;

DROP INDEX IF EXISTS idx_pr_status;
DROP INDEX IF EXISTS idx_pr_author_id;
DROP TABLE IF EXISTS pull_requests;

DROP INDEX IF EXISTS idx_users_team_name;
DROP TABLE IF EXISTS users;

DROP TABLE IF EXISTS teams;
-- +goose StatementEnd
