package auth

import "testing"

// ── Token tests ────────────────────────────────────────────────────

func TestNewTokenReturnsNonEmpty(t *testing.T) {
	token, err := NewToken()
	if err != nil {
		t.Fatalf("NewToken() error: %v", err)
	}
	if token == "" {
		t.Fatal("NewToken() returned empty string")
	}
}

func TestNewTokenIsUnique(t *testing.T) {
	token1, err := NewToken()
	if err != nil {
		t.Fatalf("NewToken() error: %v", err)
	}
	token2, err := NewToken()
	if err != nil {
		t.Fatalf("NewToken() error: %v", err)
	}
	if token1 == token2 {
		t.Fatal("two calls to NewToken() returned the same value")
	}
}

func TestHashTokenDeterministic(t *testing.T) {
	token := "test-token-12345"
	hash1 := HashToken(token)
	hash2 := HashToken(token)
	if hash1 != hash2 {
		t.Fatalf("HashToken() not deterministic: %s != %s", hash1, hash2)
	}
}

func TestHashTokenDifferentInputs(t *testing.T) {
	hash1 := HashToken("token-a")
	hash2 := HashToken("token-b")
	if hash1 == hash2 {
		t.Fatal("different tokens produced the same hash")
	}
}

func TestHashTokenNotEmpty(t *testing.T) {
	hash := HashToken("any-token")
	if hash == "" {
		t.Fatal("HashToken() returned empty string")
	}
}

// ── Password tests ─────────────────────────────────────────────────

func TestHashAndCheckPassword(t *testing.T) {
	password := "TestP@ssw0rd!"
	hashed, err := HashPassword(password)
	if err != nil {
		t.Fatalf("HashPassword() error: %v", err)
	}
	if hashed == "" {
		t.Fatal("HashPassword() returned empty string")
	}
	if hashed == password {
		t.Fatal("HashPassword() returned plaintext")
	}

	if err := CheckPassword(hashed, password); err != nil {
		t.Fatalf("CheckPassword() failed for correct password: %v", err)
	}
}

func TestCheckPasswordWrongInput(t *testing.T) {
	hashed, err := HashPassword("CorrectP@ss1!")
	if err != nil {
		t.Fatalf("HashPassword() error: %v", err)
	}

	if err := CheckPassword(hashed, "WrongP@ss1!"); err == nil {
		t.Fatal("CheckPassword() should fail for wrong password")
	}
}

// ── Password validation tests ──────────────────────────────────────

func TestValidatePasswordTooShort(t *testing.T) {
	err := validatePassword("Ab1!")
	if err == nil {
		t.Fatal("expected error for short password")
	}
}

func TestValidatePasswordNoUpper(t *testing.T) {
	err := validatePassword("abcdef1234!")
	if err == nil {
		t.Fatal("expected error for no uppercase")
	}
}

func TestValidatePasswordNoLower(t *testing.T) {
	err := validatePassword("ABCDEF1234!")
	if err == nil {
		t.Fatal("expected error for no lowercase")
	}
}

func TestValidatePasswordNoDigit(t *testing.T) {
	err := validatePassword("Abcdefghij!")
	if err == nil {
		t.Fatal("expected error for no digit")
	}
}

func TestValidatePasswordNoSymbol(t *testing.T) {
	err := validatePassword("Abcdefghi1")
	if err == nil {
		t.Fatal("expected error for no symbol")
	}
}

func TestValidatePasswordCommon(t *testing.T) {
	err := validatePassword("Password123!")
	if err == nil {
		t.Fatal("expected error for common password substring")
	}
}

func TestValidatePasswordValid(t *testing.T) {
	err := validatePassword("MyStr0ng!Pass")
	if err != nil {
		t.Fatalf("expected valid password, got error: %v", err)
	}
}
