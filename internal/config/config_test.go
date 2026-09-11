package config

import "testing"

func TestJWTConfigValidateRejectsShortSecret(t *testing.T) {
	if err := (JWTConfig{Secret: "short", ExpiresIn: 86400}).Validate(); err == nil {
		t.Fatal("expected short JWT secret to be rejected")
	}
}

func TestJWTConfigValidateAcceptsUsableConfig(t *testing.T) {
	if err := (JWTConfig{Secret: "01234567890123456789012345678901", ExpiresIn: 86400}).Validate(); err != nil {
		t.Fatalf("expected config to be valid: %v", err)
	}
}
