// Package gateway provides API gateway functionality for the CRM application.
package gateway

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"io/ioutil"
	"math/big"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// ============================================================================
// TLS Configuration
// ============================================================================

// TLSConfig holds TLS configuration options.
type TLSConfig struct {
	// Certificate and key files
	CertFile string
	KeyFile  string

	// CA certificate for client verification
	CAFile string

	// Auto-generated self-signed certificates (for development)
	AutoCert     bool
	AutoCertDir  string
	AutoCertHost string

	// TLS version constraints
	MinVersion uint16
	MaxVersion uint16

	// Cipher suites (empty = use defaults)
	CipherSuites []uint16

	// Client authentication
	ClientAuth tls.ClientAuthType

	// ALPN protocols
	NextProtos []string

	// Session tickets
	SessionTicketsDisabled bool

	// Certificate reload interval (0 = disabled)
	ReloadInterval time.Duration

	// OCSP stapling
	EnableOCSPStapling bool
}

// DefaultTLSConfig returns secure default TLS configuration.
func DefaultTLSConfig() TLSConfig {
	return TLSConfig{
		MinVersion: tls.VersionTLS12,
		MaxVersion: tls.VersionTLS13,
		CipherSuites: []uint16{
			// TLS 1.3 cipher suites (automatically used when TLS 1.3)
			tls.TLS_AES_128_GCM_SHA256,
			tls.TLS_AES_256_GCM_SHA384,
			tls.TLS_CHACHA20_POLY1305_SHA256,
			// TLS 1.2 cipher suites
			tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
			tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
			tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305,
			tls.TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305,
		},
		ClientAuth:             tls.NoClientCert,
		NextProtos:             []string{"h2", "http/1.1"},
		SessionTicketsDisabled: false,
		EnableOCSPStapling:     true,
	}
}

// DevelopmentTLSConfig returns TLS configuration for development.
func DevelopmentTLSConfig() TLSConfig {
	config := DefaultTLSConfig()
	config.AutoCert = true
	config.AutoCertDir = "./.certs"
	config.AutoCertHost = "localhost"
	return config
}

// ============================================================================
// TLS Manager
// ============================================================================

// TLSManager manages TLS certificates and configuration.
type TLSManager struct {
	config     TLSConfig
	tlsConfig  *tls.Config
	cert       *tls.Certificate
	certPool   *x509.CertPool
	mu         sync.RWMutex
	stopReload chan struct{}
}

// NewTLSManager creates a new TLS manager.
func NewTLSManager(config TLSConfig) (*TLSManager, error) {
	manager := &TLSManager{
		config:     config,
		stopReload: make(chan struct{}),
	}

	if err := manager.initialize(); err != nil {
		return nil, err
	}

	// Start certificate reload goroutine if enabled
	if config.ReloadInterval > 0 {
		go manager.reloadCertificates()
	}

	return manager, nil
}

// initialize sets up the TLS configuration.
func (m *TLSManager) initialize() error {
	// Auto-generate certificates if configured
	if m.config.AutoCert {
		if err := m.generateSelfSignedCert(); err != nil {
			return fmt.Errorf("failed to generate self-signed certificate: %w", err)
		}
	}

	// Load certificate
	if err := m.loadCertificate(); err != nil {
		return fmt.Errorf("failed to load certificate: %w", err)
	}

	// Load CA certificate pool if configured
	if m.config.CAFile != "" {
		if err := m.loadCAPool(); err != nil {
			return fmt.Errorf("failed to load CA pool: %w", err)
		}
	}

	// Build TLS config
	m.buildTLSConfig()

	return nil
}

// generateSelfSignedCert generates a self-signed certificate for development.
func (m *TLSManager) generateSelfSignedCert() error {
	// Create certificate directory if it doesn't exist
	if err := os.MkdirAll(m.config.AutoCertDir, 0755); err != nil {
		return fmt.Errorf("failed to create cert directory: %w", err)
	}

	certPath := filepath.Join(m.config.AutoCertDir, "server.crt")
	keyPath := filepath.Join(m.config.AutoCertDir, "server.key")

	// Check if certificates already exist
	if _, err := os.Stat(certPath); err == nil {
		if _, err := os.Stat(keyPath); err == nil {
			m.config.CertFile = certPath
			m.config.KeyFile = keyPath
			return nil
		}
	}

	// Generate private key
	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return fmt.Errorf("failed to generate private key: %w", err)
	}

	// Generate serial number
	serialNumber, err := rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1), 128))
	if err != nil {
		return fmt.Errorf("failed to generate serial number: %w", err)
	}

	// Build certificate template
	template := x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			Organization: []string{"CRM Development"},
			CommonName:   m.config.AutoCertHost,
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().Add(365 * 24 * time.Hour), // 1 year
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
		IsCA:                  false,
	}

	// Add SANs
	hosts := strings.Split(m.config.AutoCertHost, ",")
	for _, h := range hosts {
		h = strings.TrimSpace(h)
		if ip := net.ParseIP(h); ip != nil {
			template.IPAddresses = append(template.IPAddresses, ip)
		} else {
			template.DNSNames = append(template.DNSNames, h)
		}
	}

	// Create certificate
	derBytes, err := x509.CreateCertificate(rand.Reader, &template, &template, &privateKey.PublicKey, privateKey)
	if err != nil {
		return fmt.Errorf("failed to create certificate: %w", err)
	}

	// Write certificate
	certFile, err := os.Create(certPath)
	if err != nil {
		return fmt.Errorf("failed to create cert file: %w", err)
	}
	defer certFile.Close()

	if err := pem.Encode(certFile, &pem.Block{Type: "CERTIFICATE", Bytes: derBytes}); err != nil {
		return fmt.Errorf("failed to write certificate: %w", err)
	}

	// Write private key
	keyFile, err := os.OpenFile(keyPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return fmt.Errorf("failed to create key file: %w", err)
	}
	defer keyFile.Close()

	keyBytes, err := x509.MarshalECPrivateKey(privateKey)
	if err != nil {
		return fmt.Errorf("failed to marshal private key: %w", err)
	}

	if err := pem.Encode(keyFile, &pem.Block{Type: "EC PRIVATE KEY", Bytes: keyBytes}); err != nil {
		return fmt.Errorf("failed to write private key: %w", err)
	}

	m.config.CertFile = certPath
	m.config.KeyFile = keyPath

	return nil
}

// loadCertificate loads the TLS certificate.
func (m *TLSManager) loadCertificate() error {
	if m.config.CertFile == "" || m.config.KeyFile == "" {
		return fmt.Errorf("certificate and key files are required")
	}

	cert, err := tls.LoadX509KeyPair(m.config.CertFile, m.config.KeyFile)
	if err != nil {
		return fmt.Errorf("failed to load certificate: %w", err)
	}

	m.mu.Lock()
	m.cert = &cert
	m.mu.Unlock()

	return nil
}

// loadCAPool loads the CA certificate pool for client verification.
func (m *TLSManager) loadCAPool() error {
	caData, err := ioutil.ReadFile(m.config.CAFile)
	if err != nil {
		return fmt.Errorf("failed to read CA file: %w", err)
	}

	pool := x509.NewCertPool()
	if !pool.AppendCertsFromPEM(caData) {
		return fmt.Errorf("failed to parse CA certificate")
	}

	m.mu.Lock()
	m.certPool = pool
	m.mu.Unlock()

	return nil
}

// buildTLSConfig builds the tls.Config.
func (m *TLSManager) buildTLSConfig() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.tlsConfig = &tls.Config{
		MinVersion:            m.config.MinVersion,
		MaxVersion:            m.config.MaxVersion,
		CipherSuites:          m.config.CipherSuites,
		ClientAuth:            m.config.ClientAuth,
		ClientCAs:             m.certPool,
		NextProtos:            m.config.NextProtos,
		SessionTicketsDisabled: m.config.SessionTicketsDisabled,

		// Use GetCertificate for dynamic certificate loading
		GetCertificate: m.getCertificate,

		// Prefer server cipher suites
		PreferServerCipherSuites: true,
	}
}

// getCertificate returns the certificate for a given client hello.
func (m *TLSManager) getCertificate(hello *tls.ClientHelloInfo) (*tls.Certificate, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.cert == nil {
		return nil, fmt.Errorf("no certificate loaded")
	}

	return m.cert, nil
}

// reloadCertificates periodically reloads certificates.
func (m *TLSManager) reloadCertificates() {
	ticker := time.NewTicker(m.config.ReloadInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if err := m.loadCertificate(); err != nil {
				// Log error but continue
				fmt.Printf("Failed to reload certificate: %v\n", err)
			}
		case <-m.stopReload:
			return
		}
	}
}

// GetTLSConfig returns the TLS configuration.
func (m *TLSManager) GetTLSConfig() *tls.Config {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.tlsConfig
}

// Close stops the certificate reload goroutine.
func (m *TLSManager) Close() {
	close(m.stopReload)
}

// ============================================================================
// TLS Server
// ============================================================================

// TLSServer wraps an HTTP server with TLS.
type TLSServer struct {
	server     *http.Server
	tlsManager *TLSManager
	config     TLSConfig
}

// NewTLSServer creates a new TLS-enabled HTTP server.
func NewTLSServer(addr string, handler http.Handler, config TLSConfig) (*TLSServer, error) {
	tlsManager, err := NewTLSManager(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create TLS manager: %w", err)
	}

	server := &http.Server{
		Addr:      addr,
		Handler:   handler,
		TLSConfig: tlsManager.GetTLSConfig(),

		// Timeouts
		ReadTimeout:       30 * time.Second,
		ReadHeaderTimeout: 10 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       120 * time.Second,

		// Max header size
		MaxHeaderBytes: 1 << 20, // 1 MB
	}

	return &TLSServer{
		server:     server,
		tlsManager: tlsManager,
		config:     config,
	}, nil
}

// ListenAndServeTLS starts the TLS server.
func (s *TLSServer) ListenAndServeTLS() error {
	return s.server.ListenAndServeTLS(s.config.CertFile, s.config.KeyFile)
}

// Shutdown gracefully shuts down the server.
func (s *TLSServer) Shutdown(timeout time.Duration) error {
	s.tlsManager.Close()

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	return s.server.Shutdown(ctx)
}

// ============================================================================
// TLS Redirect Middleware
// ============================================================================

// TLSRedirect redirects HTTP to HTTPS.
func TLSRedirect(httpsPort string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		host := r.Host

		// Remove port if present
		if idx := strings.Index(host, ":"); idx != -1 {
			host = host[:idx]
		}

		// Build HTTPS URL
		target := "https://" + host
		if httpsPort != "443" {
			target += ":" + httpsPort
		}
		target += r.URL.RequestURI()

		http.Redirect(w, r, target, http.StatusMovedPermanently)
	})
}

// TLSRedirectMiddleware wraps a handler with TLS redirect for HTTP requests.
func TLSRedirectMiddleware(httpsPort string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Check if already HTTPS
			if r.TLS != nil || r.Header.Get("X-Forwarded-Proto") == "https" {
				next.ServeHTTP(w, r)
				return
			}

			// Redirect to HTTPS
			TLSRedirect(httpsPort).ServeHTTP(w, r)
		})
	}
}

// ============================================================================
// HSTS (HTTP Strict Transport Security)
// ============================================================================

// HSTSConfig holds HSTS configuration.
type HSTSConfig struct {
	MaxAge            int  // Max age in seconds
	IncludeSubDomains bool // Include subdomains
	Preload           bool // Enable HSTS preload
}

// DefaultHSTSConfig returns default HSTS configuration.
func DefaultHSTSConfig() HSTSConfig {
	return HSTSConfig{
		MaxAge:            31536000, // 1 year
		IncludeSubDomains: true,
		Preload:           false,
	}
}

// HSTSMiddleware adds HSTS headers to responses.
func HSTSMiddleware(config HSTSConfig) func(http.Handler) http.Handler {
	header := fmt.Sprintf("max-age=%d", config.MaxAge)
	if config.IncludeSubDomains {
		header += "; includeSubDomains"
	}
	if config.Preload {
		header += "; preload"
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Strict-Transport-Security", header)
			next.ServeHTTP(w, r)
		})
	}
}

// ============================================================================
// Security Headers
// ============================================================================

// SecurityHeadersMiddleware adds security headers to responses.
func SecurityHeadersMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Content Security Policy
			w.Header().Set("Content-Security-Policy", "default-src 'self'")

			// X-Content-Type-Options
			w.Header().Set("X-Content-Type-Options", "nosniff")

			// X-Frame-Options
			w.Header().Set("X-Frame-Options", "DENY")

			// X-XSS-Protection
			w.Header().Set("X-XSS-Protection", "1; mode=block")

			// Referrer-Policy
			w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")

			// Permissions-Policy
			w.Header().Set("Permissions-Policy", "geolocation=(), microphone=(), camera=()")

			next.ServeHTTP(w, r)
		})
	}
}

// ============================================================================
// mTLS (Mutual TLS) Support
// ============================================================================

// MTLSConfig holds mTLS configuration.
type MTLSConfig struct {
	TLSConfig
	RequireClientCert bool
	AllowedCNs        []string // Allowed Common Names
	AllowedOUs        []string // Allowed Organizational Units
}

// DefaultMTLSConfig returns default mTLS configuration.
func DefaultMTLSConfig() MTLSConfig {
	config := MTLSConfig{
		TLSConfig:         DefaultTLSConfig(),
		RequireClientCert: true,
	}
	config.ClientAuth = tls.RequireAndVerifyClientCert
	return config
}

// MTLSVerifier verifies client certificates in mTLS.
type MTLSVerifier struct {
	config MTLSConfig
}

// NewMTLSVerifier creates a new mTLS verifier.
func NewMTLSVerifier(config MTLSConfig) *MTLSVerifier {
	return &MTLSVerifier{config: config}
}

// VerifyMiddleware returns middleware that verifies client certificates.
func (v *MTLSVerifier) VerifyMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.TLS == nil {
				http.Error(w, "TLS required", http.StatusForbidden)
				return
			}

			if len(r.TLS.PeerCertificates) == 0 {
				http.Error(w, "Client certificate required", http.StatusForbidden)
				return
			}

			cert := r.TLS.PeerCertificates[0]

			// Verify Common Name if configured
			if len(v.config.AllowedCNs) > 0 {
				allowed := false
				for _, cn := range v.config.AllowedCNs {
					if cert.Subject.CommonName == cn {
						allowed = true
						break
					}
				}
				if !allowed {
					http.Error(w, "Invalid client certificate CN", http.StatusForbidden)
					return
				}
			}

			// Verify Organizational Unit if configured
			if len(v.config.AllowedOUs) > 0 {
				allowed := false
				for _, ou := range cert.Subject.OrganizationalUnit {
					for _, allowedOU := range v.config.AllowedOUs {
						if ou == allowedOU {
							allowed = true
							break
						}
					}
				}
				if !allowed {
					http.Error(w, "Invalid client certificate OU", http.StatusForbidden)
					return
				}
			}

			next.ServeHTTP(w, r)
		})
	}
}
