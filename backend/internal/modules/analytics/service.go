package analytics

import (
	"context"

	"github.com/google/uuid"
)

// All analytics endpoints are admin-only (enforced by the router), so the
// service is a thin pass-through with no per-actor scoping logic.
type Service struct {
	repo *Repo
}

func NewService(repo *Repo) *Service { return &Service{repo: repo} }

func (s *Service) OverallSummary(ctx context.Context, month, year int) (*OverallSummary, error) {
	return s.repo.OverallSummary(ctx, month, year)
}

func (s *Service) ActivitySummary(ctx context.Context, activityID uuid.UUID, month, year int) (*ActivitySummary, error) {
	return s.repo.ActivitySummary(ctx, activityID, month, year)
}

func (s *Service) PerfectAttendance(ctx context.Context, activityID uuid.UUID, month, year int) ([]StudentAttendanceSummary, error) {
	return s.repo.PerfectAttendanceStudents(ctx, activityID, month, year)
}

func (s *Service) MonthlyReport(ctx context.Context, activityID uuid.UUID, month, year int) ([]StudentAttendanceSummary, error) {
	return s.repo.MonthlyStudentReport(ctx, activityID, month, year)
}
