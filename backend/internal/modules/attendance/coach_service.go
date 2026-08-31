package attendance

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"attendance-backend/internal/authz"
)

var (
	ErrOutsideGeofence = errors.New("you are not within the required location radius for this class")
	// ErrNoGeofenceConfigured means this class's activity has no usable
	// location on record — either none at all, or more than one with
	// nothing on the batch to disambiguate which applies. Either way, a
	// coach cannot GPS check in until an admin configures it.
	ErrNoGeofenceConfigured = errors.New("no geofence location is configured for this class yet — ask an admin to add one")
)

type CoachService struct {
	repo *CoachAttendanceRepo
	pool *pgxpool.Pool
}

func NewCoachService(repo *CoachAttendanceRepo, pool *pgxpool.Pool) *CoachService {
	return &CoachService{repo: repo, pool: pool}
}

// CheckIn is only ever performed by the coach on their own behalf (never
// by admin, and never for a class they aren't authorized to touch — the
// same AuthorizeClassAccess guard used for marking student attendance, so
// a substitute coach can check in for their assigned substitute class too).
// A check-in outside the configured geofence radius is rejected outright
// rather than silently recorded as unverified, since an unenforced geofence
// isn't a geofence.
func (s *CoachService) CheckIn(ctx context.Context, actor *authz.Actor, classID uuid.UUID, lat, lng float64) (*CoachAttendanceRecord, error) {
	if err := authz.AuthorizeClassAccess(ctx, s.pool, actor, classID); err != nil {
		return nil, err
	}

	loc, err := s.repo.LocationForClass(ctx, classID)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNoGeofenceConfigured
	}
	if err != nil {
		return nil, err
	}

	distance := HaversineDistanceMeters(lat, lng, loc.Latitude, loc.Longitude)
	verified := distance <= float64(loc.RadiusMeters)
	if !verified {
		return nil, ErrOutsideGeofence
	}

	id, err := s.repo.CheckIn(ctx, *actor.CoachID, classID, lat, lng, verified)
	if err != nil {
		return nil, err
	}
	return &CoachAttendanceRecord{ID: id, CoachID: *actor.CoachID, CheckInVerified: verified}, nil
}

func (s *CoachService) CheckOut(ctx context.Context, actor *authz.Actor, classID uuid.UUID, lat, lng float64) error {
	if err := authz.AuthorizeClassAccess(ctx, s.pool, actor, classID); err != nil {
		return err
	}

	loc, err := s.repo.LocationForClass(ctx, classID)
	if errors.Is(err, pgx.ErrNoRows) {
		return ErrNoGeofenceConfigured
	}
	if err != nil {
		return err
	}
	distance := HaversineDistanceMeters(lat, lng, loc.Latitude, loc.Longitude)
	verified := distance <= float64(loc.RadiusMeters)
	if !verified {
		return ErrOutsideGeofence
	}

	return s.repo.CheckOut(ctx, *actor.CoachID, classID, lat, lng, verified)
}

// List is admin-only monitoring, or a coach viewing their own history.
func (s *CoachService) List(ctx context.Context, actor *authz.Actor, coachFilter *uuid.UUID, f CoachAttendanceFilter, limit, offset int) ([]CoachAttendanceRecord, int64, error) {
	if !actor.IsAdmin() {
		f.CoachID = actor.CoachID
	} else {
		f.CoachID = coachFilter
	}
	return s.repo.List(ctx, f, limit, offset)
}
