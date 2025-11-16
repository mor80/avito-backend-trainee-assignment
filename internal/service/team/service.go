package team

import (
	"context"
	"fmt"
	"strings"

	"mor80/service-reviewer/internal/model"
	"mor80/service-reviewer/internal/service"
)

type TeamService struct {
	teamRepo service.TeamRepository
	userRepo service.UserRepository
}

func New(teamRepo service.TeamRepository, userRepo service.UserRepository) *TeamService {
	return &TeamService{
		teamRepo: teamRepo,
		userRepo: userRepo,
	}
}

func (s *TeamService) Create(ctx context.Context, team model.Team) (*model.Team, error) {
	if err := validateTeam(team); err != nil {
		return nil, fmt.Errorf("team service: %w", err)
	}

	exists, err := s.teamRepo.Exists(ctx, team.Name)
	if err != nil {
		return nil, fmt.Errorf("team service: %w", err)
	}
	if exists {
		return nil, model.ErrTeamExists
	}

	if err := s.teamRepo.Create(ctx, team.Name); err != nil {
		return nil, fmt.Errorf("team service: %w", err)
	}

	if err := s.userRepo.Upsert(ctx, membersToUsers(team)); err != nil {
		return nil, fmt.Errorf("team service: %w", err)
	}

	return s.teamRepo.GetByName(ctx, team.Name)
}

func (s *TeamService) Get(ctx context.Context, teamName string) (*model.Team, error) {
	if err := validateName(teamName); err != nil {
		return nil, fmt.Errorf("team service: %w", err)
	}

	team, err := s.teamRepo.GetByName(ctx, teamName)
	if err != nil {
		return nil, fmt.Errorf("team service: get: %w", err)
	}

	return team, nil
}

func validateTeam(team model.Team) error {
	if err := validateName(team.Name); err != nil {
		return err
	}

	for _, member := range team.Members {
		if err := validateTeamMember(member); err != nil {
			return err
		}
	}

	return nil
}

func validateName(teamName string) error {
	if strings.TrimSpace(teamName) == "" {
		return fmt.Errorf("team_name is required")
	}

	return nil
}

func validateTeamMember(member model.TeamMember) error {
	if strings.TrimSpace(member.ID) == "" {
		return fmt.Errorf("member.user_id is required")
	}

	if strings.TrimSpace(member.Username) == "" {
		return fmt.Errorf("member.username is required")
	}

	return nil
}

func membersToUsers(team model.Team) []model.User {
	users := make([]model.User, len(team.Members))

	for i, member := range team.Members {
		users[i] = model.User{
			ID:       member.ID,
			Username: member.Username,
			TeamName: team.Name,
			IsActive: member.IsActive,
		}
	}

	return users
}
