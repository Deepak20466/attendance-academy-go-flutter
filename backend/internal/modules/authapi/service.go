package authapi

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"

	"attendance-backend/internal/auth"
)

var (
	ErrInvalidCredentials = errors.New("invalid email or password")
	ErrAccountDisabled    = errors.New("account is deactivated")
	ErrInvalidRefresh     = errors.New("invalid or expired refresh token")
)

type Service struct {
	repo *Repo
	tm   *auth.TokenManager
}

func NewService(repo *Repo, tm *auth.TokenManager) *Service {
	return &Service{repo: repo, tm: tm}
}

type TokenPair struct {
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token"`
	ExpiresAt    time.Time `json:"expires_at"`
	UserID       uuid.UUID `json:"user_id"`
	Name         string    `json:"name"`
	Email        string    `json:"email"`
	Role         string    `json:"role"`
}

func (s *Service) Login(ctx context.Context, email, password string) (*TokenPair, error) {
	user, err := s.repo.FindByEmail(ctx, email)
	if err != nil {
		return nil, ErrInvalidCredentials
	}
	if !auth.CheckPassword(user.PasswordHash, password) {
		return nil, ErrInvalidCredentials
	}
	if !user.IsActive {
		return nil, ErrAccountDisabled
	}
	return s.issueTokens(ctx, user.ID, user.Name, user.Email, user.RoleName)
}

func (s *Service) Refresh(ctx context.Context, rawRefreshToken string) (*TokenPair, error) {
	hash := auth.HashRefreshToken(rawRefreshToken)
	rt, err := s.repo.FindRefreshToken(ctx, hash)
	if err != nil {
		return nil, ErrInvalidRefresh
	}
	if rt.Revoked || time.Now().After(rt.ExpiresAt) {
		return nil, ErrInvalidRefresh
	}

	// Rotate: revoke the used token, then issue a brand new pair.
	if err := s.repo.RevokeRefreshToken(ctx, hash); err != nil {
		return nil, err
	}

	u, err := s.findUserByID(ctx, rt.UserID)
	if err != nil {
		return nil, ErrInvalidRefresh
	}
	if !u.IsActive {
		return nil, ErrAccountDisabled
	}

	return s.issueTokens(ctx, u.ID, u.Name, u.Email, u.RoleName)
}

func (s *Service) Logout(ctx context.Context, rawRefreshToken string) error {
	hash := auth.HashRefreshToken(rawRefreshToken)
	return s.repo.RevokeRefreshToken(ctx, hash)
}

func (s *Service) UpdateFCMToken(ctx context.Context, userID uuid.UUID, token string) error {
	return s.repo.UpdateFCMToken(ctx, userID, token)
}

func (s *Service) issueTokens(ctx context.Context, userID uuid.UUID, name, email, role string) (*TokenPair, error) {
	access, err := s.tm.GenerateAccessToken(userID)
	if err != nil {
		return nil, err
	}
	rawRefresh, refreshHash, expiresAt, err := s.tm.GenerateRefreshToken()
	if err != nil {
		return nil, err
	}
	if err := s.repo.StoreRefreshToken(ctx, userID, refreshHash, expiresAt); err != nil {
		return nil, err
	}
	return &TokenPair{
		AccessToken:  access,
		RefreshToken: rawRefresh,
		ExpiresAt:    expiresAt,
		UserID:       userID,
		Name:         name,
		Email:        email,
		Role:         role,
	}, nil
}

func (s *Service) findUserByID(ctx context.Context, id uuid.UUID) (*UserRecord, error) {
	// Re-use FindByEmail's row shape via a small dedicated query.
	return s.repo.findByID(ctx, id)
}
