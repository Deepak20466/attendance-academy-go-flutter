# Academy Attendance & Management System

Full-stack attendance/academy management app: Go REST API + PostgreSQL backend, Flutter mobile app.

## Layout

```
backend/    Go API, migrations, background scheduler (internal/, cmd/)
app/        Flutter app (lib/core, lib/features/*)
```

## Backend setup

1. Install PostgreSQL and create a database:
   ```
   createdb attendance_db
   ```
2. Copy `backend/.env.example` to `backend/.env` and fill in `DATABASE_URL`, `JWT_ACCESS_SECRET`, `JWT_REFRESH_SECRET` (use `openssl rand -hex 32` for the secrets). FCM settings are optional — leaving them blank disables push notifications but the app still records in-app notifications.
3. Fetch dependencies and run:
   ```
   cd backend
   go mod tidy
   go run ./cmd/api
   ```
   This applies all SQL migrations automatically on startup (tracked in a `schema_migrations` table).
4. Create the first admin account (there is no public signup endpoint by design — every account is provisioned by an existing admin):
   ```
   go run ./cmd/seed -email admin@example.com -password "changeme123" -name "Admin"
   ```
5. Optional: run the scheduler as its own process instead of in-process with the API:
   ```
   go run ./cmd/worker
   ```

The API listens on `:8080` by default (`PORT` in `.env`). Health check: `GET /health`.

### Backend structure

- `internal/authz` — the single choke point for activity-scoped authorization. Every service method touching student/coach/class/attendance data calls `AuthorizeActivity`, `AuthorizeClassAccess`, or `AuthorizeStudentAccess`. A coach's role and activity assignments are re-loaded from the database on **every request** (never trusted from the JWT), so revoking access takes effect immediately.
- `internal/modules/*` — one package per feature (auth, users, students, coaches, activities, classes, attendance, leaves, fees, salary, notifications, analytics, audit), each with `repo.go` / `service.go` / `handler.go`.
- `internal/scheduler` — in-process background jobs (no external cron/queue): 15-minute missing-attendance reminders, end-of-day report, 10th-of-month salary acknowledgement generation, end-of-month pending-fee report.
- `migrations/` — plain numbered `.up.sql`/`.down.sql` files, applied by a small embedded runner in `internal/db/migrate.go`.

## Flutter app setup

1. ```
   cd app
   flutter pub get
   ```
2. Point the app at your backend (defaults to `http://10.0.2.2:8080/api/v1`, i.e. localhost from the Android emulator):
   ```
   flutter run --dart-define=API_BASE_URL=http://YOUR_IP:8080/api/v1
   ```
3. **Push notifications (optional):** to enable FCM, run `flutterfire configure` to generate `firebase_options.dart` and add the platform config files (`google-services.json` / `GoogleService-Info.plist`). Without this, the app works fully except for push notifications — in-app notifications still populate via the `/notifications` list.
4. **Location permission:** the coach check-in/out screen requests location permission at runtime via `geolocator`/`permission_handler`. Make sure the Android/iOS permission entries these plugins require are present (the Flutter plugin installers add most of this automatically; see each plugin's README if you hit a permission-denial on a real device).

### Flutter structure

- `lib/core` — networking (`Dio` with JWT-attach + single-retry-on-401 refresh), secure token storage, theming, router (`go_router`, redirects based on auth state), shared widgets.
- `lib/features/*` — one folder per feature (`auth`, `dashboard`, `students`, `coaches`, `activities`, `classes`, `attendance`, `leaves`, `fees`, `salary`, `notifications`, `analytics`), each with `data/` (models + Riverpod-exposed repository) and `presentation/` (screens).
- Role-based navigation: `HomeShell` shows a different bottom nav and set of tabs for `admin` vs `coach` sessions; the backend enforces the real authorization regardless of what the app requests.

## Known environment note (Windows)

If your Windows username or install paths contain spaces (e.g. `C:\Users\K Deepak`), a couple of Flutter's newer plugins (`path_provider_android`/`path_provider_foundation`, pulled in transitively) use Dart's experimental "native assets" build hooks, which currently mis-handle spaces in paths on Windows and fail with `'C:\Users\...' is not recognized as an internal or external command`. This project's `pubspec.yaml` pins `flutter_secure_storage: 8.0.0` and `geolocator: 10.1.1` specifically to avoid pulling in that dependency chain — if you upgrade either package later and hit this error again, that's why.

## Security model summary

- JWT access tokens (short-lived) + rotating refresh tokens (hashed with SHA-256 before storage, revoked on use).
- Passwords hashed with bcrypt.
- Every list/detail endpoint for activity-scoped data filters by activity **in SQL**, not just in a pre-check — defense in depth against a bug in any single authorization layer.
- Substitute coaches are authorized only for the exact class they've been assigned to cover, via the `substitutions` table with `status = 'active'`, checked by `authz.AuthorizeClassAccess` on every attendance/check-in action.
- All admin actions that change state (substitutions, leave approvals) write to `audit_logs`.
- Rate limiting: 300 req/min general, 10 req/min on `/auth/login` and `/auth/refresh`.
