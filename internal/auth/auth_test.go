package auth

import (
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestHashPassword(t *testing.T) {
	password := "testpassword123"

	hash, err := HashPassword(password)
	if err != nil {
		t.Fatalf("HashPassword failed: %v", err)
	}
	if hash == "" {
		t.Error("Expected non-empty hash")
	}
	if hash == password {
		t.Error("Hash should not equal original password")
	}

	// Test that the same password generates different hashes (due to salt)
	hash2, err := HashPassword(password)
	if err != nil {
		t.Fatalf("HashPassword failed on second call: %v", err)
	}
	if hash == hash2 {
		t.Error("Expected different hashes for same password due to salt")
	}
}

func TestCheckPasswordHash(t *testing.T) {
	password := "testpassword123"
	wrongPassword := "wrongpassword"

	hash, err := HashPassword(password)
	if err != nil {
		t.Fatalf("HashPassword failed: %v", err)
	}

	// Test correct password
	err = CheckPasswordHash(password, hash)
	if err != nil {
		t.Errorf("CheckPasswordHash should succeed with correct password: %v", err)
	}

	// Test wrong password
	err = CheckPasswordHash(wrongPassword, hash)
	if err == nil {
		t.Error("CheckPasswordHash should fail with wrong password")
	}
}

func TestMakeJWT(t *testing.T) {
	userID := uuid.New()
	tokenSecret := "test-secret"
	expiresIn := time.Hour

	token, err := MakeJWT(userID, tokenSecret, expiresIn)
	if err != nil {
		t.Fatalf("MakeJWT failed: %v", err)
	}
	if token == "" {
		t.Error("Expected non-empty token")
	}

	// JWT should have 3 parts separated by dots
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		t.Errorf("Expected JWT to have 3 parts, got %d", len(parts))
	}
}

func TestValidateJWT_ValidToken(t *testing.T) {
	userID := uuid.New()
	tokenSecret := "test-secret"
	expiresIn := time.Hour

	// Create a token
	tokenString, err := MakeJWT(userID, tokenSecret, expiresIn)
	if err != nil {
		t.Fatalf("MakeJWT failed: %v", err)
	}

	// Validate the token
	validatedUserID, err := ValidateJWT(tokenString, tokenSecret)
	if err != nil {
		t.Fatalf("ValidateJWT failed: %v", err)
	}
	if validatedUserID != userID {
		t.Errorf("Expected user ID %v, got %v", userID, validatedUserID)
	}
}

func TestValidateJWT_ExpiredToken(t *testing.T) {
	userID := uuid.New()
	tokenSecret := "test-secret"
	expiresIn := -time.Hour // Token expired 1 hour ago

	// Create an expired token
	tokenString, err := MakeJWT(userID, tokenSecret, expiresIn)
	if err != nil {
		t.Fatalf("MakeJWT failed: %v", err)
	}

	// Validate the expired token
	_, err = ValidateJWT(tokenString, tokenSecret)
	if err == nil {
		t.Error("Expected error for expired token")
	}
	if !strings.Contains(err.Error(), "token is expired") {
		t.Errorf("Expected 'token is expired' error, got: %v", err)
	}
}

func TestValidateJWT_WrongSecret(t *testing.T) {
	userID := uuid.New()
	correctSecret := "correct-secret"
	wrongSecret := "wrong-secret"
	expiresIn := time.Hour

	// Create a token with correct secret
	tokenString, err := MakeJWT(userID, correctSecret, expiresIn)
	if err != nil {
		t.Fatalf("MakeJWT failed: %v", err)
	}

	// Try to validate with wrong secret
	_, err = ValidateJWT(tokenString, wrongSecret)
	if err == nil {
		t.Error("Expected error when validating with wrong secret")
	}
	if !strings.Contains(err.Error(), "signature is invalid") {
		t.Errorf("Expected 'signature is invalid' error, got: %v", err)
	}
}

func TestValidateJWT_InvalidTokenFormat(t *testing.T) {
	tokenSecret := "test-secret"

	tests := []struct {
		name        string
		tokenString string
	}{
		{
			name:        "empty token",
			tokenString: "",
		},
		{
			name:        "malformed token",
			tokenString: "invalid.jwt.token",
		},
		{
			name:        "not enough parts",
			tokenString: "invalid.jwt",
		},
		{
			name:        "random string",
			tokenString: "this-is-not-a-jwt",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := ValidateJWT(tt.tokenString, tokenSecret)
			if err == nil {
				t.Errorf("Expected error for invalid token: %s", tt.name)
			}
		})
	}
}

func TestValidateJWT_EmptySecret(t *testing.T) {
	userID := uuid.New()
	tokenSecret := "test-secret"
	expiresIn := time.Hour

	// Create a token with a secret
	tokenString, err := MakeJWT(userID, tokenSecret, expiresIn)
	if err != nil {
		t.Fatalf("MakeJWT failed: %v", err)
	}

	// Try to validate with empty secret
	_, err = ValidateJWT(tokenString, "")
	if err == nil {
		t.Error("Expected error when validating with empty secret")
	}
}

func TestMakeJWT_EmptySecret(t *testing.T) {
	userID := uuid.New()
	expiresIn := time.Hour

	// Try to create token with empty secret
	token, err := MakeJWT(userID, "", expiresIn)
	if err != nil {
		t.Fatalf("MakeJWT with empty secret failed: %v", err)
	}
	if token == "" {
		t.Error("Expected non-empty token even with empty secret")
	}

	// But validation should still work with empty secret
	validatedUserID, err := ValidateJWT(token, "")
	if err != nil {
		t.Fatalf("ValidateJWT with empty secret failed: %v", err)
	}
	if validatedUserID != userID {
		t.Errorf("Expected user ID %v, got %v", userID, validatedUserID)
	}
}

func TestJWT_RoundTrip_MultipleTokens(t *testing.T) {
	tokenSecret := "test-secret"
	expiresIn := time.Hour

	// Test multiple users
	users := []uuid.UUID{
		uuid.New(),
		uuid.New(),
		uuid.New(),
	}

	tokens := make([]string, len(users))

	// Create tokens for all users
	for i, userID := range users {
		token, err := MakeJWT(userID, tokenSecret, expiresIn)
		if err != nil {
			t.Fatalf("MakeJWT failed for user %d: %v", i, err)
		}
		tokens[i] = token
	}

	// Validate all tokens
	for i, token := range tokens {
		validatedUserID, err := ValidateJWT(token, tokenSecret)
		if err != nil {
			t.Fatalf("ValidateJWT failed for token %d: %v", i, err)
		}
		if validatedUserID != users[i] {
			t.Errorf("Token %d: expected user ID %v, got %v", i, users[i], validatedUserID)
		}
	}
}

func TestJWT_ShortExpiration(t *testing.T) {
	userID := uuid.New()
	tokenSecret := "test-secret"
	expiresIn := 1 * time.Second // Increased for test reliability

	// Create token
	tokenString, err := MakeJWT(userID, tokenSecret, expiresIn)
	if err != nil {
		t.Fatalf("MakeJWT failed: %v", err)
	}

	// Validate immediately (should work)
	validatedUserID, err := ValidateJWT(tokenString, tokenSecret)
	if err != nil {
		t.Fatalf("ValidateJWT failed immediately: %v", err)
	}
	if validatedUserID != userID {
		t.Errorf("Expected user ID %v, got %v", userID, validatedUserID)
	}

	// Wait for token to expire
	time.Sleep(2 * time.Second) // Increased to ensure expiration

	// Validate expired token (should fail)
	_, err = ValidateJWT(tokenString, tokenSecret)
	if err == nil {
		t.Error("Expected error for expired token after sleep")
	}
}

func TestJWT_DifferentSecrets(t *testing.T) {
	userID := uuid.New()
	expiresIn := time.Hour

	secrets := []string{
		"secret1",
		"secret2",
		"very-long-secret-key-12345",
		"short",
	}

	// Each secret should create a different token
	tokens := make([]string, len(secrets))
	for i, secret := range secrets {
		token, err := MakeJWT(userID, secret, expiresIn)
		if err != nil {
			t.Fatalf("MakeJWT failed for secret %d: %v", i, err)
		}
		tokens[i] = token

		// Validate with correct secret
		validatedUserID, err := ValidateJWT(token, secret)
		if err != nil {
			t.Fatalf("ValidateJWT failed for secret %d: %v", i, err)
		}
		if validatedUserID != userID {
			t.Errorf("Secret %d: expected user ID %v, got %v", i, userID, validatedUserID)
		}
	}

	// Each token should be different
	for i := 0; i < len(tokens); i++ {
		for j := i + 1; j < len(tokens); j++ {
			if tokens[i] == tokens[j] {
				t.Errorf("Tokens %d and %d should be different but are the same", i, j)
			}
		}
	}

	// Cross-validation should fail
	for i, token := range tokens {
		for j, secret := range secrets {
			if i != j {
				_, err := ValidateJWT(token, secret)
				if err == nil {
					t.Errorf("Token created with secret %d should not validate with secret %d", i, j)
				}
			}
		}
	}
}
