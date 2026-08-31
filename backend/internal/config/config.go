package config

import (
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	Env                   string
	Port                  string
	DatabaseURL           string
	JWTAccessSecret       string
	JWTRefreshSecret      string
	AccessTokenTTL        time.Duration
	RefreshTokenTTL       time.Duration
	FCMServerKey          string
	FCMProjectID          string
	FCMCredentialsFile    string
	GeofenceDefaultRadius int
	AttendanceReminderMin int
}

func Load() *Config {
	_ = godotenv.Load()

	return &Config{
		Env:                   getEnv("APP_ENV", "development"),
		Port:                  getEnv("PORT", "8080"),
		DatabaseURL:           getEnv("DATABASE_URL", ""),
		JWTAccessSecret:       getEnv("JWT_ACCESS_SECRET", ""),
		JWTRefreshSecret:      getEnv("JWT_REFRESH_SECRET", ""),
		AccessTokenTTL:        time.Duration(getEnvInt("JWT_ACCESS_TTL_MIN", 15)) * time.Minute,
		RefreshTokenTTL:       time.Duration(getEnvInt("JWT_REFRESH_TTL_DAYS", 30)) * 24 * time.Hour,
		FCMServerKey:          getEnv("FCM_SERVER_KEY", ""),
		FCMProjectID:          getEnv("FCM_PROJECT_ID", ""),
		FCMCredentialsFile:    getEnv("FCM_CREDENTIALS_FILE", ""),
		GeofenceDefaultRadius: getEnvInt("GEOFENCE_DEFAULT_RADIUS_METERS", 100),
		AttendanceReminderMin: getEnvInt("ATTENDANCE_REMINDER_MINUTES", 15),
	}
}

func getEnv(key, fallback string) string {
	if v, ok := os.LookupEnv(key); ok && v != "" {
		return v
	}
	return fallback
}

func getEnvInt(key string, fallback int) int {
	if v, ok := os.LookupEnv(key); ok {
		if i, err := strconv.Atoi(v); err == nil {
			return i
		}
	}
	return fallback
}
