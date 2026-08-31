package locations

import (
	"context"

	"github.com/google/uuid"

	"attendance-backend/internal/authz"
)

// Creation/update/delete is admin-only (route-gated); listing is available
// to any authenticated user but scoped to activities they're allowed to
// see, same as every other activity-scoped resource.
type Service struct {
	repo *Repo
}

func NewService(repo *Repo) *Service { return &Service{repo: repo} }

func (s *Service) Create(ctx context.Context, in UpsertInput) (uuid.UUID, error) {
	return s.repo.Create(ctx, in)
}

func (s *Service) Update(ctx context.Context, id uuid.UUID, in UpsertInput) error {
	return s.repo.Update(ctx, id, in)
}

func (s *Service) Delete(ctx context.Context, id uuid.UUID) error {
	return s.repo.Delete(ctx, id)
}

func (s *Service) ListByActivity(ctx context.Context, actor *authz.Actor, activityID uuid.UUID) ([]Location, error) {
	if err := authz.AuthorizeActivity(actor, activityID); err != nil {
		return nil, err
	}
	return s.repo.ListByActivity(ctx, activityID)
}
