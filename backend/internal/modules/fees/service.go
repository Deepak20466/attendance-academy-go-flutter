package fees

import (
	"context"
	"time"

	"github.com/google/uuid"

	"attendance-backend/internal/authz"
)

type Service struct {
	repo *Repo
}

func NewService(repo *Repo) *Service { return &Service{repo: repo} }

func (s *Service) Create(ctx context.Context, in CreateInput) (uuid.UUID, error) {
	return s.repo.Create(ctx, in)
}

func (s *Service) GenerateForActivity(ctx context.Context, activityID uuid.UUID, amount float64, dueDate time.Time, month, year int) (int64, error) {
	return s.repo.GenerateForActivity(ctx, activityID, amount, dueDate, month, year)
}

func (s *Service) MarkPaid(ctx context.Context, id uuid.UUID, paidDate time.Time, method string) error {
	return s.repo.MarkPaid(ctx, id, paidDate, method)
}

func (s *Service) Get(ctx context.Context, actor *authz.Actor, id uuid.UUID) (*Fee, error) {
	fee, err := s.repo.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	if err := authz.AuthorizeActivity(actor, fee.ActivityID); err != nil {
		return nil, err
	}
	return fee, nil
}

func (s *Service) List(ctx context.Context, actor *authz.Actor, activityFilter, studentFilter *uuid.UUID, status string, limit, offset int) ([]Fee, int64, error) {
	f := ListFilter{ActivityID: activityFilter, StudentID: studentFilter, Status: status}
	if !actor.IsAdmin() {
		f.AllowedActivityIDs = actor.ActivityIDList()
		if activityFilter != nil && !actor.HasActivity(*activityFilter) {
			return []Fee{}, 0, nil
		}
	}
	return s.repo.List(ctx, f, limit, offset)
}

func (s *Service) PendingSummary(ctx context.Context) ([]PendingSummary, error) {
	return s.repo.PendingSummaryByActivity(ctx)
}

func (s *Service) MarkOverdue(ctx context.Context) error {
	return s.repo.MarkOverdue(ctx)
}
