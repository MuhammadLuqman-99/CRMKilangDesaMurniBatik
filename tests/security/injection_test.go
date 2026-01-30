// Package security contains injection tests for the CRM system.
package security

import (
	"net/http"
	"net/url"
	"strings"
	"testing"
)

// ============================================================================
// SQL Injection Tests
// ============================================================================

// TestSQLInjection tests for SQL injection vulnerabilities.
func TestSQLInjection(t *testing.T) {
	server := NewSecurityTestServer()
	defer server.Close()
	server.Seed()

	client := NewSecurityTestClient(server.Server.URL)
	report := NewSecurityReport()

	// Login to get token
	resp, body, _ := client.DoRequest("POST", "/api/v1/auth/login", map[string]interface{}{
		"email":     "user1@tenant1.com",
		"password":  "Password123!",
		"tenant_id": getTenantID(server, "tenant-one"),
	}, nil)

	if resp.StatusCode == http.StatusOK {
		var loginResp map[string]interface{}
		parseJSON(body, &loginResp)
		client.SetToken(loginResp["access_token"].(string))
	}

	// SQL Injection payloads
	sqlPayloads := []string{
		"' OR '1'='1",
		"'; DROP TABLE users;--",
		"' UNION SELECT * FROM users--",
		"1' OR '1'='1' /*",
		"admin'--",
		"' OR 1=1--",
		"'; WAITFOR DELAY '0:0:5'--",
		"1; SELECT * FROM users",
		"' HAVING 1=1--",
		"' GROUP BY 1--",
		"') OR ('1'='1",
		"1' AND '1'='1",
		"' OR ''='",
		"'/*",
		"*/OR/*",
		"1' ORDER BY 1--",
		"1' ORDER BY 100--",
	}

	t.Run("Login SQL Injection", func(t *testing.T) {
		for _, payload := range sqlPayloads {
			resp, _, _ := client.DoRequest("POST", "/api/v1/auth/login", map[string]interface{}{
				"email":     payload,
				"password":  "anything",
				"tenant_id": "test",
			}, nil)

			// Should not return 200 OK with SQL injection
			passed := resp.StatusCode != http.StatusOK
			report.AddResult(VulnerabilityReport{
				TestName:       "Login SQL Injection: " + truncate(payload, 30),
				Category:       "SQL Injection",
				Severity:       "Critical",
				Passed:         passed,
				Description:    "SQL injection attempt in login email field",
				Recommendation: "Use parameterized queries",
				Evidence:       payload,
			})

			if !passed {
				t.Errorf("SQL injection vulnerability: %s returned %d", payload, resp.StatusCode)
			}
		}
	})

	t.Run("Search SQL Injection", func(t *testing.T) {
		for _, payload := range sqlPayloads {
			encoded := url.QueryEscape(payload)
			resp, _, _ := client.DoRequest("GET", "/api/v1/customers/search?q="+encoded, nil, nil)

			// Should handle gracefully (not crash or expose data)
			passed := resp.StatusCode != http.StatusInternalServerError
			report.AddResult(VulnerabilityReport{
				TestName:       "Search SQL Injection: " + truncate(payload, 30),
				Category:       "SQL Injection",
				Severity:       "High",
				Passed:         passed,
				Description:    "SQL injection attempt in search query",
				Recommendation: "Sanitize and validate search input",
				Evidence:       payload,
			})
		}
	})

	t.Run("ID Parameter SQL Injection", func(t *testing.T) {
		for _, payload := range sqlPayloads {
			encoded := url.PathEscape(payload)
			resp, _, _ := client.DoRequest("GET", "/api/v1/customers/"+encoded, nil, nil)

			// Should return 404 Not Found, not error or success
			passed := resp.StatusCode == http.StatusNotFound || resp.StatusCode == http.StatusBadRequest
			report.AddResult(VulnerabilityReport{
				TestName:       "ID SQL Injection: " + truncate(payload, 30),
				Category:       "SQL Injection",
				Severity:       "Critical",
				Passed:         passed,
				Description:    "SQL injection in ID parameter",
				Recommendation: "Validate ID format before query",
				Evidence:       payload,
			})
		}
	})

	report.Print()
}

// ============================================================================
// NoSQL Injection Tests
// ============================================================================

// TestNoSQLInjection tests for NoSQL injection vulnerabilities.
func TestNoSQLInjection(t *testing.T) {
	server := NewSecurityTestServer()
	defer server.Close()
	server.Seed()

	client := NewSecurityTestClient(server.Server.URL)
	report := NewSecurityReport()

	// Login to get token
	loginAndSetToken(client, server)

	// NoSQL Injection payloads
	nosqlPayloads := []map[string]interface{}{
		{"$gt": ""},
		{"$ne": nil},
		{"$where": "this.password.length > 0"},
		{"$regex": ".*"},
		{"$exists": true},
		{"$or": []map[string]interface{}{{"a": 1}, {"b": 2}}},
	}

	t.Run("Login NoSQL Injection", func(t *testing.T) {
		for _, payload := range nosqlPayloads {
			resp, _, _ := client.DoRequest("POST", "/api/v1/auth/login", map[string]interface{}{
				"email":     payload,
				"password":  "anything",
				"tenant_id": "test",
			}, nil)

			passed := resp.StatusCode != http.StatusOK
			report.AddResult(VulnerabilityReport{
				TestName:       "Login NoSQL Injection",
				Category:       "NoSQL Injection",
				Severity:       "Critical",
				Passed:         passed,
				Description:    "NoSQL injection attempt in login",
				Recommendation: "Validate input types and sanitize operators",
			})
		}
	})

	t.Run("Query NoSQL Injection", func(t *testing.T) {
		// Test in request body
		resp, _, _ := client.DoRequest("POST", "/api/v1/customers", map[string]interface{}{
			"name": map[string]interface{}{"$gt": ""},
			"code": "TEST",
		}, nil)

		passed := resp.StatusCode != http.StatusCreated || resp.StatusCode == http.StatusBadRequest
		report.AddResult(VulnerabilityReport{
			TestName:       "Body NoSQL Injection",
			Category:       "NoSQL Injection",
			Severity:       "High",
			Passed:         passed,
			Description:    "NoSQL injection in request body",
			Recommendation: "Validate input types strictly",
		})
	})

	report.Print()
}

// ============================================================================
// XSS (Cross-Site Scripting) Tests
// ============================================================================

// TestXSSInjection tests for XSS vulnerabilities.
func TestXSSInjection(t *testing.T) {
	server := NewSecurityTestServer()
	defer server.Close()
	server.Seed()

	client := NewSecurityTestClient(server.Server.URL)
	report := NewSecurityReport()

	loginAndSetToken(client, server)

	// XSS payloads
	xssPayloads := []string{
		"<script>alert('XSS')</script>",
		"<img src=x onerror=alert('XSS')>",
		"<svg onload=alert('XSS')>",
		"javascript:alert('XSS')",
		"<body onload=alert('XSS')>",
		"<iframe src='javascript:alert(1)'>",
		"<input onfocus=alert('XSS') autofocus>",
		"<marquee onstart=alert('XSS')>",
		"<video><source onerror=alert('XSS')>",
		"<details open ontoggle=alert('XSS')>",
		"\"><script>alert('XSS')</script>",
		"'><script>alert('XSS')</script>",
		"<scr<script>ipt>alert('XSS')</scr</script>ipt>",
		"<ScRiPt>alert('XSS')</ScRiPt>",
		"<script/src=data:,alert('XSS')>",
		"<math><maction actiontype=statusline#http://evil.com>XSS",
	}

	t.Run("Stored XSS in Customer Name", func(t *testing.T) {
		for _, payload := range xssPayloads {
			// Create customer with XSS payload
			resp, body, _ := client.DoRequest("POST", "/api/v1/customers", map[string]interface{}{
				"code":  "XSS-TEST",
				"name":  payload,
				"email": "xss@test.com",
			}, nil)

			if resp.StatusCode == http.StatusCreated {
				var customer map[string]interface{}
				parseJSON(body, &customer)

				// Check if payload is stored and returned escaped
				if name, ok := customer["name"].(string); ok {
					containsScript := strings.Contains(name, "<script>")
					passed := !containsScript || name != payload

					report.AddResult(VulnerabilityReport{
						TestName:       "Stored XSS: " + truncate(payload, 30),
						Category:       "XSS",
						Severity:       "High",
						Passed:         passed,
						Description:    "Stored XSS in customer name field",
						Recommendation: "Encode output and sanitize input",
						Evidence:       payload,
					})
				}
			}
		}
	})

	t.Run("XSS in Search Query", func(t *testing.T) {
		for _, payload := range xssPayloads {
			encoded := url.QueryEscape(payload)
			resp, body, _ := client.DoRequest("GET", "/api/v1/customers/search?q="+encoded, nil, nil)

			if resp.StatusCode == http.StatusOK {
				// Check if response contains unescaped script tags
				containsScript := strings.Contains(string(body), "<script>")
				passed := !containsScript

				report.AddResult(VulnerabilityReport{
					TestName:       "Reflected XSS: " + truncate(payload, 30),
					Category:       "XSS",
					Severity:       "Medium",
					Passed:         passed,
					Description:    "Reflected XSS in search response",
					Recommendation: "Encode all output",
					Evidence:       payload,
				})
			}
		}
	})

	t.Run("XSS Security Headers", func(t *testing.T) {
		resp, _, _ := client.DoRequest("GET", "/api/v1/customers", nil, nil)

		// Check for XSS protection headers
		xssHeader := resp.Header.Get("X-XSS-Protection")
		cspHeader := resp.Header.Get("Content-Security-Policy")
		contentType := resp.Header.Get("X-Content-Type-Options")

		report.AddResult(VulnerabilityReport{
			TestName:       "X-XSS-Protection Header",
			Category:       "XSS",
			Severity:       "Low",
			Passed:         xssHeader != "",
			Description:    "X-XSS-Protection header should be set",
			Recommendation: "Set X-XSS-Protection: 1; mode=block",
		})

		report.AddResult(VulnerabilityReport{
			TestName:       "Content-Security-Policy Header",
			Category:       "XSS",
			Severity:       "Medium",
			Passed:         cspHeader != "",
			Description:    "CSP header should be set",
			Recommendation: "Implement Content-Security-Policy",
		})

		report.AddResult(VulnerabilityReport{
			TestName:       "X-Content-Type-Options Header",
			Category:       "XSS",
			Severity:       "Low",
			Passed:         contentType == "nosniff",
			Description:    "X-Content-Type-Options should be nosniff",
			Recommendation: "Set X-Content-Type-Options: nosniff",
		})
	})

	report.Print()
}

// ============================================================================
// Command Injection Tests
// ============================================================================

// TestCommandInjection tests for command injection vulnerabilities.
func TestCommandInjection(t *testing.T) {
	server := NewSecurityTestServer()
	defer server.Close()
	server.Seed()

	client := NewSecurityTestClient(server.Server.URL)
	report := NewSecurityReport()

	loginAndSetToken(client, server)

	// Command injection payloads
	cmdPayloads := []string{
		"; ls -la",
		"| cat /etc/passwd",
		"& whoami",
		"`id`",
		"$(whoami)",
		"; sleep 5",
		"| sleep 5 |",
		"\n/bin/ls",
		"; ping -c 1 127.0.0.1",
		"|| true",
		"&& echo vulnerable",
		"; echo vulnerable #",
		"$(cat /etc/passwd)",
		"`cat /etc/passwd`",
		"%0a id",
	}

	t.Run("Command Injection in Fields", func(t *testing.T) {
		for _, payload := range cmdPayloads {
			resp, body, _ := client.DoRequest("POST", "/api/v1/customers", map[string]interface{}{
				"code":  payload,
				"name":  "Test",
				"email": "test@test.com",
			}, nil)

			// Check response for command output indicators
			bodyStr := string(body)
			hasCommandOutput := strings.Contains(bodyStr, "root:") ||
				strings.Contains(bodyStr, "uid=") ||
				strings.Contains(bodyStr, "vulnerable")

			passed := !hasCommandOutput
			report.AddResult(VulnerabilityReport{
				TestName:       "Command Injection: " + truncate(payload, 30),
				Category:       "Command Injection",
				Severity:       "Critical",
				Passed:         passed,
				Description:    "Command injection attempt",
				Recommendation: "Never pass user input to shell commands",
				Evidence:       payload,
			})

			if !passed {
				t.Errorf("Command injection vulnerability with payload: %s", payload)
			}

			_ = resp // Use resp if needed
		}
	})

	report.Print()
}

// ============================================================================
// LDAP Injection Tests
// ============================================================================

// TestLDAPInjection tests for LDAP injection vulnerabilities.
func TestLDAPInjection(t *testing.T) {
	server := NewSecurityTestServer()
	defer server.Close()
	server.Seed()

	client := NewSecurityTestClient(server.Server.URL)
	report := NewSecurityReport()

	// LDAP injection payloads
	ldapPayloads := []string{
		"*",
		"*)(&",
		"*)(uid=*))(|(uid=*",
		"admin)(&)",
		"*)(objectClass=*",
		"admin)(|(password=*))",
		"*))%00",
		")(cn=*)(|(cn=*",
	}

	t.Run("LDAP Injection in Login", func(t *testing.T) {
		for _, payload := range ldapPayloads {
			resp, _, _ := client.DoRequest("POST", "/api/v1/auth/login", map[string]interface{}{
				"email":     payload,
				"password":  "test",
				"tenant_id": "test",
			}, nil)

			passed := resp.StatusCode != http.StatusOK
			report.AddResult(VulnerabilityReport{
				TestName:       "LDAP Injection: " + truncate(payload, 30),
				Category:       "LDAP Injection",
				Severity:       "High",
				Passed:         passed,
				Description:    "LDAP injection attempt in authentication",
				Recommendation: "Escape LDAP special characters",
				Evidence:       payload,
			})
		}
	})

	report.Print()
}

// ============================================================================
// XML/XXE Injection Tests
// ============================================================================

// TestXXEInjection tests for XXE vulnerabilities.
func TestXXEInjection(t *testing.T) {
	server := NewSecurityTestServer()
	defer server.Close()
	server.Seed()

	client := NewSecurityTestClient(server.Server.URL)
	report := NewSecurityReport()

	loginAndSetToken(client, server)

	// XXE payloads
	xxePayloads := []string{
		`<?xml version="1.0"?><!DOCTYPE foo [<!ENTITY xxe SYSTEM "file:///etc/passwd">]><foo>&xxe;</foo>`,
		`<?xml version="1.0"?><!DOCTYPE foo [<!ENTITY xxe SYSTEM "http://evil.com/xxe">]><foo>&xxe;</foo>`,
		`<?xml version="1.0"?><!DOCTYPE foo [<!ENTITY % xxe SYSTEM "file:///etc/passwd">%xxe;]>`,
	}

	t.Run("XXE in Request Body", func(t *testing.T) {
		for _, payload := range xxePayloads {
			resp, body, _ := client.DoRequest("POST", "/api/v1/customers", payload, map[string]string{
				"Content-Type": "application/xml",
			})

			// Should reject XML or not process XXE
			bodyStr := string(body)
			hasPasswdContent := strings.Contains(bodyStr, "root:")

			passed := !hasPasswdContent && resp.StatusCode != http.StatusOK
			report.AddResult(VulnerabilityReport{
				TestName:       "XXE Injection",
				Category:       "XXE",
				Severity:       "Critical",
				Passed:         passed,
				Description:    "XXE injection attempt",
				Recommendation: "Disable external entities in XML parser",
				Evidence:       truncate(payload, 50),
			})
		}
	})

	report.Print()
}

// ============================================================================
// Path Traversal Tests
// ============================================================================

// TestPathTraversal tests for path traversal vulnerabilities.
func TestPathTraversal(t *testing.T) {
	server := NewSecurityTestServer()
	defer server.Close()
	server.Seed()

	client := NewSecurityTestClient(server.Server.URL)
	report := NewSecurityReport()

	loginAndSetToken(client, server)

	// Path traversal payloads
	pathPayloads := []string{
		"../../../etc/passwd",
		"..\\..\\..\\windows\\system32\\config\\sam",
		"....//....//....//etc/passwd",
		"..%2f..%2f..%2fetc/passwd",
		"..%252f..%252f..%252fetc/passwd",
		"%2e%2e%2f%2e%2e%2f%2e%2e%2fetc/passwd",
		"....\/....\/....\/etc/passwd",
		"..%c0%af..%c0%af..%c0%afetc/passwd",
		"/etc/passwd%00.jpg",
		"....//....//....//etc/passwd%00",
	}

	t.Run("Path Traversal in ID", func(t *testing.T) {
		for _, payload := range pathPayloads {
			encoded := url.PathEscape(payload)
			resp, body, _ := client.DoRequest("GET", "/api/v1/customers/"+encoded, nil, nil)

			// Should not expose file contents
			bodyStr := string(body)
			hasFileContent := strings.Contains(bodyStr, "root:") ||
				strings.Contains(bodyStr, "[boot loader]")

			passed := !hasFileContent
			report.AddResult(VulnerabilityReport{
				TestName:       "Path Traversal: " + truncate(payload, 30),
				Category:       "Path Traversal",
				Severity:       "High",
				Passed:         passed,
				Description:    "Path traversal attempt",
				Recommendation: "Validate and sanitize file paths",
				Evidence:       payload,
			})

			_ = resp
		}
	})

	report.Print()
}

// ============================================================================
// Header Injection Tests
// ============================================================================

// TestHeaderInjection tests for HTTP header injection vulnerabilities.
func TestHeaderInjection(t *testing.T) {
	server := NewSecurityTestServer()
	defer server.Close()
	server.Seed()

	client := NewSecurityTestClient(server.Server.URL)
	report := NewSecurityReport()

	// Header injection payloads
	headerPayloads := []string{
		"value\r\nX-Injected: header",
		"value\nSet-Cookie: stolen=true",
		"value\r\n\r\n<html>injected</html>",
		"value%0d%0aX-Injected: header",
		"value%0aSet-Cookie: stolen=true",
	}

	t.Run("Header Injection in Request", func(t *testing.T) {
		for _, payload := range headerPayloads {
			resp, _, _ := client.DoRequest("POST", "/api/v1/auth/login", map[string]interface{}{
				"email":     "test@test.com",
				"password":  payload,
				"tenant_id": "test",
			}, nil)

			// Check if injected headers appear in response
			injectedHeader := resp.Header.Get("X-Injected")
			passed := injectedHeader == ""

			report.AddResult(VulnerabilityReport{
				TestName:       "Header Injection: " + truncate(payload, 30),
				Category:       "Header Injection",
				Severity:       "Medium",
				Passed:         passed,
				Description:    "HTTP header injection attempt",
				Recommendation: "Sanitize CRLF characters from input",
				Evidence:       payload,
			})
		}
	})

	report.Print()
}

// ============================================================================
// Helper Functions
// ============================================================================

func loginAndSetToken(client *SecurityTestClient, server *SecurityTestServer) {
	resp, body, _ := client.DoRequest("POST", "/api/v1/auth/login", map[string]interface{}{
		"email":     "user1@tenant1.com",
		"password":  "Password123!",
		"tenant_id": getTenantID(server, "tenant-one"),
	}, nil)

	if resp.StatusCode == http.StatusOK {
		var loginResp map[string]interface{}
		parseJSON(body, &loginResp)
		if token, ok := loginResp["access_token"].(string); ok {
			client.SetToken(token)
		}
	}
}

func getTenantID(server *SecurityTestServer, slug string) string {
	server.DataStore.mu.RLock()
	defer server.DataStore.mu.RUnlock()

	for _, t := range server.DataStore.tenants {
		if t.Slug == slug {
			return t.ID
		}
	}
	return ""
}

func parseJSON(data []byte, v interface{}) {
	json.Unmarshal(data, v)
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}
