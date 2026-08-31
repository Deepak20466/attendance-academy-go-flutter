package scheduler

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"attendance-backend/internal/modules/classes"
	"attendance-backend/internal/modules/fees"
	"attendance-backend/internal/modules/notifications"
	"attendance-backend/internal/modules/salary"
)

// Scheduler runs the app's background jobs in-process on plain time.Ticker
// loops. There is no external cron/queue infrastructure by design (no
// Docker, no Redis) — the jobs are cheap, idempotent DB queries run at
// most every few minutes, which is more than adequate at this scale.
type Scheduler struct {
	classRepo        *classes.ClassRepo
	substitutionRepo *classes.SubstitutionRepo
	feesRepo         *fees.Repo
	salaryRepo       *salary.Repo
	notifSvc         *notifications.Service

	attendanceReminderMinutes int

	state *stateStore
}

func New(
	pool *pgxpool.Pool,
	classRepo *classes.ClassRepo,
	substitutionRepo *classes.SubstitutionRepo,
	feesRepo *fees.Repo,
	salaryRepo *salary.Repo,
	notifSvc *notifications.Service,
	attendanceReminderMinutes int,
) *Scheduler {
	return &Scheduler{
		classRepo:                 classRepo,
		substitutionRepo:          substitutionRepo,
		feesRepo:                  feesRepo,
		salaryRepo:                salaryRepo,
		notifSvc:                  notifSvc,
		attendanceReminderMinutes: attendanceReminderMinutes,
		state:                     newStateStore(pool),
	}
}

// Run blocks until ctx is cancelled, ticking every minute. Each job below
// decides for itself whether it actually has work to do this minute.
func (s *Scheduler) Run(ctx context.Context) {
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()

	log.Println("scheduler: started")
	for {
		select {
		case <-ctx.Done():
			log.Println("scheduler: stopped")
			return
		case now := <-ticker.C:
			s.tick(ctx, now)
		}
	}
}

func (s *Scheduler) tick(ctx context.Context, now time.Time) {
	if err := s.runAttendanceReminders(ctx); err != nil {
		log.Printf("scheduler: attendance reminder job failed: %v", err)
	}

	// End-of-day report: fire once, shortly after 9 PM local time. The
	// claim is persisted in the DB so a mid-window restart can't re-send it.
	if now.Hour() == 21 && now.Minute() < 5 {
		s.runOnce(ctx, "eod_report", now.Format("2006-01-02"), func() error {
			return s.runEndOfDayReport(ctx, now)
		})
	}

	// Salary acknowledgement reminder: fire once on the 10th of the month.
	if now.Day() == 10 {
		s.runOnce(ctx, "salary_reminder", now.Format("2006-01"), func() error {
			return s.runSalaryReminder(ctx, now)
		})
	}

	// End-of-month pending fee report to admins: fire once on the last day
	// of the month.
	if isLastDayOfMonth(now) {
		s.runOnce(ctx, "fee_report", now.Format("2006-01"), func() error {
			return s.runFeeReport(ctx)
		})
	}
}

// runOnce claims (jobKey, period) in the persisted state store and, only
// if this call is the one that won the claim, runs fn. If fn fails, the
// claim is released so a later tick within the same day/month retries
// rather than silently skipping that period's report forever.
func (s *Scheduler) runOnce(ctx context.Context, jobKey, period string, fn func() error) {
	claimed, err := s.state.tryClaim(ctx, jobKey, period)
	if err != nil {
		log.Printf("scheduler: %s: failed to claim state: %v", jobKey, err)
		return
	}
	if !claimed {
		return
	}
	if err := fn(); err != nil {
		log.Printf("scheduler: %s job failed: %v", jobKey, err)
		if releaseErr := s.state.release(ctx, jobKey); releaseErr != nil {
			log.Printf("scheduler: %s: failed to release claim after error: %v", jobKey, releaseErr)
		}
	}
}

func isLastDayOfMonth(t time.Time) bool {
	return t.AddDate(0, 0, 1).Day() == 1
}

// runAttendanceReminders notifies whoever is currently responsible for a
// class (the substitute if one is active, otherwise the original coach)
// when attendance hasn't been submitted N minutes after the class ended.
func (s *Scheduler) runAttendanceReminders(ctx context.Context) error {
	pending, err := s.classRepo.PendingAttendance(ctx, s.attendanceReminderMinutes)
	if err != nil {
		return err
	}

	for _, class := range pending {
		responsibleCoachID := class.CoachID
		sub, err := s.substitutionRepo.GetActiveForClass(ctx, class.ID)
		if err == nil && sub != nil {
			responsibleCoachID = sub.SubstituteCoachID
		} else if err != nil && err != pgx.ErrNoRows {
			log.Printf("scheduler: lookup substitution for class %s: %v", class.ID, err)
		}

		s.notifSvc.NotifyCoach(ctx, responsibleCoachID,
			"Attendance reminder",
			fmt.Sprintf("Attendance for your class on %s has not been submitted yet.", class.ClassDate.Format("Jan 2")),
			"attendance_reminder",
			map[string]string{"class_id": class.ID.String()},
		)

		if err := s.classRepo.MarkReminderSent(ctx, class.ID); err != nil {
			log.Printf("scheduler: mark reminder sent for class %s: %v", class.ID, err)
		}
	}
	return nil
}

// runEndOfDayReport tells every admin which classes today still have no
// attendance recorded.
func (s *Scheduler) runEndOfDayReport(ctx context.Context, now time.Time) error {
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	missing, err := s.classRepo.MissingAttendanceForDate(ctx, today)
	if err != nil {
		return err
	}

	body := fmt.Sprintf("%d class(es) today have missing attendance.", len(missing))
	if len(missing) == 0 {
		body = "All classes today have attendance recorded."
	}

	s.notifSvc.NotifyAllAdmins(ctx, "End-of-day attendance report", body, "eod_report", map[string]string{
		"date":          today.Format("2006-01-02"),
		"missing_count": fmt.Sprintf("%d", len(missing)),
	})
	return nil
}

// runSalaryReminder generates the current month's pending acknowledgement
// records (idempotent — ON CONFLICT DO NOTHING in the repo) and notifies
// every coach to acknowledge.
func (s *Scheduler) runSalaryReminder(ctx context.Context, now time.Time) error {
	month, year := int(now.Month()), now.Year()
	if _, err := s.salaryRepo.CreateForAllActiveCoaches(ctx, month, year); err != nil {
		return err
	}

	acks, _, err := s.salaryRepo.ListForPeriod(ctx, month, year, "pending", 1000, 0)
	if err != nil {
		return err
	}
	for _, ack := range acks {
		s.notifSvc.NotifyCoach(ctx, ack.CoachID,
			"Salary acknowledgement due",
			"Please acknowledge receipt of your salary for this month.",
			"salary_reminder",
			map[string]string{"acknowledgement_id": ack.ID.String()},
		)
	}
	return nil
}

// runFeeReport tells admins the total pending/overdue fees per activity at
// month end.
func (s *Scheduler) runFeeReport(ctx context.Context) error {
	if err := s.feesRepo.MarkOverdue(ctx); err != nil {
		return err
	}
	summary, err := s.feesRepo.PendingSummaryByActivity(ctx)
	if err != nil {
		return err
	}

	total := 0.0
	count := int64(0)
	for _, row := range summary {
		total += row.PendingTotal
		count += row.PendingCount
	}

	s.notifSvc.NotifyAllAdmins(ctx, "Monthly pending fees report",
		fmt.Sprintf("%d fee record(s) pending across all activities, totalling %.2f.", count, total),
		"fee_report", map[string]string{"pending_count": fmt.Sprintf("%d", count)},
	)
	return nil
}
