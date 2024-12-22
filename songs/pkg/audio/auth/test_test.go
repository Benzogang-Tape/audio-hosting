package auth_test

import (
	"crypto/ed25519"
	"encoding/base64"
	"testing"
	"time"

	"github.com/Benzogang-Tape/audio-hosting/songs/pkg/audio/auth"
	"github.com/google/uuid"
)

func TestEverything(t *testing.T) {
	// Failed initialization
	_, err := auth.NewSigner("private key", time.Minute)
	if err == nil {
		t.Fatalf("NewSigner should fail with invalid private key, got nil")
	}

	_, err = auth.NewParser("public key")
	if err == nil {
		t.Fatalf("NewParser should fail with invalid public key, got nil")
	}

	// Good initialization
	pub, priv := generateKeys(t)

	signer, err := auth.NewSigner(priv, time.Minute)
	if err != nil {
		t.Fatalf("NewSigner failed: %v", err)
	}

	parser, err := auth.NewParser(pub)
	if err != nil {
		t.Fatalf("NewParser failed: %v", err)
	}

	// Signing token
	sub := uuid.New()
	isArtist := true

	token, err := signer.Sign(auth.Token{Subject: sub, IsArtist: isArtist})
	if err != nil {
		t.Fatalf("Sign failed: %v", err)
	}

	// Parsing token
	claims, err := parser.Parse(token)
	if err != nil {
		t.Fatalf("ParseVerify failed: %v", err)
	}

	// Checking claims after parsing
	if claims.Subject != sub {
		t.Fatalf("Subject mismatch: got %s, want %s", claims.Subject, sub)
	}

	if claims.IsArtist != isArtist {
		t.Fatalf("IsArtist mismatch: got %t, want %t", claims.IsArtist, isArtist)
	}

	newPub, _ := generateKeys(t)

	// Using new public key
	invalidKeyParser, err := auth.NewParser(newPub)
	if err != nil {
		t.Fatalf("NewParser failed with new public key: %v", err)
	}

	_, err = invalidKeyParser.Parse(token)
	if err == nil {
		t.Fatalf("ParseVerify should fail with invalid public key, got nil")
	}

	// Token expired
	expiredToken, err := signer.Sign(auth.Token{
		Subject:  sub,
		IsArtist: isArtist,
		Exp:      time.Now().Add(-time.Minute).Unix(),
	})
	if err != nil {
		t.Fatalf("Sign failed for expired token: %v", err)
	}

	_, err = parser.Parse(expiredToken)
	if err == nil {
		t.Fatalf("ParseVerify should fail with expired token, got nil")
	}
}

func generateKeys(t *testing.T) (string, string) {
	pub, priv, err := ed25519.GenerateKey(nil)
	if err != nil {
		t.Fatalf("Failed to generate keys: %v", err)
	}

	encodedPriv := base64.StdEncoding.EncodeToString(priv)
	encodedPub := base64.StdEncoding.EncodeToString(pub)

	return encodedPub, encodedPriv
}
