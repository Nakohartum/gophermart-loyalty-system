package postgres

import (
	"context"
	"testing"

	"github.com/Nakohartum/gophermart-loyalty-system/internal/apperrors"
)

func TestTextNewDatabase(t *testing.T) {
	testCases := []struct {
		name      string
		secretKey string
	}{
		{name: "creates database", secretKey: "secret"},
		{name: "empty secret", secretKey: ""},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			database := New(tc.secretKey)
			if database == nil {
				t.Fatal("New returned nil")
			}
			if database.secretKey != tc.secretKey {
				t.Fatalf("secretKey = %q, want %q", database.secretKey, tc.secretKey)
			}
		})
	}
}

func TestTextCloseConnection(t *testing.T) {
	testCases := []struct {
		name        string
		connection  *Database
		expectedErr error
	}{
		{name: "no connection", connection: &Database{}, expectedErr: apperrors.ErrNoConnectionToClose},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.connection.CloseConnection(context.Background())
			if err != tc.expectedErr {
				t.Fatalf("CloseConnection error = %v, want %v", err, tc.expectedErr)
			}
		})
	}
}

func TestTextHashPassword(t *testing.T) {
	testCases := []struct {
		name      string
		secretKey string
		password  string
	}{
		{name: "normal values", secretKey: "secret", password: "password"},
		{name: "empty password", secretKey: "secret", password: ""},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			hash1, err := hashPassword(tc.secretKey, tc.password)
			if err != nil {
				t.Fatalf("hashPassword returned error: %v", err)
			}
			hash2, err := hashPassword(tc.secretKey, tc.password)
			if err != nil {
				t.Fatalf("hashPassword returned error: %v", err)
			}

			if hash1 == "" {
				t.Fatal("hashPassword returned empty hash")
			}
			if hash1 != hash2 {
				t.Fatalf("expected deterministic hash, got %q and %q", hash1, hash2)
			}
		})
	}
}

func TestTextRunMigrations(t *testing.T) {
	testCases := []struct {
		name        string
		database    *Database
		expectedErr error
	}{
		{name: "no connection", database: &Database{}, expectedErr: apperrors.ErrNoConnectionToRunMigrations},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.database.runMigrations(context.Background(), "migrations")
			if err != tc.expectedErr {
				t.Fatalf("runMigrations error = %v, want %v", err, tc.expectedErr)
			}
		})
	}
}

func TestTextMigrationVersionFromFilename(t *testing.T) {
	testCases := []struct {
		name     string
		filename string
		expected string
	}{
		{name: "valid filename", filename: "000001_create_user_table.up.sql", expected: "000001"},
		{name: "without underscore", filename: "invalid.sql", expected: "invalid.sql"},
		{name: "empty filename", filename: "", expected: ""},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if got := migrationVersionFromFilename(tc.filename); got != tc.expected {
				t.Fatalf("migrationVersionFromFilename(%q) = %q, want %q", tc.filename, got, tc.expected)
			}
		})
	}
}
