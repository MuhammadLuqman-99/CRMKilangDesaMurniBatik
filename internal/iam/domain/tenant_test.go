// Package domain contains the domain layer for the IAM service.
package domain

import (
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestTenantStatus_IsValid(t *testing.T) {
	tests := []struct {
		status   TenantStatus
		expected bool
	}{
		{TenantStatusActive, true},
		{TenantStatusInactive, true},
		{TenantStatusSuspended, true},
		{TenantStatusPending, true},
		{TenantStatusTrial, true},
		{TenantStatus("invalid"), false},
		{TenantStatus(""), false},
	}

	for _, tt := range tests {
		t.Run(string(tt.status), func(t *testing.T) {
			if tt.status.IsValid() != tt.expected {
				t.Errorf("TenantStatus.IsValid() = %v, want %v", tt.status.IsValid(), tt.expected)
			}
		})
	}
}

func TestTenantStatus_String(t *testing.T) {
	status := TenantStatusActive
	if status.String() != "active" {
		t.Errorf("TenantStatus.String() = %v, want %v", status.String(), "active")
	}
}

func TestTenantPlan_IsValid(t *testing.T) {
	tests := []struct {
		plan     TenantPlan
		expected bool
	}{
		{TenantPlanFree, true},
		{TenantPlanStarter, true},
		{TenantPlanPro, true},
		{TenantPlanEnterprise, true},
		{TenantPlan("invalid"), false},
		{TenantPlan(""), false},
	}

	for _, tt := range tests {
		t.Run(string(tt.plan), func(t *testing.T) {
			if tt.plan.IsValid() != tt.expected {
				t.Errorf("TenantPlan.IsValid() = %v, want %v", tt.plan.IsValid(), tt.expected)
			}
		})
	}
}

func TestTenantPlan_String(t *testing.T) {
	plan := TenantPlanPro
	if plan.String() != "pro" {
		t.Errorf("TenantPlan.String() = %v, want %v", plan.String(), "pro")
	}
}

func TestTenantPlan_MaxUsers(t *testing.T) {
	tests := []struct {
		plan     TenantPlan
		expected int
	}{
		{TenantPlanFree, 3},
		{TenantPlanStarter, 10},
		{TenantPlanPro, 50},
		{TenantPlanEnterprise, -1}, // Unlimited
		{TenantPlan("invalid"), 0},
	}

	for _, tt := range tests {
		t.Run(string(tt.plan), func(t *testing.T) {
			if tt.plan.MaxUsers() != tt.expected {
				t.Errorf("TenantPlan.MaxUsers() = %v, want %v", tt.plan.MaxUsers(), tt.expected)
			}
		})
	}
}

func TestTenantPlan_MaxContacts(t *testing.T) {
	tests := []struct {
		plan     TenantPlan
		expected int
	}{
		{TenantPlanFree, 100},
		{TenantPlanStarter, 1000},
		{TenantPlanPro, 10000},
		{TenantPlanEnterprise, -1}, // Unlimited
		{TenantPlan("invalid"), 0},
	}

	for _, tt := range tests {
		t.Run(string(tt.plan), func(t *testing.T) {
			if tt.plan.MaxContacts() != tt.expected {
				t.Errorf("TenantPlan.MaxContacts() = %v, want %v", tt.plan.MaxContacts(), tt.expected)
			}
		})
	}
}

func TestDefaultTenantSettings(t *testing.T) {
	settings := DefaultTenantSettings()

	if settings.Timezone != "UTC" {
		t.Errorf("DefaultTenantSettings() Timezone = %v, want UTC", settings.Timezone)
	}
	if settings.DateFormat != "YYYY-MM-DD" {
		t.Errorf("DefaultTenantSettings() DateFormat = %v", settings.DateFormat)
	}
	if settings.Currency != "USD" {
		t.Errorf("DefaultTenantSettings() Currency = %v, want USD", settings.Currency)
	}
	if settings.Language != "en" {
		t.Errorf("DefaultTenantSettings() Language = %v, want en", settings.Language)
	}
	if !settings.NotificationsEmail {
		t.Error("DefaultTenantSettings() NotificationsEmail should be true")
	}
	if settings.Custom == nil {
		t.Error("DefaultTenantSettings() Custom should be initialized")
	}
}

func TestNewTenant(t *testing.T) {
	tests := []struct {
		name        string
		tenantName  string
		slug        string
		wantErr     bool
		expectedErr error
	}{
		{
			name:       "valid tenant",
			tenantName: "Test Tenant",
			slug:       "test-tenant",
			wantErr:    false,
		},
		{
			name:       "valid slug with numbers",
			tenantName: "Tenant 123",
			slug:       "tenant-123",
			wantErr:    false,
		},
		{
			name:        "empty name returns error",
			tenantName:  "",
			slug:        "test-slug",
			wantErr:     true,
			expectedErr: ErrTenantNameRequired,
		},
		{
			name:        "whitespace only name returns error",
			tenantName:  "   ",
			slug:        "test-slug",
			wantErr:     true,
			expectedErr: ErrTenantNameRequired,
		},
		{
			name:        "name too long returns error",
			tenantName:  strings.Repeat("a", 256),
			slug:        "test-slug",
			wantErr:     true,
			expectedErr: ErrTenantNameTooLong,
		},
		{
			name:        "empty slug returns error",
			tenantName:  "Test Tenant",
			slug:        "",
			wantErr:     true,
			expectedErr: ErrTenantSlugRequired,
		},
		{
			name:        "slug too long returns error",
			tenantName:  "Test Tenant",
			slug:        strings.Repeat("a", 101),
			wantErr:     true,
			expectedErr: ErrTenantSlugTooLong,
		},
		{
			name:        "invalid slug format with uppercase",
			tenantName:  "Test Tenant",
			slug:        "Test-Slug",
			wantErr:     false, // Will be converted to lowercase
		},
		{
			name:        "invalid slug with special chars",
			tenantName:  "Test Tenant",
			slug:        "test_slug!",
			wantErr:     true,
			expectedErr: ErrTenantSlugInvalid,
		},
		{
			name:        "invalid slug starting with hyphen",
			tenantName:  "Test Tenant",
			slug:        "-test-slug",
			wantErr:     true,
			expectedErr: ErrTenantSlugInvalid,
		},
		{
			name:       "valid slug with multiple hyphens",
			tenantName: "Test Tenant",
			slug:       "test-tenant-slug",
			wantErr:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tenant, err := NewTenant(tt.tenantName, tt.slug)
			if tt.wantErr {
				if err == nil {
					t.Errorf("NewTenant() expected error, got nil")
				}
				if tt.expectedErr != nil && err != tt.expectedErr {
					t.Errorf("NewTenant() error = %v, expectedErr = %v", err, tt.expectedErr)
				}
			} else {
				if err != nil {
					t.Errorf("NewTenant() unexpected error = %v", err)
				}
				if tenant == nil {
					t.Fatal("NewTenant() returned nil tenant")
				}
				if tenant.ID == uuid.Nil {
					t.Error("NewTenant() should generate ID")
				}
				if tenant.Status() != TenantStatusPending {
					t.Errorf("NewTenant() status = %v, want %v", tenant.Status(), TenantStatusPending)
				}
				if tenant.Plan() != TenantPlanFree {
					t.Errorf("NewTenant() plan = %v, want %v", tenant.Plan(), TenantPlanFree)
				}
				// Check domain event was added
				events := tenant.GetDomainEvents()
				if len(events) != 1 {
					t.Errorf("NewTenant() should add TenantCreatedEvent, got %d events", len(events))
				}
			}
		})
	}
}

func TestNewTenantWithPlan(t *testing.T) {
	tenant, err := NewTenantWithPlan("Test Tenant", "test-tenant", TenantPlanPro)
	if err != nil {
		t.Fatalf("NewTenantWithPlan() unexpected error = %v", err)
	}

	if tenant.Plan() != TenantPlanPro {
		t.Errorf("NewTenantWithPlan() plan = %v, want %v", tenant.Plan(), TenantPlanPro)
	}
}

func TestNewTenantWithPlan_InvalidPlan(t *testing.T) {
	_, err := NewTenantWithPlan("Test Tenant", "test-tenant", TenantPlan("invalid"))
	if err != ErrTenantPlanInvalid {
		t.Errorf("NewTenantWithPlan() with invalid plan should return ErrTenantPlanInvalid, got %v", err)
	}
}

func TestReconstructTenant(t *testing.T) {
	id := uuid.New()
	settings := DefaultTenantSettings()
	metadata := map[string]interface{}{"key": "value"}
	createdAt := time.Now().Add(-time.Hour)
	updatedAt := time.Now()
	deletedAt := time.Now().Add(-time.Minute)

	tenant := ReconstructTenant(
		id,
		"Reconstructed Tenant",
		"reconstructed-tenant",
		TenantStatusActive,
		TenantPlanPro,
		settings,
		metadata,
		createdAt,
		updatedAt,
		&deletedAt,
	)

	if tenant.ID != id {
		t.Errorf("ReconstructTenant() ID = %v, want %v", tenant.ID, id)
	}
	if tenant.Name() != "Reconstructed Tenant" {
		t.Errorf("ReconstructTenant() Name = %v", tenant.Name())
	}
	if tenant.Slug() != "reconstructed-tenant" {
		t.Errorf("ReconstructTenant() Slug = %v", tenant.Slug())
	}
	if tenant.Status() != TenantStatusActive {
		t.Errorf("ReconstructTenant() Status = %v", tenant.Status())
	}
	if tenant.Plan() != TenantPlanPro {
		t.Errorf("ReconstructTenant() Plan = %v", tenant.Plan())
	}
	if tenant.IsDeleted() != true {
		t.Error("ReconstructTenant() should be deleted")
	}
	// Reconstructed tenant should have no events
	if len(tenant.GetDomainEvents()) != 0 {
		t.Error("ReconstructTenant() should not have domain events")
	}
}

func TestReconstructTenant_NilMetadata(t *testing.T) {
	tenant := ReconstructTenant(
		uuid.New(),
		"Tenant",
		"tenant",
		TenantStatusActive,
		TenantPlanFree,
		DefaultTenantSettings(),
		nil,
		time.Now(),
		time.Now(),
		nil,
	)

	if tenant.Metadata() == nil {
		t.Error("ReconstructTenant() should initialize nil metadata")
	}
}

func TestTenant_Getters(t *testing.T) {
	tenant, _ := NewTenant("Test Tenant", "test-tenant")

	if tenant.Name() != "Test Tenant" {
		t.Errorf("Tenant.Name() = %v", tenant.Name())
	}
	if tenant.Slug() != "test-tenant" {
		t.Errorf("Tenant.Slug() = %v", tenant.Slug())
	}
	if tenant.Status() != TenantStatusPending {
		t.Errorf("Tenant.Status() = %v", tenant.Status())
	}
	if tenant.Plan() != TenantPlanFree {
		t.Errorf("Tenant.Plan() = %v", tenant.Plan())
	}
}

func TestTenant_IsActive(t *testing.T) {
	tests := []struct {
		status   TenantStatus
		expected bool
	}{
		{TenantStatusActive, true},
		{TenantStatusTrial, true},
		{TenantStatusPending, false},
		{TenantStatusInactive, false},
		{TenantStatusSuspended, false},
	}

	for _, tt := range tests {
		t.Run(string(tt.status), func(t *testing.T) {
			tenant := ReconstructTenant(
				uuid.New(), "Tenant", "tenant", tt.status, TenantPlanFree,
				DefaultTenantSettings(), nil, time.Now(), time.Now(), nil,
			)
			if tenant.IsActive() != tt.expected {
				t.Errorf("Tenant.IsActive() = %v, want %v", tenant.IsActive(), tt.expected)
			}
		})
	}
}

func TestTenant_IsSuspended(t *testing.T) {
	tenant := ReconstructTenant(
		uuid.New(), "Tenant", "tenant", TenantStatusSuspended, TenantPlanFree,
		DefaultTenantSettings(), nil, time.Now(), time.Now(), nil,
	)

	if !tenant.IsSuspended() {
		t.Error("Tenant.IsSuspended() should return true for suspended tenant")
	}

	tenant2 := ReconstructTenant(
		uuid.New(), "Tenant", "tenant", TenantStatusActive, TenantPlanFree,
		DefaultTenantSettings(), nil, time.Now(), time.Now(), nil,
	)

	if tenant2.IsSuspended() {
		t.Error("Tenant.IsSuspended() should return false for active tenant")
	}
}

func TestTenant_UpdateName(t *testing.T) {
	tenant, _ := NewTenant("Original Name", "test-tenant")
	tenant.ClearDomainEvents()

	err := tenant.UpdateName("New Name")
	if err != nil {
		t.Fatalf("Tenant.UpdateName() unexpected error = %v", err)
	}

	if tenant.Name() != "New Name" {
		t.Errorf("Tenant.UpdateName() name = %v, want %v", tenant.Name(), "New Name")
	}

	events := tenant.GetDomainEvents()
	if len(events) != 1 {
		t.Errorf("Tenant.UpdateName() should add event, got %d events", len(events))
	}
}

func TestTenant_UpdateName_Validation(t *testing.T) {
	tenant, _ := NewTenant("Original Name", "test-tenant")

	// Empty name
	err := tenant.UpdateName("")
	if err != ErrTenantNameRequired {
		t.Errorf("Tenant.UpdateName() with empty name should return ErrTenantNameRequired, got %v", err)
	}

	// Name too long
	err = tenant.UpdateName(strings.Repeat("a", 256))
	if err != ErrTenantNameTooLong {
		t.Errorf("Tenant.UpdateName() with long name should return ErrTenantNameTooLong, got %v", err)
	}
}

func TestTenant_UpdateSettings(t *testing.T) {
	tenant, _ := NewTenant("Test Tenant", "test-tenant")
	tenant.ClearDomainEvents()

	newSettings := TenantSettings{
		Timezone:           "America/New_York",
		DateFormat:         "MM/DD/YYYY",
		Currency:           "EUR",
		Language:           "de",
		NotificationsEmail: false,
		Custom:             map[string]interface{}{"custom_key": "custom_value"},
	}

	tenant.UpdateSettings(newSettings)

	if tenant.Settings().Timezone != "America/New_York" {
		t.Errorf("Tenant.UpdateSettings() Timezone = %v", tenant.Settings().Timezone)
	}
	if tenant.Settings().Currency != "EUR" {
		t.Errorf("Tenant.UpdateSettings() Currency = %v", tenant.Settings().Currency)
	}

	events := tenant.GetDomainEvents()
	if len(events) != 1 {
		t.Errorf("Tenant.UpdateSettings() should add event, got %d events", len(events))
	}
}

func TestTenant_UpdateSettingsPartial(t *testing.T) {
	tenant, _ := NewTenant("Test Tenant", "test-tenant")

	updates := map[string]interface{}{
		"timezone":            "Europe/London",
		"currency":            "GBP",
		"notifications_email": false,
		"custom_field":        "custom_value",
	}

	tenant.UpdateSettingsPartial(updates)

	if tenant.Settings().Timezone != "Europe/London" {
		t.Errorf("Tenant.UpdateSettingsPartial() Timezone = %v", tenant.Settings().Timezone)
	}
	if tenant.Settings().Currency != "GBP" {
		t.Errorf("Tenant.UpdateSettingsPartial() Currency = %v", tenant.Settings().Currency)
	}
	if tenant.Settings().NotificationsEmail {
		t.Error("Tenant.UpdateSettingsPartial() NotificationsEmail should be false")
	}
	if tenant.Settings().Custom["custom_field"] != "custom_value" {
		t.Error("Tenant.UpdateSettingsPartial() should store custom field")
	}
}

func TestTenant_Activate(t *testing.T) {
	tenant, _ := NewTenant("Test Tenant", "test-tenant")
	tenant.ClearDomainEvents()

	err := tenant.Activate()
	if err != nil {
		t.Fatalf("Tenant.Activate() unexpected error = %v", err)
	}

	if tenant.Status() != TenantStatusActive {
		t.Errorf("Tenant.Activate() status = %v, want %v", tenant.Status(), TenantStatusActive)
	}

	events := tenant.GetDomainEvents()
	if len(events) != 1 {
		t.Errorf("Tenant.Activate() should add event, got %d events", len(events))
	}

	// Activating already active tenant should be idempotent
	tenant.ClearDomainEvents()
	err = tenant.Activate()
	if err != nil {
		t.Errorf("Tenant.Activate() on active tenant should be idempotent, got error = %v", err)
	}
	if len(tenant.GetDomainEvents()) != 0 {
		t.Error("Tenant.Activate() on active tenant should not add event")
	}
}

func TestTenant_Activate_DeletedTenant(t *testing.T) {
	deletedAt := time.Now()
	tenant := ReconstructTenant(
		uuid.New(), "Tenant", "tenant", TenantStatusPending, TenantPlanFree,
		DefaultTenantSettings(), nil, time.Now(), time.Now(), &deletedAt,
	)

	err := tenant.Activate()
	if err != ErrTenantDeleted {
		t.Errorf("Tenant.Activate() on deleted tenant should return ErrTenantDeleted, got %v", err)
	}
}

func TestTenant_Deactivate(t *testing.T) {
	tenant := ReconstructTenant(
		uuid.New(), "Tenant", "tenant", TenantStatusActive, TenantPlanFree,
		DefaultTenantSettings(), nil, time.Now(), time.Now(), nil,
	)
	tenant.ClearDomainEvents()

	err := tenant.Deactivate()
	if err != nil {
		t.Fatalf("Tenant.Deactivate() unexpected error = %v", err)
	}

	if tenant.Status() != TenantStatusInactive {
		t.Errorf("Tenant.Deactivate() status = %v, want %v", tenant.Status(), TenantStatusInactive)
	}

	events := tenant.GetDomainEvents()
	if len(events) != 1 {
		t.Errorf("Tenant.Deactivate() should add event, got %d events", len(events))
	}
}

func TestTenant_Suspend(t *testing.T) {
	tenant := ReconstructTenant(
		uuid.New(), "Tenant", "tenant", TenantStatusActive, TenantPlanFree,
		DefaultTenantSettings(), nil, time.Now(), time.Now(), nil,
	)
	tenant.ClearDomainEvents()

	err := tenant.Suspend("Payment overdue")
	if err != nil {
		t.Fatalf("Tenant.Suspend() unexpected error = %v", err)
	}

	if tenant.Status() != TenantStatusSuspended {
		t.Errorf("Tenant.Suspend() status = %v, want %v", tenant.Status(), TenantStatusSuspended)
	}

	reason, ok := tenant.Metadata()["suspend_reason"].(string)
	if !ok || reason != "Payment overdue" {
		t.Errorf("Tenant.Suspend() should store reason in metadata")
	}

	_, ok = tenant.Metadata()["suspended_at"]
	if !ok {
		t.Error("Tenant.Suspend() should store suspended_at in metadata")
	}

	events := tenant.GetDomainEvents()
	if len(events) != 1 {
		t.Errorf("Tenant.Suspend() should add event, got %d events", len(events))
	}
}

func TestTenant_Unsuspend(t *testing.T) {
	tenant := ReconstructTenant(
		uuid.New(), "Tenant", "tenant", TenantStatusSuspended, TenantPlanFree,
		DefaultTenantSettings(), map[string]interface{}{
			"suspend_reason": "Payment overdue",
			"suspended_at":   time.Now(),
		}, time.Now(), time.Now(), nil,
	)

	err := tenant.Unsuspend()
	if err != nil {
		t.Fatalf("Tenant.Unsuspend() unexpected error = %v", err)
	}

	if tenant.Status() != TenantStatusActive {
		t.Errorf("Tenant.Unsuspend() status = %v, want %v", tenant.Status(), TenantStatusActive)
	}

	if _, ok := tenant.Metadata()["suspend_reason"]; ok {
		t.Error("Tenant.Unsuspend() should remove suspend_reason from metadata")
	}
	if _, ok := tenant.Metadata()["suspended_at"]; ok {
		t.Error("Tenant.Unsuspend() should remove suspended_at from metadata")
	}
}

func TestTenant_Unsuspend_NotSuspended(t *testing.T) {
	tenant := ReconstructTenant(
		uuid.New(), "Tenant", "tenant", TenantStatusActive, TenantPlanFree,
		DefaultTenantSettings(), nil, time.Now(), time.Now(), nil,
	)

	err := tenant.Unsuspend()
	if err != ErrTenantNotSuspended {
		t.Errorf("Tenant.Unsuspend() on non-suspended tenant should return ErrTenantNotSuspended, got %v", err)
	}
}

func TestTenant_StartTrial(t *testing.T) {
	tenant, _ := NewTenant("Test Tenant", "test-tenant")
	tenant.ClearDomainEvents()

	err := tenant.StartTrial(14)
	if err != nil {
		t.Fatalf("Tenant.StartTrial() unexpected error = %v", err)
	}

	if tenant.Status() != TenantStatusTrial {
		t.Errorf("Tenant.StartTrial() status = %v, want %v", tenant.Status(), TenantStatusTrial)
	}

	_, ok := tenant.Metadata()["trial_started_at"]
	if !ok {
		t.Error("Tenant.StartTrial() should store trial_started_at in metadata")
	}

	_, ok = tenant.Metadata()["trial_ends_at"]
	if !ok {
		t.Error("Tenant.StartTrial() should store trial_ends_at in metadata")
	}

	events := tenant.GetDomainEvents()
	if len(events) != 1 {
		t.Errorf("Tenant.StartTrial() should add event, got %d events", len(events))
	}
}

func TestTenant_StartTrial_AlreadyActivated(t *testing.T) {
	tenant := ReconstructTenant(
		uuid.New(), "Tenant", "tenant", TenantStatusActive, TenantPlanFree,
		DefaultTenantSettings(), nil, time.Now(), time.Now(), nil,
	)

	err := tenant.StartTrial(14)
	if err != ErrTenantAlreadyActivated {
		t.Errorf("Tenant.StartTrial() on active tenant should return ErrTenantAlreadyActivated, got %v", err)
	}
}

func TestTenant_IsTrialExpired(t *testing.T) {
	// Trial not expired
	tenant := ReconstructTenant(
		uuid.New(), "Tenant", "tenant", TenantStatusTrial, TenantPlanFree,
		DefaultTenantSettings(), map[string]interface{}{
			"trial_ends_at": time.Now().Add(24 * time.Hour),
		}, time.Now(), time.Now(), nil,
	)

	if tenant.IsTrialExpired() {
		t.Error("Tenant.IsTrialExpired() should return false for non-expired trial")
	}

	// Trial expired
	tenant2 := ReconstructTenant(
		uuid.New(), "Tenant", "tenant", TenantStatusTrial, TenantPlanFree,
		DefaultTenantSettings(), map[string]interface{}{
			"trial_ends_at": time.Now().Add(-24 * time.Hour),
		}, time.Now(), time.Now(), nil,
	)

	if !tenant2.IsTrialExpired() {
		t.Error("Tenant.IsTrialExpired() should return true for expired trial")
	}

	// Not on trial
	tenant3 := ReconstructTenant(
		uuid.New(), "Tenant", "tenant", TenantStatusActive, TenantPlanFree,
		DefaultTenantSettings(), nil, time.Now(), time.Now(), nil,
	)

	if tenant3.IsTrialExpired() {
		t.Error("Tenant.IsTrialExpired() should return false for non-trial tenant")
	}
}

func TestTenant_UpgradePlan(t *testing.T) {
	tenant := ReconstructTenant(
		uuid.New(), "Tenant", "tenant", TenantStatusActive, TenantPlanFree,
		DefaultTenantSettings(), nil, time.Now(), time.Now(), nil,
	)
	tenant.ClearDomainEvents()

	err := tenant.UpgradePlan(TenantPlanPro)
	if err != nil {
		t.Fatalf("Tenant.UpgradePlan() unexpected error = %v", err)
	}

	if tenant.Plan() != TenantPlanPro {
		t.Errorf("Tenant.UpgradePlan() plan = %v, want %v", tenant.Plan(), TenantPlanPro)
	}

	events := tenant.GetDomainEvents()
	if len(events) != 1 {
		t.Errorf("Tenant.UpgradePlan() should add event, got %d events", len(events))
	}
}

func TestTenant_UpgradePlan_InvalidPlan(t *testing.T) {
	tenant := ReconstructTenant(
		uuid.New(), "Tenant", "tenant", TenantStatusActive, TenantPlanFree,
		DefaultTenantSettings(), nil, time.Now(), time.Now(), nil,
	)

	err := tenant.UpgradePlan(TenantPlan("invalid"))
	if err != ErrTenantPlanInvalid {
		t.Errorf("Tenant.UpgradePlan() with invalid plan should return ErrTenantPlanInvalid, got %v", err)
	}
}

func TestTenant_DowngradePlan(t *testing.T) {
	tenant := ReconstructTenant(
		uuid.New(), "Tenant", "tenant", TenantStatusActive, TenantPlanPro,
		DefaultTenantSettings(), nil, time.Now(), time.Now(), nil,
	)
	tenant.ClearDomainEvents()

	err := tenant.DowngradePlan(TenantPlanStarter)
	if err != nil {
		t.Fatalf("Tenant.DowngradePlan() unexpected error = %v", err)
	}

	if tenant.Plan() != TenantPlanStarter {
		t.Errorf("Tenant.DowngradePlan() plan = %v, want %v", tenant.Plan(), TenantPlanStarter)
	}
}

func TestTenant_Delete(t *testing.T) {
	tenant, _ := NewTenant("Test Tenant", "test-tenant")
	tenant.ClearDomainEvents()

	tenant.Delete()

	if !tenant.IsDeleted() {
		t.Error("Tenant.Delete() should mark tenant as deleted")
	}

	events := tenant.GetDomainEvents()
	if len(events) != 1 {
		t.Errorf("Tenant.Delete() should add event, got %d events", len(events))
	}

	// Deleting again should be idempotent
	tenant.ClearDomainEvents()
	tenant.Delete()
	if len(tenant.GetDomainEvents()) != 0 {
		t.Error("Tenant.Delete() on deleted tenant should be idempotent")
	}
}

func TestTenant_CanAddUser(t *testing.T) {
	tests := []struct {
		plan         TenantPlan
		currentCount int
		expected     bool
	}{
		{TenantPlanFree, 2, true},
		{TenantPlanFree, 3, false},
		{TenantPlanStarter, 9, true},
		{TenantPlanStarter, 10, false},
		{TenantPlanPro, 49, true},
		{TenantPlanPro, 50, false},
		{TenantPlanEnterprise, 1000, true}, // Unlimited
	}

	for _, tt := range tests {
		t.Run(string(tt.plan), func(t *testing.T) {
			tenant := ReconstructTenant(
				uuid.New(), "Tenant", "tenant", TenantStatusActive, tt.plan,
				DefaultTenantSettings(), nil, time.Now(), time.Now(), nil,
			)
			if tenant.CanAddUser(tt.currentCount) != tt.expected {
				t.Errorf("Tenant.CanAddUser() = %v, want %v", tenant.CanAddUser(tt.currentCount), tt.expected)
			}
		})
	}
}

func TestTenant_CanAddContact(t *testing.T) {
	tests := []struct {
		plan         TenantPlan
		currentCount int
		expected     bool
	}{
		{TenantPlanFree, 99, true},
		{TenantPlanFree, 100, false},
		{TenantPlanStarter, 999, true},
		{TenantPlanStarter, 1000, false},
		{TenantPlanEnterprise, 100000, true}, // Unlimited
	}

	for _, tt := range tests {
		t.Run(string(tt.plan), func(t *testing.T) {
			tenant := ReconstructTenant(
				uuid.New(), "Tenant", "tenant", TenantStatusActive, tt.plan,
				DefaultTenantSettings(), nil, time.Now(), time.Now(), nil,
			)
			if tenant.CanAddContact(tt.currentCount) != tt.expected {
				t.Errorf("Tenant.CanAddContact() = %v, want %v", tenant.CanAddContact(tt.currentCount), tt.expected)
			}
		})
	}
}

func TestTenant_MetadataManagement(t *testing.T) {
	tenant, _ := NewTenant("Test Tenant", "test-tenant")

	// Set metadata
	tenant.SetMetadata("key1", "value1")
	tenant.SetMetadata("key2", 123)

	// Get metadata
	val, ok := tenant.GetMetadata("key1")
	if !ok || val != "value1" {
		t.Errorf("Tenant.GetMetadata() = %v, %v, want value1, true", val, ok)
	}

	// Get non-existent metadata
	_, ok = tenant.GetMetadata("nonexistent")
	if ok {
		t.Error("Tenant.GetMetadata() should return false for non-existent key")
	}

	// Delete metadata
	tenant.DeleteMetadata("key1")
	_, ok = tenant.GetMetadata("key1")
	if ok {
		t.Error("Tenant.DeleteMetadata() should remove key")
	}
}

func TestTenant_MetadataManagement_NilMetadata(t *testing.T) {
	tenant := ReconstructTenant(
		uuid.New(), "Tenant", "tenant", TenantStatusActive, TenantPlanFree,
		DefaultTenantSettings(), nil, time.Now(), time.Now(), nil,
	)

	// Setting metadata should initialize the map
	tenant.SetMetadata("key", "value")
	val, ok := tenant.GetMetadata("key")
	if !ok || val != "value" {
		t.Error("Tenant.SetMetadata() should work with nil metadata")
	}

	// Delete from nil metadata should not panic
	tenant2 := ReconstructTenant(
		uuid.New(), "Tenant", "tenant", TenantStatusActive, TenantPlanFree,
		DefaultTenantSettings(), nil, time.Now(), time.Now(), nil,
	)
	tenant2.DeleteMetadata("key") // Should not panic
}

func TestTenant_SlugConversion(t *testing.T) {
	tenant, err := NewTenant("Test Tenant", "TEST-TENANT")
	if err != nil {
		t.Fatalf("NewTenant() unexpected error = %v", err)
	}

	if tenant.Slug() != "test-tenant" {
		t.Errorf("NewTenant() should convert slug to lowercase, got %v", tenant.Slug())
	}
}
