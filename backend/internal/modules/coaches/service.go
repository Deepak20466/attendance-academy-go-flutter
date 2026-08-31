package coaches

import (
	"context"

	"github.com/google/uuid"

	"attendance-backend/internal/auth"
	"attendance-backend/internal/authz"
)

type Service struct {
	repo *Repo
}

func NewService(repo *Repo) *Service { return &Service{repo: repo} }

type CreateCoachInput struct {
	Name          string
	Email         string
	Phone         string
	Password      string
	EmployeeCode  string
	MonthlySalary float64
	ActivityIDs   []uuid.UUID
}

// userCreatorFn lets the coaches service create the backing user account
// without importing the users package directly (avoids a module cycle);
// wired up from main.go.
type userCreatorFn func(ctx context.Context, name, email, phone, passwordHash string) (uuid.UUID, error)

type ServiceWithUserCreation struct {
	*Service
	createUser userCreatorFn
}

func NewServiceWithUserCreation(repo *Repo, createUser userCreatorFn) *ServiceWithUserCreation {
	return &ServiceWithUserCreation{Service: NewService(repo), createUser: createUser}
}

func (s *ServiceWithUserCreation) CreateCoach(ctx context.Context, in CreateCoachInput) (uuid.UUID, error) {
	hash, err := auth.HashPassword(in.Password)
	if err != nil {
		return uuid.Nil, err
	}
	userID, err := s.createUser(ctx, in.Name, in.Email, in.Phone, hash)
	if err != nil {
		return uuid.Nil, err
	}
	return s.repo.Create(ctx, CreateInput{
		UserID:        userID,
		EmployeeCode:  in.EmployeeCode,
		MonthlySalary: in.MonthlySalary,
		ActivityIDs:   in.ActivityIDs,
	})
}

func (s *Service) Get(ctx context.Context, id uuid.UUID) (*Coach, error) {
	return s.repo.Get(ctx, id)
}

func (s *Service) GetSelf(ctx context.Context, userID uuid.UUID) (*Coach, error) {
	return s.repo.GetByUserID(ctx, userID)
}

// List returns all coaches for admins. Coaches are not permitted to list
// other coaches at all (only their own profile via GetSelf), since coach
// rosters for other activities are not their concern.
func (s *Service) List(ctx context.Context, activityID *uuid.UUID, limit, offset int) ([]Coach, int64, error) {
	return s.repo.List(ctx, activityID, limit, offset)
}

func (s *Service) SetActivities(ctx context.Context, coachID uuid.UUID, activityIDs []uuid.UUID) error {
	return s.repo.SetActivities(ctx, coachID, activityIDs)
}

func (s *Service) SetActive(ctx context.Context, id uuid.UUID, active bool) error {
	return s.repo.SetActive(ctx, id, active)
}

func (s *Service) UpdateSalary(ctx context.Context, id uuid.UUID, salary float64) error {
	return s.repo.UpdateSalary(ctx, id, salary)
}

// EnsureCanViewCoach lets a coach fetch only their own profile.
func EnsureCanViewCoach(actor *authz.Actor, coachID uuid.UUID) error {
	return authz.AuthorizeCoachSelf(actor, coachID)
}
