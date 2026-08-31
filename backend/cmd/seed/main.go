// Command seed creates the first admin account. There is no public signup
// endpoint by design (every account is provisioned by an existing admin),
// so this bootstraps the very first one directly against the database.
//
// Usage:
//
//	go run ./cmd/seed -name "Admin" -email admin@example.com -password "changeme123"
package main

import (
	"context"
	"flag"
	"log"

	"attendance-backend/internal/auth"
	"attendance-backend/internal/config"
	"attendance-backend/internal/db"
)

func main() {
	name := flag.String("name", "Admin", "admin display name")
	email := flag.String("email", "", "admin email (required)")
	password := flag.String("password", "", "admin password, at least 8 characters (required)")
	flag.Parse()

	if *email == "" || len(*password) < 8 {
		log.Fatal("usage: go run ./cmd/seed -email you@example.com -password yourpassword (8+ chars)")
	}

	cfg := config.Load()
	ctx := context.Background()

	pool, err := db.NewPool(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("database connection failed: %v", err)
	}
	defer pool.Close()

	if err := db.RunMigrations(ctx, pool, "migrations"); err != nil {
		log.Fatalf("migrations failed: %v", err)
	}

	hash, err := auth.HashPassword(*password)
	if err != nil {
		log.Fatalf("hash password: %v", err)
	}

	var id string
	err = pool.QueryRow(ctx, `
		INSERT INTO users (role_id, name, email, password_hash)
		SELECT id, $1, $2, $3 FROM roles WHERE name = 'admin'
		RETURNING id
	`, *name, *email, hash).Scan(&id)
	if err != nil {
		log.Fatalf("create admin (does this email already exist?): %v", err)
	}

	log.Printf("admin account created: %s (%s)", *name, *email)
}
