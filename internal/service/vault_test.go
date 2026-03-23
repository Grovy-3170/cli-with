package service

import (
	"os"
	"path/filepath"
	"testing"
)

func setupTestVaultDir(t *testing.T) (string, func()) {
	t.Helper()
	tempDir := t.TempDir()
	vaultDir := filepath.Join(tempDir, "vaults")

	origVaultDir := os.Getenv("WITH_VAULT_DIR")
	os.Setenv("WITH_VAULT_DIR", vaultDir)

	cleanup := func() {
		os.Setenv("WITH_VAULT_DIR", origVaultDir)
	}

	return vaultDir, cleanup
}

func TestNewVaultService(t *testing.T) {
	_, cleanup := setupTestVaultDir(t)
	defer cleanup()

	svc, err := NewVaultService()
	if err != nil {
		t.Fatalf("NewVaultService failed: %v", err)
	}

	if svc == nil {
		t.Fatal("Expected non-nil service")
	}

	if svc.config == nil {
		t.Fatal("Expected non-nil config")
	}
}

func TestVaultService_CreateVault(t *testing.T) {
	vaultDir, cleanup := setupTestVaultDir(t)
	defer cleanup()

	svc, err := NewVaultService()
	if err != nil {
		t.Fatalf("NewVaultService failed: %v", err)
	}

	err = svc.CreateVault("testuser", "testpassword")
	if err != nil {
		t.Fatalf("CreateVault failed: %v", err)
	}

	// Verify vault file exists
	vaultPath := filepath.Join(vaultDir, "testuser.vault")
	if _, err := os.Stat(vaultPath); os.IsNotExist(err) {
		t.Error("Vault file was not created")
	}

	// Verify VaultExists returns true
	if !svc.VaultExists("testuser") {
		t.Error("VaultExists should return true after CreateVault")
	}
}

func TestVaultService_CreateVault_AlreadyExists(t *testing.T) {
	_, cleanup := setupTestVaultDir(t)
	defer cleanup()

	svc, err := NewVaultService()
	if err != nil {
		t.Fatalf("NewVaultService failed: %v", err)
	}

	// Create first vault
	err = svc.CreateVault("testuser", "testpassword")
	if err != nil {
		t.Fatalf("First CreateVault failed: %v", err)
	}

	// Try to create again - should fail
	err = svc.CreateVault("testuser", "testpassword")
	if err == nil {
		t.Error("Expected error when vault already exists")
	}
}

func TestVaultService_LoadAndDecrypt(t *testing.T) {
	_, cleanup := setupTestVaultDir(t)
	defer cleanup()

	svc, err := NewVaultService()
	if err != nil {
		t.Fatalf("NewVaultService failed: %v", err)
	}

	// Create vault first
	err = svc.CreateVault("testuser", "testpassword")
	if err != nil {
		t.Fatalf("CreateVault failed: %v", err)
	}

	// Load and decrypt
	keys, vault, err := svc.LoadAndDecrypt("testuser", "testpassword")
	if err != nil {
		t.Fatalf("LoadAndDecrypt failed: %v", err)
	}

	if vault == nil {
		t.Error("Expected non-nil vault")
	}

	if keys == nil {
		t.Error("Expected non-nil keys map")
	}
}

func TestVaultService_LoadAndDecrypt_WrongPassword(t *testing.T) {
	_, cleanup := setupTestVaultDir(t)
	defer cleanup()

	svc, err := NewVaultService()
	if err != nil {
		t.Fatalf("NewVaultService failed: %v", err)
	}

	err = svc.CreateVault("testuser", "correctpassword")
	if err != nil {
		t.Fatalf("CreateVault failed: %v", err)
	}

	_, _, err = svc.LoadAndDecrypt("testuser", "wrongpassword")
	if err == nil {
		t.Error("Expected error with wrong password")
	}
}

func TestVaultService_LoadAndDecrypt_VaultNotFound(t *testing.T) {
	_, cleanup := setupTestVaultDir(t)
	defer cleanup()

	svc, err := NewVaultService()
	if err != nil {
		t.Fatalf("NewVaultService failed: %v", err)
	}

	_, _, err = svc.LoadAndDecrypt("nonexistent", "password")
	if err == nil {
		t.Error("Expected error when vault doesn't exist")
	}
}

func TestVaultService_UpdateAndSave(t *testing.T) {
	_, cleanup := setupTestVaultDir(t)
	defer cleanup()

	svc, err := NewVaultService()
	if err != nil {
		t.Fatalf("NewVaultService failed: %v", err)
	}

	err = svc.CreateVault("testuser", "testpassword")
	if err != nil {
		t.Fatalf("CreateVault failed: %v", err)
	}

	keys, vault, err := svc.LoadAndDecrypt("testuser", "testpassword")
	if err != nil {
		t.Fatalf("LoadAndDecrypt failed: %v", err)
	}

	// Add a key
	keys["API_KEY"] = "testval123"

	err = svc.UpdateAndSave("testuser", "testpassword", vault, keys)
	if err != nil {
		t.Fatalf("UpdateAndSave failed: %v", err)
	}

	// Reload and verify
	keys2, _, err := svc.LoadAndDecrypt("testuser", "testpassword")
	if err != nil {
		t.Fatalf("Second LoadAndDecrypt failed: %v", err)
	}

	if keys2["API_KEY"] != "testval123" {
		t.Errorf("Expected API_KEY=testval123, got %s", keys2["API_KEY"])
	}
}

func TestVaultService_DeleteVault(t *testing.T) {
	vaultDir, cleanup := setupTestVaultDir(t)
	defer cleanup()

	svc, err := NewVaultService()
	if err != nil {
		t.Fatalf("NewVaultService failed: %v", err)
	}

	err = svc.CreateVault("testuser", "testpassword")
	if err != nil {
		t.Fatalf("CreateVault failed: %v", err)
	}

	// Verify exists
	vaultPath := filepath.Join(vaultDir, "testuser.vault")
	if _, err := os.Stat(vaultPath); os.IsNotExist(err) {
		t.Fatal("Vault should exist before delete")
	}

	// Delete
	err = svc.DeleteVault("testuser")
	if err != nil {
		t.Fatalf("DeleteVault failed: %v", err)
	}

	// Verify gone
	if _, err := os.Stat(vaultPath); !os.IsNotExist(err) {
		t.Error("Vault file should be deleted")
	}

	if svc.VaultExists("testuser") {
		t.Error("VaultExists should return false after delete")
	}
}

func TestVaultService_GetVaultPath(t *testing.T) {
	vaultDir, cleanup := setupTestVaultDir(t)
	defer cleanup()

	svc, err := NewVaultService()
	if err != nil {
		t.Fatalf("NewVaultService failed: %v", err)
	}

	path := svc.GetVaultPath("alice")
	expected := filepath.Join(vaultDir, "alice.vault")

	if path != expected {
		t.Errorf("Expected %s, got %s", expected, path)
	}
}

func TestVaultService_GetVaultDir(t *testing.T) {
	vaultDir, cleanup := setupTestVaultDir(t)
	defer cleanup()

	svc, err := NewVaultService()
	if err != nil {
		t.Fatalf("NewVaultService failed: %v", err)
	}

	if svc.GetVaultDir() != vaultDir {
		t.Errorf("Expected %s, got %s", vaultDir, svc.GetVaultDir())
	}
}
