package activities

import (
	"context"

	"github.com/google/uuid"

	"attendance-backend/internal/authz"
)

type Service struct {
	repo *Repo
}

func NewService(repo *Repo) *Service { return &Service{repo: repo} }

func (s *Service) Create(ctx context.Context, name, description string) (uuid.UUID, error) {
	return s.repo.Create(ctx, name, description)
}

func (s *Service) Update(ctx context.Context, id uuid.UUID, name, description string, isActive bool) error {
	return s.repo.Update(ctx, id, name, description, isActive)
}

func (s *Service) Get(ctx context.Context, id uuid.UUID) (*Activity, error) {
	return s.repo.Get(ctx, id)
}

// List returns every activity for admins, but only the activities the
// calling coach is assigned to. A coach must never see the existence of
// activities outside their own assignment.
func (s *Service) List(ctx context.Context, actor *authz.Actor, onlyActive bool) ([]Activity, error) {
	if actor.IsAdmin() {
		return s.repo.ListAll(ctx, onlyActive)
	}
	return s.repo.ListByIDs(ctx, actor.ActivityIDList())
}
