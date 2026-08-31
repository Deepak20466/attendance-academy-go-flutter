package main

import (
	"context"
	"log"
	"net/http"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"
	"github.com/google/uuid"

	"attendance-backend/internal/audit"
	"attendance-backend/internal/auth"
	"attendance-backend/internal/config"
	"attendance-backend/internal/db"
	"attendance-backend/internal/httpapi"
	appmw "attendance-backend/internal/middleware"
	notifinfra "attendance-backend/internal/notifications"
	"attendance-backend/internal/scheduler"

	"attendance-backend/internal/modules/activities"
	analyticsmod "attendance-backend/internal/modules/analytics"
	"attendance-backend/internal/modules/attendance"
	auditmod "attendance-backend/internal/modules/audit"
	"attendance-backend/internal/modules/authapi"
	"attendance-backend/internal/modules/classes"
	"attendance-backend/internal/modules/coaches"
	"attendance-backend/internal/modules/fees"
	"attendance-backend/internal/modules/leaves"
	"attendance-backend/internal/modules/locations"
	notifmod "attendance-backend/internal/modules/notifications"
	"attendance-backend/internal/modules/salary"
	"attendance-backend/internal/modules/students"
	"attendance-backend/internal/modules/users"
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

	migrationsDir := filepath.Join("migrations")
	if err := db.RunMigrations(ctx, pool, migrationsDir); err != nil {
		log.Fatalf("migrations failed: %v", err)
	}

	tokenManager := auth.NewTokenManager(cfg.JWTAccessSecret, cfg.JWTRefreshSecret, cfg.AccessTokenTTL, cfg.RefreshTokenTTL)
	auditLogger := audit.NewLogger(pool)

	// ---- FCM (optional: nil client degrades gracefully to in-app-only notifications) ----
	var fcmClient *notifinfra.FCMClient
	if cfg.FCMProjectID != "" && cfg.FCMCredentialsFile != "" {
		fcmClient, err = notifinfra.NewFCMClient(cfg.FCMProjectID, cfg.FCMCredentialsFile)
		if err != nil {
			log.Printf("warning: FCM not configured: %v", err)
		}
	}

	// ---- repos ----
	usersRepo := users.NewRepo(pool)
	authRepo := authapi.NewRepo(pool)
	activitiesRepo := activities.NewRepo(pool)
	coachesRepo := coaches.NewRepo(pool)
	studentsRepo := students.NewRepo(pool)
	batchRepo := classes.NewBatchRepo(pool)
	classRepo := classes.NewClassRepo(pool)
	substitutionRepo := classes.NewSubstitutionRepo(pool)
	studentAttendanceRepo := attendance.NewStudentAttendanceRepo(pool)
	coachAttendanceRepo := attendance.NewCoachAttendanceRepo(pool)
	leavesRepo := leaves.NewRepo(pool)
	locationsRepo := locations.NewRepo(pool)
	salaryRepo := salary.NewRepo(pool)
	feesRepo := fees.NewRepo(pool)
	notifRepo := notifmod.NewRepo(pool)
	auditRepo := auditmod.NewRepo(pool)
	analyticsRepo := analyticsmod.NewRepo(pool)

	// ---- services ----
	authSvc := authapi.NewService(authRepo, tokenManager)
	usersSvc := users.NewService(usersRepo)
	activitiesSvc := activities.NewService(activitiesRepo)
	coachesSvc := coaches.NewServiceWithUserCreation(coachesRepo, func(ctx context.Context, name, email, phone, hash string) (uuid.UUID, error) {
		return usersRepo.Create(ctx, "coach", name, email, phone, hash)
	})
	notifSvc := notifmod.NewService(notifRepo, fcmClient)

	batchSvc := classes.NewBatchService(batchRepo)
	classSvc := classes.NewClassService(classRepo, pool)
	substitutionSvc := classes.NewSubstitutionService(substitutionRepo, classRepo, auditLogger)
	studentsSvc := students.NewService(studentsRepo)
	studentAttendanceSvc := attendance.NewStudentService(studentAttendanceRepo, pool, classRepo)
	coachAttendanceSvc := attendance.NewCoachService(coachAttendanceRepo, pool)
	leavesSvc := leaves.NewService(leavesRepo, auditLogger)
	locationsSvc := locations.NewService(locationsRepo)
	salarySvc := salary.NewService(salaryRepo)
	feesSvc := fees.NewService(feesRepo)
	analyticsSvc := analyticsmod.NewService(analyticsRepo)

	// ---- handlers ----
	authHandler := authapi.NewHandler(authSvc)
	usersHandler := users.NewHandler(usersSvc)
	activitiesHandler := activities.NewHandler(activitiesSvc)
	coachesHandler := coaches.NewHandler(coachesSvc)
	batchHandler := classes.NewBatchHandler(batchSvc)
	classHandler := classes.NewClassHandler(classSvc)
	substitutionHandler := classes.NewSubstitutionHandler(substitutionSvc)
	studentsHandler := students.NewHandler(studentsSvc)
	studentAttendanceHandler := attendance.NewStudentHandler(studentAttendanceSvc)
	coachAttendanceHandler := attendance.NewCoachHandler(coachAttendanceSvc)
	leavesHandler := leaves.NewHandler(leavesSvc)
	locationsHandler := locations.NewHandler(locationsSvc)
	salaryHandler := salary.NewHandler(salarySvc)
	feesHandler := fees.NewHandler(feesSvc)
	notifHandler := notifmod.NewHandler(notifSvc)
	auditHandler := auditmod.NewHandler(auditRepo)
	analyticsHandler := analyticsmod.NewHandler(analyticsSvc)

	// ---- background scheduler ----
	sched := scheduler.New(pool, classRepo, substitutionRepo, feesRepo, salaryRepo, notifSvc, cfg.AttendanceReminderMin)
	go sched.Run(ctx)

	// ---- router ----
	r := chi.NewRouter()
	r.Use(chimw.RequestID)
	r.Use(chimw.RealIP)
	r.Use(chimw.Logger)
	r.Use(chimw.Recoverer)
	r.Use(appmw.CORS())
	r.Use(appmw.RateLimit)

	r.Route("/api/v1", func(api chi.Router) {
		api.Route("/auth", func(a chi.Router) {
			a.With(appmw.AuthRateLimit).Post("/login", authHandler.Login)
			a.With(appmw.AuthRateLimit).Post("/refresh", authHandler.Refresh)
			a.Post("/logout", authHandler.Logout)

			a.Group(func(ar chi.Router) {
				ar.Use(appmw.RequireAuth(tokenManager, pool))
				ar.Get("/me", authHandler.Me)
				ar.Put("/fcm-token", authHandler.UpdateFCMToken)
			})
		})

		api.Group(func(protected chi.Router) {
			protected.Use(appmw.RequireAuth(tokenManager, pool))

			protected.Put("/users/me/password", usersHandler.ChangePassword)
			protected.Group(func(admin chi.Router) {
				admin.Use(appmw.RequireAdmin)
				admin.Post("/users", usersHandler.CreateAdmin)
				admin.Get("/users", usersHandler.List)
				admin.Post("/users/{id}/activate", usersHandler.Activate)
				admin.Post("/users/{id}/deactivate", usersHandler.Deactivate)
			})

			protected.Get("/activities", activitiesHandler.List)
			protected.Get("/activities/{id}", activitiesHandler.Get)
			protected.Group(func(admin chi.Router) {
				admin.Use(appmw.RequireAdmin)
				admin.Post("/activities", activitiesHandler.Create)
				admin.Put("/activities/{id}", activitiesHandler.Update)
			})

			protected.Get("/coaches/me", coachesHandler.Me)
			protected.Get("/coaches/{id}", coachesHandler.Get)
			protected.Group(func(admin chi.Router) {
				admin.Use(appmw.RequireAdmin)
				admin.Post("/coaches", coachesHandler.Create)
				admin.Get("/coaches", coachesHandler.List)
				admin.Put("/coaches/{id}/activities", coachesHandler.SetActivities)
				admin.Put("/coaches/{id}/active", coachesHandler.SetActive)
				admin.Put("/coaches/{id}/salary", coachesHandler.UpdateSalary)
			})

			protected.Get("/locations", locationsHandler.ListByActivity)
			protected.Group(func(admin chi.Router) {
				admin.Use(appmw.RequireAdmin)
				admin.Post("/locations", locationsHandler.Create)
				admin.Put("/locations/{id}", locationsHandler.Update)
				admin.Delete("/locations/{id}", locationsHandler.Delete)
			})

			protected.Get("/batches", batchHandler.List)
			protected.Get("/batches/{id}", batchHandler.Get)
			protected.Group(func(admin chi.Router) {
				admin.Use(appmw.RequireAdmin)
				admin.Post("/batches", batchHandler.Create)
				admin.Put("/batches/{id}", batchHandler.Update)
			})

			protected.Get("/classes", classHandler.List)
			protected.Get("/classes/{id}", classHandler.Get)
			protected.Get("/classes/{id}/roster", classHandler.Roster)
			protected.With(appmw.RequireAdmin).Post("/classes", classHandler.Create)

			protected.Get("/substitutions/mine", substitutionHandler.ListMine)
			protected.Group(func(admin chi.Router) {
				admin.Use(appmw.RequireAdmin)
				admin.Post("/substitutions", substitutionHandler.Create)
				admin.Post("/substitutions/{id}/cancel", substitutionHandler.Cancel)
				admin.Get("/substitutions", substitutionHandler.ListAll)
			})

			protected.Get("/students", studentsHandler.List)
			protected.Get("/students/{id}", studentsHandler.Get)
			protected.Post("/students", studentsHandler.Create)
			protected.Put("/students/{id}", studentsHandler.Update)

			protected.Post("/attendance/students/mark", studentAttendanceHandler.MarkBulk)
			protected.Get("/attendance/students/class/{classId}", studentAttendanceHandler.ListForClass)
			protected.Get("/attendance/students/student/{studentId}/history", studentAttendanceHandler.History)
			protected.Get("/attendance/students/student/{studentId}/monthly", studentAttendanceHandler.MonthlyPercentage)

			protected.Post("/attendance/coaches/check-in", coachAttendanceHandler.CheckIn)
			protected.Post("/attendance/coaches/check-out", coachAttendanceHandler.CheckOut)
			protected.Get("/attendance/coaches", coachAttendanceHandler.List)

			protected.Post("/leaves", leavesHandler.Apply)
			protected.Post("/leaves/{id}/cancel", leavesHandler.Cancel)
			protected.Get("/leaves/mine", leavesHandler.ListMine)
			protected.Group(func(admin chi.Router) {
				admin.Use(appmw.RequireAdmin)
				admin.Post("/leaves/{id}/approve", leavesHandler.Approve)
				admin.Post("/leaves/{id}/reject", leavesHandler.Reject)
				admin.Get("/leaves", leavesHandler.ListAll)
			})

			protected.Post("/salary/{id}/acknowledge", salaryHandler.Acknowledge)
			protected.Get("/salary/mine", salaryHandler.ListMine)
			protected.Group(func(admin chi.Router) {
				admin.Use(appmw.RequireAdmin)
				admin.Post("/salary/generate", salaryHandler.Generate)
				admin.Get("/salary", salaryHandler.ListForPeriod)
			})

			protected.Get("/fees", feesHandler.List)
			protected.Get("/fees/{id}", feesHandler.Get)
			protected.Group(func(admin chi.Router) {
				admin.Use(appmw.RequireAdmin)
				admin.Post("/fees", feesHandler.Create)
				admin.Post("/fees/generate", feesHandler.Generate)
				admin.Post("/fees/{id}/paid", feesHandler.MarkPaid)
				admin.Get("/fees-pending-summary", feesHandler.PendingSummary)
			})

			protected.Get("/notifications", notifHandler.List)
			protected.Post("/notifications/{id}/read", notifHandler.MarkRead)

			protected.Group(func(admin chi.Router) {
				admin.Use(appmw.RequireAdmin)
				admin.Get("/analytics/overall", analyticsHandler.Overall)
				admin.Get("/analytics/activity", analyticsHandler.ActivitySummary)
				admin.Get("/analytics/perfect-attendance", analyticsHandler.PerfectAttendance)
				admin.Get("/analytics/monthly-report", analyticsHandler.MonthlyReport)
				admin.Get("/analytics/monthly-report.csv", analyticsHandler.MonthlyReportCSV)
				admin.Get("/audit-logs", auditHandler.List)
			})
		})
	})

	r.Get("/health", func(w http.ResponseWriter, req *http.Request) {
		httpapi.JSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})

	srv := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		log.Printf("attendance API listening on :%s", cfg.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server error: %v", err)
		}
	}()

	<-ctx.Done()
	log.Println("shutting down...")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	_ = srv.Shutdown(shutdownCtx)
}
