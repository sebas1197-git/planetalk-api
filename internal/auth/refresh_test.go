package auth

import "testing"

func TestHashRefreshTokenDeterministic(t *testing.T) {
	a := hashRefreshToken("abc")
	if a != hashRefreshToken("abc") {
		t.Error("hash should be deterministic for the same input")
	}
	if a == hashRefreshToken("abd") {
		t.Error("different inputs should hash differently")
	}
	if len(a) != 64 { // sha256 hex
		t.Errorf("hash length = %d, want 64", len(a))
	}
}

func TestGenerateRefreshTokenUnique(t *testing.T) {
	t1, err := generateRefreshToken()
	if err != nil {
		t.Fatal(err)
	}
	t2, _ := generateRefreshToken()
	if t1 == t2 {
		t.Error("generated tokens should be unique")
	}
	if len(t1) != 64 { // 32 random bytes -> 64 hex chars
		t.Errorf("token length = %d, want 64", len(t1))
	}
}
