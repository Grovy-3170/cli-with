package storage

import (
	"os"
	"path/filepath"
	"testing"
)

func TestCreate_Decrypt_Roundtrip(t *testing.T) {
	keys := map[string]string{
		"ANTHROPIC_API_KEY": "test-ant-12345",
		"OPENAI_API_KEY":    "test-openai-67890",
	}

	vault, err := Create("testpassword", keys)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	if vault.Version != VaultVersion {
		t.Errorf("Expected version %d, got %d", VaultVersion, vault.Version)
	}

	if vault.Format != VaultFormat {
		t.Errorf("Expected format %s, got %s", VaultFormat, vault.Format)
	}

	decryptedKeys, err := vault.Decrypt("testpassword")
	if err != nil {
		t.Fatalf("Decrypt failed: %v", err)
	}

	if len(decryptedKeys) != len(keys) {
		t.Errorf("Expected %d keys, got %d", len(keys), len(decryptedKeys))
	}

	for k, v := range keys {
		if decryptedKeys[k] != v {
			t.Errorf("Key %s: expected %s, got %s", k, v, decryptedKeys[k])
		}
	}
}

func TestDecrypt_WrongPassword(t *testing.T) {
	keys := map[string]string{
		"TEST_KEY": "test_value",
	}

	vault, err := Create("correctpassword", keys)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	_, err = vault.Decrypt("wrongpassword")
	if err != ErrInvalidPassword {
		t.Errorf("Expected ErrInvalidPassword, got %v", err)
	}
}

func TestCreate_EmptyKeys(t *testing.T) {
	vault, err := Create("password", nil)
	if err != nil {
		t.Fatalf("Create with nil keys failed: %v", err)
	}

	decryptedKeys, err := vault.Decrypt("password")
	if err != nil {
		t.Fatalf("Decrypt failed: %v", err)
	}

	if len(decryptedKeys) != 0 {
		t.Errorf("Expected 0 keys, got %d", len(decryptedKeys))
	}
}

func TestCreate_EmptyMapKeys(t *testing.T) {
	vault, err := Create("password", map[string]string{})
	if err != nil {
		t.Fatalf("Create with empty map failed: %v", err)
	}

	decryptedKeys, err := vault.Decrypt("password")
	if err != nil {
		t.Fatalf("Decrypt failed: %v", err)
	}

	if len(decryptedKeys) != 0 {
		t.Errorf("Expected 0 keys, got %d", len(decryptedKeys))
	}
}

func TestSave_Load_Roundtrip(t *testing.T) {
	tempDir := t.TempDir()
	vaultPath := filepath.Join(tempDir, "test.vault")

	keys := map[string]string{
		"API_KEY": "testval123",
	}

	vault, err := Create("password", keys)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	err = vault.Save(vaultPath)
	if err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	info, err := os.Stat(vaultPath)
	if err != nil {
		t.Fatalf("Stat failed: %v", err)
	}

	if info.Mode().Perm() != 0600 {
		t.Errorf("Expected permissions 0600, got %o", info.Mode().Perm())
	}

	loadedVault, err := Load(vaultPath)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	if loadedVault.Version != vault.Version {
		t.Errorf("Version mismatch: expected %d, got %d", vault.Version, loadedVault.Version)
	}

	decryptedKeys, err := loadedVault.Decrypt("password")
	if err != nil {
		t.Fatalf("Decrypt failed: %v", err)
	}

	if decryptedKeys["API_KEY"] != "testval123" {
		t.Errorf("Key mismatch: expected testval123, got %s", decryptedKeys["API_KEY"])
	}
}

func TestLoad_InvalidPermissions(t *testing.T) {
	tempDir := t.TempDir()
	vaultPath := filepath.Join(tempDir, "test.vault")

	keys := map[string]string{
		"API_KEY": "testval123",
	}

	vault, _ := Create("password", keys)
	vault.Save(vaultPath)

	os.Chmod(vaultPath, 0644)

	_, err := Load(vaultPath)
	if err == nil {
		t.Error("Expected error for invalid permissions")
	}
	if !os.IsPermission(err) && err != ErrInvalidPermissions {
		t.Logf("Got expected permission error: %v", err)
	}
}

func TestLoad_NonexistentFile(t *testing.T) {
	_, err := Load("/nonexistent/path/vault")
	if err == nil {
		t.Error("Expected error for nonexistent file")
	}
}

func TestLoad_CorruptedFile(t *testing.T) {
	tempDir := t.TempDir()
	vaultPath := filepath.Join(tempDir, "test.vault")

	os.WriteFile(vaultPath, []byte("corrupted data"), 0600)

	_, err := Load(vaultPath)
	if err == nil {
		t.Error("Expected error for corrupted file")
	}
}

func TestVault_ConcurrentAccess(t *testing.T) {
	tempDir := t.TempDir()
	vaultPath := filepath.Join(tempDir, "test.vault")

	keys := map[string]string{
		"KEY1": "value1",
	}

	vault, _ := Create("password", keys)
	vault.Save(vaultPath)

	done := make(chan bool, 10)
	for i := 0; i < 10; i++ {
		go func() {
			v, err := Load(vaultPath)
			if err != nil {
				t.Errorf("Load failed: %v", err)
			}
			_, err = v.Decrypt("password")
			if err != nil {
				t.Errorf("Decrypt failed: %v", err)
			}
			done <- true
		}()
	}

	for i := 0; i < 10; i++ {
		<-done
	}
}

func BenchmarkCreate(b *testing.B) {
	keys := map[string]string{
		"API_KEY": "testval123",
	}

	for i := 0; i < b.N; i++ {
		_, err := Create("password", keys)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkDecrypt(b *testing.B) {
	keys := map[string]string{
		"API_KEY": "testval123",
	}

	vault, _ := Create("password", keys)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := vault.Decrypt("password")
		if err != nil {
			b.Fatal(err)
		}
	}
}
