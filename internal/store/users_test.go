package store

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"testing"
	"time"

	"github.com/google/uuid"
	_ "github.com/mattn/go-sqlite3"
)

func setupTestDB(t *testing.T) *sql.DB {
	t.Helper()

	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("failed to open test database: %v", err)
	}

	queries := []string{
		`CREATE TABLE users (
			id TEXT PRIMARY KEY,
			username TEXT UNIQUE NOT NULL,
			email TEXT UNIQUE NOT NULL,
			password BLOB NOT NULL,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			is_active BOOLEAN NOT NULL DEFAULT FALSE,
			role_id INTEGER DEFAULT 1
		)`,
		`CREATE TABLE user_invitations (
			token TEXT PRIMARY KEY,
			id TEXT NOT NULL,
			expiry TIMESTAMP NOT NULL,
			FOREIGN KEY (id) REFERENCES users(id)
		)`,
	}

	for _, query := range queries {
		if _, err := db.Exec(query); err != nil {
			t.Fatalf("failed to create test table: %v", err)
		}
	}

	return db
}

func TestUsersPostgresStore_Activate(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	store := &UsersPostgresStore{db: db}
	ctx := context.Background()

	userID := uuid.New()
	token := "test-activation-token"
	hash := sha256.Sum256([]byte(token))
	hashToken := hex.EncodeToString(hash[:])

	_, err := db.Exec(`
		INSERT INTO users (id, username, email, password, is_active, created_at, updated_at) 
		VALUES (?, ?, ?, ?, ?, ?, ?)`,
		userID.String(), "testuser", "test@example.com", []byte("password"), false, time.Now(), time.Now())
	if err != nil {
		t.Fatalf("failed to insert test user: %v", err)
	}


	expiry := time.Now().Add(24 * time.Hour)
	_, err = db.Exec(`
		INSERT INTO user_invitations (token, id, expiry) 
		VALUES (?, ?, ?)`,
		hashToken, userID.String(), expiry)
	if err != nil {
		t.Fatalf("failed to insert test invitation: %v", err)
	}

	type args struct {
		ctx   context.Context
		token string
	}

	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "success",
			args: args{
				ctx:   ctx,
				token: token,
			},
			wantErr: false,
		},
		{
			name: "token not found",
			args: args{
				ctx:   ctx,
				token: "invalid-token",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := store.Activate(tt.args.ctx, tt.args.token); (err != nil) != tt.wantErr {
				t.Errorf("UsersPostgresStore.Activate() error = %v, wantErr %v", err, tt.wantErr)
			}

			if !tt.wantErr {
				var isActive bool
				err := db.QueryRow("SELECT is_active FROM users WHERE id = ?", userID.String()).Scan(&isActive)
				if err != nil {
					t.Fatalf("failed to query user status: %v", err)
				}
				if !isActive {
					t.Error("user should be active after successful activation")
				}

				var count int
				err = db.QueryRow("SELECT COUNT(*) FROM user_invitations WHERE id = ?", userID.String()).Scan(&count)
				if err != nil {
					t.Fatalf("failed to query invitations: %v", err)
				}
				if count != 0 {
					t.Error("invitation should be deleted after successful activation")
				}
			}
		})
	}
}
