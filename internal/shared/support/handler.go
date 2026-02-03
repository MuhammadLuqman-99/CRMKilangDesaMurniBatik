// Package support provides HTTP handlers for the support system.
package support

import (
	"embed"
	"encoding/json"
	"io/fs"
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

//go:embed faq_data.json
var faqDataFS embed.FS

// SupportHandler handles support-related HTTP requests.
type SupportHandler struct {
	faqCategories []FAQCategory
	articles      []KBArticle
}

// FAQCategory represents a FAQ category with questions.
type FAQCategory struct {
	ID        string        `json:"id"`
	Name      string        `json:"name"`
	Order     int           `json:"order"`
	Questions []FAQQuestion `json:"questions"`
}

// FAQQuestion represents a single FAQ item.
type FAQQuestion struct {
	ID       string   `json:"id"`
	Question string   `json:"question"`
	Answer   string   `json:"answer"`
	Tags     []string `json:"tags"`
	Views    int      `json:"views"`
	Helpful  int      `json:"helpful"`
	Updated  string   `json:"updated"`
}

// KBArticle represents a knowledge base article.
type KBArticle struct {
	ID          string   `json:"id"`
	Title       string   `json:"title"`
	Slug        string   `json:"slug"`
	Category    string   `json:"category"`
	Content     string   `json:"content"`
	Summary     string   `json:"summary"`
	Tags        []string `json:"tags"`
	Views       int      `json:"views"`
	Helpful     int      `json:"helpful"`
	NotHelpful  int      `json:"not_helpful"`
	AuthorID    string   `json:"author_id"`
	Published   bool     `json:"published"`
	CreatedAt   string   `json:"created_at"`
	UpdatedAt   string   `json:"updated_at"`
}

// SupportTicket represents a support ticket.
type SupportTicket struct {
	ID          uuid.UUID `json:"id"`
	TenantID    uuid.UUID `json:"tenant_id"`
	UserID      uuid.UUID `json:"user_id"`
	Subject     string    `json:"subject"`
	Description string    `json:"description"`
	Priority    string    `json:"priority"` // low, medium, high, critical
	Status      string    `json:"status"`   // open, in_progress, waiting, resolved, closed
	Category    string    `json:"category"`
	Messages    []TicketMessage `json:"messages"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	ResolvedAt  *time.Time `json:"resolved_at,omitempty"`
}

// TicketMessage represents a message in a ticket thread.
type TicketMessage struct {
	ID        uuid.UUID `json:"id"`
	AuthorID  uuid.UUID `json:"author_id"`
	Content   string    `json:"content"`
	IsStaff   bool      `json:"is_staff"`
	CreatedAt time.Time `json:"created_at"`
}

// NewSupportHandler creates a new SupportHandler.
func NewSupportHandler() *SupportHandler {
	h := &SupportHandler{
		faqCategories: getDefaultFAQCategories(),
		articles:      getDefaultKBArticles(),
	}
	return h
}

// RegisterRoutes registers all support routes.
func (h *SupportHandler) RegisterRoutes(r chi.Router) {
	r.Route("/api/v1/support", func(r chi.Router) {
		// FAQ endpoints (public)
		r.Get("/faq", h.HandleListFAQ)
		r.Get("/faq/search", h.HandleSearchFAQ)
		r.Get("/faq/{id}", h.HandleGetFAQ)
		r.Post("/faq/{id}/helpful", h.HandleFAQHelpful)

		// Knowledge base endpoints (public)
		r.Get("/kb", h.HandleListArticles)
		r.Get("/kb/search", h.HandleSearchArticles)
		r.Get("/kb/categories", h.HandleListKBCategories)
		r.Get("/kb/{slug}", h.HandleGetArticle)
		r.Post("/kb/{slug}/feedback", h.HandleArticleFeedback)

		// Support ticket endpoints (authenticated)
		r.Post("/tickets", h.HandleCreateTicket)
		r.Get("/tickets", h.HandleListTickets)
		r.Get("/tickets/{id}", h.HandleGetTicket)
		r.Post("/tickets/{id}/messages", h.HandleAddMessage)
		r.Put("/tickets/{id}/status", h.HandleUpdateStatus)

		// Contact endpoint
		r.Post("/contact", h.HandleContactForm)
	})
}

// HandleListFAQ returns all FAQ categories and questions.
func (h *SupportHandler) HandleListFAQ(w http.ResponseWriter, r *http.Request) {
	category := r.URL.Query().Get("category")

	if category != "" {
		for _, cat := range h.faqCategories {
			if strings.EqualFold(cat.ID, category) {
				writeJSON(w, http.StatusOK, cat)
				return
			}
		}
		writeJSON(w, http.StatusOK, FAQCategory{})
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"categories": h.faqCategories,
	})
}

// HandleSearchFAQ searches FAQ questions.
func (h *SupportHandler) HandleSearchFAQ(w http.ResponseWriter, r *http.Request) {
	query := strings.ToLower(r.URL.Query().Get("q"))
	if query == "" {
		writeError(w, http.StatusBadRequest, "Query parameter 'q' is required")
		return
	}

	var results []FAQQuestion
	for _, cat := range h.faqCategories {
		for _, q := range cat.Questions {
			if strings.Contains(strings.ToLower(q.Question), query) ||
				strings.Contains(strings.ToLower(q.Answer), query) {
				results = append(results, q)
			}
		}
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"query":   query,
		"results": results,
		"total":   len(results),
	})
}

// HandleGetFAQ returns a specific FAQ question.
func (h *SupportHandler) HandleGetFAQ(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	for _, cat := range h.faqCategories {
		for _, q := range cat.Questions {
			if q.ID == id {
				writeJSON(w, http.StatusOK, q)
				return
			}
		}
	}

	writeError(w, http.StatusNotFound, "FAQ not found")
}

// HandleFAQHelpful marks a FAQ as helpful.
func (h *SupportHandler) HandleFAQHelpful(w http.ResponseWriter, r *http.Request) {
	// In production, this would update the database
	writeJSON(w, http.StatusOK, map[string]string{
		"message": "Thank you for your feedback!",
	})
}

// HandleListArticles returns knowledge base articles.
func (h *SupportHandler) HandleListArticles(w http.ResponseWriter, r *http.Request) {
	category := r.URL.Query().Get("category")

	var articles []KBArticle
	for _, a := range h.articles {
		if !a.Published {
			continue
		}
		if category != "" && !strings.EqualFold(a.Category, category) {
			continue
		}
		articles = append(articles, a)
	}

	// Sort by views (popular first)
	sort.Slice(articles, func(i, j int) bool {
		return articles[i].Views > articles[j].Views
	})

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"articles": articles,
		"total":    len(articles),
	})
}

// HandleSearchArticles searches knowledge base articles.
func (h *SupportHandler) HandleSearchArticles(w http.ResponseWriter, r *http.Request) {
	query := strings.ToLower(r.URL.Query().Get("q"))
	if query == "" {
		writeError(w, http.StatusBadRequest, "Query parameter 'q' is required")
		return
	}

	var results []KBArticle
	for _, a := range h.articles {
		if !a.Published {
			continue
		}
		if strings.Contains(strings.ToLower(a.Title), query) ||
			strings.Contains(strings.ToLower(a.Summary), query) ||
			strings.Contains(strings.ToLower(a.Content), query) {
			results = append(results, a)
		}
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"query":   query,
		"results": results,
		"total":   len(results),
	})
}

// HandleListKBCategories returns knowledge base categories.
func (h *SupportHandler) HandleListKBCategories(w http.ResponseWriter, r *http.Request) {
	categories := make(map[string]int)
	for _, a := range h.articles {
		if a.Published {
			categories[a.Category]++
		}
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"categories": categories,
	})
}

// HandleGetArticle returns a specific article by slug.
func (h *SupportHandler) HandleGetArticle(w http.ResponseWriter, r *http.Request) {
	slug := chi.URLParam(r, "slug")

	for _, a := range h.articles {
		if a.Slug == slug && a.Published {
			// Increment view count (in production, update database)
			writeJSON(w, http.StatusOK, a)
			return
		}
	}

	writeError(w, http.StatusNotFound, "Article not found")
}

// HandleArticleFeedback handles article feedback (helpful/not helpful).
func (h *SupportHandler) HandleArticleFeedback(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Helpful bool   `json:"helpful"`
		Comment string `json:"comment,omitempty"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{
		"message": "Thank you for your feedback!",
	})
}

// HandleCreateTicket creates a new support ticket.
func (h *SupportHandler) HandleCreateTicket(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Subject     string `json:"subject"`
		Description string `json:"description"`
		Priority    string `json:"priority"`
		Category    string `json:"category"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.Subject == "" || req.Description == "" {
		writeError(w, http.StatusBadRequest, "Subject and description are required")
		return
	}

	ticket := SupportTicket{
		ID:          uuid.New(),
		TenantID:    uuid.Nil, // Would be extracted from auth context
		UserID:      uuid.Nil, // Would be extracted from auth context
		Subject:     req.Subject,
		Description: req.Description,
		Priority:    defaultIfEmpty(req.Priority, "medium"),
		Status:      "open",
		Category:    defaultIfEmpty(req.Category, "general"),
		Messages: []TicketMessage{
			{
				ID:        uuid.New(),
				AuthorID:  uuid.Nil,
				Content:   req.Description,
				IsStaff:   false,
				CreatedAt: time.Now(),
			},
		},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	writeJSON(w, http.StatusCreated, ticket)
}

// HandleListTickets returns user's support tickets.
func (h *SupportHandler) HandleListTickets(w http.ResponseWriter, r *http.Request) {
	// In production, fetch from database filtered by user/tenant
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"tickets": []SupportTicket{},
		"total":   0,
	})
}

// HandleGetTicket returns a specific ticket.
func (h *SupportHandler) HandleGetTicket(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	_, err := uuid.Parse(idStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "Invalid ticket ID")
		return
	}

	writeError(w, http.StatusNotFound, "Ticket not found")
}

// HandleAddMessage adds a message to a ticket.
func (h *SupportHandler) HandleAddMessage(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Content string `json:"content"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.Content == "" {
		writeError(w, http.StatusBadRequest, "Content is required")
		return
	}

	message := TicketMessage{
		ID:        uuid.New(),
		AuthorID:  uuid.Nil,
		Content:   req.Content,
		IsStaff:   false,
		CreatedAt: time.Now(),
	}

	writeJSON(w, http.StatusCreated, message)
}

// HandleUpdateStatus updates ticket status.
func (h *SupportHandler) HandleUpdateStatus(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Status string `json:"status"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	validStatuses := []string{"open", "in_progress", "waiting", "resolved", "closed"}
	valid := false
	for _, s := range validStatuses {
		if s == req.Status {
			valid = true
			break
		}
	}

	if !valid {
		writeError(w, http.StatusBadRequest, "Invalid status")
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{
		"status": req.Status,
	})
}

// HandleContactForm handles contact form submissions.
func (h *SupportHandler) HandleContactForm(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Name    string `json:"name"`
		Email   string `json:"email"`
		Subject string `json:"subject"`
		Message string `json:"message"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.Email == "" || req.Message == "" {
		writeError(w, http.StatusBadRequest, "Email and message are required")
		return
	}

	// In production, send email and create ticket
	writeJSON(w, http.StatusOK, map[string]string{
		"message": "Thank you for contacting us. We'll respond within 24-48 hours.",
	})
}

// Helper functions

func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}

func defaultIfEmpty(s, def string) string {
	if s == "" {
		return def
	}
	return s
}

// Default data

func getDefaultFAQCategories() []FAQCategory {
	return []FAQCategory{
		{
			ID:    "getting-started",
			Name:  "Getting Started",
			Order: 1,
			Questions: []FAQQuestion{
				{ID: "gs-1", Question: "How do I create an account?", Answer: "Visit crmplatform.my and click 'Sign Up'. Enter your email, create a password, and verify your email.", Tags: []string{"account", "signup"}, Updated: "2026-02"},
				{ID: "gs-2", Question: "What browsers are supported?", Answer: "We support Chrome, Firefox, Edge, and Safari (latest 2 versions).", Tags: []string{"browser", "compatibility"}, Updated: "2026-02"},
				{ID: "gs-3", Question: "How do I invite team members?", Answer: "Go to Settings > Team Management > Invite User. Enter their email and select a role.", Tags: []string{"team", "invite"}, Updated: "2026-02"},
			},
		},
		{
			ID:    "account",
			Name:  "Account Management",
			Order: 2,
			Questions: []FAQQuestion{
				{ID: "acc-1", Question: "How do I reset my password?", Answer: "Click 'Forgot Password' on the login page and follow the instructions sent to your email.", Tags: []string{"password", "reset"}, Updated: "2026-02"},
				{ID: "acc-2", Question: "How do I enable two-factor authentication?", Answer: "Go to Settings > Security > Enable 2FA. Scan the QR code with your authenticator app.", Tags: []string{"2fa", "security"}, Updated: "2026-02"},
				{ID: "acc-3", Question: "Can I delete my account?", Answer: "Yes, go to Settings > Account > Delete Account. This action is permanent.", Tags: []string{"delete", "account"}, Updated: "2026-02"},
			},
		},
		{
			ID:    "billing",
			Name:  "Billing & Subscription",
			Order: 3,
			Questions: []FAQQuestion{
				{ID: "bill-1", Question: "What payment methods are accepted?", Answer: "We accept credit/debit cards, FPX, and e-wallets (GrabPay, Touch 'n Go).", Tags: []string{"payment", "billing"}, Updated: "2026-02"},
				{ID: "bill-2", Question: "How do I upgrade my plan?", Answer: "Go to Settings > Billing > Change Plan and select your desired plan.", Tags: []string{"upgrade", "plan"}, Updated: "2026-02"},
				{ID: "bill-3", Question: "Is there a refund policy?", Answer: "Yes, monthly plans offer pro-rata refunds within 14 days, annual within 30 days.", Tags: []string{"refund", "cancellation"}, Updated: "2026-02"},
			},
		},
		{
			ID:    "features",
			Name:  "Features & Usage",
			Order: 4,
			Questions: []FAQQuestion{
				{ID: "feat-1", Question: "What is lead scoring?", Answer: "Lead scoring assigns points to leads based on their likelihood to convert, helping prioritize follow-ups.", Tags: []string{"leads", "scoring"}, Updated: "2026-02"},
				{ID: "feat-2", Question: "How do I convert a lead to a customer?", Answer: "Open the lead, click 'Convert', fill in opportunity details, and confirm the conversion.", Tags: []string{"leads", "conversion"}, Updated: "2026-02"},
				{ID: "feat-3", Question: "Can I import data from another CRM?", Answer: "Yes, we support CSV imports and direct migrations from Salesforce, HubSpot, Zoho, and Pipedrive.", Tags: []string{"import", "migration"}, Updated: "2026-02"},
			},
		},
	}
}

func getDefaultKBArticles() []KBArticle {
	return []KBArticle{
		{
			ID:        "kb-001",
			Title:     "Getting Started with CRM Platform",
			Slug:      "getting-started",
			Category:  "Getting Started",
			Summary:   "Learn how to set up your account and start using CRM Platform effectively.",
			Content:   "Welcome to CRM Platform! This guide will help you get started...",
			Tags:      []string{"beginner", "setup", "onboarding"},
			Views:     1250,
			Helpful:   98,
			Published: true,
			CreatedAt: "2026-01-15",
			UpdatedAt: "2026-02-01",
		},
		{
			ID:        "kb-002",
			Title:     "Lead Management Best Practices",
			Slug:      "lead-management-best-practices",
			Category:  "Leads",
			Summary:   "Optimize your lead management process with these proven strategies.",
			Content:   "Effective lead management is crucial for sales success...",
			Tags:      []string{"leads", "best-practices", "sales"},
			Views:     890,
			Helpful:   85,
			Published: true,
			CreatedAt: "2026-01-20",
			UpdatedAt: "2026-02-01",
		},
		{
			ID:        "kb-003",
			Title:     "Setting Up Email Integration",
			Slug:      "email-integration-setup",
			Category:  "Integrations",
			Summary:   "Connect your email account to sync contacts and track communications.",
			Content:   "Email integration allows you to sync your inbox with CRM Platform...",
			Tags:      []string{"email", "integration", "gmail", "outlook"},
			Views:     720,
			Helpful:   92,
			Published: true,
			CreatedAt: "2026-01-22",
			UpdatedAt: "2026-02-01",
		},
		{
			ID:        "kb-004",
			Title:     "Understanding Sales Pipelines",
			Slug:      "sales-pipelines-explained",
			Category:  "Sales",
			Summary:   "Master the art of pipeline management to improve your close rates.",
			Content:   "A sales pipeline visualizes your sales process...",
			Tags:      []string{"pipeline", "deals", "sales"},
			Views:     650,
			Helpful:   88,
			Published: true,
			CreatedAt: "2026-01-25",
			UpdatedAt: "2026-02-01",
		},
		{
			ID:        "kb-005",
			Title:     "API Quick Start Guide",
			Slug:      "api-quick-start",
			Category:  "Developers",
			Summary:   "Get started with the CRM Platform API in minutes.",
			Content:   "Our REST API allows you to integrate CRM Platform with your applications...",
			Tags:      []string{"api", "developers", "integration"},
			Views:     420,
			Helpful:   95,
			Published: true,
			CreatedAt: "2026-01-28",
			UpdatedAt: "2026-02-01",
		},
	}
}
