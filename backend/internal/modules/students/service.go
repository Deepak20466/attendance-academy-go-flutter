package students

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"attendance-backend/internal/authz"
)

// ErrBatchNotFound is returned when a request references a batch_id that
// doesn't exist — distinct from authz.ErrForbidden so the handler can tell
// "no such batch" apart from "batch exists but you can't use it".
var ErrBatchNotFound = errors.New("batch not found")

type Service struct {
	repo *Repo
}

func NewService(repo *Repo) *Service { return &Service{repo: repo} }

func (s *Service) Create(ctx context.Context, actor *authz.Actor, in CreateInput) (uuid.UUID, error) {
	// activity_id is always resolved from the batch server-side, never
	// trusted from the request, so it can't be spoofed independently of
	// batch_id to slip past the activity check below.
	activityID, err := s.repo.GetBatchActivityID(ctx, in.BatchID)
	if errors.Is(err, pgx.ErrNoRows) {
		return uuid.Nil, ErrBatchNotFound
	}
	if err != nil {
		return uuid.Nil, err
	}
	if err := authz.AuthorizeActivity(actor, activityID); err != nil {
		return uuid.Nil, err
	}
	in.ActivityID = activityID
	return s.repo.Create(ctx, in)
}

func (s *Service) Get(ctx context.Context, actor *authz.Actor, id uuid.UUID) (*Student, error) {
	student, err := s.repo.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	if err := authz.AuthorizeActivity(actor, student.ActivityID); err != nil {
		return nil, err
	}
	return student, nil
}

func (s *Service) List(ctx context.Context, actor *authz.Actor, activityFilter, batchFilter *uuid.UUID, limit, offset int) ([]Student, int64, error) {
	var allowed []uuid.UUID
	if !actor.IsAdmin() {
		allowed = actor.ActivityIDList()
		// A coach requesting a specific activity they don't belong to gets
		// an empty result, not another coach's data and not an error that
		// would reveal the activity exists.
		if activityFilter != nil && !actor.HasActivity(*activityFilter) {
			return []Student{}, 0, nil
		}
	}
	return s.repo.List(ctx, allowed, activityFilter, batchFilter, limit, offset)
}

func (s *Service) Update(ctx context.Context, actor *authz.Actor, id uuid.UUID, in CreateInput, isActive bool) error {
	existing, err := s.repo.Get(ctx, id)
	if err != nil {
		return err
	}
	if err := authz.AuthorizeActivity(actor, existing.ActivityID); err != nil {
		return err
	}
	if in.BatchID != existing.BatchID {
		newActivityID, err := s.repo.GetBatchActivityID(ctx, in.BatchID)
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrBatchNotFound
		}
		if err != nil {
			return err
		}
		if err := authz.AuthorizeActivity(actor, newActivityID); err != nil {
			return err
		}
	}
	return s.repo.Update(ctx, id, in, isActive)
}
