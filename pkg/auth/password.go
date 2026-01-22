// Package auth provides authentication and authorization utilities for the CRM application.
package auth

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"fmt"
	"strings"

	"golang.org/x/crypto/argon2"
)

// PasswordConfig holds the configuration for password hashing.
type PasswordConfig struct {
	Time    uint32 // Number of iterations
	Memory  uint32 // Memory in KiB
	Threads uint8  // Number of threads
	KeyLen  uint32 // Length of the derived key
	SaltLen uint32 // Length of the salt
}

// DefaultPasswordConfig returns the default password configuration.
func DefaultPasswordConfig() *PasswordConfig {
	return &PasswordConfig{
		Time:    3,
		Memory:  64 * 1024, // 64 MiB
		Threads: 4,
		KeyLen:  32,
		SaltLen: 16,
	}
}

// PasswordHasher handles password hashing and verification.
type PasswordHasher struct {
	config *PasswordConfig
}

// NewPasswordHasher creates a new password hasher with the given configuration.
func NewPasswordHasher(config *PasswordConfig) *PasswordHasher {
	if config == nil {
		config = DefaultPasswordConfig()
	}
	return &PasswordHasher{config: config}
}

// Hash generates a hash for the given password.
func (h *PasswordHasher) Hash(password string) (string, error) {
	// Generate a random salt
	salt := make([]byte, h.config.SaltLen)
	if _, err := rand.Read(salt); err != nil {
		return "", fmt.Errorf("failed to generate salt: %w", err)
	}

	// Generate the hash using Argon2id
	hash := argon2.IDKey(
		[]byte(password),
		salt,
		h.config.Time,
		h.config.Memory,
		h.config.Threads,
		h.config.KeyLen,
	)

	// Encode the hash in the standard format
	// $argon2id$v=19$m=65536,t=3,p=4$salt$hash
	encoded := fmt.Sprintf(
		"$argon2id$v=%d$m=%d,t=%d,p=%d$%s$%s",
		argon2.Version,
		h.config.Memory,
		h.config.Time,
		h.config.Threads,
		base64.RawStdEncoding.EncodeToString(salt),
		base64.RawStdEncoding.EncodeToString(hash),
	)

	return encoded, nil
}

// Verify checks if the password matches the hash.
func (h *PasswordHasher) Verify(password, encodedHash string) (bool, error) {
	// Parse the encoded hash
	config, salt, hash, err := h.decodeHash(encodedHash)
	if err != nil {
		return false, err
	}

	// Generate a hash with the same parameters
	otherHash := argon2.IDKey(
		[]byte(password),
		salt,
		config.Time,
		config.Memory,
		config.Threads,
		config.KeyLen,
	)

	// Compare the hashes in constant time
	if subtle.ConstantTimeCompare(hash, otherHash) == 1 {
		return true, nil
	}

	return false, nil
}

// decodeHash decodes an encoded hash string.
func (h *PasswordHasher) decodeHash(encodedHash string) (*PasswordConfig, []byte, []byte, error) {
	parts := strings.Split(encodedHash, "$")
	if len(parts) != 6 {
		return nil, nil, nil, fmt.Errorf("invalid hash format")
	}

	if parts[1] != "argon2id" {
		return nil, nil, nil, fmt.Errorf("unsupported algorithm: %s", parts[1])
	}

	var version int
	if _, err := fmt.Sscanf(parts[2], "v=%d", &version); err != nil {
		return nil, nil, nil, fmt.Errorf("invalid version: %w", err)
	}

	if version != argon2.Version {
		return nil, nil, nil, fmt.Errorf("incompatible version: %d", version)
	}

	config := &PasswordConfig{}
	if _, err := fmt.Sscanf(parts[3], "m=%d,t=%d,p=%d", &config.Memory, &config.Time, &config.Threads); err != nil {
		return nil, nil, nil, fmt.Errorf("invalid parameters: %w", err)
	}

	salt, err := base64.RawStdEncoding.DecodeString(parts[4])
	if err != nil {
		return nil, nil, nil, fmt.Errorf("invalid salt: %w", err)
	}

	hash, err := base64.RawStdEncoding.DecodeString(parts[5])
	if err != nil {
		return nil, nil, nil, fmt.Errorf("invalid hash: %w", err)
	}

	config.KeyLen = uint32(len(hash))
	config.SaltLen = uint32(len(salt))

	return config, salt, hash, nil
}

// NeedsRehash checks if the hash needs to be rehashed with updated parameters.
func (h *PasswordHasher) NeedsRehash(encodedHash string) (bool, error) {
	config, _, _, err := h.decodeHash(encodedHash)
	if err != nil {
		return true, err
	}

	// Check if any parameters have changed
	if config.Time != h.config.Time ||
		config.Memory != h.config.Memory ||
		config.Threads != h.config.Threads ||
		config.KeyLen != h.config.KeyLen {
		return true, nil
	}

	return false, nil
}

// ValidatePasswordStrength checks if a password meets minimum requirements.
func ValidatePasswordStrength(password string) error {
	if len(password) < 8 {
		return fmt.Errorf("password must be at least 8 characters long")
	}

	var hasUpper, hasLower, hasDigit, hasSpecial bool

	for _, char := range password {
		switch {
		case 'A' <= char && char <= 'Z':
			hasUpper = true
		case 'a' <= char && char <= 'z':
			hasLower = true
		case '0' <= char && char <= '9':
			hasDigit = true
		case strings.ContainsRune("!@#$%^&*()_+-=[]{}|;':\",./<>?", char):
			hasSpecial = true
		}
	}

	if !hasUpper {
		return fmt.Errorf("password must contain at least one uppercase letter")
	}
	if !hasLower {
		return fmt.Errorf("password must contain at least one lowercase letter")
	}
	if !hasDigit {
		return fmt.Errorf("password must contain at least one digit")
	}
	if !hasSpecial {
		return fmt.Errorf("password must contain at least one special character")
	}

	return nil
}

// GenerateRandomPassword generates a random password of the specified length.
func GenerateRandomPassword(length int) (string, error) {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789!@#$%^&*"

	if length < 8 {
		length = 8
	}

	b := make([]byte, length)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("failed to generate random password: %w", err)
	}

	for i := range b {
		b[i] = charset[int(b[i])%len(charset)]
	}

	return string(b), nil
}

// GenerateSecureToken generates a cryptographically secure random token.
func GenerateSecureToken(length int) (string, error) {
	b := make([]byte, length)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("failed to generate token: %w", err)
	}
	return base64.URLEncoding.EncodeToString(b), nil
}
