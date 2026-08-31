package salary

import (
	"context"
	"errors"

	"github.com/google/uuid"

	"attendance-backend/internal/authz"
)

var ErrAlreadyAcknowledged = errors.New("salary already acknowledged")

type Service struct {
	repo *Repo
}

func NewService(repo *Repo) *Service { return &Service{repo: repo} }

func (s *Service) GenerateForPeriod(ctx context.Context, month, year int) (int64, error) {
	return s.repo.CreateForAllActiveCoaches(ctx, month, year)
}

func (s *Service) Acknowledge(ctx context.Context, actor *authz.Actor, id uuid.UUID) error {
	ack, err := s.repo.Get(ctx, id)
	if err != nil {
		return err
	}
	if err := authz.AuthorizeCoachSelf(actor, ack.CoachID); err != nil {
		return err
	}
	if ack.Status == "acknowledged" {
		return ErrAlreadyAcknowledged
	}
	return s.repo.Acknowledge(ctx, id)
}

func (s *Service) ListMine(ctx context.Context, coachID uuid.UUID, limit, offset int) ([]Acknowledgement, int64, error) {
	return s.repo.ListForCoach(ctx, coachID, limit, offset)
}

func (s *Service) ListForPeriod(ctx context.Context, month, year int, statusFilter string, limit, offset int) ([]Acknowledgement, int64, error) {
	return s.repo.ListForPeriod(ctx, month, year, statusFilter, limit, offset)
}
