package users

import (
	"context"
	"errors"

	"github.com/google/uuid"

	"attendance-backend/internal/auth"
)

var ErrWrongPassword = errors.New("current password is incorrect")

type Service struct {
	repo *Repo
}

func NewService(repo *Repo) *Service { return &Service{repo: repo} }

// CreateAdmin creates an additional admin account. Coach accounts are
// created via the coaches module, which also provisions the coach profile
// and activity assignments in one transaction.
func (s *Service) CreateAdmin(ctx context.Context, name, email, phone, password string) (uuid.UUID, error) {
	hash, err := auth.HashPassword(password)
	if err != nil {
		return uuid.Nil, err
	}
	return s.repo.Create(ctx, "admin", name, email, phone, hash)
}

func (s *Service) List(ctx context.Context, roleFilter string, limit, offset int) ([]User, int64, error) {
	return s.repo.List(ctx, roleFilter, limit, offset)
}

func (s *Service) SetActive(ctx context.Context, id uuid.UUID, active bool) error {
	return s.repo.SetActive(ctx, id, active)
}

func (s *Service) ChangePassword(ctx context.Context, userID uuid.UUID, currentPassword, newPassword string) error {
	hash, err := s.repo.GetPasswordHash(ctx, userID)
	if err != nil {
		return err
	}
	if !auth.CheckPassword(hash, currentPassword) {
		return ErrWrongPassword
	}
	newHash, err := auth.HashPassword(newPassword)
	if err != nil {
		return err
	}
	return s.repo.UpdatePasswordHash(ctx, userID, newHash)
}
