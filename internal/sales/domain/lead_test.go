// Package domain contains the domain layer for the Sales Pipeline service.
package domain

import (
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestLeadStatus_IsValid(t *testing.T) {
	tests := []struct {
		status   LeadStatus
		expected bool
	}{
		{LeadStatusNew, true},
		{LeadStatusContacted, true},
		{LeadStatusQualified, true},
		{LeadStatusUnqualified, true},
		{LeadStatusConverted, true},
		{LeadStatusNurturing, true},
		{LeadStatus("invalid"), false},
		{LeadStatus(""), false},
	}

	for _, tt := range tests {
		t.Run(string(tt.status), func(t *testing.T) {
			if tt.status.IsValid() != tt.expected {
				t.Errorf("LeadStatus.IsValid() = %v, want %v", tt.status.IsValid(), tt.expected)
			}
		})
	}
}

func TestLeadStatus_CanConvert(t *testing.T) {
	tests := []struct {
		status   LeadStatus
		expected bool
	}{
		{LeadStatusQualified, true},
		{LeadStatusNew, false},
		{LeadStatusContacted, false},
		{LeadStatusConverted, false},
		{LeadStatusUnqualified, false},
		{LeadStatusNurturing, false},
	}

	for _, tt := range tests {
		t.Run(string(tt.status), func(t *testing.T) {
			if tt.status.CanConvert() != tt.expected {
				t.Errorf("LeadStatus.CanConvert() = %v, want %v", tt.status.CanConvert(), tt.expected)
			}
		})
	}
}

func TestValidLeadStatuses(t *testing.T) {
	statuses := ValidLeadStatuses()
	if len(statuses) != 6 {
		t.Errorf("ValidLeadStatuses() len = %v, want 6", len(statuses))
	}
}

func TestLeadSource_IsValid(t *testing.T) {
	tests := []struct {
		source   LeadSource
		expected bool
	}{
		{LeadSourceWebsite, true},
		{LeadSourceReferral, true},
		{LeadSourceSocialMedia, true},
		{LeadSourceAdvertising, true},
		{LeadSourceTradeShow, true},
		{LeadSourceColdCall, true},
		{LeadSourceEmail, true},
		{LeadSourcePartner, true},
		{LeadSourceOther, true},
		{LeadSource("invalid"), false},
		{LeadSource(""), false},
	}

	for _, tt := range tests {
		t.Run(string(tt.source), func(t *testing.T) {
			if tt.source.IsValid() != tt.expected {
				t.Errorf("LeadSource.IsValid() = %v, want %v", tt.source.IsValid(), tt.expected)
			}
		})
	}
}

func TestValidLeadSources(t *testing.T) {
	sources := ValidLeadSources()
	if len(sources) != 9 {
		t.Errorf("ValidLeadSources() len = %v, want 9", len(sources))
	}
}

func TestLeadRating_IsValid(t *testing.T) {
	tests := []struct {
		rating   LeadRating
		expected bool
	}{
		{LeadRatingHot, true},
		{LeadRatingWarm, true},
		{LeadRatingCold, true},
		{LeadRating("invalid"), false},
		{LeadRating(""), false},
	}

	for _, tt := range tests {
		t.Run(string(tt.rating), func(t *testing.T) {
			if tt.rating.IsValid() != tt.expected {
				t.Errorf("LeadRating.IsValid() = %v, want %v", tt.rating.IsValid(), tt.expected)
			}
		})
	}
}

func TestValidLeadRatings(t *testing.T) {
	ratings := ValidLeadRatings()
	if len(ratings) != 3 {
		t.Errorf("ValidLeadRatings() len = %v, want 3", len(ratings))
	}
}

func TestLeadContact_FullName(t *testing.T) {
	tests := []struct {
		name      string
		contact   LeadContact
		expected  string
	}{
		{
			name: "full name with both parts",
			contact: LeadContact{
				FirstName: "John",
				LastName:  "Doe",
			},
			expected: "John Doe",
		},
		{
			name: "first name only",
			contact: LeadContact{
				FirstName: "John",
				LastName:  "",
			},
			expected: "John",
		},
		{
			name: "empty names",
			contact: LeadContact{
				FirstName: "",
				LastName:  "",
			},
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.contact.FullName() != tt.expected {
				t.Errorf("LeadContact.FullName() = %v, want %v", tt.contact.FullName(), tt.expected)
			}
		})
	}
}

func TestNewLead(t *testing.T) {
	tenantID := uuid.New()
	createdBy := uuid.New()

	contact := LeadContact{
		FirstName: "John",
		LastName:  "Doe",
		Email:     "john.doe@example.com",
		Phone:     "+1234567890",
	}

	company := LeadCompany{
		Name:     "Acme Corp",
		Website:  "https://acme.com",
		Industry: "Technology",
	}

	tests := []struct {
		name      string
		tenantID  uuid.UUID
		contact   LeadContact
		company   LeadCompany
		source    LeadSource
		createdBy uuid.UUID
		wantErr   bool
		errMsg    string
	}{
		{
			name:      "valid lead",
			tenantID:  tenantID,
			contact:   contact,
			company:   company,
			source:    LeadSourceWebsite,
			createdBy: createdBy,
			wantErr:   false,
		},
		{
			name:     "missing contact first name",
			tenantID: tenantID,
			contact: LeadContact{
				FirstName: "",
				Email:     "john@example.com",
			},
			company:   company,
			source:    LeadSourceWebsite,
			createdBy: createdBy,
			wantErr:   true,
			errMsg:    "contact first name is required",
		},
		{
			name:     "missing contact email",
			tenantID: tenantID,
			contact: LeadContact{
				FirstName: "John",
				Email:     "",
			},
			company:   company,
			source:    LeadSourceWebsite,
			createdBy: createdBy,
			wantErr:   true,
			errMsg:    "contact email is required",
		},
		{
			name:     "missing company name",
			tenantID: tenantID,
			contact:  contact,
			company: LeadCompany{
				Name: "",
			},
			source:    LeadSourceWebsite,
			createdBy: createdBy,
			wantErr:   true,
			errMsg:    "company name is required",
		},
		{
			name:      "invalid source defaults to other",
			tenantID:  tenantID,
			contact:   contact,
			company:   company,
			source:    LeadSource("invalid"),
			createdBy: createdBy,
			wantErr:   false, // Should default to Other
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lead, err := NewLead(tt.tenantID, tt.contact, tt.company, tt.source, tt.createdBy)
			if tt.wantErr {
				if err == nil {
					t.Errorf("NewLead() expected error, got nil")
				}
				if err != nil && err.Error() != tt.errMsg {
					t.Errorf("NewLead() error = %v, want %v", err.Error(), tt.errMsg)
				}
			} else {
				if err != nil {
					t.Errorf("NewLead() unexpected error = %v", err)
				}
				if lead == nil {
					t.Fatal("NewLead() returned nil lead")
				}
				if lead.ID == uuid.Nil {
					t.Error("NewLead() should generate ID")
				}
				if lead.TenantID != tt.tenantID {
					t.Errorf("NewLead() TenantID = %v, want %v", lead.TenantID, tt.tenantID)
				}
				if lead.Status != LeadStatusNew {
					t.Errorf("NewLead() Status = %v, want %v", lead.Status, LeadStatusNew)
				}
				if lead.Rating != LeadRatingCold {
					t.Errorf("NewLead() Rating = %v, want %v", lead.Rating, LeadRatingCold)
				}
				if lead.Version != 1 {
					t.Errorf("NewLead() Version = %v, want 1", lead.Version)
				}
				// Check domain event was added
				events := lead.GetEvents()
				if len(events) != 1 {
					t.Errorf("NewLead() should add LeadCreatedEvent, got %d events", len(events))
				}
			}
		})
	}
}

func createTestLead(t *testing.T) *Lead {
	tenantID := uuid.New()
	createdBy := uuid.New()

	contact := LeadContact{
		FirstName: "John",
		LastName:  "Doe",
		Email:     "john.doe@example.com",
	}

	company := LeadCompany{
		Name: "Acme Corp",
	}

	lead, err := NewLead(tenantID, contact, company, LeadSourceWebsite, createdBy)
	if err != nil {
		t.Fatalf("createTestLead() failed: %v", err)
	}
	lead.ClearEvents()
	return lead
}

func TestLead_Update(t *testing.T) {
	lead := createTestLead(t)

	newContact := LeadContact{
		FirstName: "Jane",
		LastName:  "Smith",
		Email:     "jane@example.com",
	}
	newCompany := LeadCompany{
		Name: "New Corp",
	}

	lead.Update(newContact, newCompany, LeadSourceReferral, LeadRatingWarm)

	if lead.Contact.FirstName != "Jane" {
		t.Errorf("Lead.Update() Contact.FirstName = %v, want Jane", lead.Contact.FirstName)
	}
	if lead.Company.Name != "New Corp" {
		t.Errorf("Lead.Update() Company.Name = %v", lead.Company.Name)
	}
	if lead.Source != LeadSourceReferral {
		t.Errorf("Lead.Update() Source = %v, want %v", lead.Source, LeadSourceReferral)
	}
	if lead.Rating != LeadRatingWarm {
		t.Errorf("Lead.Update() Rating = %v, want %v", lead.Rating, LeadRatingWarm)
	}

	events := lead.GetEvents()
	if len(events) != 1 {
		t.Errorf("Lead.Update() should add event, got %d events", len(events))
	}
}

func TestLead_SetEstimatedValue(t *testing.T) {
	lead := createTestLead(t)
	value, _ := NewMoneyFromFloat(10000, "USD")

	lead.SetEstimatedValue(value)

	if lead.EstimatedValue.Amount != value.Amount {
		t.Errorf("Lead.SetEstimatedValue() Amount = %v, want %v", lead.EstimatedValue.Amount, value.Amount)
	}
}

func TestLead_AssignOwner(t *testing.T) {
	lead := createTestLead(t)
	ownerID := uuid.New()

	lead.AssignOwner(ownerID, "John Owner")

	if lead.OwnerID == nil || *lead.OwnerID != ownerID {
		t.Errorf("Lead.AssignOwner() OwnerID = %v, want %v", lead.OwnerID, ownerID)
	}
	if lead.OwnerName != "John Owner" {
		t.Errorf("Lead.AssignOwner() OwnerName = %v", lead.OwnerName)
	}

	events := lead.GetEvents()
	if len(events) != 1 {
		t.Errorf("Lead.AssignOwner() should add event, got %d events", len(events))
	}
}

func TestLead_UnassignOwner(t *testing.T) {
	lead := createTestLead(t)
	ownerID := uuid.New()
	lead.AssignOwner(ownerID, "John Owner")
	lead.ClearEvents()

	lead.UnassignOwner()

	if lead.OwnerID != nil {
		t.Error("Lead.UnassignOwner() OwnerID should be nil")
	}
	if lead.OwnerName != "" {
		t.Error("Lead.UnassignOwner() OwnerName should be empty")
	}
}

func TestLead_MarkContacted(t *testing.T) {
	lead := createTestLead(t)

	err := lead.MarkContacted()
	if err != nil {
		t.Fatalf("Lead.MarkContacted() unexpected error = %v", err)
	}

	if lead.Status != LeadStatusContacted {
		t.Errorf("Lead.MarkContacted() Status = %v, want %v", lead.Status, LeadStatusContacted)
	}
	if lead.LastContactedAt == nil {
		t.Error("Lead.MarkContacted() should set LastContactedAt")
	}

	events := lead.GetEvents()
	if len(events) != 1 {
		t.Errorf("Lead.MarkContacted() should add event, got %d events", len(events))
	}
}

func TestLead_MarkContacted_AlreadyConverted(t *testing.T) {
	lead := createTestLead(t)
	lead.Status = LeadStatusConverted

	err := lead.MarkContacted()
	if err != ErrLeadAlreadyConverted {
		t.Errorf("Lead.MarkContacted() on converted lead should return ErrLeadAlreadyConverted, got %v", err)
	}
}

func TestLead_Qualify(t *testing.T) {
	lead := createTestLead(t)
	lead.Status = LeadStatusContacted

	err := lead.Qualify()
	if err != nil {
		t.Fatalf("Lead.Qualify() unexpected error = %v", err)
	}

	if lead.Status != LeadStatusQualified {
		t.Errorf("Lead.Qualify() Status = %v, want %v", lead.Status, LeadStatusQualified)
	}

	events := lead.GetEvents()
	if len(events) != 1 {
		t.Errorf("Lead.Qualify() should add event, got %d events", len(events))
	}
}

func TestLead_Qualify_AlreadyConverted(t *testing.T) {
	lead := createTestLead(t)
	lead.Status = LeadStatusConverted

	err := lead.Qualify()
	if err != ErrLeadAlreadyConverted {
		t.Errorf("Lead.Qualify() on converted lead should return ErrLeadAlreadyConverted, got %v", err)
	}
}

func TestLead_Qualify_AlreadyQualified(t *testing.T) {
	lead := createTestLead(t)
	lead.Status = LeadStatusQualified

	err := lead.Qualify()
	if err != ErrLeadAlreadyQualified {
		t.Errorf("Lead.Qualify() on qualified lead should return ErrLeadAlreadyQualified, got %v", err)
	}
}

func TestLead_Disqualify(t *testing.T) {
	lead := createTestLead(t)
	disqualifiedBy := uuid.New()

	err := lead.Disqualify("Not a good fit", "Additional notes", disqualifiedBy)
	if err != nil {
		t.Fatalf("Lead.Disqualify() unexpected error = %v", err)
	}

	if lead.Status != LeadStatusUnqualified {
		t.Errorf("Lead.Disqualify() Status = %v, want %v", lead.Status, LeadStatusUnqualified)
	}
	if lead.DisqualifyInfo == nil {
		t.Fatal("Lead.Disqualify() should set DisqualifyInfo")
	}
	if lead.DisqualifyInfo.Reason != "Not a good fit" {
		t.Errorf("Lead.Disqualify() Reason = %v", lead.DisqualifyInfo.Reason)
	}
	if lead.DisqualifyInfo.DisqualifiedBy != disqualifiedBy {
		t.Errorf("Lead.Disqualify() DisqualifiedBy = %v", lead.DisqualifyInfo.DisqualifiedBy)
	}

	events := lead.GetEvents()
	if len(events) != 1 {
		t.Errorf("Lead.Disqualify() should add event, got %d events", len(events))
	}
}

func TestLead_Disqualify_AlreadyConverted(t *testing.T) {
	lead := createTestLead(t)
	lead.Status = LeadStatusConverted

	err := lead.Disqualify("Reason", "Notes", uuid.New())
	if err != ErrLeadAlreadyConverted {
		t.Errorf("Lead.Disqualify() on converted lead should return ErrLeadAlreadyConverted, got %v", err)
	}
}

func TestLead_Disqualify_AlreadyDisqualified(t *testing.T) {
	lead := createTestLead(t)
	lead.Status = LeadStatusUnqualified

	err := lead.Disqualify("Reason", "Notes", uuid.New())
	if err != ErrLeadAlreadyDisqualified {
		t.Errorf("Lead.Disqualify() on disqualified lead should return ErrLeadAlreadyDisqualified, got %v", err)
	}
}

func TestLead_StartNurturing(t *testing.T) {
	lead := createTestLead(t)

	err := lead.StartNurturing()
	if err != nil {
		t.Fatalf("Lead.StartNurturing() unexpected error = %v", err)
	}

	if lead.Status != LeadStatusNurturing {
		t.Errorf("Lead.StartNurturing() Status = %v, want %v", lead.Status, LeadStatusNurturing)
	}
}

func TestLead_StartNurturing_AlreadyConverted(t *testing.T) {
	lead := createTestLead(t)
	lead.Status = LeadStatusConverted

	err := lead.StartNurturing()
	if err != ErrLeadAlreadyConverted {
		t.Errorf("Lead.StartNurturing() on converted lead should return ErrLeadAlreadyConverted, got %v", err)
	}
}

func TestLead_ConvertToOpportunity(t *testing.T) {
	lead := createTestLead(t)
	lead.Status = LeadStatusQualified
	lead.ClearEvents()

	opportunityID := uuid.New()
	convertedBy := uuid.New()
	customerID := uuid.New()
	contactID := uuid.New()

	err := lead.ConvertToOpportunity(opportunityID, convertedBy, &customerID, &contactID)
	if err != nil {
		t.Fatalf("Lead.ConvertToOpportunity() unexpected error = %v", err)
	}

	if lead.Status != LeadStatusConverted {
		t.Errorf("Lead.ConvertToOpportunity() Status = %v, want %v", lead.Status, LeadStatusConverted)
	}
	if lead.ConversionInfo == nil {
		t.Fatal("Lead.ConvertToOpportunity() should set ConversionInfo")
	}
	if lead.ConversionInfo.OpportunityID != opportunityID {
		t.Errorf("Lead.ConvertToOpportunity() OpportunityID = %v", lead.ConversionInfo.OpportunityID)
	}
	if lead.ConversionInfo.ConvertedBy != convertedBy {
		t.Errorf("Lead.ConvertToOpportunity() ConvertedBy = %v", lead.ConversionInfo.ConvertedBy)
	}
	if *lead.ConversionInfo.CustomerID != customerID {
		t.Errorf("Lead.ConvertToOpportunity() CustomerID = %v", lead.ConversionInfo.CustomerID)
	}

	events := lead.GetEvents()
	if len(events) != 1 {
		t.Errorf("Lead.ConvertToOpportunity() should add event, got %d events", len(events))
	}
}

func TestLead_ConvertToOpportunity_NotQualified(t *testing.T) {
	lead := createTestLead(t)
	// Status is New, not Qualified

	err := lead.ConvertToOpportunity(uuid.New(), uuid.New(), nil, nil)
	if err != ErrLeadNotQualified {
		t.Errorf("Lead.ConvertToOpportunity() on non-qualified lead should return ErrLeadNotQualified, got %v", err)
	}
}

func TestLead_ConvertToOpportunity_AlreadyConverted(t *testing.T) {
	lead := createTestLead(t)
	lead.Status = LeadStatusConverted

	err := lead.ConvertToOpportunity(uuid.New(), uuid.New(), nil, nil)
	if err != ErrLeadAlreadyConverted {
		t.Errorf("Lead.ConvertToOpportunity() on converted lead should return ErrLeadAlreadyConverted, got %v", err)
	}
}

func TestLead_RevertConversion(t *testing.T) {
	lead := createTestLead(t)
	lead.Status = LeadStatusQualified
	lead.ConvertToOpportunity(uuid.New(), uuid.New(), nil, nil)
	lead.ClearEvents()

	err := lead.RevertConversion()
	if err != nil {
		t.Fatalf("Lead.RevertConversion() unexpected error = %v", err)
	}

	if lead.Status != LeadStatusQualified {
		t.Errorf("Lead.RevertConversion() Status = %v, want %v", lead.Status, LeadStatusQualified)
	}
	if lead.ConversionInfo != nil {
		t.Error("Lead.RevertConversion() should clear ConversionInfo")
	}
}

func TestLead_RevertConversion_NotConverted(t *testing.T) {
	lead := createTestLead(t)

	err := lead.RevertConversion()
	if err != ErrLeadNotConverted {
		t.Errorf("Lead.RevertConversion() on non-converted lead should return ErrLeadNotConverted, got %v", err)
	}
}

func TestLead_UpdateScore(t *testing.T) {
	lead := createTestLead(t)

	lead.UpdateScore(50, 30, "Initial scoring")

	if lead.Score.Score != 80 {
		t.Errorf("Lead.UpdateScore() Score = %v, want 80", lead.Score.Score)
	}
	if lead.Score.Demographic != 50 {
		t.Errorf("Lead.UpdateScore() Demographic = %v, want 50", lead.Score.Demographic)
	}
	if lead.Score.Behavioral != 30 {
		t.Errorf("Lead.UpdateScore() Behavioral = %v, want 30", lead.Score.Behavioral)
	}
	if lead.Rating != LeadRatingHot {
		t.Errorf("Lead.UpdateScore() Rating = %v, want %v (score >= 70)", lead.Rating, LeadRatingHot)
	}
	if len(lead.Score.ScoreHistory) != 1 {
		t.Errorf("Lead.UpdateScore() should add to history, got %d entries", len(lead.Score.ScoreHistory))
	}

	events := lead.GetEvents()
	if len(events) != 1 {
		t.Errorf("Lead.UpdateScore() should add event, got %d events", len(events))
	}
}

func TestLead_UpdateScore_RatingThresholds(t *testing.T) {
	tests := []struct {
		name           string
		demographic    int
		behavioral     int
		expectedScore  int
		expectedRating LeadRating
	}{
		{"hot", 50, 30, 80, LeadRatingHot},
		{"warm", 30, 20, 50, LeadRatingWarm},
		{"cold", 10, 10, 20, LeadRatingCold},
		{"max cap", 80, 80, 100, LeadRatingHot},
		{"min cap", -10, -10, 0, LeadRatingCold},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lead := createTestLead(t)
			lead.UpdateScore(tt.demographic, tt.behavioral, "test")

			if lead.Score.Score != tt.expectedScore {
				t.Errorf("Lead.UpdateScore() Score = %v, want %v", lead.Score.Score, tt.expectedScore)
			}
			if lead.Rating != tt.expectedRating {
				t.Errorf("Lead.UpdateScore() Rating = %v, want %v", lead.Rating, tt.expectedRating)
			}
		})
	}
}

func TestLead_UpdateScore_HistoryLimit(t *testing.T) {
	lead := createTestLead(t)

	// Add 15 score entries
	for i := 0; i < 15; i++ {
		lead.UpdateScore(i, i, "scoring")
	}

	// History should be capped at 10
	if len(lead.Score.ScoreHistory) != 10 {
		t.Errorf("Lead.UpdateScore() history len = %v, want 10 (max)", len(lead.Score.ScoreHistory))
	}
}

func TestLead_RecordEngagement(t *testing.T) {
	lead := createTestLead(t)

	tests := []struct {
		engagementType string
		field          string
	}{
		{"email_opened", "EmailsOpened"},
		{"email_clicked", "EmailsClicked"},
		{"web_visit", "WebVisits"},
		{"form_submission", "FormSubmissions"},
	}

	for _, tt := range tests {
		t.Run(tt.engagementType, func(t *testing.T) {
			lead.RecordEngagement(tt.engagementType)
		})
	}

	if lead.Engagement.EmailsOpened != 1 {
		t.Errorf("Lead.RecordEngagement() EmailsOpened = %v, want 1", lead.Engagement.EmailsOpened)
	}
	if lead.Engagement.EmailsClicked != 1 {
		t.Errorf("Lead.RecordEngagement() EmailsClicked = %v, want 1", lead.Engagement.EmailsClicked)
	}
	if lead.Engagement.WebVisits != 1 {
		t.Errorf("Lead.RecordEngagement() WebVisits = %v, want 1", lead.Engagement.WebVisits)
	}
	if lead.Engagement.FormSubmissions != 1 {
		t.Errorf("Lead.RecordEngagement() FormSubmissions = %v, want 1", lead.Engagement.FormSubmissions)
	}
	if lead.Engagement.LastEngagement == nil {
		t.Error("Lead.RecordEngagement() should set LastEngagement")
	}
}

func TestLead_SetNextFollowUp(t *testing.T) {
	lead := createTestLead(t)
	followUp := time.Now().Add(24 * time.Hour)

	lead.SetNextFollowUp(followUp)

	if lead.NextFollowUp == nil {
		t.Fatal("Lead.SetNextFollowUp() should set NextFollowUp")
	}
	if !lead.NextFollowUp.Equal(followUp) {
		t.Errorf("Lead.SetNextFollowUp() NextFollowUp = %v, want %v", lead.NextFollowUp, followUp)
	}
}

func TestLead_ClearNextFollowUp(t *testing.T) {
	lead := createTestLead(t)
	lead.SetNextFollowUp(time.Now().Add(24 * time.Hour))

	lead.ClearNextFollowUp()

	if lead.NextFollowUp != nil {
		t.Error("Lead.ClearNextFollowUp() should clear NextFollowUp")
	}
}

func TestLead_Tags(t *testing.T) {
	lead := createTestLead(t)

	// Add tags
	lead.AddTag("important")
	lead.AddTag("follow-up")
	lead.AddTag("important") // Duplicate should be ignored

	if len(lead.Tags) != 2 {
		t.Errorf("Lead.AddTag() len = %v, want 2", len(lead.Tags))
	}

	// Remove tag
	lead.RemoveTag("important")

	if len(lead.Tags) != 1 {
		t.Errorf("Lead.RemoveTag() len = %v, want 1", len(lead.Tags))
	}
	if lead.Tags[0] != "follow-up" {
		t.Errorf("Lead.RemoveTag() remaining tag = %v", lead.Tags[0])
	}

	// Remove non-existent tag should not panic
	lead.RemoveTag("nonexistent")
}

func TestLead_SetCustomField(t *testing.T) {
	lead := createTestLead(t)

	lead.SetCustomField("priority", "high")
	lead.SetCustomField("budget", 50000)

	if lead.CustomFields["priority"] != "high" {
		t.Errorf("Lead.SetCustomField() priority = %v", lead.CustomFields["priority"])
	}
	if lead.CustomFields["budget"] != 50000 {
		t.Errorf("Lead.SetCustomField() budget = %v", lead.CustomFields["budget"])
	}
}

func TestLead_Delete(t *testing.T) {
	lead := createTestLead(t)

	err := lead.Delete()
	if err != nil {
		t.Fatalf("Lead.Delete() unexpected error = %v", err)
	}

	if !lead.IsDeleted() {
		t.Error("Lead.Delete() should mark lead as deleted")
	}
	if lead.DeletedAt == nil {
		t.Error("Lead.Delete() should set DeletedAt")
	}

	events := lead.GetEvents()
	if len(events) != 1 {
		t.Errorf("Lead.Delete() should add event, got %d events", len(events))
	}
}

func TestLead_Delete_AlreadyConverted(t *testing.T) {
	lead := createTestLead(t)
	lead.Status = LeadStatusConverted

	err := lead.Delete()
	if err != ErrLeadAlreadyConverted {
		t.Errorf("Lead.Delete() on converted lead should return ErrLeadAlreadyConverted, got %v", err)
	}
}

func TestLead_Restore(t *testing.T) {
	lead := createTestLead(t)
	lead.Delete()
	lead.ClearEvents()

	lead.Restore()

	if lead.IsDeleted() {
		t.Error("Lead.Restore() should restore lead")
	}
	if lead.DeletedAt != nil {
		t.Error("Lead.Restore() should clear DeletedAt")
	}
}

func TestLead_StatusChecks(t *testing.T) {
	lead := createTestLead(t)

	// IsConverted
	if lead.IsConverted() {
		t.Error("Lead.IsConverted() should return false for new lead")
	}
	lead.Status = LeadStatusConverted
	if !lead.IsConverted() {
		t.Error("Lead.IsConverted() should return true for converted lead")
	}

	// IsQualified
	lead.Status = LeadStatusNew
	if lead.IsQualified() {
		t.Error("Lead.IsQualified() should return false for new lead")
	}
	lead.Status = LeadStatusQualified
	if !lead.IsQualified() {
		t.Error("Lead.IsQualified() should return true for qualified lead")
	}

	// CanConvert
	lead.Status = LeadStatusNew
	if lead.CanConvert() {
		t.Error("Lead.CanConvert() should return false for new lead")
	}
	lead.Status = LeadStatusQualified
	if !lead.CanConvert() {
		t.Error("Lead.CanConvert() should return true for qualified lead")
	}
	lead.Delete()
	if lead.CanConvert() {
		t.Error("Lead.CanConvert() should return false for deleted lead")
	}
}

func TestLead_DaysCalculations(t *testing.T) {
	lead := createTestLead(t)
	// Set CreatedAt to 5 days ago
	lead.CreatedAt = time.Now().Add(-5 * 24 * time.Hour)
	lead.UpdatedAt = time.Now().Add(-2 * 24 * time.Hour)

	days := lead.DaysInCurrentStatus()
	if days < 1 || days > 3 {
		t.Errorf("Lead.DaysInCurrentStatus() = %v, expected ~2", days)
	}

	// DaysSinceLastContact with no contact
	days = lead.DaysSinceLastContact()
	if days < 4 || days > 6 {
		t.Errorf("Lead.DaysSinceLastContact() = %v, expected ~5", days)
	}

	// DaysSinceLastContact with contact
	contactTime := time.Now().Add(-1 * 24 * time.Hour)
	lead.LastContactedAt = &contactTime
	days = lead.DaysSinceLastContact()
	if days < 0 || days > 2 {
		t.Errorf("Lead.DaysSinceLastContact() = %v, expected ~1", days)
	}
}

func TestLead_Events(t *testing.T) {
	lead := createTestLead(t)
	lead.ClearEvents()

	// GetEvents should return empty after clear
	if len(lead.GetEvents()) != 0 {
		t.Error("Lead.GetEvents() should return empty after ClearEvents")
	}

	// Add event manually
	lead.AddEvent(&LeadCreatedEvent{})
	if len(lead.GetEvents()) != 1 {
		t.Error("Lead.AddEvent() should add event")
	}

	// Clear events
	lead.ClearEvents()
	if len(lead.GetEvents()) != 0 {
		t.Error("Lead.ClearEvents() should clear all events")
	}
}
