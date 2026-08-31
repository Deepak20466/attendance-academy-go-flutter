package notifications

import (
	"context"
	"log"

	"github.com/google/uuid"

	notifinfra "attendance-backend/internal/notifications"
)

type Service struct {
	repo *Repo
	fcm  *notifinfra.FCMClient // nil if FCM is not configured; pushes are then skipped gracefully
}

func NewService(repo *Repo, fcm *notifinfra.FCMClient) *Service {
	return &Service{repo: repo, fcm: fcm}
}

// Notify always records an in-app notification first (source of truth for
// the notification center), then best-effort pushes via FCM. A push
// failure (missing token, FCM outage) never fails the caller's request.
func (s *Service) Notify(ctx context.Context, userID uuid.UUID, title, body, notifType string, data map[string]string) error {
	if _, err := s.repo.Create(ctx, userID, title, body, notifType, data); err != nil {
		return err
	}

	if s.fcm == nil {
		return nil
	}
	token, err := s.repo.FCMTokenForUser(ctx, userID)
	if err != nil || token == "" {
		return nil
	}
	if err := s.fcm.Send(ctx, token, title, body, data); err != nil {
		log.Printf("notifications: FCM push failed for user %s: %v", userID, err)
	}
	return nil
}

func (s *Service) NotifyAllAdmins(ctx context.Context, title, body, notifType string, data map[string]string) {
	admins, err := s.repo.AdminUserIDs(ctx)
	if err != nil {
		log.Printf("notifications: failed to load admins: %v", err)
		return
	}
	for _, adminID := range admins {
		if err := s.Notify(ctx, adminID, title, body, notifType, data); err != nil {
			log.Printf("notifications: failed to notify admin %s: %v", adminID, err)
		}
	}
}

func (s *Service) NotifyCoach(ctx context.Context, coachID uuid.UUID, title, body, notifType string, data map[string]string) {
	userID, err := s.repo.UserIDForCoach(ctx, coachID)
	if err != nil {
		log.Printf("notifications: failed to resolve user for coach %s: %v", coachID, err)
		return
	}
	if err := s.Notify(ctx, userID, title, body, notifType, data); err != nil {
		log.Printf("notifications: failed to notify coach %s: %v", coachID, err)
	}
}

func (s *Service) ListForUser(ctx context.Context, userID uuid.UUID, limit, offset int) ([]Notification, int64, error) {
	return s.repo.ListForUser(ctx, userID, limit, offset)
}

func (s *Service) MarkRead(ctx context.Context, id, userID uuid.UUID) error {
	return s.repo.MarkRead(ctx, id, userID)
}
