// Package postgres contains PostgreSQL repository implementations for the sales service.
package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
)

// ============================================================================
// Transaction Management
// ============================================================================

// txKey is the context key for database transactions.
type txKey struct{}

// getTxFromContext retrieves a transaction from context if present.
func getTxFromContext(ctx context.Context) *sqlx.Tx {
	if tx, ok := ctx.Value(txKey{}).(*sqlx.Tx); ok {
		return tx
	}
	return nil
}

// setTxToContext stores a transaction in context.
func setTxToContext(ctx context.Context, tx *sqlx.Tx) context.Context {
	return context.WithValue(ctx, txKey{}, tx)
}

// getExecutor returns either the transaction from context or the database connection.
func getExecutor(ctx context.Context, db *sqlx.DB) sqlx.ExtContext {
	if tx := getTxFromContext(ctx); tx != nil {
		return tx
	}
	return db
}

// ============================================================================
// Tenant Context
// ============================================================================

// tenantContextKey is the context key for tenant ID.
type tenantContextKey struct{}

// SetTenantContext sets the tenant context for row-level security.
func SetTenantContext(ctx context.Context, db sqlx.ExtContext, tenantID string) error {
	_, err := db.ExecContext(ctx, "SET LOCAL app.tenant_id = $1", tenantID)
	return err
}

// GetTenantFromContext retrieves the tenant ID from context.
func GetTenantFromContext(ctx context.Context) string {
	if tid, ok := ctx.Value(tenantContextKey{}).(string); ok {
		return tid
	}
	return ""
}

// WithTenantContext adds tenant ID to context.
func WithTenantContext(ctx context.Context, tenantID string) context.Context {
	return context.WithValue(ctx, tenantContextKey{}, tenantID)
}

// ============================================================================
// Transaction Manager
// ============================================================================

// TransactionManager manages database transactions.
type TransactionManager struct {
	db *sqlx.DB
}

// NewTransactionManager creates a new TransactionManager.
func NewTransactionManager(db *sqlx.DB) *TransactionManager {
	return &TransactionManager{db: db}
}

// WithTransaction executes a function within a transaction.
func (tm *TransactionManager) WithTransaction(ctx context.Context, fn func(ctx context.Context) error) error {
	// Check if already in transaction
	if getTxFromContext(ctx) != nil {
		return fn(ctx)
	}

	tx, err := tm.db.BeginTxx(ctx, &sql.TxOptions{
		Isolation: sql.LevelReadCommitted,
	})
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	txCtx := setTxToContext(ctx, tx)

	defer func() {
		if p := recover(); p != nil {
			_ = tx.Rollback()
			panic(p)
		}
	}()

	if err := fn(txCtx); err != nil {
		if rbErr := tx.Rollback(); rbErr != nil {
			return fmt.Errorf("rollback failed: %v, original error: %w", rbErr, err)
		}
		return err
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// WithSerializableTransaction executes a function within a serializable transaction.
func (tm *TransactionManager) WithSerializableTransaction(ctx context.Context, fn func(ctx context.Context) error) error {
	tx, err := tm.db.BeginTxx(ctx, &sql.TxOptions{
		Isolation: sql.LevelSerializable,
	})
	if err != nil {
		return fmt.Errorf("failed to begin serializable transaction: %w", err)
	}

	txCtx := setTxToContext(ctx, tx)

	defer func() {
		if p := recover(); p != nil {
			_ = tx.Rollback()
			panic(p)
		}
	}()

	if err := fn(txCtx); err != nil {
		if rbErr := tx.Rollback(); rbErr != nil {
			return fmt.Errorf("rollback failed: %v, original error: %w", rbErr, err)
		}
		return err
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit serializable transaction: %w", err)
	}

	return nil
}

// BeginTx starts a new transaction.
func (tm *TransactionManager) BeginTx(ctx context.Context) (context.Context, error) {
	tx, err := tm.db.BeginTxx(ctx, &sql.TxOptions{
		Isolation: sql.LevelReadCommitted,
	})
	if err != nil {
		return ctx, fmt.Errorf("failed to begin transaction: %w", err)
	}

	return setTxToContext(ctx, tx), nil
}

// CommitTx commits the current transaction.
func (tm *TransactionManager) CommitTx(ctx context.Context) error {
	tx := getTxFromContext(ctx)
	if tx == nil {
		return fmt.Errorf("no transaction in context")
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// RollbackTx rolls back the current transaction.
func (tm *TransactionManager) RollbackTx(ctx context.Context) error {
	tx := getTxFromContext(ctx)
	if tx == nil {
		return fmt.Errorf("no transaction in context")
	}

	if err := tx.Rollback(); err != nil {
		return fmt.Errorf("failed to rollback transaction: %w", err)
	}

	return nil
}

// ============================================================================
// Query Builder Helpers
// ============================================================================

// QueryBuilder helps construct dynamic SQL queries.
type QueryBuilder struct {
	baseQuery  string
	conditions []string
	args       []interface{}
	orderBy    string
	limit      int
	offset     int
}

// NewQueryBuilder creates a new QueryBuilder.
func NewQueryBuilder(baseQuery string) *QueryBuilder {
	return &QueryBuilder{
		baseQuery:  baseQuery,
		conditions: make([]string, 0),
		args:       make([]interface{}, 0),
	}
}

// Where adds a WHERE condition.
func (qb *QueryBuilder) Where(condition string, args ...interface{}) *QueryBuilder {
	qb.conditions = append(qb.conditions, condition)
	qb.args = append(qb.args, args...)
	return qb
}

// WhereIn adds a WHERE IN condition.
func (qb *QueryBuilder) WhereIn(column string, values []uuid.UUID) *QueryBuilder {
	if len(values) == 0 {
		return qb
	}
	placeholders := make([]string, len(values))
	for i, v := range values {
		placeholders[i] = fmt.Sprintf("$%d", len(qb.args)+i+1)
		qb.args = append(qb.args, v)
	}
	qb.conditions = append(qb.conditions, fmt.Sprintf("%s IN (%s)", column, strings.Join(placeholders, ",")))
	return qb
}

// WhereInStrings adds a WHERE IN condition for strings.
func (qb *QueryBuilder) WhereInStrings(column string, values []string) *QueryBuilder {
	if len(values) == 0 {
		return qb
	}
	placeholders := make([]string, len(values))
	for i, v := range values {
		placeholders[i] = fmt.Sprintf("$%d", len(qb.args)+i+1)
		qb.args = append(qb.args, v)
	}
	qb.conditions = append(qb.conditions, fmt.Sprintf("%s IN (%s)", column, strings.Join(placeholders, ",")))
	return qb
}

// OrderBy sets the ORDER BY clause.
func (qb *QueryBuilder) OrderBy(column, direction string) *QueryBuilder {
	if direction != "asc" && direction != "desc" {
		direction = "desc"
	}
	qb.orderBy = fmt.Sprintf("ORDER BY %s %s", column, strings.ToUpper(direction))
	return qb
}

// Limit sets the LIMIT clause.
func (qb *QueryBuilder) Limit(limit int) *QueryBuilder {
	qb.limit = limit
	return qb
}

// Offset sets the OFFSET clause.
func (qb *QueryBuilder) Offset(offset int) *QueryBuilder {
	qb.offset = offset
	return qb
}

// Build constructs the final SQL query.
func (qb *QueryBuilder) Build() (string, []interface{}) {
	query := qb.baseQuery

	if len(qb.conditions) > 0 {
		if strings.Contains(strings.ToUpper(query), "WHERE") {
			query += " AND " + strings.Join(qb.conditions, " AND ")
		} else {
			query += " WHERE " + strings.Join(qb.conditions, " AND ")
		}
	}

	if qb.orderBy != "" {
		query += " " + qb.orderBy
	}

	if qb.limit > 0 {
		query += fmt.Sprintf(" LIMIT %d", qb.limit)
	}

	if qb.offset > 0 {
		query += fmt.Sprintf(" OFFSET %d", qb.offset)
	}

	return query, qb.args
}

// BuildCount constructs a count query.
func (qb *QueryBuilder) BuildCount() (string, []interface{}) {
	// Replace SELECT ... FROM with SELECT COUNT(*) FROM
	countQuery := "SELECT COUNT(*) " + qb.baseQuery[strings.Index(strings.ToUpper(qb.baseQuery), "FROM"):]

	if len(qb.conditions) > 0 {
		if strings.Contains(strings.ToUpper(countQuery), "WHERE") {
			countQuery += " AND " + strings.Join(qb.conditions, " AND ")
		} else {
			countQuery += " WHERE " + strings.Join(qb.conditions, " AND ")
		}
	}

	return countQuery, qb.args
}

// Args returns the current arguments count for parameter numbering.
func (qb *QueryBuilder) Args() []interface{} {
	return qb.args
}

// NextParam returns the next parameter placeholder number.
func (qb *QueryBuilder) NextParam() int {
	return len(qb.args) + 1
}

// ============================================================================
// JSON Helpers
// ============================================================================

// NullableJSON represents a nullable JSON field.
type NullableJSON struct {
	sql.NullString
}

// Scan implements the sql.Scanner interface.
func (nj *NullableJSON) Scan(value interface{}) error {
	if value == nil {
		nj.Valid = false
		return nil
	}

	switch v := value.(type) {
	case []byte:
		nj.String = string(v)
		nj.Valid = true
	case string:
		nj.String = v
		nj.Valid = true
	default:
		return fmt.Errorf("unsupported type: %T", value)
	}

	return nil
}

// MarshalTo marshals the value to the target.
func (nj NullableJSON) MarshalTo(target interface{}) error {
	if !nj.Valid || nj.String == "" {
		return nil
	}
	return json.Unmarshal([]byte(nj.String), target)
}

// ToJSON converts a value to JSON string.
func ToJSON(v interface{}) (string, error) {
	if v == nil {
		return "{}", nil
	}
	data, err := json.Marshal(v)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// FromJSON converts a JSON string to a map.
func FromJSON(s string) (map[string]interface{}, error) {
	if s == "" {
		return nil, nil
	}
	var result map[string]interface{}
	if err := json.Unmarshal([]byte(s), &result); err != nil {
		return nil, err
	}
	return result, nil
}

// ============================================================================
// Array Helpers
// ============================================================================

// StringArray is a helper for PostgreSQL text arrays.
type StringArray []string

// Scan implements the sql.Scanner interface.
func (a *StringArray) Scan(src interface{}) error {
	return pq.Array((*[]string)(a)).Scan(src)
}

// UUIDArray is a helper for PostgreSQL UUID arrays.
type UUIDArray []uuid.UUID

// Scan implements the sql.Scanner interface.
func (a *UUIDArray) Scan(src interface{}) error {
	if src == nil {
		*a = nil
		return nil
	}

	// Handle as string array and convert
	var strArray []string
	if err := pq.Array(&strArray).Scan(src); err != nil {
		return err
	}

	result := make([]uuid.UUID, 0, len(strArray))
	for _, s := range strArray {
		id, err := uuid.Parse(s)
		if err != nil {
			return fmt.Errorf("failed to parse UUID: %w", err)
		}
		result = append(result, id)
	}
	*a = result
	return nil
}

// ============================================================================
// Time Helpers
// ============================================================================

// NullTime is a wrapper around sql.NullTime for easier usage.
type NullTime struct {
	sql.NullTime
}

// TimePtr returns a pointer to the time if valid, nil otherwise.
func (nt NullTime) TimePtr() *time.Time {
	if !nt.Valid {
		return nil
	}
	return &nt.Time
}

// NewNullTime creates a NullTime from a time pointer.
func NewNullTime(t *time.Time) NullTime {
	if t == nil {
		return NullTime{sql.NullTime{Valid: false}}
	}
	return NullTime{sql.NullTime{Time: *t, Valid: true}}
}

// ============================================================================
// Error Helpers
// ============================================================================

// IsNotFoundError checks if an error is a not found error.
func IsNotFoundError(err error) bool {
	return err == sql.ErrNoRows
}

// IsUniqueViolation checks if an error is a unique constraint violation.
func IsUniqueViolation(err error) bool {
	if pqErr, ok := err.(*pq.Error); ok {
		return pqErr.Code == "23505"
	}
	return false
}

// IsForeignKeyViolation checks if an error is a foreign key violation.
func IsForeignKeyViolation(err error) bool {
	if pqErr, ok := err.(*pq.Error); ok {
		return pqErr.Code == "23503"
	}
	return false
}

// IsCheckViolation checks if an error is a check constraint violation.
func IsCheckViolation(err error) bool {
	if pqErr, ok := err.(*pq.Error); ok {
		return pqErr.Code == "23514"
	}
	return false
}

// ============================================================================
// Sorting Helpers
// ============================================================================

// ValidateSortColumn validates and returns a safe column name for sorting.
func ValidateSortColumn(column string, allowedColumns map[string]string) string {
	if mapped, ok := allowedColumns[column]; ok {
		return mapped
	}
	return "created_at"
}

// ValidateSortOrder validates and returns a safe sort order.
func ValidateSortOrder(order string) string {
	order = strings.ToLower(order)
	if order == "asc" || order == "desc" {
		return order
	}
	return "desc"
}
