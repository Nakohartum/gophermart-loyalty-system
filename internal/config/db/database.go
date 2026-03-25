package db

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"

	"github.com/jackc/pgx/v5"
)

type PgDatabase struct {
	connection *pgx.Conn
	mu         sync.Mutex
}

const migrationsTableName = "schema_migrations"

func NewPgDatabase() *PgDatabase {
	return &PgDatabase{}
}

func (pg *PgDatabase) OpenConnection(ctx context.Context, databaseUri string) error {
	conn, err := pgx.Connect(ctx, databaseUri)
	if err != nil {
		return err
	}
	pg.connection = conn

	if err := pg.runMigrations(ctx, "migrations"); err != nil {
		return err
	}
	return err
}

func (pg *PgDatabase) runMigrations(ctx context.Context, dir string) error {
	if pg.connection == nil {
		return errNoConnectionToRunMigrations
	}
	if err := pg.ensureMigrationsTable(ctx); err != nil {
		return err
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		return err
	}

	files := make([]string, 0)

	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		name := e.Name()

		if strings.HasSuffix(name, ".up.sql") {
			files = append(files, filepath.Join(dir, name))
		}
	}
	sort.Strings(files)

	appliedVersions, err := pg.appliedMigrationVersions(ctx)
	if err != nil {
		return err
	}

	for _, f := range files {
		version := migrationVersionFromFilename(filepath.Base(f))
		if version == "" {
			return fmt.Errorf("invalid migration filename %s", f)
		}
		if _, ok := appliedVersions[version]; ok {
			continue
		}

		sqlBytes, err := os.ReadFile(f)
		if err != nil {
			return fmt.Errorf("read migration %s: %w", f, err)
		}
		tx, err := pg.connection.Begin(ctx)
		if err != nil {
			return fmt.Errorf("begin migration %s: %w", f, err)
		}

		if _, err := tx.Exec(ctx, string(sqlBytes)); err != nil {
			_ = tx.Rollback(ctx)
			return fmt.Errorf("apply migration %s: %w", f, err)
		}

		if _, err := tx.Exec(
			ctx,
			fmt.Sprintf("INSERT INTO %s (version, name) VALUES ($1, $2)", migrationsTableName),
			version,
			filepath.Base(f),
		); err != nil {
			_ = tx.Rollback(ctx)
			return fmt.Errorf("save migration %s version: %w", f, err)
		}

		if err := tx.Commit(ctx); err != nil {
			return fmt.Errorf("commit migration %s: %w", f, err)
		}
	}
	return nil
}

func (pg *PgDatabase) ensureMigrationsTable(ctx context.Context) error {
	query := fmt.Sprintf(`
		CREATE TABLE IF NOT EXISTS %s (
			version TEXT PRIMARY KEY,
			name TEXT NOT NULL,
			applied_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		)
	`, migrationsTableName)

	if _, err := pg.connection.Exec(ctx, query); err != nil {
		return fmt.Errorf("ensure migrations table: %w", err)
	}
	return nil
}

func (pg *PgDatabase) appliedMigrationVersions(ctx context.Context) (map[string]struct{}, error) {
	rows, err := pg.connection.Query(ctx, fmt.Sprintf("SELECT version FROM %s", migrationsTableName))
	if err != nil {
		return nil, fmt.Errorf("load applied migrations: %w", err)
	}
	defer rows.Close()

	versions := make(map[string]struct{})
	for rows.Next() {
		var version string
		if err := rows.Scan(&version); err != nil {
			return nil, fmt.Errorf("scan applied migration version: %w", err)
		}
		versions[version] = struct{}{}
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate applied migrations: %w", err)
	}

	return versions, nil
}

func migrationVersionFromFilename(name string) string {
	parts := strings.SplitN(name, "_", 2)
	if len(parts) == 0 {
		return ""
	}
	return parts[0]
}

func (pg *PgDatabase) CloseConnection(ctx context.Context) error {
	pg.mu.Lock()
	defer pg.mu.Unlock()
	if pg.connection == nil {
		return errNoConnectionToClose
	}
	return pg.connection.Close(ctx)
}

func (pg *PgDatabase) CheckConnection(ctx context.Context) error {
	return pg.connection.Ping(ctx)
}
