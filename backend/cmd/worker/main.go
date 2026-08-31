// Command worker runs only the background scheduler (attendance reminders,
// end-of-day report, salary reminders, fee reports) as its own process,
// for deployments that want to scale/restart it independently of the API.
// Running the API alone is also fine — it starts the same scheduler
// in-process — this binary is purely optional.
package main

import (
	"context"
	"log"
	"os/signal"
	"syscall"

	"attendance-backend/internal/config"
	"attendance-backend/internal/db"
	"attendance-backend/internal/modules/classes"
	"attendance-backend/internal/modules/fees"
	"attendance-backend/internal/modules/notifications"
	"attendance-backend/internal/modules/salary"
	notifinfra "attendance-backend/internal/notifications"
	"attendance-backend/internal/scheduler"
)

func main() {
	cfg := config.Load()
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	pool, err := db.NewPool(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("database connection failed: %v", err)
	}
	defer pool.Close()

	var fcmClient *notifinfra.FCMClient
	if cfg.FCMProjectID != "" && cfg.FCMCredentialsFile != "" {
		fcmClient, err = notifinfra.NewFCMClient(cfg.FCMProjectID, cfg.FCMCredentialsFile)
		if err != nil {
			log.Printf("warning: FCM not configured: %v", err)
		}
	}

	classRepo := classes.NewClassRepo(pool)
	substitutionRepo := classes.NewSubstitutionRepo(pool)
	feesRepo := fees.NewRepo(pool)
	salaryRepo := salary.NewRepo(pool)
	notifRepo := notifications.NewRepo(pool)
	notifSvc := notifications.NewService(notifRepo, fcmClient)

	sched := scheduler.New(pool, classRepo, substitutionRepo, feesRepo, salaryRepo, notifSvc, cfg.AttendanceReminderMin)
	sched.Run(ctx)
}
