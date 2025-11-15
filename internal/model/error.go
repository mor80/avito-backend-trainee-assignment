package model

import "fmt"

type ErrorCode string

const (
	ErrorCodeTeamExists  ErrorCode = "TEAM_EXISTS"
	ErrorCodePRExists    ErrorCode = "PR_EXISTS"
	ErrorCodePRMerged    ErrorCode = "PR_MERGED"
	ErrorCodeNotAssigned ErrorCode = "NOT_ASSIGNED"
	ErrorCodeNoCandidate ErrorCode = "NO_CANDIDATE"
	ErrorCodeNotFound    ErrorCode = "NOT_FOUND"
)

type DomainError struct {
	Code    ErrorCode
	Message string
}

func (e DomainError) Error() string {
	if e.Message == "" {
		return string(e.Code)
	}

	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

func NewDomainError(code ErrorCode, message string) DomainError {
	return DomainError{
		Code:    code,
		Message: message,
	}
}

var (
	ErrTeamExists  = DomainError{Code: ErrorCodeTeamExists, Message: "team already exists"}
	ErrPRExists    = DomainError{Code: ErrorCodePRExists, Message: "pull request already exists"}
	ErrPRMerged    = DomainError{Code: ErrorCodePRMerged, Message: "pull request already merged"}
	ErrNotAssigned = DomainError{Code: ErrorCodeNotAssigned, Message: "reviewer is not assigned to this pull request"}
	ErrNoCandidate = DomainError{Code: ErrorCodeNoCandidate, Message: "no replacement candidate available"}
	ErrNotFound    = DomainError{Code: ErrorCodeNotFound, Message: "resource not found"}
)
