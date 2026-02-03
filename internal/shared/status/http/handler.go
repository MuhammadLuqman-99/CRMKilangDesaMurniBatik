// Package http provides HTTP handlers for the status page.
package http

import (
	"encoding/json"
	"html/template"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/kilang-desa-murni/crm/internal/shared/status/domain"
	"github.com/kilang-desa-murni/crm/internal/shared/status/service"
)

// StatusHandler handles status page HTTP requests.
type StatusHandler struct {
	statusService *service.StatusService
	templates     *template.Template
}

// NewStatusHandler creates a new StatusHandler.
func NewStatusHandler(statusService *service.StatusService) *StatusHandler {
	tmpl := template.Must(template.New("status").Parse(statusPageTemplate))
	return &StatusHandler{
		statusService: statusService,
		templates:     tmpl,
	}
}

// RegisterRoutes registers all status page routes.
func (h *StatusHandler) RegisterRoutes(r chi.Router) {
	// Public status page
	r.Get("/status", h.HandleStatusPage)
	r.Get("/status.json", h.HandleStatusJSON)

	// API routes
	r.Route("/api/v1/status", func(r chi.Router) {
		// Public endpoints
		r.Get("/", h.HandleGetStatus)
		r.Get("/summary", h.HandleGetSummary)
		r.Get("/services", h.HandleListServices)
		r.Get("/services/{id}", h.HandleGetService)
		r.Get("/services/{id}/metrics", h.HandleGetServiceMetrics)
		r.Get("/incidents", h.HandleListIncidents)
		r.Get("/incidents/{id}", h.HandleGetIncident)
		r.Get("/maintenances", h.HandleListMaintenances)
		r.Get("/history", h.HandleGetHistory)

		// Subscriber endpoints
		r.Post("/subscribe", h.HandleSubscribe)
		r.Get("/verify", h.HandleVerifySubscription)
		r.Delete("/unsubscribe/{id}", h.HandleUnsubscribe)

		// Admin endpoints (should be protected by auth middleware)
		r.Route("/admin", func(r chi.Router) {
			r.Post("/incidents", h.HandleCreateIncident)
			r.Put("/incidents/{id}", h.HandleUpdateIncident)
			r.Post("/maintenances", h.HandleScheduleMaintenance)
		})
	})
}

// HandleStatusPage serves the HTML status page.
func (h *StatusHandler) HandleStatusPage(w http.ResponseWriter, r *http.Request) {
	summary := h.statusService.GetStatusSummary(r.Context())

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("Cache-Control", "public, max-age=60")

	if err := h.templates.ExecuteTemplate(w, "status", summary); err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}

// HandleStatusJSON serves the status as JSON for external monitoring.
func (h *StatusHandler) HandleStatusJSON(w http.ResponseWriter, r *http.Request) {
	summary := h.statusService.GetStatusSummary(r.Context())
	writeJSON(w, http.StatusOK, summary)
}

// HandleGetStatus returns the overall status.
func (h *StatusHandler) HandleGetStatus(w http.ResponseWriter, r *http.Request) {
	summary := h.statusService.GetStatusSummary(r.Context())
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"status":      summary.OverallStatus,
		"last_updated": summary.LastUpdated,
	})
}

// HandleGetSummary returns the full status summary.
func (h *StatusHandler) HandleGetSummary(w http.ResponseWriter, r *http.Request) {
	summary := h.statusService.GetStatusSummary(r.Context())
	writeJSON(w, http.StatusOK, summary)
}

// HandleListServices returns all monitored services.
func (h *StatusHandler) HandleListServices(w http.ResponseWriter, r *http.Request) {
	summary := h.statusService.GetStatusSummary(r.Context())

	// Group by service group
	grouped := make(map[string][]domain.Service)
	for _, svc := range summary.Services {
		grouped[svc.Group] = append(grouped[svc.Group], svc)
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"services": summary.Services,
		"grouped":  grouped,
	})
}

// HandleGetService returns a specific service.
func (h *StatusHandler) HandleGetService(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "Invalid service ID")
		return
	}

	summary := h.statusService.GetStatusSummary(r.Context())
	for _, svc := range summary.Services {
		if svc.ID == id {
			writeJSON(w, http.StatusOK, svc)
			return
		}
	}

	writeError(w, http.StatusNotFound, "Service not found")
}

// HandleGetServiceMetrics returns metrics for a service.
func (h *StatusHandler) HandleGetServiceMetrics(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "Invalid service ID")
		return
	}

	metrics, err := h.statusService.GetServiceMetrics(id)
	if err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, metrics)
}

// HandleListIncidents returns all incidents.
func (h *StatusHandler) HandleListIncidents(w http.ResponseWriter, r *http.Request) {
	activeOnly := r.URL.Query().Get("active") == "true"
	limit := 50
	if l := r.URL.Query().Get("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil {
			limit = parsed
		}
	}

	incidents := h.statusService.GetAllIncidents(activeOnly, limit)
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"incidents": incidents,
		"total":     len(incidents),
	})
}

// HandleGetIncident returns a specific incident.
func (h *StatusHandler) HandleGetIncident(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "Invalid incident ID")
		return
	}

	incident, err := h.statusService.GetIncident(id)
	if err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, incident)
}

// HandleListMaintenances returns scheduled maintenances.
func (h *StatusHandler) HandleListMaintenances(w http.ResponseWriter, r *http.Request) {
	upcomingOnly := r.URL.Query().Get("upcoming") != "false"
	maintenances := h.statusService.GetScheduledMaintenances(upcomingOnly)
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"maintenances": maintenances,
	})
}

// HandleGetHistory returns historical status data.
func (h *StatusHandler) HandleGetHistory(w http.ResponseWriter, r *http.Request) {
	days := 90
	if d := r.URL.Query().Get("days"); d != "" {
		if parsed, err := strconv.Atoi(d); err == nil && parsed > 0 && parsed <= 365 {
			days = parsed
		}
	}

	incidents := h.statusService.GetAllIncidents(false, 0)

	// Filter to requested timeframe
	cutoff := time.Now().AddDate(0, 0, -days)
	filtered := make([]*domain.Incident, 0)
	for _, inc := range incidents {
		if inc.CreatedAt.After(cutoff) {
			filtered = append(filtered, inc)
		}
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"incidents": filtered,
		"days":      days,
	})
}

// HandleSubscribe handles subscription requests.
func (h *StatusHandler) HandleSubscribe(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Email       string `json:"email"`
		Phone       string `json:"phone"`
		Preferences struct {
			NotifyIncidents    bool `json:"notify_incidents"`
			NotifyMaintenance  bool `json:"notify_maintenance"`
			NotifyResolved     bool `json:"notify_resolved"`
			EmailNotifications bool `json:"email_notifications"`
			SMSNotifications   bool `json:"sms_notifications"`
		} `json:"preferences"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.Email == "" && req.Phone == "" {
		writeError(w, http.StatusBadRequest, "Email or phone is required")
		return
	}

	prefs := domain.SubscriberPreferences{
		NotifyIncidents:    req.Preferences.NotifyIncidents,
		NotifyMaintenance:  req.Preferences.NotifyMaintenance,
		NotifyResolved:     req.Preferences.NotifyResolved,
		EmailNotifications: req.Preferences.EmailNotifications,
		SMSNotifications:   req.Preferences.SMSNotifications,
	}

	// Default preferences
	if !prefs.NotifyIncidents && !prefs.NotifyMaintenance && !prefs.NotifyResolved {
		prefs.NotifyIncidents = true
		prefs.NotifyResolved = true
	}
	if !prefs.EmailNotifications && !prefs.SMSNotifications {
		prefs.EmailNotifications = req.Email != ""
		prefs.SMSNotifications = req.Phone != ""
	}

	subscriber, err := h.statusService.Subscribe(r.Context(), req.Email, req.Phone, prefs)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, map[string]interface{}{
		"id":      subscriber.ID,
		"message": "Please check your email/phone to verify your subscription",
	})
}

// HandleVerifySubscription verifies a subscription.
func (h *StatusHandler) HandleVerifySubscription(w http.ResponseWriter, r *http.Request) {
	token := r.URL.Query().Get("token")
	if token == "" {
		writeError(w, http.StatusBadRequest, "Token is required")
		return
	}

	if err := h.statusService.VerifySubscriber(token); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{
		"message": "Subscription verified successfully",
	})
}

// HandleUnsubscribe handles unsubscription requests.
func (h *StatusHandler) HandleUnsubscribe(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "Invalid subscriber ID")
		return
	}

	if err := h.statusService.Unsubscribe(id); err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// HandleCreateIncident creates a new incident (admin only).
func (h *StatusHandler) HandleCreateIncident(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Title            string   `json:"title"`
		Severity         string   `json:"severity"`
		AffectedServices []string `json:"affected_services"`
		Message          string   `json:"message"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.Title == "" || req.Message == "" {
		writeError(w, http.StatusBadRequest, "Title and message are required")
		return
	}

	severity := domain.IncidentSeverity(req.Severity)
	if severity != domain.SeverityCritical && severity != domain.SeverityMajor && severity != domain.SeverityMinor {
		severity = domain.SeverityMajor
	}

	serviceIDs := make([]uuid.UUID, 0, len(req.AffectedServices))
	for _, idStr := range req.AffectedServices {
		if id, err := uuid.Parse(idStr); err == nil {
			serviceIDs = append(serviceIDs, id)
		}
	}

	incident, err := h.statusService.CreateIncident(r.Context(), req.Title, severity, serviceIDs, req.Message)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, incident)
}

// HandleUpdateIncident updates an incident (admin only).
func (h *StatusHandler) HandleUpdateIncident(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "Invalid incident ID")
		return
	}

	var req struct {
		Status    string `json:"status"`
		Message   string `json:"message"`
		UpdatedBy string `json:"updated_by"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	status := domain.IncidentStatus(req.Status)
	if status != domain.IncidentInvestigating && status != domain.IncidentIdentified &&
		status != domain.IncidentMonitoring && status != domain.IncidentResolved {
		writeError(w, http.StatusBadRequest, "Invalid status")
		return
	}

	updatedBy := req.UpdatedBy
	if updatedBy == "" {
		updatedBy = "Admin"
	}

	incident, err := h.statusService.UpdateIncident(r.Context(), id, status, req.Message, updatedBy)
	if err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, incident)
}

// HandleScheduleMaintenance schedules a maintenance (admin only).
func (h *StatusHandler) HandleScheduleMaintenance(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Title            string    `json:"title"`
		Description      string    `json:"description"`
		AffectedServices []string  `json:"affected_services"`
		ScheduledStart   time.Time `json:"scheduled_start"`
		ScheduledEnd     time.Time `json:"scheduled_end"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.Title == "" {
		writeError(w, http.StatusBadRequest, "Title is required")
		return
	}

	serviceIDs := make([]uuid.UUID, 0, len(req.AffectedServices))
	for _, idStr := range req.AffectedServices {
		if id, err := uuid.Parse(idStr); err == nil {
			serviceIDs = append(serviceIDs, id)
		}
	}

	maintenance, err := h.statusService.ScheduleMaintenance(
		r.Context(), req.Title, req.Description, serviceIDs, req.ScheduledStart, req.ScheduledEnd,
	)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, maintenance)
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

// HTML template for status page
const statusPageTemplate = `{{define "status"}}<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <meta name="description" content="CRM Platform System Status - Real-time service availability">
    <title>System Status - CRM Platform</title>
    <style>
        :root {
            --operational: #22c55e;
            --degraded: #eab308;
            --partial: #f97316;
            --major: #ef4444;
            --maintenance: #3b82f6;
            --bg: #0f172a;
            --card: #1e293b;
            --text: #f1f5f9;
            --muted: #94a3b8;
            --border: #334155;
        }
        * { box-sizing: border-box; margin: 0; padding: 0; }
        body {
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
            background: var(--bg);
            color: var(--text);
            line-height: 1.6;
        }
        .container { max-width: 900px; margin: 0 auto; padding: 2rem; }
        header { text-align: center; margin-bottom: 3rem; }
        h1 { font-size: 2rem; margin-bottom: 0.5rem; }
        .overall-status {
            display: inline-flex;
            align-items: center;
            gap: 0.5rem;
            padding: 0.75rem 1.5rem;
            border-radius: 9999px;
            font-weight: 600;
            font-size: 1.1rem;
            margin: 1rem 0;
        }
        .overall-status.operational { background: rgba(34, 197, 94, 0.2); color: var(--operational); }
        .overall-status.degraded { background: rgba(234, 179, 8, 0.2); color: var(--degraded); }
        .overall-status.partial_outage { background: rgba(249, 115, 22, 0.2); color: var(--partial); }
        .overall-status.major_outage { background: rgba(239, 68, 68, 0.2); color: var(--major); }
        .status-dot {
            width: 12px;
            height: 12px;
            border-radius: 50%;
            display: inline-block;
        }
        .status-dot.operational { background: var(--operational); }
        .status-dot.degraded { background: var(--degraded); }
        .status-dot.partial_outage { background: var(--partial); }
        .status-dot.major_outage { background: var(--major); }
        .status-dot.maintenance { background: var(--maintenance); }
        .card {
            background: var(--card);
            border-radius: 12px;
            padding: 1.5rem;
            margin-bottom: 1.5rem;
            border: 1px solid var(--border);
        }
        .card h2 { font-size: 1.25rem; margin-bottom: 1rem; }
        .service-group { margin-bottom: 1.5rem; }
        .service-group h3 { font-size: 0.875rem; color: var(--muted); margin-bottom: 0.75rem; text-transform: uppercase; letter-spacing: 0.05em; }
        .service-item {
            display: flex;
            justify-content: space-between;
            align-items: center;
            padding: 0.75rem 0;
            border-bottom: 1px solid var(--border);
        }
        .service-item:last-child { border-bottom: none; }
        .service-name { font-weight: 500; }
        .service-status {
            display: flex;
            align-items: center;
            gap: 0.5rem;
            font-size: 0.875rem;
        }
        .uptime-stats {
            display: grid;
            grid-template-columns: repeat(4, 1fr);
            gap: 1rem;
            text-align: center;
        }
        .uptime-stat .value { font-size: 2rem; font-weight: 700; color: var(--operational); }
        .uptime-stat .label { font-size: 0.75rem; color: var(--muted); text-transform: uppercase; }
        .incident {
            padding: 1rem;
            border-radius: 8px;
            margin-bottom: 1rem;
            border-left: 4px solid;
        }
        .incident.critical { border-color: var(--major); background: rgba(239, 68, 68, 0.1); }
        .incident.major { border-color: var(--partial); background: rgba(249, 115, 22, 0.1); }
        .incident.minor { border-color: var(--degraded); background: rgba(234, 179, 8, 0.1); }
        .incident h4 { margin-bottom: 0.5rem; }
        .incident .meta { font-size: 0.875rem; color: var(--muted); }
        .incident .updates { margin-top: 1rem; padding-top: 1rem; border-top: 1px solid var(--border); }
        .update-item { margin-bottom: 0.75rem; padding-left: 1rem; border-left: 2px solid var(--border); }
        .update-item .time { font-size: 0.75rem; color: var(--muted); }
        .no-incidents { color: var(--muted); text-align: center; padding: 2rem; }
        .subscribe-form { display: flex; gap: 0.5rem; margin-top: 1rem; }
        .subscribe-form input {
            flex: 1;
            padding: 0.75rem 1rem;
            border: 1px solid var(--border);
            border-radius: 8px;
            background: var(--bg);
            color: var(--text);
        }
        .subscribe-form button {
            padding: 0.75rem 1.5rem;
            background: var(--operational);
            color: white;
            border: none;
            border-radius: 8px;
            cursor: pointer;
            font-weight: 600;
        }
        .footer { text-align: center; margin-top: 3rem; color: var(--muted); font-size: 0.875rem; }
        @media (max-width: 640px) {
            .container { padding: 1rem; }
            .uptime-stats { grid-template-columns: repeat(2, 1fr); }
        }
    </style>
</head>
<body>
    <div class="container">
        <header>
            <h1>CRM Platform Status</h1>
            <div class="overall-status {{.OverallStatus}}">
                <span class="status-dot {{.OverallStatus}}"></span>
                {{if eq .OverallStatus "operational"}}All Systems Operational{{end}}
                {{if eq .OverallStatus "degraded"}}Degraded Performance{{end}}
                {{if eq .OverallStatus "partial_outage"}}Partial System Outage{{end}}
                {{if eq .OverallStatus "major_outage"}}Major System Outage{{end}}
                {{if eq .OverallStatus "maintenance"}}Under Maintenance{{end}}
            </div>
            <p style="color: var(--muted); font-size: 0.875rem;">
                Last updated: {{.LastUpdated.Format "Jan 2, 2006 3:04 PM MST"}}
            </p>
        </header>

        <div class="card">
            <h2>Uptime</h2>
            <div class="uptime-stats">
                <div class="uptime-stat">
                    <div class="value">{{printf "%.2f" .UptimeLast24h}}%</div>
                    <div class="label">Last 24 Hours</div>
                </div>
                <div class="uptime-stat">
                    <div class="value">{{printf "%.2f" .UptimeLast7d}}%</div>
                    <div class="label">Last 7 Days</div>
                </div>
                <div class="uptime-stat">
                    <div class="value">{{printf "%.2f" .UptimeLast30d}}%</div>
                    <div class="label">Last 30 Days</div>
                </div>
                <div class="uptime-stat">
                    <div class="value">{{printf "%.2f" .UptimeLast90d}}%</div>
                    <div class="label">Last 90 Days</div>
                </div>
            </div>
        </div>

        <div class="card">
            <h2>Services</h2>
            {{range $group, $services := groupServices .Services}}
            <div class="service-group">
                <h3>{{$group}}</h3>
                {{range $services}}
                <div class="service-item">
                    <span class="service-name">{{.Name}}</span>
                    <span class="service-status">
                        <span class="status-dot {{.Status}}"></span>
                        {{if eq .Status "operational"}}Operational{{end}}
                        {{if eq .Status "degraded"}}Degraded{{end}}
                        {{if eq .Status "partial_outage"}}Partial Outage{{end}}
                        {{if eq .Status "major_outage"}}Major Outage{{end}}
                        {{if eq .Status "maintenance"}}Maintenance{{end}}
                    </span>
                </div>
                {{end}}
            </div>
            {{end}}
        </div>

        {{if .ActiveIncidents}}
        <div class="card">
            <h2>Active Incidents</h2>
            {{range .ActiveIncidents}}
            <div class="incident {{.Severity}}">
                <h4>{{.Title}}</h4>
                <div class="meta">
                    Started {{.StartedAt.Format "Jan 2, 2006 3:04 PM"}} •
                    Status: {{.Status}}
                </div>
                {{if .Updates}}
                <div class="updates">
                    {{range .Updates}}
                    <div class="update-item">
                        <div class="time">{{.CreatedAt.Format "Jan 2, 3:04 PM"}}</div>
                        <div>{{.Message}}</div>
                    </div>
                    {{end}}
                </div>
                {{end}}
            </div>
            {{end}}
        </div>
        {{else}}
        <div class="card">
            <h2>Incidents</h2>
            <div class="no-incidents">
                <p>No active incidents</p>
            </div>
        </div>
        {{end}}

        {{if .UpcomingMaintenance}}
        <div class="card">
            <h2>Scheduled Maintenance</h2>
            {{range .UpcomingMaintenance}}
            <div class="incident minor">
                <h4>{{.Title}}</h4>
                <div class="meta">
                    {{.ScheduledStart.Format "Jan 2, 2006 3:04 PM"}} -
                    {{.ScheduledEnd.Format "3:04 PM MST"}}
                </div>
                {{if .Description}}<p style="margin-top: 0.5rem;">{{.Description}}</p>{{end}}
            </div>
            {{end}}
        </div>
        {{end}}

        <div class="card">
            <h2>Subscribe to Updates</h2>
            <p style="color: var(--muted); margin-bottom: 1rem;">
                Get notified about incidents and scheduled maintenance.
            </p>
            <form class="subscribe-form" id="subscribeForm">
                <input type="email" name="email" placeholder="Enter your email" required>
                <button type="submit">Subscribe</button>
            </form>
        </div>

        <footer class="footer">
            <p>&copy; 2026 Kilang Desa Murni Batik. All rights reserved.</p>
            <p style="margin-top: 0.5rem;">
                <a href="/terms" style="color: var(--muted);">Terms</a> •
                <a href="/privacy" style="color: var(--muted);">Privacy</a> •
                <a href="/sla" style="color: var(--muted);">SLA</a>
            </p>
        </footer>
    </div>

    <script>
        document.getElementById('subscribeForm').addEventListener('submit', async (e) => {
            e.preventDefault();
            const email = e.target.email.value;
            try {
                const res = await fetch('/api/v1/status/subscribe', {
                    method: 'POST',
                    headers: { 'Content-Type': 'application/json' },
                    body: JSON.stringify({ email, preferences: { notify_incidents: true, notify_resolved: true, email_notifications: true } })
                });
                if (res.ok) {
                    alert('Please check your email to verify your subscription.');
                    e.target.reset();
                } else {
                    const data = await res.json();
                    alert(data.error || 'Subscription failed');
                }
            } catch (err) {
                alert('An error occurred. Please try again.');
            }
        });

        // Auto-refresh every 60 seconds
        setTimeout(() => location.reload(), 60000);
    </script>
</body>
</html>{{end}}`
