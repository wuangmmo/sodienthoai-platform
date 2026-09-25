package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
	"golang.org/x/crypto/bcrypt"
)

func main() {
	dbURL := os.Getenv("DATABASE_URL")
	password := os.Getenv("ADMIN_PASSWORD")
	email := strings.ToLower(strings.TrimSpace(os.Getenv("ADMIN_EMAIL")))
	name := strings.TrimSpace(os.Getenv("ADMIN_NAME"))
	role := strings.TrimSpace(os.Getenv("ADMIN_ROLE"))
	if dbURL == "" { log.Fatal("DATABASE_URL is required") }
	if email == "" { log.Fatal("ADMIN_EMAIL is required") }
	if len(password) < 12 { log.Fatal("ADMIN_PASSWORD must be at least 12 characters") }
	if name == "" { name = "Administrator" }
	if role == "" { role = "super-admin" }
	if role != "admin" && role != "super-admin" { log.Fatal("ADMIN_ROLE must be admin or super-admin") }

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil { log.Fatal(err) }
	db, err := sql.Open("pgx", dbURL)
	if err != nil { log.Fatal(err) }
	defer db.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	var id string
	err = db.QueryRowContext(ctx, `INSERT INTO admin_users(name,email,password_hash,role,status)
		VALUES($1,$2,$3,$4,'active')
		ON CONFLICT (LOWER(email)) DO UPDATE SET
			name=EXCLUDED.name,password_hash=EXCLUDED.password_hash,role=EXCLUDED.role,status='active',updated_at=NOW()
		RETURNING id::text`, name,email,string(hash),role).Scan(&id)
	if err != nil { log.Fatal(err) }
	fmt.Printf("admin ready: %s (%s)\n", email, id)
}
