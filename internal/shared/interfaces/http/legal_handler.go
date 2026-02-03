// Package http provides shared HTTP handlers for public-facing endpoints.
package http

import (
	"embed"
	"encoding/json"
	"html"
	"html/template"
	"io/fs"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
)

//go:embed templates/*.html
var templateFS embed.FS

// LegalHandler handles legal document routes.
type LegalHandler struct {
	templates      *template.Template
	termsContent   []byte
	privacyContent []byte
	slaContent     []byte
	lastModified   time.Time
}

// LegalDocument represents a legal document.
type LegalDocument struct {
	Title       string
	Content     template.HTML
	LastUpdated string
	Version     string
}

// LegalAPIResponse represents the API response for legal documents.
type LegalAPIResponse struct {
	DocumentType string `json:"document_type"`
	Version      string `json:"version"`
	LastUpdated  string `json:"last_updated"`
	Format       string `json:"format"`
	Content      string `json:"content"`
}

// NewLegalHandler creates a new LegalHandler.
func NewLegalHandler(docsFS fs.FS) (*LegalHandler, error) {
	// Parse templates
	tmpl, err := template.ParseFS(templateFS, "templates/*.html")
	if err != nil {
		// Create a basic template if not found
		tmpl = template.Must(template.New("legal").Parse(defaultLegalTemplate))
	}

	handler := &LegalHandler{
		templates:    tmpl,
		lastModified: time.Now(),
	}

	// Load legal documents
	if docsFS != nil {
		handler.termsContent, _ = fs.ReadFile(docsFS, "legal/terms-of-service.md")
		handler.privacyContent, _ = fs.ReadFile(docsFS, "legal/privacy-policy.md")
		handler.slaContent, _ = fs.ReadFile(docsFS, "legal/sla.md")
	}

	return handler, nil
}

// RegisterRoutes registers the legal routes.
func (h *LegalHandler) RegisterRoutes(r chi.Router) {
	r.Get("/terms", h.HandleTerms)
	r.Get("/terms-of-service", h.HandleTerms)
	r.Get("/privacy", h.HandlePrivacy)
	r.Get("/privacy-policy", h.HandlePrivacy)
	r.Get("/sla", h.HandleSLA)
	r.Get("/service-level-agreement", h.HandleSLA)

	// API endpoints for raw/structured data
	r.Get("/api/v1/legal/terms", h.HandleTermsAPI)
	r.Get("/api/v1/legal/privacy", h.HandlePrivacyAPI)
	r.Get("/api/v1/legal/sla", h.HandleSLAAPI)
}

// HandleTerms serves the Terms of Service page.
func (h *LegalHandler) HandleTerms(w http.ResponseWriter, r *http.Request) {
	h.serveLegalDocument(w, r, "Terms of Service", h.termsContent, defaultTermsContent)
}

// HandlePrivacy serves the Privacy Policy page.
func (h *LegalHandler) HandlePrivacy(w http.ResponseWriter, r *http.Request) {
	h.serveLegalDocument(w, r, "Privacy Policy", h.privacyContent, defaultPrivacyContent)
}

// HandleSLA serves the Service Level Agreement page.
func (h *LegalHandler) HandleSLA(w http.ResponseWriter, r *http.Request) {
	h.serveLegalDocument(w, r, "Service Level Agreement", h.slaContent, defaultSLAContent)
}

// HandleTermsAPI serves the Terms of Service as JSON/Markdown.
func (h *LegalHandler) HandleTermsAPI(w http.ResponseWriter, r *http.Request) {
	h.serveLegalAPI(w, r, "terms-of-service", h.termsContent, defaultTermsContent)
}

// HandlePrivacyAPI serves the Privacy Policy as JSON/Markdown.
func (h *LegalHandler) HandlePrivacyAPI(w http.ResponseWriter, r *http.Request) {
	h.serveLegalAPI(w, r, "privacy-policy", h.privacyContent, defaultPrivacyContent)
}

// HandleSLAAPI serves the SLA as JSON/Markdown.
func (h *LegalHandler) HandleSLAAPI(w http.ResponseWriter, r *http.Request) {
	h.serveLegalAPI(w, r, "sla", h.slaContent, defaultSLAContent)
}

func (h *LegalHandler) serveLegalDocument(w http.ResponseWriter, r *http.Request, title string, content []byte, defaultContent string) {
	// Use loaded content or default
	mdContent := content
	if len(mdContent) == 0 {
		mdContent = []byte(defaultContent)
	}

	// Convert markdown to HTML
	htmlContent := simpleMarkdownToHTML(string(mdContent))

	doc := LegalDocument{
		Title:       title,
		Content:     template.HTML(htmlContent),
		LastUpdated: "February 2026",
		Version:     "1.0",
	}

	// Check Accept header for content negotiation
	accept := r.Header.Get("Accept")
	if strings.Contains(accept, "application/json") {
		h.serveLegalAPI(w, r, strings.ToLower(strings.ReplaceAll(title, " ", "-")), content, defaultContent)
		return
	}

	if strings.Contains(accept, "text/markdown") {
		w.Header().Set("Content-Type", "text/markdown; charset=utf-8")
		w.Write(mdContent)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("Cache-Control", "public, max-age=3600")
	w.Header().Set("Last-Modified", h.lastModified.UTC().Format(http.TimeFormat))

	err := h.templates.ExecuteTemplate(w, "legal", doc)
	if err != nil {
		// Fallback to basic HTML
		w.Write([]byte(basicHTMLWrapper(title, htmlContent)))
	}
}

func (h *LegalHandler) serveLegalAPI(w http.ResponseWriter, r *http.Request, docType string, content []byte, defaultContent string) {
	mdContent := content
	if len(mdContent) == 0 {
		mdContent = []byte(defaultContent)
	}

	accept := r.Header.Get("Accept")

	if strings.Contains(accept, "text/markdown") {
		w.Header().Set("Content-Type", "text/markdown; charset=utf-8")
		w.Write(mdContent)
		return
	}

	if strings.Contains(accept, "text/html") {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Write([]byte(simpleMarkdownToHTML(string(mdContent))))
		return
	}

	// Default to JSON
	w.Header().Set("Content-Type", "application/json")

	response := LegalAPIResponse{
		DocumentType: docType,
		Version:      "1.0",
		LastUpdated:  "2026-02",
		Format:       "html",
		Content:      simpleMarkdownToHTML(string(mdContent)),
	}

	json.NewEncoder(w).Encode(response)
}

// simpleMarkdownToHTML converts basic markdown to HTML without external dependencies.
func simpleMarkdownToHTML(md string) string {
	// Escape HTML first
	content := html.EscapeString(md)

	// Process markdown elements

	// Headers
	content = regexp.MustCompile(`(?m)^######\s+(.+)$`).ReplaceAllString(content, `<h6>$1</h6>`)
	content = regexp.MustCompile(`(?m)^#####\s+(.+)$`).ReplaceAllString(content, `<h5>$1</h5>`)
	content = regexp.MustCompile(`(?m)^####\s+(.+)$`).ReplaceAllString(content, `<h4>$1</h4>`)
	content = regexp.MustCompile(`(?m)^###\s+(.+)$`).ReplaceAllString(content, `<h3>$1</h3>`)
	content = regexp.MustCompile(`(?m)^##\s+(.+)$`).ReplaceAllString(content, `<h2>$1</h2>`)
	content = regexp.MustCompile(`(?m)^#\s+(.+)$`).ReplaceAllString(content, `<h1>$1</h1>`)

	// Horizontal rules
	content = regexp.MustCompile(`(?m)^---+$`).ReplaceAllString(content, `<hr>`)

	// Bold
	content = regexp.MustCompile(`\*\*(.+?)\*\*`).ReplaceAllString(content, `<strong>$1</strong>`)

	// Italic
	content = regexp.MustCompile(`\*(.+?)\*`).ReplaceAllString(content, `<em>$1</em>`)

	// Inline code
	content = regexp.MustCompile("`([^`]+)`").ReplaceAllString(content, `<code>$1</code>`)

	// Code blocks
	content = regexp.MustCompile("(?s)```([a-z]*)\\n(.+?)```").ReplaceAllString(content, `<pre><code class="language-$1">$2</code></pre>`)

	// Links
	content = regexp.MustCompile(`\[([^\]]+)\]\(([^)]+)\)`).ReplaceAllString(content, `<a href="$2" target="_blank" rel="noopener">$1</a>`)

	// Tables (basic support)
	content = processMarkdownTables(content)

	// Lists
	content = processMarkdownLists(content)

	// Paragraphs - wrap remaining text blocks
	lines := strings.Split(content, "\n")
	var result []string
	var inParagraph bool
	var paragraphContent []string

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		// Skip if already an HTML element
		if strings.HasPrefix(trimmed, "<h") ||
			strings.HasPrefix(trimmed, "</h") ||
			strings.HasPrefix(trimmed, "<hr") ||
			strings.HasPrefix(trimmed, "<ul") ||
			strings.HasPrefix(trimmed, "</ul") ||
			strings.HasPrefix(trimmed, "<ol") ||
			strings.HasPrefix(trimmed, "</ol") ||
			strings.HasPrefix(trimmed, "<li") ||
			strings.HasPrefix(trimmed, "</li") ||
			strings.HasPrefix(trimmed, "<table") ||
			strings.HasPrefix(trimmed, "</table") ||
			strings.HasPrefix(trimmed, "<thead") ||
			strings.HasPrefix(trimmed, "<tbody") ||
			strings.HasPrefix(trimmed, "<tr") ||
			strings.HasPrefix(trimmed, "<th") ||
			strings.HasPrefix(trimmed, "<td") ||
			strings.HasPrefix(trimmed, "<pre") ||
			strings.HasPrefix(trimmed, "</pre") ||
			strings.HasPrefix(trimmed, "<blockquote") ||
			strings.HasPrefix(trimmed, "</blockquote") {
			if inParagraph && len(paragraphContent) > 0 {
				result = append(result, "<p>"+strings.Join(paragraphContent, " ")+"</p>")
				paragraphContent = nil
				inParagraph = false
			}
			result = append(result, line)
			continue
		}

		if trimmed == "" {
			if inParagraph && len(paragraphContent) > 0 {
				result = append(result, "<p>"+strings.Join(paragraphContent, " ")+"</p>")
				paragraphContent = nil
			}
			inParagraph = false
			continue
		}

		inParagraph = true
		paragraphContent = append(paragraphContent, trimmed)
	}

	if len(paragraphContent) > 0 {
		result = append(result, "<p>"+strings.Join(paragraphContent, " ")+"</p>")
	}

	return strings.Join(result, "\n")
}

func processMarkdownLists(content string) string {
	lines := strings.Split(content, "\n")
	var result []string
	var inList bool
	var listType string

	for i, line := range lines {
		trimmed := strings.TrimSpace(line)

		// Unordered list
		if strings.HasPrefix(trimmed, "- ") || strings.HasPrefix(trimmed, "* ") {
			if !inList || listType != "ul" {
				if inList {
					result = append(result, "</"+listType+">")
				}
				result = append(result, "<ul>")
				inList = true
				listType = "ul"
			}
			item := strings.TrimPrefix(strings.TrimPrefix(trimmed, "- "), "* ")
			result = append(result, "<li>"+item+"</li>")
			continue
		}

		// Ordered list
		if matched, _ := regexp.MatchString(`^\d+\.\s`, trimmed); matched {
			if !inList || listType != "ol" {
				if inList {
					result = append(result, "</"+listType+">")
				}
				result = append(result, "<ol>")
				inList = true
				listType = "ol"
			}
			item := regexp.MustCompile(`^\d+\.\s`).ReplaceAllString(trimmed, "")
			result = append(result, "<li>"+item+"</li>")
			continue
		}

		// End list if we hit a non-list line
		if inList && trimmed != "" {
			result = append(result, "</"+listType+">")
			inList = false
			listType = ""
		}

		result = append(result, lines[i])
	}

	// Close any open list
	if inList {
		result = append(result, "</"+listType+">")
	}

	return strings.Join(result, "\n")
}

func processMarkdownTables(content string) string {
	lines := strings.Split(content, "\n")
	var result []string
	var inTable bool
	var headerProcessed bool

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		// Check if this is a table row
		if strings.HasPrefix(trimmed, "|") && strings.HasSuffix(trimmed, "|") {
			// Check if this is a separator row
			if regexp.MustCompile(`^\|[\s\-:|]+\|$`).MatchString(trimmed) {
				if inTable {
					result = append(result, "</thead><tbody>")
					headerProcessed = true
				}
				continue
			}

			// Start table if needed
			if !inTable {
				result = append(result, "<table>")
				result = append(result, "<thead>")
				inTable = true
				headerProcessed = false
			}

			// Process cells
			cells := strings.Split(strings.Trim(trimmed, "|"), "|")
			cellTag := "td"
			if !headerProcessed {
				cellTag = "th"
			}

			row := "<tr>"
			for _, cell := range cells {
				row += "<" + cellTag + ">" + strings.TrimSpace(cell) + "</" + cellTag + ">"
			}
			row += "</tr>"
			result = append(result, row)
			continue
		}

		// End table if we hit a non-table line
		if inTable && trimmed != "" {
			if headerProcessed {
				result = append(result, "</tbody>")
			} else {
				result = append(result, "</thead>")
			}
			result = append(result, "</table>")
			inTable = false
			headerProcessed = false
		}

		result = append(result, line)
	}

	// Close any open table
	if inTable {
		if headerProcessed {
			result = append(result, "</tbody>")
		}
		result = append(result, "</table>")
	}

	return strings.Join(result, "\n")
}

func basicHTMLWrapper(title, content string) string {
	return `<!DOCTYPE html>
<html lang="en">
<head>
	<meta charset="UTF-8">
	<meta name="viewport" content="width=device-width, initial-scale=1.0">
	<title>` + html.EscapeString(title) + ` - CRM Platform</title>
	<style>
		body {
			font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Oxygen, Ubuntu, sans-serif;
			line-height: 1.6;
			max-width: 800px;
			margin: 0 auto;
			padding: 2rem;
			color: #333;
		}
		h1, h2, h3 { color: #1a1a1a; margin-top: 2rem; }
		h1 { border-bottom: 2px solid #e0e0e0; padding-bottom: 0.5rem; }
		h2 { border-bottom: 1px solid #e0e0e0; padding-bottom: 0.3rem; }
		table { border-collapse: collapse; width: 100%; margin: 1rem 0; }
		th, td { border: 1px solid #ddd; padding: 0.75rem; text-align: left; }
		th { background-color: #f5f5f5; }
		code { background: #f4f4f4; padding: 0.2rem 0.4rem; border-radius: 3px; }
		pre { background: #f4f4f4; padding: 1rem; overflow-x: auto; border-radius: 5px; }
		a { color: #0066cc; }
		blockquote { border-left: 4px solid #ddd; margin: 1rem 0; padding-left: 1rem; color: #666; }
		.footer { margin-top: 3rem; padding-top: 1rem; border-top: 1px solid #e0e0e0; font-size: 0.9rem; color: #666; }
		ul, ol { margin: 1rem 0; padding-left: 2rem; }
		li { margin-bottom: 0.5rem; }
	</style>
</head>
<body>
	<nav style="margin-bottom: 2rem;">
		<a href="/">Home</a> |
		<a href="/terms">Terms of Service</a> |
		<a href="/privacy">Privacy Policy</a> |
		<a href="/sla">SLA</a>
	</nav>
	<main>` + content + `</main>
	<footer class="footer">
		<p>&copy; 2026 Kilang Desa Murni Batik. All rights reserved.</p>
	</footer>
</body>
</html>`
}

// Default content when files are not available
const defaultTermsContent = `# Terms of Service

Please refer to our complete Terms of Service document.

Contact legal@crmplatform.my for more information.
`

const defaultPrivacyContent = `# Privacy Policy

Please refer to our complete Privacy Policy document.

Contact privacy@crmplatform.my for more information.
`

const defaultSLAContent = `# Service Level Agreement

Please refer to our complete SLA document.

Contact support@crmplatform.my for more information.
`

const defaultLegalTemplate = `{{define "legal"}}<!DOCTYPE html>
<html lang="en">
<head>
	<meta charset="UTF-8">
	<meta name="viewport" content="width=device-width, initial-scale=1.0">
	<title>{{.Title}} - CRM Platform</title>
	<style>
		body {
			font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Oxygen, Ubuntu, sans-serif;
			line-height: 1.6;
			max-width: 800px;
			margin: 0 auto;
			padding: 2rem;
			color: #333;
		}
		h1, h2, h3 { color: #1a1a1a; margin-top: 2rem; }
		h1 { border-bottom: 2px solid #e0e0e0; padding-bottom: 0.5rem; }
		h2 { border-bottom: 1px solid #e0e0e0; padding-bottom: 0.3rem; }
		table { border-collapse: collapse; width: 100%; margin: 1rem 0; }
		th, td { border: 1px solid #ddd; padding: 0.75rem; text-align: left; }
		th { background-color: #f5f5f5; }
		code { background: #f4f4f4; padding: 0.2rem 0.4rem; border-radius: 3px; }
		pre { background: #f4f4f4; padding: 1rem; overflow-x: auto; border-radius: 5px; }
		a { color: #0066cc; }
		blockquote { border-left: 4px solid #ddd; margin: 1rem 0; padding-left: 1rem; color: #666; }
		.footer { margin-top: 3rem; padding-top: 1rem; border-top: 1px solid #e0e0e0; font-size: 0.9rem; color: #666; }
		ul, ol { margin: 1rem 0; padding-left: 2rem; }
		li { margin-bottom: 0.5rem; }
	</style>
</head>
<body>
	<nav style="margin-bottom: 2rem;">
		<a href="/">Home</a> |
		<a href="/terms">Terms of Service</a> |
		<a href="/privacy">Privacy Policy</a> |
		<a href="/sla">SLA</a>
	</nav>
	<main>{{.Content}}</main>
	<footer class="footer">
		<p>Version {{.Version}} | Last Updated: {{.LastUpdated}}</p>
		<p>&copy; 2026 Kilang Desa Murni Batik. All rights reserved.</p>
	</footer>
</body>
</html>{{end}}`
