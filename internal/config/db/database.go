package db

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"

	"github.com/Nakohartum/gophermart-loyalty-system/internal/model"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type PgDatabase struct {
	connection *pgx.Conn
	mu         sync.Mutex
	secretKey  string
}

const migrationsTableName = "schema_migrations"

func NewPgDatabase(secretKey string) *PgDatabase {
	return &PgDatabase{
		secretKey: secretKey,
	}
}

func (pg *PgDatabase) OpenConnection(ctx context.Context, databaseUri string) error {
	conn, err := pgx.Connect(ctx, databaseUri)
	if err != nil {
		return err
	}
	pg.connection = conn

	if err := pg.runMigrations(ctx, "..\\..\\migrations"); err != nil {
		return err
	}
	return err
}

func (pg *PgDatabase) CloseConnection(ctx context.Context) error {
	pg.mu.Lock()
	defer pg.mu.Unlock()
	if pg.connection == nil {
		return ErrNoConnectionToClose
	}
	return pg.connection.Close(ctx)
}

func (pg *PgDatabase) CheckConnection(ctx context.Context) error {
	return pg.connection.Ping(ctx)
}

func (pg *PgDatabase) RegisterUser(ctx context.Context, login, password string) (int64, error) {
	pass, err := hashPassword(pg.secretKey, password)
	if err != nil {
		return 0, err
	}
	tx, err := pg.connection.Begin(ctx)
	if err != nil {
		return 0, err
	}
	var userId int64
	err = tx.QueryRow(ctx, "INSERT INTO \"users\" (\"login\", \"password_hash\") VALUES($1, $2) RETURNING \"id\"", login, pass).Scan(&userId)
	if err != nil {
		_ = tx.Rollback(ctx)

		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return 0, ErrUserAlreadyExists
		}

		return 0, err
	}
	return userId, tx.Commit(ctx)
}



func (pg *PgDatabase) AuthenticateUser(ctx context.Context, login, password string) (int64, error) {
	var authResponse model.AuthResponse

	err := pg.connection.QueryRow(ctx, "SELECT \"id\", \"password_hash\" FROM \"users\" WHERE login = $1", login).Scan(&authResponse.UserId, &authResponse.Password)
	if err != nil {
		return 0, err
	}
	inputHash, err := hashPassword(pg.secretKey, password)
	if err != nil {
		return 0, err
	}

	if authResponse.Password != inputHash {
		return 0, ErrPasswordNotMatch
	}
	return authResponse.UserId, nil
}

func (pg *PgDatabase) CreateOrder(ctx context.Context, order model.Order, userId int64) (string, error) {
	_, err := pg.connection.Exec(ctx, "INSERT INTO \"orders\" (\"number\", \"user_id\", \"status\", \"accrual\", \"uploaded_at\") "+
		"VALUES ($1, $2, $3, $4, $5)", order.Number, userId, order.Status, order.Accrual, order.UploadedAt)
	if err == nil {
		return order.Number, nil
	}
	
	var pgErr *pgconn.PgError

	if errors.As(err, &pgErr) && pgErr.Code == "23505" {
		var existingUserId int64

		err = pg.connection.QueryRow(ctx, `SELECT user_id FROM orders WHERE number = $1`, order.Number).Scan(&existingUserId)
		if err != nil {
			return "", err
		}
		if existingUserId == userId {
			return "", ErrOrderAlreadyUploadedByUser
		}
		return "", ErrOrderAlreadyUploadedByAnotherUser
	}
	return "", err
}

func (pg *PgDatabase) GetOrdersForProcessing(ctx context.Context) ([]model.Order, error) {
	rows, err := pg.connection.Query(
		ctx,
		`SELECT user_id, number, status, accrual, uploaded_at
		 FROM orders
		 WHERE status IN ($1, $2)
		 ORDER BY uploaded_at ASC`,
		model.NEW,
		model.PROCESSING,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	orders := make([]model.Order, 0)
	for rows.Next() {
		var order model.Order
		if err := rows.Scan(&order.UserID, &order.Number, &order.Status, &order.Accrual, &order.UploadedAt); err != nil {
			return nil, err
		}
		orders = append(orders, order)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return orders, nil
}

func (pg *PgDatabase) UpdateOrder(ctx context.Context, userId int64, number string, status model.Status, accrual *float64) error {
	var prevStatus model.Status
	tx, err := pg.connection.Begin(ctx)
	if err != nil {
		return err
	}
	row := tx.QueryRow(ctx, "SELECT status FROM orders WHERE number = $1 AND user_id = $2", number, userId)
	err = row.Scan(&prevStatus)
	if err != nil {
		return ErrOrderNotFound
	}
	if prevStatus == model.PROCESSED || prevStatus == model.INVALID {
		tx.Rollback(ctx)
		return ErrOrderAlreadyProcessed
	}

	tag, err := tx.Exec(ctx, `UPDATE orders SET status = $1, accrual = $2 WHERE number = $3 AND user_id = $4`, status, accrual, number, userId)
	if err != nil {
		tx.Rollback(ctx)
		return err
	}
	if tag.RowsAffected() == 0 {
		tx.Rollback(ctx)
		return ErrOrderNotFound
	}
	if status == model.PROCESSED && accrual != nil {
		tag, err = tx.Exec(ctx, `UPDATE users SET current_balance = current_balance + $1 WHERE id = $2`, accrual, userId)
		if err != nil {
			tx.Rollback(ctx)
			return err
		}
		if tag.RowsAffected() == 0 {
			tx.Rollback(ctx)
			return ErrUserNotFound
		}
	}
	return tx.Commit(ctx)
}

func (pg *PgDatabase) GetListOfUploadedOrders(ctx context.Context, userId int64) []model.OrderResponse {
	rows, err := pg.connection.Query(ctx, `SELECT number, status, accrual, uploaded_at FROM orders WHERE user_id = $1 ORDER BY uploaded_at ASC`, userId)
	if err != nil {
		return nil
	}
	defer rows.Close()

	orders := make([]model.OrderResponse, 0)
	for rows.Next() {
		var odrder model.OrderResponse
		if err := rows.Scan(&odrder.Number, &odrder.Status, &odrder.Accrual, &odrder.UploadedAt); err != nil {
			return nil
		}
		orders = append(orders, odrder)
	}
	return orders
}

func hashPassword(secretKey, password string) (string, error) {
	h := hmac.New(sha256.New, []byte(secretKey))
	_, err := h.Write([]byte(password))
	if err != nil {
		return "", err
	}
	pass := hex.EncodeToString(h.Sum(nil))
	return pass, nil
}

func (pg *PgDatabase) runMigrations(ctx context.Context, dir string) error {
	if pg.connection == nil {
		return ErrNoConnectionToRunMigrations
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
