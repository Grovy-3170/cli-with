package main

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"testing"
	"time"

	"github.com/Grovy-3170/cli-with/internal/commands"
	"github.com/Grovy-3170/cli-with/internal/storage"
)

// Test mocks for password and confirmation input

type testPasswordReader struct {
	input []string
	index int
}

func (t *testPasswordReader) ReadPassword(prompt string) (string, error) {
	if t.index >= len(t.input) {
		return "", fmt.Errorf("no more password inputs")
	}
	password := t.input[t.index]
	t.index++
	return password, nil
}

type testConfirmationReader struct {
	input []string
	index int
}

func (t *testConfirmationReader) ReadConfirmation(prompt string) (string, error) {
	if t.index >= len(t.input) {
		return "", fmt.Errorf("no more confirmation inputs")
	}
	response := t.input[t.index]
	t.index++
	return response, nil
}

func TestInitCommand_CreatesVault(t *testing.T) {
	tempDir := t.TempDir()
	vaultDir := filepath.Join(tempDir, "vaults")

	origVaultDir := os.Getenv("WITH_VAULT_DIR")
	os.Setenv("WITH_VAULT_DIR", vaultDir)
	defer os.Setenv("WITH_VAULT_DIR", origVaultDir)

	passwordInput = &testPasswordReader{
		input: []string{"testpassword", "testpassword"},
	}

	rootCmd := newRootCmd()
	rootCmd.SetArgs([]string{"init", "--user", "testuser"})
	err := rootCmd.Execute()
	if err != nil {
		t.Fatalf("Init command failed: %v", err)
	}

	vaultPath := filepath.Join(vaultDir, "testuser.vault")
	if _, err := os.Stat(vaultPath); os.IsNotExist(err) {
		t.Error("Vault file was not created")
	}
}

func TestInitCommand_VaultPermissions(t *testing.T) {
	tempDir := t.TempDir()
	vaultDir := filepath.Join(tempDir, "vaults")

	origVaultDir := os.Getenv("WITH_VAULT_DIR")
	os.Setenv("WITH_VAULT_DIR", vaultDir)
	defer os.Setenv("WITH_VAULT_DIR", origVaultDir)

	passwordInput = &testPasswordReader{
		input: []string{"testpassword", "testpassword"},
	}

	rootCmd := newRootCmd()
	rootCmd.SetArgs([]string{"init", "--user", "testuser"})
	err := rootCmd.Execute()
	if err != nil {
		t.Fatalf("Init command failed: %v", err)
	}

	vaultPath := filepath.Join(vaultDir, "testuser.vault")
	info, err := os.Stat(vaultPath)
	if err != nil {
		t.Fatalf("Failed to stat vault file: %v", err)
	}

	if info.Mode().Perm() != 0600 {
		t.Errorf("Expected vault permissions 0600, got %o", info.Mode().Perm())
	}
}

func TestInitCommand_DirectoryPermissions(t *testing.T) {
	tempDir := t.TempDir()
	vaultDir := filepath.Join(tempDir, "vaults")

	origVaultDir := os.Getenv("WITH_VAULT_DIR")
	os.Setenv("WITH_VAULT_DIR", vaultDir)
	defer os.Setenv("WITH_VAULT_DIR", origVaultDir)

	passwordInput = &testPasswordReader{
		input: []string{"testpassword", "testpassword"},
	}

	rootCmd := newRootCmd()
	rootCmd.SetArgs([]string{"init", "--user", "testuser"})
	err := rootCmd.Execute()
	if err != nil {
		t.Fatalf("Init command failed: %v", err)
	}

	info, err := os.Stat(vaultDir)
	if err != nil {
		t.Fatalf("Failed to stat vault directory: %v", err)
	}

	if info.Mode().Perm() != 0700 {
		t.Errorf("Expected directory permissions 0700, got %o", info.Mode().Perm())
	}
}

func TestInitCommand_VaultAlreadyExists(t *testing.T) {
	tempDir := t.TempDir()
	vaultDir := filepath.Join(tempDir, "vaults")

	origVaultDir := os.Getenv("WITH_VAULT_DIR")
	os.Setenv("WITH_VAULT_DIR", vaultDir)
	defer os.Setenv("WITH_VAULT_DIR", origVaultDir)

	if err := os.MkdirAll(vaultDir, 0700); err != nil {
		t.Fatalf("Failed to create vault directory: %v", err)
	}

	vaultPath := filepath.Join(vaultDir, "existinguser.vault")
	vault, err := storage.Create("oldpassword", nil)
	if err != nil {
		t.Fatalf("Failed to create existing vault: %v", err)
	}
	if err := vault.Save(vaultPath); err != nil {
		t.Fatalf("Failed to save existing vault: %v", err)
	}

	passwordInput = &testPasswordReader{
		input: []string{"testpassword", "testpassword"},
	}

	rootCmd := newRootCmd()
	rootCmd.SetArgs([]string{"init", "--user", "existinguser"})
	err = rootCmd.Execute()
	if err == nil {
		t.Error("Expected error when vault already exists")
	}

	if !strings.Contains(err.Error(), "already exists") {
		t.Errorf("Expected error message about vault already existing, got: %s", err.Error())
	}
}

func TestInitCommand_PasswordMismatch(t *testing.T) {
	tempDir := t.TempDir()
	vaultDir := filepath.Join(tempDir, "vaults")

	origVaultDir := os.Getenv("WITH_VAULT_DIR")
	os.Setenv("WITH_VAULT_DIR", vaultDir)
	defer os.Setenv("WITH_VAULT_DIR", origVaultDir)

	passwordInput = &testPasswordReader{
		input: []string{"password1", "password2"},
	}

	rootCmd := newRootCmd()
	rootCmd.SetArgs([]string{"init", "--user", "testuser"})
	err := rootCmd.Execute()
	if err == nil {
		t.Error("Expected error when passwords don't match")
	}

	if !strings.Contains(err.Error(), "passwords do not match") {
		t.Errorf("Expected error message about password mismatch, got: %s", err.Error())
	}
}

func TestInitCommand_EmptyPassword(t *testing.T) {
	tempDir := t.TempDir()
	vaultDir := filepath.Join(tempDir, "vaults")

	origVaultDir := os.Getenv("WITH_VAULT_DIR")
	os.Setenv("WITH_VAULT_DIR", vaultDir)
	defer os.Setenv("WITH_VAULT_DIR", origVaultDir)

	passwordInput = &testPasswordReader{
		input: []string{"", ""},
	}

	rootCmd := newRootCmd()
	rootCmd.SetArgs([]string{"init", "--user", "testuser"})
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("Init command failed with empty password: %v", err)
	}

	vaultPath := filepath.Join(vaultDir, "testuser.vault")
	if _, err := os.Stat(vaultPath); os.IsNotExist(err) {
		t.Error("Vault file was not created with empty password")
	}
}

func TestListCommand_ShowsKeyNames(t *testing.T) {
	tempDir := t.TempDir()
	vaultDir := filepath.Join(tempDir, "vaults")

	origVaultDir := os.Getenv("WITH_VAULT_DIR")
	os.Setenv("WITH_VAULT_DIR", vaultDir)
	defer os.Setenv("WITH_VAULT_DIR", origVaultDir)

	if err := os.MkdirAll(vaultDir, 0700); err != nil {
		t.Fatalf("Failed to create vault directory: %v", err)
	}

	testKeys := map[string]string{
		"API_KEY":      "test-key-1",
		"DATABASE_URL": "db-host:5432",
		"JWT_SECRET":   "test-jwt-token",
	}

	vault, err := storage.Create("testpassword", testKeys)
	if err != nil {
		t.Fatalf("Failed to create vault: %v", err)
	}

	vaultPath := filepath.Join(vaultDir, "listuser.vault")
	if err := vault.Save(vaultPath); err != nil {
		t.Fatalf("Failed to save vault: %v", err)
	}

	passwordInput = &testPasswordReader{
		input: []string{"testpassword"},
	}

	rootCmd := newRootCmd()
	rootCmd.SetArgs([]string{"list", "--user", "listuser"})
	err = rootCmd.Execute()
	if err != nil {
		t.Fatalf("List command failed: %v", err)
	}
}

func TestListCommand_EmptyVault(t *testing.T) {
	tempDir := t.TempDir()
	vaultDir := filepath.Join(tempDir, "vaults")

	origVaultDir := os.Getenv("WITH_VAULT_DIR")
	os.Setenv("WITH_VAULT_DIR", vaultDir)
	defer os.Setenv("WITH_VAULT_DIR", origVaultDir)

	if err := os.MkdirAll(vaultDir, 0700); err != nil {
		t.Fatalf("Failed to create vault directory: %v", err)
	}

	vault, err := storage.Create("testpassword", nil)
	if err != nil {
		t.Fatalf("Failed to create vault: %v", err)
	}

	vaultPath := filepath.Join(vaultDir, "emptyuser.vault")
	if err := vault.Save(vaultPath); err != nil {
		t.Fatalf("Failed to save vault: %v", err)
	}

	passwordInput = &testPasswordReader{
		input: []string{"testpassword"},
	}

	rootCmd := newRootCmd()
	rootCmd.SetArgs([]string{"list", "--user", "emptyuser"})
	err = rootCmd.Execute()
	if err != nil {
		t.Fatalf("List command failed: %v", err)
	}
}

func TestGetCommand_OutputsCorrectValue(t *testing.T) {
	tempDir := t.TempDir()
	vaultDir := filepath.Join(tempDir, "vaults")

	origVaultDir := os.Getenv("WITH_VAULT_DIR")
	os.Setenv("WITH_VAULT_DIR", vaultDir)
	defer os.Setenv("WITH_VAULT_DIR", origVaultDir)

	if err := os.MkdirAll(vaultDir, 0700); err != nil {
		t.Fatalf("Failed to create vault directory: %v", err)
	}

	testKeys := map[string]string{
		"MY_KEY":    "my-test-value",
		"OTHER_KEY": "other-value",
	}

	vault, err := storage.Create("testpassword", testKeys)
	if err != nil {
		t.Fatalf("Failed to create vault: %v", err)
	}

	vaultPath := filepath.Join(vaultDir, "getuser.vault")
	if err := vault.Save(vaultPath); err != nil {
		t.Fatalf("Failed to save vault: %v", err)
	}

	passwordInput = &testPasswordReader{
		input: []string{"testpassword"},
	}

	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	rootCmd := newRootCmd()
	rootCmd.SetArgs([]string{"get", "--user", "getuser", "MY_KEY"})
	err = rootCmd.Execute()
	if err != nil {
		os.Stdout = oldStdout
		t.Fatalf("Get command failed: %v", err)
	}

	w.Close()
	var buf bytes.Buffer
	buf.ReadFrom(r)
	os.Stdout = oldStdout

	output := strings.TrimSpace(buf.String())
	if output != "my-test-value" {
		t.Errorf("Expected output 'my-test-value', got '%s'", output)
	}
}

func TestGetCommand_KeyNotFound(t *testing.T) {
	tempDir := t.TempDir()
	vaultDir := filepath.Join(tempDir, "vaults")

	origVaultDir := os.Getenv("WITH_VAULT_DIR")
	os.Setenv("WITH_VAULT_DIR", vaultDir)
	defer os.Setenv("WITH_VAULT_DIR", origVaultDir)

	if err := os.MkdirAll(vaultDir, 0700); err != nil {
		t.Fatalf("Failed to create vault directory: %v", err)
	}

	testKeys := map[string]string{
		"EXISTING_KEY": "some-value",
	}

	vault, err := storage.Create("testpassword", testKeys)
	if err != nil {
		t.Fatalf("Failed to create vault: %v", err)
	}

	vaultPath := filepath.Join(vaultDir, "getuser2.vault")
	if err := vault.Save(vaultPath); err != nil {
		t.Fatalf("Failed to save vault: %v", err)
	}

	passwordInput = &testPasswordReader{
		input: []string{"testpassword"},
	}

	rootCmd := newRootCmd()
	rootCmd.SetArgs([]string{"get", "--user", "getuser2", "NONEXISTENT_KEY"})
	err = rootCmd.Execute()
	if err == nil {
		t.Error("Expected error when key not found")
	}

	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("Expected error message about key not found, got: %s", err.Error())
	}
}

func TestRemoveCommand_RemovesKey(t *testing.T) {
	tempDir := t.TempDir()
	vaultDir := filepath.Join(tempDir, "vaults")

	origVaultDir := os.Getenv("WITH_VAULT_DIR")
	os.Setenv("WITH_VAULT_DIR", vaultDir)
	defer os.Setenv("WITH_VAULT_DIR", origVaultDir)

	if err := os.MkdirAll(vaultDir, 0700); err != nil {
		t.Fatalf("Failed to create vault directory: %v", err)
	}

	testKeys := map[string]string{
		"API_KEY":   "test-key-1",
		"OTHER_KEY": "test-key-2",
	}

	vault, err := storage.Create("testpassword", testKeys)
	if err != nil {
		t.Fatalf("Failed to create vault: %v", err)
	}

	vaultPath := filepath.Join(vaultDir, "removeuser.vault")
	if err := vault.Save(vaultPath); err != nil {
		t.Fatalf("Failed to save vault: %v", err)
	}

	passwordInput = &testPasswordReader{
		input: []string{"testpassword"},
	}

	rootCmd := newRootCmd()
	rootCmd.SetArgs([]string{"remove", "--user", "removeuser", "API_KEY"})
	err = rootCmd.Execute()
	if err != nil {
		t.Fatalf("Remove command failed: %v", err)
	}

	vault, err = storage.Load(vaultPath)
	if err != nil {
		t.Fatalf("Failed to load vault after removal: %v", err)
	}

	keys, err := vault.Decrypt("testpassword")
	if err != nil {
		t.Fatalf("Failed to decrypt vault: %v", err)
	}

	if _, exists := keys["API_KEY"]; exists {
		t.Error("Key 'API_KEY' should have been removed")
	}

	if _, exists := keys["OTHER_KEY"]; !exists {
		t.Error("Key 'OTHER_KEY' should still exist")
	}
}

func TestRemoveCommand_KeyNotFound(t *testing.T) {
	tempDir := t.TempDir()
	vaultDir := filepath.Join(tempDir, "vaults")

	origVaultDir := os.Getenv("WITH_VAULT_DIR")
	os.Setenv("WITH_VAULT_DIR", vaultDir)
	defer os.Setenv("WITH_VAULT_DIR", origVaultDir)

	if err := os.MkdirAll(vaultDir, 0700); err != nil {
		t.Fatalf("Failed to create vault directory: %v", err)
	}

	testKeys := map[string]string{
		"EXISTING_KEY": "secret",
	}

	vault, err := storage.Create("testpassword", testKeys)
	if err != nil {
		t.Fatalf("Failed to create vault: %v", err)
	}

	vaultPath := filepath.Join(vaultDir, "removeuser2.vault")
	if err := vault.Save(vaultPath); err != nil {
		t.Fatalf("Failed to save vault: %v", err)
	}

	passwordInput = &testPasswordReader{
		input: []string{"testpassword"},
	}

	rootCmd := newRootCmd()
	rootCmd.SetArgs([]string{"remove", "--user", "removeuser2", "NONEXISTENT_KEY"})
	err = rootCmd.Execute()
	if err == nil {
		t.Error("Expected error when key not found")
	}

	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("Expected error message about key not found, got: %s", err.Error())
	}
}

func TestRemoveCommand_DeletesVaultWithConfirmation(t *testing.T) {
	tempDir := t.TempDir()
	vaultDir := filepath.Join(tempDir, "vaults")

	origVaultDir := os.Getenv("WITH_VAULT_DIR")
	os.Setenv("WITH_VAULT_DIR", vaultDir)
	defer os.Setenv("WITH_VAULT_DIR", origVaultDir)

	if err := os.MkdirAll(vaultDir, 0700); err != nil {
		t.Fatalf("Failed to create vault directory: %v", err)
	}

	testKeys := map[string]string{
		"API_KEY": "secret",
	}

	vault, err := storage.Create("testpassword", testKeys)
	if err != nil {
		t.Fatalf("Failed to create vault: %v", err)
	}

	vaultPath := filepath.Join(vaultDir, "deleteuser.vault")
	if err := vault.Save(vaultPath); err != nil {
		t.Fatalf("Failed to save vault: %v", err)
	}

	confirmationInput = &testConfirmationReader{
		input: []string{"yes"},
	}

	rootCmd := newRootCmd()
	rootCmd.SetArgs([]string{"remove", "--user", "deleteuser"})
	err = rootCmd.Execute()
	if err != nil {
		t.Fatalf("Remove command failed: %v", err)
	}

	if _, err := os.Stat(vaultPath); !os.IsNotExist(err) {
		t.Error("Vault file should have been deleted")
	}
}

func TestRemoveCommand_CancelsVaultDeletion(t *testing.T) {
	tempDir := t.TempDir()
	vaultDir := filepath.Join(tempDir, "vaults")

	origVaultDir := os.Getenv("WITH_VAULT_DIR")
	os.Setenv("WITH_VAULT_DIR", vaultDir)
	defer os.Setenv("WITH_VAULT_DIR", origVaultDir)

	if err := os.MkdirAll(vaultDir, 0700); err != nil {
		t.Fatalf("Failed to create vault directory: %v", err)
	}

	testKeys := map[string]string{
		"API_KEY": "secret",
	}

	vault, err := storage.Create("testpassword", testKeys)
	if err != nil {
		t.Fatalf("Failed to create vault: %v", err)
	}

	vaultPath := filepath.Join(vaultDir, "canceluser.vault")
	if err := vault.Save(vaultPath); err != nil {
		t.Fatalf("Failed to save vault: %v", err)
	}

	confirmationInput = &testConfirmationReader{
		input: []string{"no"},
	}

	rootCmd := newRootCmd()
	rootCmd.SetArgs([]string{"remove", "--user", "canceluser"})
	err = rootCmd.Execute()
	if err != nil {
		t.Fatalf("Remove command failed: %v", err)
	}

	if _, err := os.Stat(vaultPath); os.IsNotExist(err) {
		t.Error("Vault file should still exist after cancelled deletion")
	}
}

func TestRemoveCommand_VaultNotFound(t *testing.T) {
	tempDir := t.TempDir()
	vaultDir := filepath.Join(tempDir, "vaults")

	origVaultDir := os.Getenv("WITH_VAULT_DIR")
	os.Setenv("WITH_VAULT_DIR", vaultDir)
	defer os.Setenv("WITH_VAULT_DIR", origVaultDir)

	rootCmd := newRootCmd()
	rootCmd.SetArgs([]string{"remove", "--user", "nonexistent"})
	err := rootCmd.Execute()
	if err == nil {
		t.Error("Expected error when vault does not exist")
	}

	if !strings.Contains(err.Error(), "does not exist") {
		t.Errorf("Expected error message about vault not existing, got: %s", err.Error())
	}
}

func TestSetCommand_AddsNewKey(t *testing.T) {
	tempDir := t.TempDir()
	vaultDir := filepath.Join(tempDir, "vaults")

	origVaultDir := os.Getenv("WITH_VAULT_DIR")
	os.Setenv("WITH_VAULT_DIR", vaultDir)
	defer os.Setenv("WITH_VAULT_DIR", origVaultDir)

	if err := os.MkdirAll(vaultDir, 0700); err != nil {
		t.Fatalf("Failed to create vault directory: %v", err)
	}

	vault, err := storage.Create("testpassword", nil)
	if err != nil {
		t.Fatalf("Failed to create vault: %v", err)
	}

	vaultPath := filepath.Join(vaultDir, "setuser.vault")
	if err := vault.Save(vaultPath); err != nil {
		t.Fatalf("Failed to save vault: %v", err)
	}

	passwordInput = &testPasswordReader{
		input: []string{"testpassword", "my-test-value"},
	}

	rootCmd := newRootCmd()
	rootCmd.SetArgs([]string{"set", "--user", "setuser", "MY_KEY"})
	err = rootCmd.Execute()
	if err != nil {
		t.Fatalf("Set command failed: %v", err)
	}

	loadedVault, err := storage.Load(vaultPath)
	if err != nil {
		t.Fatalf("Failed to load vault: %v", err)
	}

	keys, err := loadedVault.Decrypt("testpassword")
	if err != nil {
		t.Fatalf("Failed to decrypt vault: %v", err)
	}

	if keys["MY_KEY"] != "my-test-value" {
		t.Errorf("Expected MY_KEY to be 'my-test-value', got '%s'", keys["MY_KEY"])
	}
}

func TestSetCommand_UpdatesExistingKey(t *testing.T) {
	tempDir := t.TempDir()
	vaultDir := filepath.Join(tempDir, "vaults")

	origVaultDir := os.Getenv("WITH_VAULT_DIR")
	os.Setenv("WITH_VAULT_DIR", vaultDir)
	defer os.Setenv("WITH_VAULT_DIR", origVaultDir)

	if err := os.MkdirAll(vaultDir, 0700); err != nil {
		t.Fatalf("Failed to create vault directory: %v", err)
	}

	testKeys := map[string]string{
		"MY_KEY":    "old-value",
		"OTHER_KEY": "other-value",
	}

	vault, err := storage.Create("testpassword", testKeys)
	if err != nil {
		t.Fatalf("Failed to create vault: %v", err)
	}

	vaultPath := filepath.Join(vaultDir, "setuser2.vault")
	if err := vault.Save(vaultPath); err != nil {
		t.Fatalf("Failed to save vault: %v", err)
	}

	passwordInput = &testPasswordReader{
		input: []string{"testpassword", "new-value"},
	}

	rootCmd := newRootCmd()
	rootCmd.SetArgs([]string{"set", "--user", "setuser2", "MY_KEY"})
	err = rootCmd.Execute()
	if err != nil {
		t.Fatalf("Set command failed: %v", err)
	}

	loadedVault, err := storage.Load(vaultPath)
	if err != nil {
		t.Fatalf("Failed to load vault: %v", err)
	}

	keys, err := loadedVault.Decrypt("testpassword")
	if err != nil {
		t.Fatalf("Failed to decrypt vault: %v", err)
	}

	if keys["MY_KEY"] != "new-value" {
		t.Errorf("Expected MY_KEY to be 'new-value', got '%s'", keys["MY_KEY"])
	}

	if keys["OTHER_KEY"] != "other-value" {
		t.Errorf("Expected OTHER_KEY to remain 'other-value', got '%s'", keys["OTHER_KEY"])
	}
}

func TestSetCommand_InvalidKeyName(t *testing.T) {
	tempDir := t.TempDir()
	vaultDir := filepath.Join(tempDir, "vaults")

	origVaultDir := os.Getenv("WITH_VAULT_DIR")
	os.Setenv("WITH_VAULT_DIR", vaultDir)
	defer os.Setenv("WITH_VAULT_DIR", origVaultDir)

	if err := os.MkdirAll(vaultDir, 0700); err != nil {
		t.Fatalf("Failed to create vault directory: %v", err)
	}

	vault, err := storage.Create("testpassword", nil)
	if err != nil {
		t.Fatalf("Failed to create vault: %v", err)
	}

	vaultPath := filepath.Join(vaultDir, "setuser3.vault")
	if err := vault.Save(vaultPath); err != nil {
		t.Fatalf("Failed to save vault: %v", err)
	}

	passwordInput = &testPasswordReader{
		input: []string{"testpassword"},
	}

	rootCmd := newRootCmd()
	rootCmd.SetArgs([]string{"set", "--user", "setuser3", "123_INVALID"})
	err = rootCmd.Execute()
	if err == nil {
		t.Error("Expected error when key name is invalid")
	}

	if !strings.Contains(err.Error(), "invalid key name") {
		t.Errorf("Expected error message about invalid key name, got: %s", err.Error())
	}
}

func TestSetCommand_VaultDoesNotExist(t *testing.T) {
	tempDir := t.TempDir()
	vaultDir := filepath.Join(tempDir, "vaults")

	origVaultDir := os.Getenv("WITH_VAULT_DIR")
	os.Setenv("WITH_VAULT_DIR", vaultDir)
	defer os.Setenv("WITH_VAULT_DIR", origVaultDir)

	passwordInput = &testPasswordReader{
		input: []string{"testpassword", "my-value"},
	}

	rootCmd := newRootCmd()
	rootCmd.SetArgs([]string{"set", "--user", "nonexistent", "MY_KEY"})
	err := rootCmd.Execute()
	if err == nil {
		t.Error("Expected error when vault does not exist")
	}

	if !strings.Contains(err.Error(), "does not exist") {
		t.Errorf("Expected error message about vault not existing, got: %s", err.Error())
	}
}

func TestSetCommand_InvalidPassword(t *testing.T) {
	tempDir := t.TempDir()
	vaultDir := filepath.Join(tempDir, "vaults")

	origVaultDir := os.Getenv("WITH_VAULT_DIR")
	os.Setenv("WITH_VAULT_DIR", vaultDir)
	defer os.Setenv("WITH_VAULT_DIR", origVaultDir)

	if err := os.MkdirAll(vaultDir, 0700); err != nil {
		t.Fatalf("Failed to create vault directory: %v", err)
	}

	vault, err := storage.Create("correctpassword", nil)
	if err != nil {
		t.Fatalf("Failed to create vault: %v", err)
	}

	vaultPath := filepath.Join(vaultDir, "setuser4.vault")
	if err := vault.Save(vaultPath); err != nil {
		t.Fatalf("Failed to save vault: %v", err)
	}

	passwordInput = &testPasswordReader{
		input: []string{"wrongpassword", "my-value"},
	}

	rootCmd := newRootCmd()
	rootCmd.SetArgs([]string{"set", "--user", "setuser4", "MY_KEY"})
	err = rootCmd.Execute()
	if err == nil {
		t.Error("Expected error when password is wrong")
	}

	if !strings.Contains(err.Error(), "decrypting vault") {
		t.Errorf("Expected error message about decrypting vault, got: %s", err.Error())
	}
}

func TestExecCommand_VaultNotFound(t *testing.T) {
	tempDir := t.TempDir()
	vaultDir := filepath.Join(tempDir, "vaults")

	// Create vault directory so we get "vault does not exist" instead of "no users found"
	if err := os.MkdirAll(vaultDir, 0700); err != nil {
		t.Fatalf("Failed to create vault dir: %v", err)
	}

	origVaultDir := os.Getenv("WITH_VAULT_DIR")
	os.Setenv("WITH_VAULT_DIR", vaultDir)
	defer os.Setenv("WITH_VAULT_DIR", origVaultDir)

	rootCmd := newRootCmd()
	rootCmd.SetArgs([]string{"exec", "--user", "nonexistent", "--", "echo", "hello"})
	err := rootCmd.Execute()
	if err == nil {
		t.Error("Expected error when vault does not exist")
	}

	if !strings.Contains(err.Error(), "does not exist") {
		t.Errorf("Expected error message about vault not existing, got: %s", err.Error())
	}
}

func TestBuildMinimalEnv_ContainsEssentialVars(t *testing.T) {
	originalEnv := os.Environ()
	defer func() {
		os.Clearenv()
		for _, env := range originalEnv {
			parts := strings.SplitN(env, "=", 2)
			if len(parts) == 2 {
				os.Setenv(parts[0], parts[1])
			}
		}
	}()

	os.Clearenv()
	os.Setenv("PATH", "/usr/bin")
	os.Setenv("HOME", "/home/test")
	os.Setenv("USER", "testuser")
	os.Setenv("SHELL", "/bin/bash")
	os.Setenv("TMPDIR", "/tmp")
	os.Setenv("CUSTOM_VAR", "should-not-appear")

	keys := map[string]string{
		"API_KEY":  "testval123",
		"DB_TOKEN": "token456",
	}

	env := commands.BuildMinimalEnv(keys)

	expectedVars := map[string]string{
		"PATH":     "/usr/bin",
		"HOME":     "/home/test",
		"USER":     "testuser",
		"SHELL":    "/bin/bash",
		"TMPDIR":   "/tmp",
		"API_KEY":  "testval123",
		"DB_TOKEN": "token456",
	}

	foundVars := make(map[string]string)
	for _, e := range env {
		parts := strings.SplitN(e, "=", 2)
		if len(parts) == 2 {
			foundVars[parts[0]] = parts[1]
		}
	}

	for key, expected := range expectedVars {
		if foundVars[key] != expected {
			t.Errorf("Expected %s=%s, got %s", key, expected, foundVars[key])
		}
	}

	if _, exists := foundVars["CUSTOM_VAR"]; exists {
		t.Error("CUSTOM_VAR should not be in minimal env")
	}
}

func TestBuildMinimalEnv_HandlesMissingVars(t *testing.T) {
	originalEnv := os.Environ()
	defer func() {
		os.Clearenv()
		for _, env := range originalEnv {
			parts := strings.SplitN(env, "=", 2)
			if len(parts) == 2 {
				os.Setenv(parts[0], parts[1])
			}
		}
	}()

	os.Clearenv()
	os.Setenv("PATH", "/usr/bin")

	keys := map[string]string{
		"SECRET_KEY": "test-val",
	}

	env := commands.BuildMinimalEnv(keys)

	foundVars := make(map[string]string)
	for _, e := range env {
		parts := strings.SplitN(e, "=", 2)
		if len(parts) == 2 {
			foundVars[parts[0]] = parts[1]
		}
	}

	if foundVars["PATH"] != "/usr/bin" {
		t.Errorf("Expected PATH=/usr/bin, got %s", foundVars["PATH"])
	}

	if foundVars["SECRET_KEY"] != "test-val" {
		t.Errorf("Expected SECRET_KEY=test-val, got %s", foundVars["SECRET_KEY"])
	}

	if _, exists := foundVars["HOME"]; exists {
		t.Error("HOME should not be present when not set in environment")
	}
}

func TestExec_Integration_EnvVars(t *testing.T) {
	tempDir := t.TempDir()
	vaultDir := filepath.Join(tempDir, "vaults")

	origVaultDir := os.Getenv("WITH_VAULT_DIR")
	os.Setenv("WITH_VAULT_DIR", vaultDir)
	defer os.Setenv("WITH_VAULT_DIR", origVaultDir)

	if err := os.MkdirAll(vaultDir, 0700); err != nil {
		t.Fatalf("Failed to create vault directory: %v", err)
	}

	testKeys := map[string]string{
		"API_KEY":  "test-secret-123",
		"DB_TOKEN": "db-token-456",
	}

	vault, err := storage.Create("testpassword", testKeys)
	if err != nil {
		t.Fatalf("Failed to create vault: %v", err)
	}

	vaultPath := filepath.Join(vaultDir, "execuser.vault")
	if err := vault.Save(vaultPath); err != nil {
		t.Fatalf("Failed to save vault: %v", err)
	}

	passwordFilePath := filepath.Join(tempDir, "password")
	if err := os.WriteFile(passwordFilePath, []byte("testpassword"), 0600); err != nil {
		t.Fatalf("Failed to write password file: %v", err)
	}

	binaryPath := filepath.Join(tempDir, "auth-test")
	buildCmd := exec.Command("go", "build", "-buildvcs=false", "-o", binaryPath, ".")
	buildCmd.Dir = "."
	if output, err := buildCmd.CombinedOutput(); err != nil {
		t.Fatalf("Failed to build binary: %v\n%s", err, output)
	}

	cmd := exec.Command(binaryPath, "--user", "execuser", "--password-file", passwordFilePath, "exec", "--", "env")
	cmd.Env = append(os.Environ(), "WITH_VAULT_DIR="+vaultDir)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		t.Fatalf("Command failed: %v\nStdout: %s\nStderr: %s", err, stdout.String(), stderr.String())
	}

	output := stdout.String()
	if !strings.Contains(output, "API_KEY=test-secret-123") {
		t.Errorf("Expected API_KEY in output, got: %s", output)
	}
	if !strings.Contains(output, "DB_TOKEN=db-token-456") {
		t.Errorf("Expected DB_TOKEN in output, got: %s", output)
	}
}

func TestExec_Integration_ExitCode(t *testing.T) {
	tempDir := t.TempDir()
	vaultDir := filepath.Join(tempDir, "vaults")

	origVaultDir := os.Getenv("WITH_VAULT_DIR")
	os.Setenv("WITH_VAULT_DIR", vaultDir)
	defer os.Setenv("WITH_VAULT_DIR", origVaultDir)

	if err := os.MkdirAll(vaultDir, 0700); err != nil {
		t.Fatalf("Failed to create vault directory: %v", err)
	}

	vault, err := storage.Create("testpassword", map[string]string{"KEY": "value"})
	if err != nil {
		t.Fatalf("Failed to create vault: %v", err)
	}

	vaultPath := filepath.Join(vaultDir, "exituser.vault")
	if err := vault.Save(vaultPath); err != nil {
		t.Fatalf("Failed to save vault: %v", err)
	}

	passwordFilePath := filepath.Join(tempDir, "password")
	if err := os.WriteFile(passwordFilePath, []byte("testpassword"), 0600); err != nil {
		t.Fatalf("Failed to write password file: %v", err)
	}

	binaryPath := filepath.Join(tempDir, "auth-test")
	buildCmd := exec.Command("go", "build", "-buildvcs=false", "-o", binaryPath, ".")
	buildCmd.Dir = "."
	if output, err := buildCmd.CombinedOutput(); err != nil {
		t.Fatalf("Failed to build binary: %v\n%s", err, output)
	}

	cmd := exec.Command(binaryPath, "--user", "exituser", "--password-file", passwordFilePath, "exec", "--", "sh", "-c", "exit 42")
	cmd.Env = append(os.Environ(), "WITH_VAULT_DIR="+vaultDir)

	err = cmd.Run()

	if err == nil {
		t.Fatal("Expected non-zero exit code, got 0")
	}

	exitErr, ok := err.(*exec.ExitError)
	if !ok {
		t.Fatalf("Expected ExitError, got: %v", err)
	}

	if exitErr.ExitCode() != 42 {
		t.Errorf("Expected exit code 42, got %d", exitErr.ExitCode())
	}
}

func TestExec_Integration_MinimalEnv(t *testing.T) {
	tempDir := t.TempDir()
	vaultDir := filepath.Join(tempDir, "vaults")

	origVaultDir := os.Getenv("WITH_VAULT_DIR")
	os.Setenv("WITH_VAULT_DIR", vaultDir)
	defer os.Setenv("WITH_VAULT_DIR", origVaultDir)

	if err := os.MkdirAll(vaultDir, 0700); err != nil {
		t.Fatalf("Failed to create vault directory: %v", err)
	}

	vault, err := storage.Create("testpassword", map[string]string{"SECRET_KEY": "test-val"})
	if err != nil {
		t.Fatalf("Failed to create vault: %v", err)
	}

	vaultPath := filepath.Join(vaultDir, "envuser.vault")
	if err := vault.Save(vaultPath); err != nil {
		t.Fatalf("Failed to save vault: %v", err)
	}

	passwordFilePath := filepath.Join(tempDir, "password")
	if err := os.WriteFile(passwordFilePath, []byte("testpassword"), 0600); err != nil {
		t.Fatalf("Failed to write password file: %v", err)
	}

	binaryPath := filepath.Join(tempDir, "auth-test")
	buildCmd := exec.Command("go", "build", "-buildvcs=false", "-o", binaryPath, ".")
	buildCmd.Dir = "."
	if output, err := buildCmd.CombinedOutput(); err != nil {
		t.Fatalf("Failed to build binary: %v\n%s", err, output)
	}

	cmd := exec.Command(binaryPath, "--user", "envuser", "--password-file", passwordFilePath, "exec", "--", "env")
	cmd.Env = append(os.Environ(),
		"WITH_VAULT_DIR="+vaultDir,
		"LEAKED_VAR=should-not-appear",
	)

	var stdout bytes.Buffer
	cmd.Stdout = &stdout

	err = cmd.Run()
	if err != nil {
		t.Fatalf("Command failed: %v", err)
	}

	output := stdout.String()
	hasSecretKey := strings.Contains(output, "SECRET_KEY=test-val")
	hasLeakedVar := strings.Contains(output, "LEAKED_VAR=should-not-appear")

	if !hasSecretKey {
		t.Error("Expected SECRET_KEY to be in environment")
	}
	if hasLeakedVar {
		t.Error("LEAKED_VAR should NOT be in environment (minimal env inheritance)")
	}
}

func TestExec_Integration_SignalForwarding(t *testing.T) {
	tempDir := t.TempDir()
	vaultDir := filepath.Join(tempDir, "vaults")

	origVaultDir := os.Getenv("WITH_VAULT_DIR")
	os.Setenv("WITH_VAULT_DIR", vaultDir)
	defer os.Setenv("WITH_VAULT_DIR", origVaultDir)

	if err := os.MkdirAll(vaultDir, 0700); err != nil {
		t.Fatalf("Failed to create vault directory: %v", err)
	}

	vault, err := storage.Create("testpassword", map[string]string{"KEY": "value"})
	if err != nil {
		t.Fatalf("Failed to create vault: %v", err)
	}

	vaultPath := filepath.Join(vaultDir, "siguser.vault")
	if err := vault.Save(vaultPath); err != nil {
		t.Fatalf("Failed to save vault: %v", err)
	}

	passwordFilePath := filepath.Join(tempDir, "password")
	if err := os.WriteFile(passwordFilePath, []byte("testpassword"), 0600); err != nil {
		t.Fatalf("Failed to write password file: %v", err)
	}

	binaryPath := filepath.Join(tempDir, "auth-test")
	buildCmd := exec.Command("go", "build", "-buildvcs=false", "-o", binaryPath, ".")
	buildCmd.Dir = "."
	if output, err := buildCmd.CombinedOutput(); err != nil {
		t.Fatalf("Failed to build binary: %v\n%s", err, output)
	}

	cmd := exec.Command(binaryPath, "--user", "siguser", "--password-file", passwordFilePath, "exec", "--", "sleep", "10")
	cmd.Env = append(os.Environ(), "WITH_VAULT_DIR="+vaultDir)

	if err := cmd.Start(); err != nil {
		t.Fatalf("Failed to start command: %v", err)
	}

	time.Sleep(200 * time.Millisecond)

	cmd.Process.Signal(syscall.SIGTERM)

	err = cmd.Wait()

	if err == nil {
		t.Fatal("Expected non-zero exit code from signal")
	}

	exitErr, ok := err.(*exec.ExitError)
	if !ok {
		t.Fatalf("Expected ExitError, got: %v", err)
	}

	if exitErr.ExitCode() == 0 {
		t.Error("Expected non-zero exit code after SIGTERM")
	}
}
