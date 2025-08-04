package infrastructure

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/elotusteam/microservice-project/shared/config"
	"github.com/elotusteam/microservice-project/shared/data"
	_ "github.com/lib/pq" // PostgreSQL driver
)

// PostgreSQLWrapper wraps sql.DB to implement data.DatabaseConnection
type PostgreSQLWrapper struct {
	db *sql.DB
}

// NewPostgreSQLConnection creates a new PostgreSQL connection
func NewPostgreSQLConnection(cfg *config.DatabaseConfig) (data.DatabaseConnection, error) {
	connStr := cfg.GetConnectionString()
	db, err := sql.Open(cfg.Driver, connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Configure connection pool
	db.SetMaxOpenConns(cfg.MaxOpenConnections)
	db.SetMaxIdleConns(cfg.MaxIdleConnections)
	db.SetConnMaxLifetime(cfg.ConnectionLifetime)

	// Test the connection
	ctx, cancel := context.WithTimeout(context.Background(), cfg.ConnectionTimeout)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return &PostgreSQLWrapper{db: db}, nil
}

// Query executes a query and returns rows
func (w *PostgreSQLWrapper) Query(ctx context.Context, query string, args ...interface{}) (data.Rows, error) {
	rows, err := w.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	return &RowsWrapper{rows: rows}, nil
}

// QueryRow executes a query and returns a single row
func (w *PostgreSQLWrapper) QueryRow(ctx context.Context, query string, args ...interface{}) data.Row {
	row := w.db.QueryRowContext(ctx, query, args...)
	return &RowWrapper{row: row}
}

// Exec executes a query without returning rows
func (w *PostgreSQLWrapper) Exec(ctx context.Context, query string, args ...interface{}) (data.Result, error) {
	result, err := w.db.ExecContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	return &ResultWrapper{result: result}, nil
}

// Prepare prepares a statement
func (w *PostgreSQLWrapper) Prepare(ctx context.Context, query string) (data.Statement, error) {
	stmt, err := w.db.PrepareContext(ctx, query)
	if err != nil {
		return nil, err
	}
	return &StatementWrapper{stmt: stmt}, nil
}

// Begin starts a transaction
func (w *PostgreSQLWrapper) Begin(ctx context.Context) (data.Transaction, error) {
	tx, err := w.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	return &TransactionWrapper{tx: tx}, nil
}

// Ping pings the database
func (w *PostgreSQLWrapper) Ping(ctx context.Context) error {
	return w.db.PingContext(ctx)
}

// Close closes the connection
func (w *PostgreSQLWrapper) Close() error {
	return w.db.Close()
}

// RowsWrapper wraps sql.Rows to implement data.Rows
type RowsWrapper struct {
	rows *sql.Rows
}

func (r *RowsWrapper) Next() bool {
	return r.rows.Next()
}

func (r *RowsWrapper) Scan(dest ...interface{}) error {
	return r.rows.Scan(dest...)
}

func (r *RowsWrapper) Close() error {
	return r.rows.Close()
}

func (r *RowsWrapper) Err() error {
	return r.rows.Err()
}

func (r *RowsWrapper) Columns() ([]string, error) {
	return r.rows.Columns()
}

// RowWrapper wraps sql.Row to implement data.Row
type RowWrapper struct {
	row *sql.Row
}

func (r *RowWrapper) Scan(dest ...interface{}) error {
	return r.row.Scan(dest...)
}

// ResultWrapper wraps sql.Result to implement data.Result
type ResultWrapper struct {
	result sql.Result
}

func (r *ResultWrapper) LastInsertId() (int64, error) {
	return r.result.LastInsertId()
}

func (r *ResultWrapper) RowsAffected() (int64, error) {
	return r.result.RowsAffected()
}

// StatementWrapper wraps sql.Stmt to implement data.Statement
type StatementWrapper struct {
	stmt *sql.Stmt
}

func (s *StatementWrapper) Query(ctx context.Context, args ...interface{}) (data.Rows, error) {
	rows, err := s.stmt.QueryContext(ctx, args...)
	if err != nil {
		return nil, err
	}
	return &RowsWrapper{rows: rows}, nil
}

func (s *StatementWrapper) QueryRow(ctx context.Context, args ...interface{}) data.Row {
	row := s.stmt.QueryRowContext(ctx, args...)
	return &RowWrapper{row: row}
}

func (s *StatementWrapper) Exec(ctx context.Context, args ...interface{}) (data.Result, error) {
	result, err := s.stmt.ExecContext(ctx, args...)
	if err != nil {
		return nil, err
	}
	return &ResultWrapper{result: result}, nil
}

func (s *StatementWrapper) Close() error {
	return s.stmt.Close()
}

// TransactionWrapper wraps sql.Tx to implement data.Transaction and data.DatabaseConnection
type TransactionWrapper struct {
	tx *sql.Tx
}

func (t *TransactionWrapper) Commit() error {
	return t.tx.Commit()
}

func (t *TransactionWrapper) Rollback() error {
	return t.tx.Rollback()
}

func (t *TransactionWrapper) Context() context.Context {
	// sql.Tx doesn't have a context method, so we return background
	return context.Background()
}

// Implement DatabaseConnection interface for TransactionWrapper
func (t *TransactionWrapper) Query(ctx context.Context, query string, args ...interface{}) (data.Rows, error) {
	rows, err := t.tx.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	return &RowsWrapper{rows: rows}, nil
}

func (t *TransactionWrapper) QueryRow(ctx context.Context, query string, args ...interface{}) data.Row {
	row := t.tx.QueryRowContext(ctx, query, args...)
	return &RowWrapper{row: row}
}

func (t *TransactionWrapper) Exec(ctx context.Context, query string, args ...interface{}) (data.Result, error) {
	result, err := t.tx.ExecContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	return &ResultWrapper{result: result}, nil
}

func (t *TransactionWrapper) Prepare(ctx context.Context, query string) (data.Statement, error) {
	stmt, err := t.tx.PrepareContext(ctx, query)
	if err != nil {
		return nil, err
	}
	return &StatementWrapper{stmt: stmt}, nil
}

func (t *TransactionWrapper) Begin(ctx context.Context) (data.Transaction, error) {
	// Nested transactions are not supported in PostgreSQL
	return nil, fmt.Errorf("nested transactions are not supported")
}

func (t *TransactionWrapper) Ping(ctx context.Context) error {
	// Transactions don't support ping, but we can execute a simple query
	_, err := t.tx.ExecContext(ctx, "SELECT 1")
	return err
}

func (t *TransactionWrapper) Close() error {
	// Transactions are closed via Commit() or Rollback()
	return nil
}
