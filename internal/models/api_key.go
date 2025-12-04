package models

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"time"
)

type APIKeyType string

const (
	APIKeyTypePublic APIKeyType = "public"
	APIKeyTypeSecret APIKeyType = "secret"
)

type Environment string

const (
	EnvironmentTest Environment = "test"
	EnvironmentLive Environment = "live"
)

type APIKey struct {
	ID          string      `json:"id"`
	MerchantID  string      `json:"merchant_id"`
	KeyHash     string      `json:"-"` // Never expose the hash
	KeyPrefix   string      `json:"key_prefix"`
	KeyType     APIKeyType  `json:"key_type"`
	Environment Environment `json:"environment"`
	IsActive    bool        `json:"is_active"`
	LastUsedAt  *time.Time  `json:"last_used_at,omitempty"`
	ExpiresAt   *time.Time  `json:"expires_at,omitempty"`
	CreatedAt   time.Time   `json:"created_at"`
}

// GenerateAPIKey creates a new API key with the given prefix (pk_ or sk_)
func GenerateAPIKey(merchantID string, keyType APIKeyType, env Environment) (*APIKey, string, error) {
	// Generate random bytes
	keyBytes := make([]byte, 32)
	if _, err := rand.Read(keyBytes); err != nil {
		return nil, "", err
	}

	// Encode to base64
	keySecret := base64.URLEncoding.EncodeToString(keyBytes)

	// Determine prefix
	var prefix string
	if keyType == APIKeyTypePublic {
		prefix = "pk_"
	} else {
		prefix = "sk_"
	}

	if env == EnvironmentTest {
		prefix += "test_"
	} else {
		prefix += "live_"
	}

	// Full key
	fullKey := prefix + keySecret

	// Hash for storage
	hash := sha256.Sum256([]byte(fullKey))
	keyHash := fmt.Sprintf("%x", hash)

	// Get key prefix for identification (first 16 chars)
	keyPrefix := fullKey
	if len(fullKey) > 16 {
		keyPrefix = fullKey[:16]
	}

	apiKey := &APIKey{
		MerchantID:  merchantID,
		KeyHash:     keyHash,
		KeyPrefix:   keyPrefix,
		KeyType:     keyType,
		Environment: env,
		IsActive:    true,
		CreatedAt:   time.Now(),
	}

	return apiKey, fullKey, nil
}
