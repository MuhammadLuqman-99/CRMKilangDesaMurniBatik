// Package usecase contains the application use cases for the Customer service.
package usecase

import (
	"bytes"
	"context"
	"io"
	"time"

	"github.com/google/uuid"

	"github.com/kilang-desa-murni/crm/internal/customer/application"
	"github.com/kilang-desa-murni/crm/internal/customer/application/dto"
	"github.com/kilang-desa-murni/crm/internal/customer/application/mapper"
	"github.com/kilang-desa-murni/crm/internal/customer/application/ports"
	"github.com/kilang-desa-murni/crm/internal/customer/domain"
)

// ============================================================================
// Import Customers Use Case
// ============================================================================

// ImportCustomersUseCase handles customer import.
type ImportCustomersUseCase struct {
	uow            domain.UnitOfWork
	importService  ports.ImportService
	eventPublisher ports.EventPublisher
	idGenerator    ports.IDGenerator
	cache          ports.CacheService
	auditLogger    ports.AuditLogger
	customerMapper *mapper.CustomerMapper
	config         ImportConfig
}

// ImportConfig holds configuration for import operations.
type ImportConfig struct {
	MaxFileSize          int64
	MaxRowsPerImport     int
	BatchSize            int
	SkipDuplicates       bool
	UpdateExisting       bool
	ValidateBeforeImport bool
	SupportedFormats     []string
}

// DefaultImportConfig returns default configuration.
func DefaultImportConfig() ImportConfig {
	return ImportConfig{
		MaxFileSize:          10 * 1024 * 1024, // 10MB
		MaxRowsPerImport:     10000,
		BatchSize:            100,
		SkipDuplicates:       true,
		UpdateExisting:       false,
		ValidateBeforeImport: true,
		SupportedFormats:     []string{"csv", "xlsx", "json"},
	}
}

// NewImportCustomersUseCase creates a new ImportCustomersUseCase.
func NewImportCustomersUseCase(
	uow domain.UnitOfWork,
	importService ports.ImportService,
	eventPublisher ports.EventPublisher,
	idGenerator ports.IDGenerator,
	cache ports.CacheService,
	auditLogger ports.AuditLogger,
	config ImportConfig,
) *ImportCustomersUseCase {
	return &ImportCustomersUseCase{
		uow:            uow,
		importService:  importService,
		eventPublisher: eventPublisher,
		idGenerator:    idGenerator,
		cache:          cache,
		auditLogger:    auditLogger,
		customerMapper: mapper.NewCustomerMapper(),
		config:         config,
	}
}

// ImportCustomersInput holds input for customer import.
type ImportCustomersInput struct {
	TenantID       uuid.UUID
	UserID         uuid.UUID
	FileName       string
	FileSize       int64
	Format         string
	Data           []byte
	FieldMapping   map[string]string
	SkipDuplicates bool
	UpdateExisting bool
	DefaultOwner   *uuid.UUID
	DefaultTags    []string
	IPAddress      string
	UserAgent      string
}

// ImportCustomersOutput holds the result of customer import.
type ImportCustomersOutput struct {
	ImportID     uuid.UUID         `json:"import_id"`
	TotalRows    int               `json:"total_rows"`
	SuccessCount int               `json:"success_count"`
	FailureCount int               `json:"failure_count"`
	SkippedCount int               `json:"skipped_count"`
	UpdatedCount int               `json:"updated_count"`
	Results      []ImportRowResult `json:"results,omitempty"`
	Errors       []string          `json:"errors,omitempty"`
	Duration     time.Duration     `json:"duration"`
}

// ImportRowResult represents the result of importing a single row.
type ImportRowResult struct {
	RowNumber  int        `json:"row_number"`
	Success    bool       `json:"success"`
	CustomerID *uuid.UUID `json:"customer_id,omitempty"`
	Action     string     `json:"action,omitempty"` // "created", "updated", "skipped"
	Errors     []string   `json:"errors,omitempty"`
	Warnings   []string   `json:"warnings,omitempty"`
}

// Execute imports customers from a file.
func (uc *ImportCustomersUseCase) Execute(ctx context.Context, input ImportCustomersInput) (*ImportCustomersOutput, error) {
	startTime := time.Now()

	// Validate input
	if err := uc.validateImportInput(input); err != nil {
		return nil, err
	}

	// Create import record
	importID := uc.idGenerator.NewID()
	importRecord := &domain.Import{
		BaseEntity: domain.NewBaseEntity(),
		TenantID:   input.TenantID,
		FileName:   input.FileName,
		FileSize:   input.FileSize,
		FileType:   input.Format,
		Status:     domain.ImportStatusProcessing,
		CreatedBy:  &input.UserID,
		StartedAt:  &startTime,
		Options: domain.ImportOptions{
			SkipDuplicates: input.SkipDuplicates,
			UpdateExisting: input.UpdateExisting,
			DefaultOwner:   input.DefaultOwner,
			DefaultStatus:  domain.CustomerStatusLead,
			DefaultType:    domain.CustomerTypeIndividual,
			DefaultSource:  domain.CustomerSourceImport,
			DefaultTags:    input.DefaultTags,
			FieldMapping:   input.FieldMapping,
		},
	}
	importRecord.ID = importID

	// Save import record
	if err := uc.uow.Imports().CreateImport(ctx, importRecord); err != nil {
		return nil, application.ErrInternalError("failed to create import record", err)
	}

	// Parse import data
	rows, err := uc.importService.ParseCustomers(ctx, bytes.NewReader(input.Data), ports.ExportFormat(input.Format), input.FieldMapping)
	if err != nil {
		uc.markImportFailed(ctx, importRecord, err.Error())
		return nil, application.ErrInvalidFormat(input.Format, uc.config.SupportedFormats)
	}

	// Validate rows if enabled
	if uc.config.ValidateBeforeImport {
		validationErrors, err := uc.importService.ValidateImportData(ctx, rows)
		if err != nil {
			uc.markImportFailed(ctx, importRecord, err.Error())
			return nil, application.ErrInternalError("validation failed", err)
		}

		// Check for critical errors
		var criticalErrors []string
		for _, ve := range validationErrors {
			if ve.Severity == "error" {
				criticalErrors = append(criticalErrors, ve.Error)
			}
		}
		if len(criticalErrors) > 0 && len(criticalErrors) > len(rows)/2 {
			uc.markImportFailed(ctx, importRecord, "too many validation errors")
			return nil, application.ErrImportFailed("too many validation errors", len(criticalErrors))
		}
	}

	// Update total rows
	importRecord.TotalRows = len(rows)
	_ = uc.uow.Imports().UpdateImport(ctx, importRecord)

	// Process rows in batches
	output := &ImportCustomersOutput{
		ImportID:  importID,
		TotalRows: len(rows),
		Results:   make([]ImportRowResult, 0, len(rows)),
	}

	for i := 0; i < len(rows); i += uc.config.BatchSize {
		end := i + uc.config.BatchSize
		if end > len(rows) {
			end = len(rows)
		}
		batch := rows[i:end]

		results := uc.processBatch(ctx, input, batch)
		for _, result := range results {
			output.Results = append(output.Results, result)
			if result.Success {
				if result.Action == "updated" {
					output.UpdatedCount++
				} else if result.Action == "skipped" {
					output.SkippedCount++
				} else {
					output.SuccessCount++
				}
			} else {
				output.FailureCount++
				output.Errors = append(output.Errors, result.Errors...)
			}
		}

		// Update progress
		importRecord.ProcessedRows = i + len(batch)
		importRecord.SuccessRows = output.SuccessCount + output.UpdatedCount
		importRecord.FailedRows = output.FailureCount
		importRecord.DuplicateRows = output.SkippedCount
		_ = uc.uow.Imports().UpdateImport(ctx, importRecord)
	}

	// Mark import as completed
	now := time.Now()
	importRecord.Status = domain.ImportStatusCompleted
	importRecord.CompletedAt = &now
	_ = uc.uow.Imports().UpdateImport(ctx, importRecord)

	output.Duration = time.Since(startTime)

	// Invalidate cache
	if uc.cache != nil {
		_ = uc.cache.InvalidateByTenant(ctx, input.TenantID)
	}

	// Audit log
	if uc.auditLogger != nil {
		_ = uc.auditLogger.LogAction(ctx, ports.AuditEntry{
			ID:         uc.idGenerator.NewID(),
			TenantID:   input.TenantID,
			UserID:     &input.UserID,
			Action:     "customers.imported",
			EntityType: "import",
			EntityID:   importID,
			Metadata: map[string]interface{}{
				"file_name":     input.FileName,
				"total_rows":    output.TotalRows,
				"success_count": output.SuccessCount,
				"failure_count": output.FailureCount,
				"skipped_count": output.SkippedCount,
				"updated_count": output.UpdatedCount,
				"duration_ms":   output.Duration.Milliseconds(),
			},
			IPAddress: input.IPAddress,
			UserAgent: input.UserAgent,
			Timestamp: time.Now().UTC(),
		})
	}

	return output, nil
}

// validateImportInput validates the import input.
func (uc *ImportCustomersUseCase) validateImportInput(input ImportCustomersInput) error {
	if input.TenantID == uuid.Nil {
		return application.ErrInvalidInput("tenant_id is required")
	}
	if input.UserID == uuid.Nil {
		return application.ErrInvalidInput("user_id is required")
	}
	if len(input.Data) == 0 {
		return application.ErrInvalidInput("data is required")
	}
	if input.FileSize > uc.config.MaxFileSize {
		return application.ErrFileTooLarge(input.FileSize, uc.config.MaxFileSize)
	}

	// Check format
	validFormat := false
	for _, f := range uc.config.SupportedFormats {
		if f == input.Format {
			validFormat = true
			break
		}
	}
	if !validFormat {
		return application.ErrInvalidFormat(input.Format, uc.config.SupportedFormats)
	}

	return nil
}

// processBatch processes a batch of import rows.
func (uc *ImportCustomersUseCase) processBatch(ctx context.Context, input ImportCustomersInput, rows []*ports.CustomerImportRow) []ImportRowResult {
	results := make([]ImportRowResult, len(rows))

	for i, row := range rows {
		result := ImportRowResult{
			RowNumber: row.RowNumber,
			Errors:    row.Errors,
			Warnings:  row.Warnings,
		}

		// Skip if row has errors
		if len(row.Errors) > 0 {
			result.Success = false
			results[i] = result
			continue
		}

		// Check for duplicates
		if input.SkipDuplicates && row.Email != "" {
			exists, _ := uc.uow.Customers().ExistsByEmail(ctx, input.TenantID, row.Email)
			if exists {
				if input.UpdateExisting {
					// Find and update existing
					existing, err := uc.uow.Customers().FindByEmail(ctx, input.TenantID, row.Email)
					if err == nil {
						uc.updateFromImportRow(existing, row, input)
						if err := uc.uow.Customers().Update(ctx, existing); err == nil {
							result.Success = true
							result.Action = "updated"
							result.CustomerID = &existing.ID
						} else {
							result.Success = false
							result.Errors = append(result.Errors, err.Error())
						}
					}
				} else {
					result.Success = true
					result.Action = "skipped"
					result.Warnings = append(result.Warnings, "duplicate email")
				}
				results[i] = result
				continue
			}
		}

		// Create new customer
		customer, err := uc.createFromImportRow(input, row)
		if err != nil {
			result.Success = false
			result.Errors = append(result.Errors, err.Error())
			results[i] = result
			continue
		}

		// Save customer
		if err := uc.uow.Customers().Create(ctx, customer); err != nil {
			result.Success = false
			result.Errors = append(result.Errors, err.Error())
		} else {
			result.Success = true
			result.Action = "created"
			result.CustomerID = &customer.ID
		}

		results[i] = result
	}

	return results
}

// createFromImportRow creates a customer from an import row.
func (uc *ImportCustomersUseCase) createFromImportRow(input ImportCustomersInput, row *ports.CustomerImportRow) (*domain.Customer, error) {
	customerType := domain.CustomerTypeIndividual
	if row.Type != "" {
		customerType = domain.CustomerType(row.Type)
	}

	builder := domain.NewCustomerBuilder(input.TenantID, row.Name, customerType).
		WithSource(domain.CustomerSourceImport).
		WithCreatedBy(input.UserID)

	if row.Email != "" {
		builder.WithEmail(row.Email)
	}
	if row.Phone != "" {
		builder.WithPhone(row.Phone, domain.PhoneTypeMobile, true)
	}
	if row.Website != "" {
		builder.WithWebsite(row.Website)
	}
	if row.Address.Line1 != "" {
		builder.WithAddress(
			row.Address.Line1,
			row.Address.City,
			row.Address.PostalCode,
			row.Address.CountryCode,
			domain.AddressTypeOffice,
		)
	}
	if row.CompanyInfo.LegalName != "" {
		builder.WithCompanyInfo(domain.CompanyInfo{
			LegalName:          row.CompanyInfo.LegalName,
			RegistrationNumber: row.CompanyInfo.RegistrationNumber,
			TaxID:              row.CompanyInfo.TaxID,
			Industry:           domain.Industry(row.CompanyInfo.Industry),
			Size:               domain.CompanySize(row.CompanyInfo.Size),
		})
	}
	if input.DefaultOwner != nil {
		builder.WithOwner(*input.DefaultOwner)
	}
	if len(input.DefaultTags) > 0 {
		builder.WithTags(input.DefaultTags...)
	}
	if len(row.Tags) > 0 {
		builder.WithTags(row.Tags...)
	}
	if row.Notes != "" {
		builder.WithNotes(row.Notes)
	}

	return builder.Build()
}

// updateFromImportRow updates a customer from an import row.
func (uc *ImportCustomersUseCase) updateFromImportRow(customer *domain.Customer, row *ports.CustomerImportRow, input ImportCustomersInput) {
	if row.Name != "" && row.Name != customer.Name {
		customer.UpdateName(row.Name)
	}
	if row.Phone != "" {
		phone, err := domain.NewPhoneNumber(row.Phone, domain.PhoneTypeMobile)
		if err == nil {
			customer.AddPhoneNumber(phone)
		}
	}
	if row.Notes != "" {
		customer.UpdateNotes(customer.Notes + "\n" + row.Notes)
	}
	customer.AuditInfo.SetUpdatedBy(input.UserID)
}

// markImportFailed marks an import as failed.
func (uc *ImportCustomersUseCase) markImportFailed(ctx context.Context, importRecord *domain.Import, errMsg string) {
	importRecord.Status = domain.ImportStatusFailed
	importRecord.ErrorMessage = errMsg
	now := time.Now()
	importRecord.CompletedAt = &now
	_ = uc.uow.Imports().UpdateImport(ctx, importRecord)
}

// ============================================================================
// Export Customers Use Case
// ============================================================================

// ExportCustomersUseCase handles customer export.
type ExportCustomersUseCase struct {
	uow            domain.UnitOfWork
	exportService  ports.ExportService
	idGenerator    ports.IDGenerator
	auditLogger    ports.AuditLogger
	customerMapper *mapper.CustomerMapper
	config         ExportConfig
}

// ExportConfig holds configuration for export operations.
type ExportConfig struct {
	MaxRowsPerExport int
	SupportedFormats []string
	DefaultFields    []string
}

// DefaultExportConfig returns default configuration.
func DefaultExportConfig() ExportConfig {
	return ExportConfig{
		MaxRowsPerExport: 50000,
		SupportedFormats: []string{"csv", "xlsx", "json"},
		DefaultFields: []string{
			"code", "name", "type", "status", "email", "phone",
			"website", "address", "tags", "owner", "created_at",
		},
	}
}

// NewExportCustomersUseCase creates a new ExportCustomersUseCase.
func NewExportCustomersUseCase(
	uow domain.UnitOfWork,
	exportService ports.ExportService,
	idGenerator ports.IDGenerator,
	auditLogger ports.AuditLogger,
	config ExportConfig,
) *ExportCustomersUseCase {
	return &ExportCustomersUseCase{
		uow:            uow,
		exportService:  exportService,
		idGenerator:    idGenerator,
		auditLogger:    auditLogger,
		customerMapper: mapper.NewCustomerMapper(),
		config:         config,
	}
}

// ExportCustomersInput holds input for customer export.
type ExportCustomersInput struct {
	TenantID  uuid.UUID
	UserID    uuid.UUID
	Format    string
	Fields    []string
	Filter    *dto.SearchCustomersRequest
	IPAddress string
	UserAgent string
}

// ExportCustomersOutput holds the result of customer export.
type ExportCustomersOutput struct {
	Data        []byte `json:"-"`
	FileName    string `json:"file_name"`
	ContentType string `json:"content_type"`
	TotalRows   int    `json:"total_rows"`
}

// Execute exports customers to a file.
func (uc *ExportCustomersUseCase) Execute(ctx context.Context, input ExportCustomersInput) (*ExportCustomersOutput, error) {
	// Validate input
	if err := uc.validateExportInput(input); err != nil {
		return nil, err
	}

	// Normalize fields
	fields := input.Fields
	if len(fields) == 0 {
		fields = uc.config.DefaultFields
	}

	// Build filter from request
	var filter domain.CustomerFilter
	filter.TenantID = &input.TenantID
	filter.Limit = uc.config.MaxRowsPerExport

	if input.Filter != nil {
		filter = uc.customerMapper.SearchRequestToFilter(input.Filter)
		filter.TenantID = &input.TenantID
		if filter.Limit == 0 || filter.Limit > uc.config.MaxRowsPerExport {
			filter.Limit = uc.config.MaxRowsPerExport
		}
	}

	// Fetch customers
	customerList, err := uc.uow.Customers().List(ctx, filter)
	if err != nil {
		return nil, application.ErrInternalError("failed to fetch customers", err)
	}

	if len(customerList.Customers) == 0 {
		return nil, application.ErrExportFailed("no customers to export")
	}

	// Export to format
	data, err := uc.exportService.ExportCustomers(ctx, input.TenantID, customerList.Customers, ports.ExportFormat(input.Format), fields)
	if err != nil {
		return nil, application.ErrExportFailed(err.Error())
	}

	// Generate filename
	timestamp := time.Now().Format("20060102_150405")
	filename := "customers_" + timestamp

	var contentType string
	switch input.Format {
	case "csv":
		filename += ".csv"
		contentType = "text/csv"
	case "xlsx":
		filename += ".xlsx"
		contentType = "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"
	case "json":
		filename += ".json"
		contentType = "application/json"
	case "pdf":
		filename += ".pdf"
		contentType = "application/pdf"
	}

	// Audit log
	if uc.auditLogger != nil {
		_ = uc.auditLogger.LogAction(ctx, ports.AuditEntry{
			ID:         uc.idGenerator.NewID(),
			TenantID:   input.TenantID,
			UserID:     &input.UserID,
			Action:     "customers.exported",
			EntityType: "export",
			EntityID:   uuid.Nil,
			Metadata: map[string]interface{}{
				"format":     input.Format,
				"total_rows": len(customerList.Customers),
				"fields":     fields,
				"file_name":  filename,
			},
			IPAddress: input.IPAddress,
			UserAgent: input.UserAgent,
			Timestamp: time.Now().UTC(),
		})
	}

	return &ExportCustomersOutput{
		Data:        data,
		FileName:    filename,
		ContentType: contentType,
		TotalRows:   len(customerList.Customers),
	}, nil
}

// validateExportInput validates the export input.
func (uc *ExportCustomersUseCase) validateExportInput(input ExportCustomersInput) error {
	if input.TenantID == uuid.Nil {
		return application.ErrInvalidInput("tenant_id is required")
	}
	if input.UserID == uuid.Nil {
		return application.ErrInvalidInput("user_id is required")
	}

	// Check format
	validFormat := false
	for _, f := range uc.config.SupportedFormats {
		if f == input.Format {
			validFormat = true
			break
		}
	}
	if !validFormat {
		return application.ErrInvalidFormat(input.Format, uc.config.SupportedFormats)
	}

	return nil
}

// ============================================================================
// Stream Export Use Case (for large exports)
// ============================================================================

// StreamExportCustomersUseCase handles streaming export for large datasets.
type StreamExportCustomersUseCase struct {
	uow           domain.UnitOfWork
	exportService ports.ExportService
	idGenerator   ports.IDGenerator
	auditLogger   ports.AuditLogger
	config        ExportConfig
}

// NewStreamExportCustomersUseCase creates a new StreamExportCustomersUseCase.
func NewStreamExportCustomersUseCase(
	uow domain.UnitOfWork,
	exportService ports.ExportService,
	idGenerator ports.IDGenerator,
	auditLogger ports.AuditLogger,
	config ExportConfig,
) *StreamExportCustomersUseCase {
	return &StreamExportCustomersUseCase{
		uow:           uow,
		exportService: exportService,
		idGenerator:   idGenerator,
		auditLogger:   auditLogger,
		config:        config,
	}
}

// StreamExportInput holds input for streaming export.
type StreamExportInput struct {
	TenantID uuid.UUID
	UserID   uuid.UUID
	Format   string
	Query    string
	Writer   io.Writer
}

// Execute streams customer export to a writer.
func (uc *StreamExportCustomersUseCase) Execute(ctx context.Context, input StreamExportInput) error {
	// Validate input
	if input.TenantID == uuid.Nil {
		return application.ErrInvalidInput("tenant_id is required")
	}
	if input.Writer == nil {
		return application.ErrInvalidInput("writer is required")
	}

	// Use export service to stream
	if err := uc.exportService.StreamExport(ctx, input.TenantID, input.Query, ports.ExportFormat(input.Format), input.Writer); err != nil {
		return application.ErrExportFailed(err.Error())
	}

	// Audit log
	if uc.auditLogger != nil {
		_ = uc.auditLogger.LogAction(ctx, ports.AuditEntry{
			ID:         uc.idGenerator.NewID(),
			TenantID:   input.TenantID,
			UserID:     &input.UserID,
			Action:     "customers.stream_exported",
			EntityType: "export",
			EntityID:   uuid.Nil,
			Metadata: map[string]interface{}{
				"format": input.Format,
				"query":  input.Query,
			},
			Timestamp: time.Now().UTC(),
		})
	}

	return nil
}
