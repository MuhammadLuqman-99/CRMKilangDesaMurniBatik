// Package fixtures provides test data fixtures for integration testing.
package fixtures

import (
	"time"

	"github.com/google/uuid"
)

// TestIDs contains commonly used test UUIDs.
var TestIDs = struct {
	TenantID1 uuid.UUID
	TenantID2 uuid.UUID
	UserID1   uuid.UUID
	UserID2   uuid.UUID
	UserID3   uuid.UUID
	RoleID1   uuid.UUID
	RoleID2   uuid.UUID
	CustomerID1 uuid.UUID
	CustomerID2 uuid.UUID
	CustomerID3 uuid.UUID
	ContactID1  uuid.UUID
	ContactID2  uuid.UUID
	LeadID1     uuid.UUID
	LeadID2     uuid.UUID
	OpportunityID1 uuid.UUID
	OpportunityID2 uuid.UUID
	DealID1     uuid.UUID
	DealID2     uuid.UUID
	PipelineID1 uuid.UUID
	NotificationID1 uuid.UUID
	TemplateID1 uuid.UUID
}{
	TenantID1:      uuid.MustParse("11111111-1111-1111-1111-111111111111"),
	TenantID2:      uuid.MustParse("22222222-2222-2222-2222-222222222222"),
	UserID1:        uuid.MustParse("aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa"),
	UserID2:        uuid.MustParse("bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb"),
	UserID3:        uuid.MustParse("cccccccc-cccc-cccc-cccc-cccccccccccc"),
	RoleID1:        uuid.MustParse("dddddddd-dddd-dddd-dddd-dddddddddddd"),
	RoleID2:        uuid.MustParse("eeeeeeee-eeee-eeee-eeee-eeeeeeeeeeee"),
	CustomerID1:    uuid.MustParse("f1111111-1111-1111-1111-111111111111"),
	CustomerID2:    uuid.MustParse("f2222222-2222-2222-2222-222222222222"),
	CustomerID3:    uuid.MustParse("f3333333-3333-3333-3333-333333333333"),
	ContactID1:     uuid.MustParse("c1111111-1111-1111-1111-111111111111"),
	ContactID2:     uuid.MustParse("c2222222-2222-2222-2222-222222222222"),
	LeadID1:        uuid.MustParse("1eadaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa"),
	LeadID2:        uuid.MustParse("1eadbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb"),
	OpportunityID1: uuid.MustParse("0777aaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa"),
	OpportunityID2: uuid.MustParse("0777bbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb"),
	DealID1:        uuid.MustParse("dea1aaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa"),
	DealID2:        uuid.MustParse("dea1bbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb"),
	PipelineID1:    uuid.MustParse("71711111-1111-1111-1111-111111111111"),
	NotificationID1: uuid.MustParse("90711111-1111-1111-1111-111111111111"),
	TemplateID1:    uuid.MustParse("7e911111-1111-1111-1111-111111111111"),
}

// TenantFixture represents a test tenant.
type TenantFixture struct {
	ID        uuid.UUID
	Name      string
	Slug      string
	Status    string
	Plan      string
	Settings  map[string]interface{}
	Metadata  map[string]interface{}
	CreatedAt time.Time
	UpdatedAt time.Time
}

// DefaultTenantFixtures returns default tenant fixtures.
func DefaultTenantFixtures() []TenantFixture {
	now := time.Now().UTC()
	return []TenantFixture{
		{
			ID:        TestIDs.TenantID1,
			Name:      "Test Tenant 1",
			Slug:      "test-tenant-1",
			Status:    "active",
			Plan:      "professional",
			Settings:  map[string]interface{}{"timezone": "UTC"},
			Metadata:  map[string]interface{}{"source": "test"},
			CreatedAt: now,
			UpdatedAt: now,
		},
		{
			ID:        TestIDs.TenantID2,
			Name:      "Test Tenant 2",
			Slug:      "test-tenant-2",
			Status:    "active",
			Plan:      "enterprise",
			Settings:  map[string]interface{}{"timezone": "Asia/Jakarta"},
			Metadata:  map[string]interface{}{"source": "test"},
			CreatedAt: now,
			UpdatedAt: now,
		},
	}
}

// UserFixture represents a test user.
type UserFixture struct {
	ID              uuid.UUID
	TenantID        uuid.UUID
	Email           string
	PasswordHash    string
	FirstName       string
	LastName        string
	Status          string
	EmailVerifiedAt *time.Time
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

// DefaultUserFixtures returns default user fixtures.
func DefaultUserFixtures() []UserFixture {
	now := time.Now().UTC()
	return []UserFixture{
		{
			ID:              TestIDs.UserID1,
			TenantID:        TestIDs.TenantID1,
			Email:           "admin@test.com",
			PasswordHash:    "$argon2id$v=19$m=65536,t=3,p=4$test$hash", // Not a real hash
			FirstName:       "Admin",
			LastName:        "User",
			Status:          "active",
			EmailVerifiedAt: &now,
			CreatedAt:       now,
			UpdatedAt:       now,
		},
		{
			ID:           TestIDs.UserID2,
			TenantID:     TestIDs.TenantID1,
			Email:        "user@test.com",
			PasswordHash: "$argon2id$v=19$m=65536,t=3,p=4$test$hash",
			FirstName:    "Regular",
			LastName:     "User",
			Status:       "active",
			CreatedAt:    now,
			UpdatedAt:    now,
		},
		{
			ID:           TestIDs.UserID3,
			TenantID:     TestIDs.TenantID2,
			Email:        "user2@test.com",
			PasswordHash: "$argon2id$v=19$m=65536,t=3,p=4$test$hash",
			FirstName:    "Tenant2",
			LastName:     "User",
			Status:       "active",
			CreatedAt:    now,
			UpdatedAt:    now,
		},
	}
}

// RoleFixture represents a test role.
type RoleFixture struct {
	ID          uuid.UUID
	TenantID    *uuid.UUID
	Name        string
	Description string
	Permissions []string
	IsSystem    bool
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// DefaultRoleFixtures returns default role fixtures.
func DefaultRoleFixtures() []RoleFixture {
	now := time.Now().UTC()
	return []RoleFixture{
		{
			ID:          TestIDs.RoleID1,
			TenantID:    nil, // System role
			Name:        "admin",
			Description: "Administrator role",
			Permissions: []string{"*"},
			IsSystem:    true,
			CreatedAt:   now,
			UpdatedAt:   now,
		},
		{
			ID:          TestIDs.RoleID2,
			TenantID:    &TestIDs.TenantID1,
			Name:        "sales_rep",
			Description: "Sales representative role",
			Permissions: []string{"customers:read", "customers:write", "leads:*"},
			IsSystem:    false,
			CreatedAt:   now,
			UpdatedAt:   now,
		},
	}
}

// CustomerFixture represents a test customer (for MongoDB).
type CustomerFixture struct {
	ID          uuid.UUID              `bson:"_id"`
	TenantID    uuid.UUID              `bson:"tenant_id"`
	Code        string                 `bson:"code"`
	Name        string                 `bson:"name"`
	Type        string                 `bson:"type"`
	Status      string                 `bson:"status"`
	Tier        string                 `bson:"tier"`
	Source      string                 `bson:"source"`
	Email       map[string]interface{} `bson:"email"`
	Tags        []string               `bson:"tags"`
	OwnerID     *uuid.UUID             `bson:"owner_id"`
	Version     int                    `bson:"version"`
	CreatedAt   time.Time              `bson:"created_at"`
	UpdatedAt   time.Time              `bson:"updated_at"`
	DeletedAt   *time.Time             `bson:"deleted_at"`
}

// DefaultCustomerFixtures returns default customer fixtures.
func DefaultCustomerFixtures() []CustomerFixture {
	now := time.Now().UTC()
	return []CustomerFixture{
		{
			ID:       TestIDs.CustomerID1,
			TenantID: TestIDs.TenantID1,
			Code:     "CUST-001",
			Name:     "Test Customer 1",
			Type:     "business",
			Status:   "active",
			Tier:     "premium",
			Source:   "referral",
			Email:    map[string]interface{}{"address": "customer1@test.com", "verified": true},
			Tags:     []string{"vip", "priority"},
			OwnerID:  &TestIDs.UserID1,
			Version:  1,
			CreatedAt: now,
			UpdatedAt: now,
		},
		{
			ID:       TestIDs.CustomerID2,
			TenantID: TestIDs.TenantID1,
			Code:     "CUST-002",
			Name:     "Test Customer 2",
			Type:     "individual",
			Status:   "active",
			Tier:     "standard",
			Source:   "website",
			Email:    map[string]interface{}{"address": "customer2@test.com", "verified": false},
			Tags:     []string{"new"},
			OwnerID:  &TestIDs.UserID2,
			Version:  1,
			CreatedAt: now,
			UpdatedAt: now,
		},
		{
			ID:       TestIDs.CustomerID3,
			TenantID: TestIDs.TenantID2,
			Code:     "CUST-003",
			Name:     "Tenant 2 Customer",
			Type:     "business",
			Status:   "active",
			Tier:     "enterprise",
			Source:   "direct",
			Email:    map[string]interface{}{"address": "customer3@test.com", "verified": true},
			Tags:     []string{"enterprise"},
			Version:  1,
			CreatedAt: now,
			UpdatedAt: now,
		},
	}
}

// ContactFixture represents a test contact.
type ContactFixture struct {
	ID         uuid.UUID              `bson:"_id"`
	TenantID   uuid.UUID              `bson:"tenant_id"`
	CustomerID uuid.UUID              `bson:"customer_id"`
	FirstName  string                 `bson:"first_name"`
	LastName   string                 `bson:"last_name"`
	Email      map[string]interface{} `bson:"email"`
	Phone      map[string]interface{} `bson:"phone"`
	Title      string                 `bson:"title"`
	Department string                 `bson:"department"`
	IsPrimary  bool                   `bson:"is_primary"`
	CreatedAt  time.Time              `bson:"created_at"`
	UpdatedAt  time.Time              `bson:"updated_at"`
}

// DefaultContactFixtures returns default contact fixtures.
func DefaultContactFixtures() []ContactFixture {
	now := time.Now().UTC()
	return []ContactFixture{
		{
			ID:         TestIDs.ContactID1,
			TenantID:   TestIDs.TenantID1,
			CustomerID: TestIDs.CustomerID1,
			FirstName:  "John",
			LastName:   "Doe",
			Email:      map[string]interface{}{"address": "john.doe@test.com"},
			Phone:      map[string]interface{}{"number": "+1234567890", "type": "mobile"},
			Title:      "CEO",
			Department: "Executive",
			IsPrimary:  true,
			CreatedAt:  now,
			UpdatedAt:  now,
		},
		{
			ID:         TestIDs.ContactID2,
			TenantID:   TestIDs.TenantID1,
			CustomerID: TestIDs.CustomerID1,
			FirstName:  "Jane",
			LastName:   "Smith",
			Email:      map[string]interface{}{"address": "jane.smith@test.com"},
			Phone:      map[string]interface{}{"number": "+1234567891", "type": "work"},
			Title:      "CTO",
			Department: "Technology",
			IsPrimary:  false,
			CreatedAt:  now,
			UpdatedAt:  now,
		},
	}
}

// LeadFixture represents a test lead.
type LeadFixture struct {
	ID           uuid.UUID
	TenantID     uuid.UUID
	CustomerID   *uuid.UUID
	Source       string
	Status       string
	Score        int
	CompanyName  string
	ContactName  string
	ContactEmail string
	ContactPhone string
	AssignedTo   *uuid.UUID
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

// DefaultLeadFixtures returns default lead fixtures.
func DefaultLeadFixtures() []LeadFixture {
	now := time.Now().UTC()
	return []LeadFixture{
		{
			ID:           TestIDs.LeadID1,
			TenantID:     TestIDs.TenantID1,
			Source:       "website",
			Status:       "new",
			Score:        75,
			CompanyName:  "Lead Company 1",
			ContactName:  "Lead Contact 1",
			ContactEmail: "lead1@test.com",
			ContactPhone: "+1234567892",
			AssignedTo:   &TestIDs.UserID1,
			CreatedAt:    now,
			UpdatedAt:    now,
		},
		{
			ID:           TestIDs.LeadID2,
			TenantID:     TestIDs.TenantID1,
			Source:       "referral",
			Status:       "qualified",
			Score:        90,
			CompanyName:  "Lead Company 2",
			ContactName:  "Lead Contact 2",
			ContactEmail: "lead2@test.com",
			ContactPhone: "+1234567893",
			AssignedTo:   &TestIDs.UserID2,
			CreatedAt:    now,
			UpdatedAt:    now,
		},
	}
}

// OpportunityFixture represents a test opportunity.
type OpportunityFixture struct {
	ID            uuid.UUID
	TenantID      uuid.UUID
	CustomerID    uuid.UUID
	LeadID        *uuid.UUID
	PipelineID    uuid.UUID
	StageID       uuid.UUID
	Name          string
	ValueAmount   int64
	ValueCurrency string
	Probability   int
	ExpectedClose time.Time
	AssignedTo    *uuid.UUID
	Status        string
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

// DefaultOpportunityFixtures returns default opportunity fixtures.
func DefaultOpportunityFixtures() []OpportunityFixture {
	now := time.Now().UTC()
	return []OpportunityFixture{
		{
			ID:            TestIDs.OpportunityID1,
			TenantID:      TestIDs.TenantID1,
			CustomerID:    TestIDs.CustomerID1,
			PipelineID:    TestIDs.PipelineID1,
			Name:          "Big Deal Opportunity",
			ValueAmount:   1000000, // $10,000.00 in cents
			ValueCurrency: "USD",
			Probability:   70,
			ExpectedClose: now.AddDate(0, 1, 0),
			AssignedTo:    &TestIDs.UserID1,
			Status:        "open",
			CreatedAt:     now,
			UpdatedAt:     now,
		},
		{
			ID:            TestIDs.OpportunityID2,
			TenantID:      TestIDs.TenantID1,
			CustomerID:    TestIDs.CustomerID2,
			LeadID:        &TestIDs.LeadID1,
			PipelineID:    TestIDs.PipelineID1,
			Name:          "Small Deal Opportunity",
			ValueAmount:   50000, // $500.00 in cents
			ValueCurrency: "USD",
			Probability:   50,
			ExpectedClose: now.AddDate(0, 0, 14),
			AssignedTo:    &TestIDs.UserID2,
			Status:        "open",
			CreatedAt:     now,
			UpdatedAt:     now,
		},
	}
}

// DealFixture represents a test deal.
type DealFixture struct {
	ID              uuid.UUID
	TenantID        uuid.UUID
	OpportunityID   uuid.UUID
	CustomerID      uuid.UUID
	Name            string
	ValueAmount     int64
	ValueCurrency   string
	Status          string
	ClosedAt        *time.Time
	AssignedTo      *uuid.UUID
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

// DefaultDealFixtures returns default deal fixtures.
func DefaultDealFixtures() []DealFixture {
	now := time.Now().UTC()
	closedAt := now.Add(-24 * time.Hour)
	return []DealFixture{
		{
			ID:            TestIDs.DealID1,
			TenantID:      TestIDs.TenantID1,
			OpportunityID: TestIDs.OpportunityID1,
			CustomerID:    TestIDs.CustomerID1,
			Name:          "Closed Deal 1",
			ValueAmount:   1000000,
			ValueCurrency: "USD",
			Status:        "won",
			ClosedAt:      &closedAt,
			AssignedTo:    &TestIDs.UserID1,
			CreatedAt:     now,
			UpdatedAt:     now,
		},
	}
}

// PipelineFixture represents a test pipeline.
type PipelineFixture struct {
	ID          uuid.UUID
	TenantID    uuid.UUID
	Name        string
	Description string
	IsDefault   bool
	Stages      []StageFixture
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// StageFixture represents a test pipeline stage.
type StageFixture struct {
	ID          uuid.UUID
	PipelineID  uuid.UUID
	Name        string
	Type        string
	Order       int
	Probability int
}

// DefaultPipelineFixtures returns default pipeline fixtures.
func DefaultPipelineFixtures() []PipelineFixture {
	now := time.Now().UTC()
	stageID1 := uuid.MustParse("57a61111-1111-1111-1111-111111111111")
	stageID2 := uuid.MustParse("57a62222-2222-2222-2222-222222222222")
	stageID3 := uuid.MustParse("57a63333-3333-3333-3333-333333333333")
	stageID4 := uuid.MustParse("57a64444-4444-4444-4444-444444444444")
	stageID5 := uuid.MustParse("57a65555-5555-5555-5555-555555555555")

	return []PipelineFixture{
		{
			ID:          TestIDs.PipelineID1,
			TenantID:    TestIDs.TenantID1,
			Name:        "Default Sales Pipeline",
			Description: "Main sales pipeline",
			IsDefault:   true,
			Stages: []StageFixture{
				{ID: stageID1, PipelineID: TestIDs.PipelineID1, Name: "Lead", Type: "open", Order: 1, Probability: 10},
				{ID: stageID2, PipelineID: TestIDs.PipelineID1, Name: "Qualified", Type: "open", Order: 2, Probability: 30},
				{ID: stageID3, PipelineID: TestIDs.PipelineID1, Name: "Proposal", Type: "open", Order: 3, Probability: 60},
				{ID: stageID4, PipelineID: TestIDs.PipelineID1, Name: "Won", Type: "won", Order: 4, Probability: 100},
				{ID: stageID5, PipelineID: TestIDs.PipelineID1, Name: "Lost", Type: "lost", Order: 5, Probability: 0},
			},
			CreatedAt: now,
			UpdatedAt: now,
		},
	}
}

// EventFixture represents a test event for event bus testing.
type EventFixture struct {
	ID            string
	Type          string
	TenantID      string
	AggregateID   string
	AggregateType string
	Version       int
	Timestamp     time.Time
	Data          map[string]interface{}
}

// DefaultEventFixtures returns default event fixtures.
func DefaultEventFixtures() []EventFixture {
	now := time.Now().UTC()
	return []EventFixture{
		{
			ID:            uuid.New().String(),
			Type:          "user.created",
			TenantID:      TestIDs.TenantID1.String(),
			AggregateID:   TestIDs.UserID1.String(),
			AggregateType: "User",
			Version:       1,
			Timestamp:     now,
			Data: map[string]interface{}{
				"email":      "newuser@test.com",
				"first_name": "New",
				"last_name":  "User",
			},
		},
		{
			ID:            uuid.New().String(),
			Type:          "customer.created",
			TenantID:      TestIDs.TenantID1.String(),
			AggregateID:   TestIDs.CustomerID1.String(),
			AggregateType: "Customer",
			Version:       1,
			Timestamp:     now,
			Data: map[string]interface{}{
				"name": "New Customer",
				"code": "CUST-NEW",
			},
		},
		{
			ID:            uuid.New().String(),
			Type:          "lead.converted",
			TenantID:      TestIDs.TenantID1.String(),
			AggregateID:   TestIDs.LeadID1.String(),
			AggregateType: "Lead",
			Version:       2,
			Timestamp:     now,
			Data: map[string]interface{}{
				"opportunity_id": TestIDs.OpportunityID1.String(),
				"converted_by":   TestIDs.UserID1.String(),
			},
		},
		{
			ID:            uuid.New().String(),
			Type:          "deal.won",
			TenantID:      TestIDs.TenantID1.String(),
			AggregateID:   TestIDs.DealID1.String(),
			AggregateType: "Deal",
			Version:       1,
			Timestamp:     now,
			Data: map[string]interface{}{
				"value_amount":   1000000,
				"value_currency": "USD",
				"customer_id":    TestIDs.CustomerID1.String(),
			},
		},
	}
}

// NewUUID generates a new UUID for testing.
func NewUUID() uuid.UUID {
	return uuid.New()
}

// TimeNow returns the current UTC time.
func TimeNow() time.Time {
	return time.Now().UTC()
}

// TimePast returns a time in the past.
func TimePast(d time.Duration) time.Time {
	return time.Now().UTC().Add(-d)
}

// TimeFuture returns a time in the future.
func TimeFuture(d time.Duration) time.Time {
	return time.Now().UTC().Add(d)
}
