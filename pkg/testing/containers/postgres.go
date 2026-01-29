// Package containers provides test container implementations for integration testing.
package containers

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

// PostgresContainer represents a PostgreSQL test container configuration.
type PostgresContainer struct {
	Host       string
	Port       string
	Database   string
	User       string
	Password   string
	SSLMode    string
	DB         *sqlx.DB
	migrations []string
}

// PostgresContainerConfig holds configuration for PostgreSQL container.
type PostgresContainerConfig struct {
	Database       string
	User           string
	Password       string
	MigrationsPath string
}

// DefaultPostgresConfig returns default PostgreSQL configuration.
func DefaultPostgresConfig() PostgresContainerConfig {
	return PostgresContainerConfig{
		Database: "crm_test",
		User:     "crm_test",
		Password: "crm_test_password",
	}
}

// NewPostgresContainer creates a new PostgreSQL container for testing.
// For integration tests, this connects to the docker-compose PostgreSQL instance.
func NewPostgresContainer(ctx context.Context, cfg PostgresContainerConfig) (*PostgresContainer, error) {
	container := &PostgresContainer{
		Host:     getEnvOrDefault("TEST_POSTGRES_HOST", "localhost"),
		Port:     getEnvOrDefault("TEST_POSTGRES_PORT", "5432"),
		Database: getEnvOrDefault("TEST_POSTGRES_DB", cfg.Database),
		User:     getEnvOrDefault("TEST_POSTGRES_USER", cfg.User),
		Password: getEnvOrDefault("TEST_POSTGRES_PASSWORD", cfg.Password),
		SSLMode:  "disable",
	}

	// Connect to the database
	connStr := container.ConnectionString()
	db, err := sqlx.ConnectContext(ctx, "postgres", connStr)
	if err != nil {
		// Try to connect to the default 'postgres' database and create test database
		adminConnStr := fmt.Sprintf(
			"host=%s port=%s user=%s password=%s dbname=postgres sslmode=%s",
			container.Host, container.Port, container.User, container.Password, container.SSLMode,
		)
		adminDB, adminErr := sqlx.ConnectContext(ctx, "postgres", adminConnStr)
		if adminErr != nil {
			return nil, fmt.Errorf("failed to connect to PostgreSQL: %w (original error: %v)", adminErr, err)
		}
		defer adminDB.Close()

		// Create the test database
		_, err = adminDB.ExecContext(ctx, fmt.Sprintf("CREATE DATABASE %s", container.Database))
		if err != nil && !strings.Contains(err.Error(), "already exists") {
			return nil, fmt.Errorf("failed to create test database: %w", err)
		}

		// Now connect to the test database
		db, err = sqlx.ConnectContext(ctx, "postgres", connStr)
		if err != nil {
			return nil, fmt.Errorf("failed to connect to test database: %w", err)
		}
	}

	container.DB = db

	// Configure connection pool
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(time.Hour)

	// Verify connection
	if err := db.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping PostgreSQL: %w", err)
	}

	return container, nil
}

// ConnectionString returns the PostgreSQL connection string.
func (c *PostgresContainer) ConnectionString() string {
	return fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		c.Host, c.Port, c.User, c.Password, c.Database, c.SSLMode,
	)
}

// GetDB returns the database connection.
func (c *PostgresContainer) GetDB() *sqlx.DB {
	return c.DB
}

// RunMigrations runs SQL migrations from the specified directory.
func (c *PostgresContainer) RunMigrations(ctx context.Context, migrationsPath string) error {
	// Read migration files
	files, err := filepath.Glob(filepath.Join(migrationsPath, "*.up.sql"))
	if err != nil {
		return fmt.Errorf("failed to read migration files: %w", err)
	}

	// Sort files to ensure order
	for _, file := range files {
		content, err := os.ReadFile(file)
		if err != nil {
			return fmt.Errorf("failed to read migration file %s: %w", file, err)
		}

		_, err = c.DB.ExecContext(ctx, string(content))
		if err != nil {
			// Ignore "already exists" errors
			if !strings.Contains(err.Error(), "already exists") &&
				!strings.Contains(err.Error(), "duplicate") {
				return fmt.Errorf("failed to execute migration %s: %w", file, err)
			}
		}

		c.migrations = append(c.migrations, file)
	}

	return nil
}

// RunMigrationSQL runs a SQL migration directly.
func (c *PostgresContainer) RunMigrationSQL(ctx context.Context, sql string) error {
	_, err := c.DB.ExecContext(ctx, sql)
	if err != nil {
		// Ignore "already exists" errors
		if !strings.Contains(err.Error(), "already exists") &&
			!strings.Contains(err.Error(), "duplicate") {
			return fmt.Errorf("failed to execute migration: %w", err)
		}
	}
	return nil
}

// TruncateTables truncates all tables in the database (for test isolation).
func (c *PostgresContainer) TruncateTables(ctx context.Context, tables ...string) error {
	if len(tables) == 0 {
		return nil
	}

	// Disable foreign key checks temporarily
	query := fmt.Sprintf("TRUNCATE TABLE %s RESTART IDENTITY CASCADE", strings.Join(tables, ", "))
	_, err := c.DB.ExecContext(ctx, query)
	if err != nil {
		return fmt.Errorf("failed to truncate tables: %w", err)
	}

	return nil
}

// CleanDatabase removes all data from the database.
func (c *PostgresContainer) CleanDatabase(ctx context.Context) error {
	// Get all table names
	query := `
		SELECT tablename
		FROM pg_tables
		WHERE schemaname = 'public'
		AND tablename NOT IN ('schema_migrations', 'goose_db_version')
	`

	var tables []string
	err := c.DB.SelectContext(ctx, &tables, query)
	if err != nil {
		return fmt.Errorf("failed to get table names: %w", err)
	}

	if len(tables) > 0 {
		return c.TruncateTables(ctx, tables...)
	}

	return nil
}

// ExecuteInTransaction executes a function within a transaction.
func (c *PostgresContainer) ExecuteInTransaction(ctx context.Context, fn func(tx *sql.Tx) error) error {
	tx, err := c.DB.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	if err := fn(tx); err != nil {
		if rbErr := tx.Rollback(); rbErr != nil {
			return fmt.Errorf("failed to rollback: %v (original error: %w)", rbErr, err)
		}
		return err
	}

	return tx.Commit()
}

// Close closes the database connection.
func (c *PostgresContainer) Close() error {
	if c.DB != nil {
		return c.DB.Close()
	}
	return nil
}

// CreateSchema creates a schema if it doesn't exist.
func (c *PostgresContainer) CreateSchema(ctx context.Context, schema string) error {
	_, err := c.DB.ExecContext(ctx, fmt.Sprintf("CREATE SCHEMA IF NOT EXISTS %s", schema))
	return err
}

// SetSearchPath sets the search path for the connection.
func (c *PostgresContainer) SetSearchPath(ctx context.Context, schemas ...string) error {
	_, err := c.DB.ExecContext(ctx, fmt.Sprintf("SET search_path TO %s", strings.Join(schemas, ", ")))
	return err
}

// WaitForReady waits for the database to be ready.
func (c *PostgresContainer) WaitForReady(ctx context.Context, timeout time.Duration) error {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return fmt.Errorf("timeout waiting for PostgreSQL to be ready")
		case <-ticker.C:
			if err := c.DB.PingContext(ctx); err == nil {
				return nil
			}
		}
	}
}

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
