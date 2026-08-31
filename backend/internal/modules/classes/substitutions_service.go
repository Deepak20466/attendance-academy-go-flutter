package classes

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"

	"attendance-backend/internal/audit"
	"attendance-backend/internal/authz"
)

var (
	ErrSelfSubstitution = errors.New("a coach cannot be assigned as their own substitute")
	ErrActiveSubExists  = errors.New("this class already has an active substitution")
)

const (
	pgUniqueViolationCode = "23505"
	pgCheckViolationCode  = "23514"
)

type SubstitutionService struct {
	repo      *SubstitutionRepo
	classRepo *ClassRepo
	audit     *audit.Logger
}

func NewSubstitutionService(repo *SubstitutionRepo, classRepo *ClassRepo, auditLogger *audit.Logger) *SubstitutionService {
	return &SubstitutionService{repo: repo, classRepo: classRepo, audit: auditLogger}
}

// Create is admin-only. It looks up the class's current coach to record as
// original_coach_id and stamps the acting admin as authorized_by, giving a
// full audit trail of who assigned the substitution and why.
func (s *SubstitutionService) Create(ctx context.Context, actor *authz.Actor, classID, substituteCoachID uuid.UUID, reason string) (uuid.UUID, error) {
	class, err := s.classRepo.Get(ctx, classID)
	if err != nil {
		return uuid.Nil, err
	}
	if class.CoachID == substituteCoachID {
		return uuid.Nil, ErrSelfSubstitution
	}
	id, err := s.repo.Create(ctx, classID, class.CoachID, substituteCoachID, actor.UserID, reason)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			switch pgErr.Code {
			case pgUniqueViolationCode:
				return uuid.Nil, ErrActiveSubExists
			case pgCheckViolationCode:
				return uuid.Nil, ErrSelfSubstitution
			}
		}
		return uuid.Nil, err
	}
	_ = s.audit.Log(ctx, &actor.UserID, "substitution.create", "substitution", &id, nil, map[string]interface{}{
		"class_id": classID, "original_coach_id": class.CoachID, "substitute_coach_id": substituteCoachID, "reason": reason,
	}, "")
	return id, nil
}

func (s *SubstitutionService) Cancel(ctx context.Context, actor *authz.Actor, id uuid.UUID) error {
	if err := s.repo.Cancel(ctx, id); err != nil {
		return err
	}
	_ = s.audit.Log(ctx, &actor.UserID, "substitution.cancel", "substitution", &id, nil, nil, "")
	return nil
}

func (s *SubstitutionService) ListMine(ctx context.Context, coachID uuid.UUID) ([]Substitution, error) {
	return s.repo.ListForCoach(ctx, coachID)
}

func (s *SubstitutionService) ListAll(ctx context.Context, limit, offset int) ([]Substitution, int64, error) {
	return s.repo.ListAll(ctx, limit, offset)
}
