package auth_test

import (
	"testing"
	"webshop/internal/auth"
)

func TestHashAndCheckPassword(t *testing.T) {
	password := "VeiligWachtwoord123!"

	hash, err := auth.HashPassword(password)
	if err != nil {
		t.Fatalf("Fout bij hashen van wachtwoord: %v", err)
	}

	if hash == password {
		t.Errorf("Hash mag niet gelijk zijn aan het platte wachtwoord")
	}

	if !auth.CheckPassword(password, hash) {
		t.Errorf("CheckPassword zou true moeten retourneren voor correct wachtwoord")
	}

	if auth.CheckPassword("FoutWachtwoord", hash) {
		t.Errorf("CheckPassword zou false moeten retourneren voor verkeerd wachtwoord")
	}
}

func TestGenerateSessionToken(t *testing.T) {
	token1, err := auth.GenerateSessionToken()
	if err != nil {
		t.Fatalf("Fout bij aanmaken sessietoken: %v", err)
	}

	if len(token1) != 64 {
		t.Errorf("Verwachtte tokenlengte van 64 tekens, kreeg %d", len(token1))
	}

	token2, err := auth.GenerateSessionToken()
	if err != nil {
		t.Fatalf("Fout bij aanmaken tweede sessietoken: %v", err)
	}

	if token1 == token2 {
		t.Errorf("Twee sessietokens mogen nooit identiek zijn")
	}
}