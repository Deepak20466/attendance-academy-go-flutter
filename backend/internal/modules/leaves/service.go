package leaves

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"

	"attendance-backend/internal/audit"
	"attendance-backend/internal/authz"
)

var (
	ErrInvalidDateRange = errors.New("end_date must not be before start_date")
	ErrNotPending       = errors.New("only pending leave requests can be reviewed or cancelled")
)

type Service struct {
	repo  *Repo
	audit *audit.Logger
}

func NewService(repo *Repo, auditLogger *audit.Logger) *Service {
	return &Service{repo: repo, audit: auditLogger}
}

func (s *Service) Apply(ctx context.Context, coachID uuid.UUID, start, end time.Time, reason string) (uuid.UUID, error) {
	if end.Before(start) {
		return uuid.Nil, ErrInvalidDateRange
	}
	return s.repo.Create(ctx, coachID, start, end, reason)
}

func (s *Service) Review(ctx context.Context, actor *authz.Actor, leaveID uuid.UUID, approve bool) error {
	leave, err := s.repo.Get(ctx, leaveID)
	if err != nil {
		return err
	}
	if leave.Status != "pending" {
		return ErrNotPending
	}
	status := "rejected"
	if approve {
		status = "approved"
	}
	if err := s.repo.UpdateStatus(ctx, leaveID, status, &actor.UserID); err != nil {
		return err
	}
	_ = s.audit.Log(ctx, &actor.UserID, "leave."+status, "leave", &leaveID, map[string]interface{}{"status": "pending"}, map[string]interface{}{"status": status}, "")
	return nil
}

// Cancel is performed by the coach themselves, only while still pending.
func (s *Service) Cancel(ctx context.Context, actor *authz.Actor, leaveID uuid.UUID) error {
	leave, err := s.repo.Get(ctx, leaveID)
	if err != nil {
		return err
	}
	if err := authz.AuthorizeCoachSelf(actor, leave.CoachID); err != nil {
		return err
	}
	if leave.Status != "pending" {
		return ErrNotPending
	}
	return s.repo.UpdateStatus(ctx, leaveID, "cancelled", nil)
}

func (s *Service) ListMine(ctx context.Context, coachID uuid.UUID, limit, offset int) ([]Leave, int64, error) {
	return s.repo.ListForCoach(ctx, coachID, limit, offset)
}

func (s *Service) ListAll(ctx context.Context, statusFilter string, limit, offset int) ([]Leave, int64, error) {
	return s.repo.ListAll(ctx, statusFilter, limit, offset)
}

func (s *Service) Get(ctx context.Context, actor *authz.Actor, id uuid.UUID) (*Leave, error) {
	leave, err := s.repo.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	if !actor.IsAdmin() {
		if err := authz.AuthorizeCoachSelf(actor, leave.CoachID); err != nil {
			return nil, err
		}
	}
	return leave, nil
}
