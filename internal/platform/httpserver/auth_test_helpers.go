package httpserver

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"time"

	"github.com/lestrrat-go/jwx/v2/jwa"
	"github.com/lestrrat-go/jwx/v2/jwk"
	"github.com/lestrrat-go/jwx/v2/jwt"
	"github.com/rs/zerolog"
)

// TestKeyPair holds a test RSA key pair for JWT signing
type TestKeyPair struct {
	PrivateKey *rsa.PrivateKey
	PublicKey  *rsa.PublicKey
	KeyID      string
}

// GenerateTestKeyPair creates a new RSA key pair for testing
func GenerateTestKeyPair() (*TestKeyPair, error) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, fmt.Errorf("failed to generate private key: %w", err)
	}

	keyID := fmt.Sprintf("test-key-%d", time.Now().Unix())

	return &TestKeyPair{
		PrivateKey: privateKey,
		PublicKey:  &privateKey.PublicKey,
		KeyID:      keyID,
	}, nil
}

// CreateMockJWT creates a JWT token with custom claims for testing
func CreateMockJWT(keyPair *TestKeyPair, claims map[string]interface{}) (string, error) {
	now := time.Now()

	// Build token with standard claims
	token := jwt.New()
	token.Set("sub", claims["sub"])
	token.Set("iat", now.Unix())
	token.Set("exp", now.Add(1*time.Hour).Unix())

	// Add custom claims
	for key, value := range claims {
		if key != "sub" && key != "iat" && key != "exp" {
			token.Set(key, value)
		}
	}

	// Convert RSA private key to JWK
	privateJWK, err := jwk.FromRaw(keyPair.PrivateKey)
	if err != nil {
		return "", fmt.Errorf("failed to create JWK from private key: %w", err)
	}

	// Set key ID
	if err := privateJWK.Set(jwk.KeyIDKey, keyPair.KeyID); err != nil {
		return "", fmt.Errorf("failed to set key ID: %w", err)
	}

	// Set algorithm
	if err := privateJWK.Set(jwk.AlgorithmKey, jwa.RS256); err != nil {
		return "", fmt.Errorf("failed to set algorithm: %w", err)
	}

	// Sign token
	signed, err := jwt.Sign(token, jwt.WithKey(jwa.RS256, privateJWK))
	if err != nil {
		return "", fmt.Errorf("failed to sign token: %w", err)
	}

	return string(signed), nil
}

// CreateExpiredJWT creates an expired JWT token for testing
func CreateExpiredJWT(keyPair *TestKeyPair, claims map[string]interface{}) (string, error) {
	now := time.Now()

	// Build token with expired exp claim
	token := jwt.New()
	token.Set("sub", claims["sub"])
	token.Set("iat", now.Add(-2*time.Hour).Unix())
	token.Set("exp", now.Add(-1*time.Hour).Unix()) // Expired 1 hour ago

	// Add custom claims
	for key, value := range claims {
		if key != "sub" && key != "iat" && key != "exp" {
			token.Set(key, value)
		}
	}

	// Convert RSA private key to JWK
	privateJWK, err := jwk.FromRaw(keyPair.PrivateKey)
	if err != nil {
		return "", fmt.Errorf("failed to create JWK from private key: %w", err)
	}

	// Set key ID
	if err := privateJWK.Set(jwk.KeyIDKey, keyPair.KeyID); err != nil {
		return "", fmt.Errorf("failed to set key ID: %w", err)
	}

	// Set algorithm
	if err := privateJWK.Set(jwk.AlgorithmKey, jwa.RS256); err != nil {
		return "", fmt.Errorf("failed to set algorithm: %w", err)
	}

	// Sign token
	signed, err := jwt.Sign(token, jwt.WithKey(jwa.RS256, privateJWK))
	if err != nil {
		return "", fmt.Errorf("failed to sign token: %w", err)
	}

	return string(signed), nil
}

// CreateMockJWKSServer creates a mock HTTP server that serves JWKS
func CreateMockJWKSServer(keyPair *TestKeyPair) (*httptest.Server, error) {
	// Convert public key to JWK
	publicJWK, err := jwk.FromRaw(keyPair.PublicKey)
	if err != nil {
		return nil, fmt.Errorf("failed to create JWK from public key: %w", err)
	}

	// Set key ID
	if err := publicJWK.Set(jwk.KeyIDKey, keyPair.KeyID); err != nil {
		return nil, fmt.Errorf("failed to set key ID: %w", err)
	}

	// Set algorithm
	if err := publicJWK.Set(jwk.AlgorithmKey, jwa.RS256); err != nil {
		return nil, fmt.Errorf("failed to set algorithm: %w", err)
	}

	// Set key use
	if err := publicJWK.Set(jwk.KeyUsageKey, "sig"); err != nil {
		return nil, fmt.Errorf("failed to set key usage: %w", err)
	}

	// Create key set
	keySet := jwk.NewSet()
	keySet.AddKey(publicJWK)

	// Create HTTP server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		// Convert keySet to JSON format
		keys := make([]map[string]interface{}, 0, keySet.Len())
		ctx := context.Background()
		for it := keySet.Keys(ctx); it.Next(ctx); {
			key := it.Pair().Value.(jwk.Key)
			// Marshal key to JSON
			raw, err := json.Marshal(key)
			if err != nil {
				continue
			}
			var keyMap map[string]interface{}
			if err := json.Unmarshal(raw, &keyMap); err != nil {
				continue
			}
			keys = append(keys, keyMap)
		}
		jwks := map[string]interface{}{
			"keys": keys,
		}
		raw, err := json.Marshal(jwks)
		if err != nil {
			http.Error(w, "Failed to marshal JWKS", http.StatusInternalServerError)
			return
		}
		w.Write(raw)
	}))

	return server, nil
}

// CreateTestRequest creates an HTTP request with optional headers
func CreateTestRequest(method, path string, headers map[string]string) *http.Request {
	req := httptest.NewRequest(method, path, nil)
	for key, value := range headers {
		req.Header.Set(key, value)
	}
	return req
}

// CreateTestRequireAuth creates a RequireAuth instance with a mock JWKS server
func CreateTestRequireAuth(keyPair *TestKeyPair) (*RequireAuth, *httptest.Server, error) {
	// Create mock JWKS server
	jwksServer, err := CreateMockJWKSServer(keyPair)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create mock JWKS server: %w", err)
	}

	// Create logger (discard output for tests)
	logger := zerolog.Nop()

	// Create RequireAuth with mock JWKS URL
	auth, err := NewRequireAuth(jwksServer.URL, logger)
	if err != nil {
		jwksServer.Close()
		return nil, nil, fmt.Errorf("failed to create RequireAuth: %w", err)
	}

	return auth, jwksServer, nil
}

// PEMToRSAPrivateKey converts a PEM-encoded private key to *rsa.PrivateKey
func PEMToRSAPrivateKey(pemData []byte) (*rsa.PrivateKey, error) {
	block, _ := pem.Decode(pemData)
	if block == nil {
		return nil, fmt.Errorf("failed to decode PEM block")
	}

	privateKey, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		// Try PKCS8 format
		key, err := x509.ParsePKCS8PrivateKey(block.Bytes)
		if err != nil {
			return nil, fmt.Errorf("failed to parse private key: %w", err)
		}
		rsaKey, ok := key.(*rsa.PrivateKey)
		if !ok {
			return nil, fmt.Errorf("key is not RSA private key")
		}
		return rsaKey, nil
	}

	return privateKey, nil
}

// CreateTestHandler creates a test HTTP handler that returns user context
func CreateTestHandler() http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		userID := ctx.Value("user_id")
		email := ctx.Value("email")
		orgID := ctx.Value("org_id")

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, `{"user_id":%q,"email":%q,"org_id":%q}`,
			userID, email, orgID)
	})
}

// AssertAuthError checks if the response is a 401 Unauthorized error
func AssertAuthError(t interface {
	Errorf(format string, args ...interface{})
}, resp *http.Response, expectedMessage string) {
	if resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("expected status 401, got %d", resp.StatusCode)
		return
	}

	// Read response body
	body := make([]byte, 1024)
	n, _ := resp.Body.Read(body)
	bodyStr := string(body[:n])

	if expectedMessage != "" && !strings.Contains(bodyStr, expectedMessage) {
		t.Errorf("expected error message to contain %q, got %q", expectedMessage, bodyStr)
	}
}

// MakeAuthenticatedRequest creates a request with a valid token
func MakeAuthenticatedRequest(method, path, token string, tokenSource TokenSource) *http.Request {
	req := CreateTestRequest(method, path, nil)
	
	switch tokenSource {
	case TokenSourceAuthorization:
		req.Header.Set("Authorization", "Bearer "+token)
	case TokenSourceClerkAuthToken:
		req.Header.Set("x-clerk-auth-token", token)
	case TokenSourceXAuthToken:
		req.Header.Set("X-Auth-Token", token)
	}
	
	return req
}

// MakeUnauthenticatedRequest creates a request without any auth headers
func MakeUnauthenticatedRequest(method, path string) *http.Request {
	return CreateTestRequest(method, path, nil)
}
