package crypto

import (
	"bytes"
	"testing"
)

func TestArgon2idDeriver_DeriveKey_ConsistentResults(t *testing.T) {
	deriver := NewArgon2idDeriver()
	password := []byte("test-password-123")
	salt := []byte("fixed-salt-16-byt")

	key1 := deriver.DeriveKey(password, salt)
	key2 := deriver.DeriveKey(password, salt)

	if len(key1) != 32 {
		t.Errorf("expected key length 32, got %d", len(key1))
	}

	if !bytes.Equal(key1, key2) {
		t.Error("same password and salt should produce same key")
	}
}

func TestArgon2idDeriver_DeriveKey_DifferentPasswords(t *testing.T) {
	deriver := NewArgon2idDeriver()
	salt := []byte("fixed-salt-16-byt")

	key1 := deriver.DeriveKey([]byte("password1"), salt)
	key2 := deriver.DeriveKey([]byte("password2"), salt)

	if bytes.Equal(key1, key2) {
		t.Error("different passwords should produce different keys")
	}
}

func TestArgon2idDeriver_DeriveKey_DifferentSalts(t *testing.T) {
	deriver := NewArgon2idDeriver()
	password := []byte("same-password")

	key1 := deriver.DeriveKey(password, []byte("salt1-16-bytes---"))
	key2 := deriver.DeriveKey(password, []byte("salt2-16-bytes---"))

	if bytes.Equal(key1, key2) {
		t.Error("different salts should produce different keys")
	}
}

func TestArgon2idDeriver_VerifyWithSalt_Success(t *testing.T) {
	deriver := NewArgon2idDeriver()
	password := []byte("correct-password")
	salt := []byte("test-salt-16-byte")

	expectedKey := deriver.DeriveKey(password, salt)

	if !deriver.VerifyWithSalt(password, salt, expectedKey) {
		t.Error("verification should succeed for correct password")
	}
}

func TestArgon2idDeriver_VerifyWithSalt_WrongPassword(t *testing.T) {
	deriver := NewArgon2idDeriver()
	salt := []byte("test-salt-16-byte")

	expectedKey := deriver.DeriveKey([]byte("correct-password"), salt)

	if deriver.VerifyWithSalt([]byte("wrong-password"), salt, expectedKey) {
		t.Error("verification should fail for wrong password")
	}
}

func TestArgon2idDeriver_VerifyWithSalt_WrongKeyLength(t *testing.T) {
	deriver := NewArgon2idDeriver()
	password := []byte("test-password")
	salt := []byte("test-salt-16-byte")

	wrongLengthKey := make([]byte, 16)

	if deriver.VerifyWithSalt(password, salt, wrongLengthKey) {
		t.Error("verification should fail for wrong key length")
	}
}

func TestArgon2idDeriver_GenerateSalt(t *testing.T) {
	deriver := NewArgon2idDeriver()

	salt1, err := deriver.GenerateSalt()
	if err != nil {
		t.Fatalf("GenerateSalt failed: %v", err)
	}

	if len(salt1) != 16 {
		t.Errorf("expected salt length 16, got %d", len(salt1))
	}

	salt2, err := deriver.GenerateSalt()
	if err != nil {
		t.Fatalf("GenerateSalt failed: %v", err)
	}

	if bytes.Equal(salt1, salt2) {
		t.Error("consecutive salts should be different (crypto/rand)")
	}
}

func TestArgon2idDeriver_Verify_WrongKeyLength(t *testing.T) {
	deriver := NewArgon2idDeriver()

	wrongLengthKey := make([]byte, 16)
	if deriver.Verify([]byte("password"), wrongLengthKey) {
		t.Error("Verify should fail for wrong key length")
	}
}

func TestArgon2idDeriver_GetKeyLength(t *testing.T) {
	deriver := NewArgon2idDeriver()
	if deriver.GetKeyLength() != 32 {
		t.Errorf("expected key length 32, got %d", deriver.GetKeyLength())
	}
}

func TestArgon2idDeriver_GetSaltLength(t *testing.T) {
	deriver := NewArgon2idDeriver()
	if deriver.GetSaltLength() != 16 {
		t.Errorf("expected salt length 16, got %d", deriver.GetSaltLength())
	}
}

func TestArgon2idDeriver_InterfaceCompliance(t *testing.T) {
	var _ KeyDerivationFunc = NewArgon2idDeriver()
}

func BenchmarkDeriveKey(b *testing.B) {
	deriver := NewArgon2idDeriver()
	password := []byte("benchmark-password")
	salt := []byte("benchmark-salt-16")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = deriver.DeriveKey(password, salt)
	}
}

func BenchmarkVerifyWithSalt(b *testing.B) {
	deriver := NewArgon2idDeriver()
	password := []byte("benchmark-password")
	salt := []byte("benchmark-salt-16")
	expectedKey := deriver.DeriveKey(password, salt)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = deriver.VerifyWithSalt(password, salt, expectedKey)
	}
}

func BenchmarkGenerateSalt(b *testing.B) {
	deriver := NewArgon2idDeriver()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = deriver.GenerateSalt()
	}
}
