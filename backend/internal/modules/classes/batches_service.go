package classes

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"

	"attendance-backend/internal/authz"
)

// ErrBatchBadReference means activity_id, default_coach_id or location_id
// doesn't point at a real row.
var ErrBatchBadReference = errors.New("activity_id, default_coach_id or location_id does not refer to an existing record")

type BatchService struct {
	repo *BatchRepo
}

func NewBatchService(repo *BatchRepo) *BatchService { return &BatchService{repo: repo} }

func (s *BatchService) Create(ctx context.Context, actor *authz.Actor, in BatchInput) (uuid.UUID, error) {
	if err := authz.AuthorizeActivity(actor, in.ActivityID); err != nil {
		return uuid.Nil, err
	}
	id, err := s.repo.Create(ctx, in)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == pgForeignKeyViolationCode {
			return uuid.Nil, ErrBatchBadReference
		}
		return uuid.Nil, err
	}
	return id, nil
}

func (s *BatchService) Get(ctx context.Context, actor *authz.Actor, id uuid.UUID) (*Batch, error) {
	b, err := s.repo.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	if err := authz.AuthorizeActivity(actor, b.ActivityID); err != nil {
		return nil, err
	}
	return b, nil
}

func (s *BatchService) List(ctx context.Context, actor *authz.Actor, activityFilter *uuid.UUID) ([]Batch, error) {
	var allowed []uuid.UUID
	if !actor.IsAdmin() {
		allowed = actor.ActivityIDList()
		if activityFilter != nil && !actor.HasActivity(*activityFilter) {
			return []Batch{}, nil
		}
	}
	return s.repo.ListByActivities(ctx, allowed, activityFilter)
}

func (s *BatchService) Update(ctx context.Context, actor *authz.Actor, id uuid.UUID, in BatchInput, isActive bool) error {
	existing, err := s.repo.Get(ctx, id)
	if err != nil {
		return err
	}
	if err := authz.AuthorizeActivity(actor, existing.ActivityID); err != nil {
		return err
	}
	if err := s.repo.Update(ctx, id, in, isActive); err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == pgForeignKeyViolationCode {
			return ErrBatchBadReference
		}
		return err
	}
	return nil
}
