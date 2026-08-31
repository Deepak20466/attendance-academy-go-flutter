package attendance

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"attendance-backend/internal/authz"
)

var ErrInvalidStatus = errors.New("invalid attendance status")
var ErrStudentNotInBatch = errors.New("student does not belong to this class's batch")

var validStatuses = map[string]bool{"present": true, "absent": true, "late": true, "excused": true}

// ClassCompleter marks a class as having its attendance submitted, closing
// the window the 15-minute reminder job watches. Implemented by
// classes.ClassRepo and wired in from main.go, kept as an interface here
// so this module doesn't need to import the classes module.
type ClassCompleter interface {
	MarkAttendanceDone(ctx context.Context, classID uuid.UUID) error
}

type StudentService struct {
	repo      *StudentAttendanceRepo
	pool      *pgxpool.Pool
	completer ClassCompleter
}

func NewStudentService(repo *StudentAttendanceRepo, pool *pgxpool.Pool, completer ClassCompleter) *StudentService {
	return &StudentService{repo: repo, pool: pool, completer: completer}
}

type MarkEntry struct {
	StudentID uuid.UUID
	Status    string
	Remarks   string
}

// MarkBulk authorizes the caller against the specific class (covers the
// substitution case), verifies every student actually belongs to that
// class's batch, then upserts each mark and flips the class to
// attendance_marked = true so the missing-attendance reminder stops
// firing for it.
func (s *StudentService) MarkBulk(ctx context.Context, actor *authz.Actor, classID uuid.UUID, entries []MarkEntry, markedBy uuid.UUID) error {
	if err := authz.AuthorizeClassAccess(ctx, s.pool, actor, classID); err != nil {
		return err
	}
	classInfo, err := s.repo.GetClassInfo(ctx, classID)
	if err != nil {
		return err
	}

	for _, e := range entries {
		if !validStatuses[e.Status] {
			return ErrInvalidStatus
		}
		ok, err := s.repo.StudentBelongsToBatch(ctx, e.StudentID, classInfo.BatchID)
		if err != nil {
			return err
		}
		if !ok {
			return ErrStudentNotInBatch
		}
	}

	for _, e := range entries {
		if err := s.repo.Upsert(ctx, MarkInput{
			ClassID: classID, StudentID: e.StudentID, ActivityID: classInfo.ActivityID,
			Status: e.Status, Remarks: e.Remarks, MarkedBy: markedBy,
		}); err != nil {
			return err
		}
	}

	return s.completer.MarkAttendanceDone(ctx, classID)
}

func (s *StudentService) ListForClass(ctx context.Context, actor *authz.Actor, classID uuid.UUID) ([]StudentAttendanceRecord, error) {
	if err := authz.AuthorizeClassAccess(ctx, s.pool, actor, classID); err != nil {
		return nil, err
	}
	return s.repo.ListForClass(ctx, classID)
}

func (s *StudentService) History(ctx context.Context, actor *authz.Actor, studentID uuid.UUID, limit, offset int) ([]StudentAttendanceRecord, int64, error) {
	if err := authz.AuthorizeStudentAccess(ctx, s.pool, actor, studentID); err != nil {
		return nil, 0, err
	}
	return s.repo.HistoryForStudent(ctx, studentID, limit, offset)
}

func (s *StudentService) MonthlyPercentage(ctx context.Context, actor *authz.Actor, studentID, batchID uuid.UUID, year, month int) (*MonthlyPercentage, error) {
	if err := authz.AuthorizeStudentAccess(ctx, s.pool, actor, studentID); err != nil {
		return nil, err
	}
	return s.repo.MonthlyPercentage(ctx, studentID, batchID, year, month)
}
