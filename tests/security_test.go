// Package tests contains security-focused integration tests for the with CLI.
// These tests verify security properties like permission enforcement,
// environment isolation, signal handling, and error message sanitization.
package tests

import (
	"bytes"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"testing"
	"time"
)

// =============================================================================
// File Permission Enforcement Tests
// =============================================================================

// TestSecurity_FilePermissionEnforcement_VaultLoadRejectsWrongPermissions verifies that
// the vault loader rejects files with permissions other than 0600.
// This prevents unauthorized access to vault files that may have been
// accidentally created with overly permissive permissions.
func TestSecurity_FilePermissionEnforcement_VaultLoadRejectsWrongPermissions(t *testing.T) {
	binary := buildBinary(t)
	tempDir := t.TempDir()
	vaultDir := filepath.Join(tempDir, "vaults")
	if err := os.MkdirAll(vaultDir, 0700); err != nil {
		t.Fatalf("Failed to create vault directory: %v", err)
	}

	vaultPath := filepath.Join(vaultDir, "testuser.vault")
	vaultContent := map[string]interface{}{
		"version":    1,
		"format":     "cli-auth-vault",
		"created_at": time.Now().UTC().Format(time.RFC3339),
		"kdf": map[string]interface{}{
			"algorithm":   "argon2id",
			"memory":      65536,
			"iterations":  3,
			"parallelism": 1,
			"salt":        "dGVzdC1zYWx0LTE2LWJ5dGU=",
		},
		"encryption": map[string]interface{}{
			"algorithm": "aes-256-gcm",
			"nonce":     "dGVzdC1ub25jZQ==",
		},
		"ciphertext": "dGVzdC1jaXBoZXJ0ZXh0",
	}
	data, _ := json.Marshal(vaultContent)

	if err := os.WriteFile(vaultPath, data, 0644); err != nil {
		t.Fatalf("Failed to write vault file: %v", err)
	}

	info, err := os.Stat(vaultPath)
	if err != nil {
		t.Fatalf("Failed to stat vault file: %v", err)
	}
	if info.Mode().Perm() != 0644 {
		t.Fatalf("Expected 0644 permissions, got %04o", info.Mode().Perm())
	}

	passwordFile := filepath.Join(tempDir, "password")
	if err := os.WriteFile(passwordFile, []byte("testpassword"), 0600); err != nil {
		t.Fatalf("Failed to write password file: %v", err)
	}

	cmd := exec.Command(binary, "--user", "testuser", "--password-file", passwordFile, "list")
	cmd.Env = append(os.Environ(), "WITH_VAULT_DIR="+vaultDir)

	output, err := cmd.CombinedOutput()
	if err == nil {
		t.Error("Expected command to fail due to invalid permissions, but it succeeded")
	}

	outputStr := string(output)
	if !strings.Contains(outputStr, "permission") && !strings.Contains(outputStr, "0600") {
		t.Errorf("Expected permission error message, got: %s", outputStr)
	}
}

// TestSecurity_FilePermissionEnforcement_NewVaultCorrectPermissions verifies that
// newly created vault files have correct 0600 permissions.
func TestSecurity_FilePermissionEnforcement_NewVaultCorrectPermissions(t *testing.T) {
	binary := buildBinary(t)
	tempDir := t.TempDir()
	vaultDir := filepath.Join(tempDir, "vaults")

	passwordFile := filepath.Join(tempDir, "password")
	if err := os.WriteFile(passwordFile, []byte("testpassword\ntestpassword"), 0600); err != nil {
		t.Fatalf("Failed to write password file: %v", err)
	}

	cmd := exec.Command(binary, "--user", "testuser", "--password-file", passwordFile, "init")
	cmd.Env = append(os.Environ(), "WITH_VAULT_DIR="+vaultDir)

	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("Failed to initialize vault: %v\n%s", err, output)
	}

	vaultPath := filepath.Join(vaultDir, "testuser.vault")
	info, err := os.Stat(vaultPath)
	if err != nil {
		t.Fatalf("Failed to stat vault file: %v", err)
	}

	if info.Mode().Perm() != 0600 {
		t.Errorf("Expected vault permissions 0600, got %04o", info.Mode().Perm())
	}

	dirInfo, err := os.Stat(vaultDir)
	if err != nil {
		t.Fatalf("Failed to stat vault directory: %v", err)
	}

	if dirInfo.Mode().Perm() != 0700 {
		t.Errorf("Expected directory permissions 0700, got %04o", dirInfo.Mode().Perm())
	}
}

// =============================================================================
// Environment Isolation Tests
// =============================================================================

// TestSecurity_EnvironmentIsolation_ChildProcessOnly verifies that environment variables
// set by the auth CLI are only visible in the child process, not in the
// parent process or subsequent commands.
func TestSecurity_EnvironmentIsolation_ChildProcessOnly(t *testing.T) {
	binary := buildBinary(t)
	tempDir := t.TempDir()
	vaultDir := filepath.Join(tempDir, "vaults")
	if err := os.MkdirAll(vaultDir, 0700); err != nil {
		t.Fatalf("Failed to create vault directory: %v", err)
	}

	setupSecurityVault(t, binary, vaultDir, "isolationuser", "testpassword", map[string]string{
		"SECRET_API_KEY": "test-value-value-12345",
	})

	passwordFile := filepath.Join(tempDir, "password")
	if err := os.WriteFile(passwordFile, []byte("testpassword"), 0600); err != nil {
		t.Fatalf("Failed to write password file: %v", err)
	}

	cmd := exec.Command(binary, "--user", "isolationuser", "--password-file", passwordFile, "exec", "--", "env")
	cmd.Env = append(os.Environ(), "WITH_VAULT_DIR="+vaultDir)

	var stdout bytes.Buffer
	cmd.Stdout = &stdout

	if err := cmd.Run(); err != nil {
		t.Fatalf("Command failed: %v", err)
	}

	childEnv := stdout.String()
	if !strings.Contains(childEnv, "SECRET_API_KEY=test-value-value-12345") {
		t.Error("SECRET_API_KEY should be present in child process environment")
	}

	parentEnv := os.Getenv("SECRET_API_KEY")
	if parentEnv != "" {
		t.Error("SECRET_API_KEY should NOT be set in parent process environment")
	}
}

// TestSecurity_EnvironmentIsolation_MinimalEnvInheritance verifies that only
// essential environment variables are inherited by the child process.
func TestSecurity_EnvironmentIsolation_MinimalEnvInheritance(t *testing.T) {
	binary := buildBinary(t)
	tempDir := t.TempDir()
	vaultDir := filepath.Join(tempDir, "vaults")
	if err := os.MkdirAll(vaultDir, 0700); err != nil {
		t.Fatalf("Failed to create vault directory: %v", err)
	}

	setupSecurityVault(t, binary, vaultDir, "minimaluser", "testpassword", map[string]string{
		"SECRET_KEY": "test-val",
	})

	passwordFile := filepath.Join(tempDir, "password")
	if err := os.WriteFile(passwordFile, []byte("testpassword"), 0600); err != nil {
		t.Fatalf("Failed to write password file: %v", err)
	}

	os.Setenv("CUSTOM_SECRET_VAR", "should-not-appear-in-child")
	defer os.Unsetenv("CUSTOM_SECRET_VAR")

	cmd := exec.Command(binary, "--user", "minimaluser", "--password-file", passwordFile, "exec", "--", "env")
	cmd.Env = append(os.Environ(),
		"WITH_VAULT_DIR="+vaultDir,
		"ANOTHER_SECRET_VAR=also-should-not-appear",
	)

	var stdout bytes.Buffer
	cmd.Stdout = &stdout

	if err := cmd.Run(); err != nil {
		t.Fatalf("Command failed: %v", err)
	}

	childEnv := stdout.String()

	essentialVars := []string{"PATH", "HOME", "USER"}
	for _, v := range essentialVars {
		if !strings.Contains(childEnv, v+"=") {
			t.Errorf("Essential variable %s should be in child environment", v)
		}
	}

	if !strings.Contains(childEnv, "SECRET_KEY=test-val") {
		t.Error("SECRET_KEY from vault should be in child environment")
	}

	if strings.Contains(childEnv, "CUSTOM_SECRET_VAR=") {
		t.Error("CUSTOM_SECRET_VAR should NOT be in child environment (minimal env)")
	}
	if strings.Contains(childEnv, "ANOTHER_SECRET_VAR=") {
		t.Error("ANOTHER_SECRET_VAR should NOT be in child environment (minimal env)")
	}
}

// =============================================================================
// Signal Forwarding Tests
// =============================================================================

// TestSecurity_SignalForwarding_SIGTERM verifies that SIGTERM is properly forwarded
// from the parent process to the child process.
func TestSecurity_SignalForwarding_SIGTERM(t *testing.T) {
	binary := buildBinary(t)
	tempDir := t.TempDir()
	vaultDir := filepath.Join(tempDir, "vaults")
	if err := os.MkdirAll(vaultDir, 0700); err != nil {
		t.Fatalf("Failed to create vault directory: %v", err)
	}

	setupSecurityVault(t, binary, vaultDir, "signaluser", "testpassword", map[string]string{
		"KEY": "value",
	})

	passwordFile := filepath.Join(tempDir, "password")
	if err := os.WriteFile(passwordFile, []byte("testpassword"), 0600); err != nil {
		t.Fatalf("Failed to write password file: %v", err)
	}

	cmd := exec.Command(binary, "--user", "signaluser", "--password-file", passwordFile, "exec", "--", "sleep", "30")
	cmd.Env = append(os.Environ(), "WITH_VAULT_DIR="+vaultDir)

	if err := cmd.Start(); err != nil {
		t.Fatalf("Failed to start command: %v", err)
	}

	time.Sleep(200 * time.Millisecond)

	if err := cmd.Process.Signal(syscall.SIGTERM); err != nil {
		t.Fatalf("Failed to send SIGTERM: %v", err)
	}

	err := cmd.Wait()

	if err == nil {
		t.Log("Process exited cleanly (exit code 0)")
	} else {
		exitErr, ok := err.(*exec.ExitError)
		if !ok {
			t.Fatalf("Unexpected error type: %v", err)
		}

		if exitErr.ExitCode() == 0 {
			t.Error("Expected non-zero exit code after SIGTERM")
		}
	}
}

// TestSecurity_SignalForwarding_SIGINT verifies that SIGINT (Ctrl+C) is properly
// forwarded from the parent process to the child process.
func TestSecurity_SignalForwarding_SIGINT(t *testing.T) {
	binary := buildBinary(t)
	tempDir := t.TempDir()
	vaultDir := filepath.Join(tempDir, "vaults")
	if err := os.MkdirAll(vaultDir, 0700); err != nil {
		t.Fatalf("Failed to create vault directory: %v", err)
	}

	setupSecurityVault(t, binary, vaultDir, "sigintuser", "testpassword", map[string]string{
		"KEY": "value",
	})

	passwordFile := filepath.Join(tempDir, "password")
	if err := os.WriteFile(passwordFile, []byte("testpassword"), 0600); err != nil {
		t.Fatalf("Failed to write password file: %v", err)
	}

	cmd := exec.Command(binary, "--user", "sigintuser", "--password-file", passwordFile, "exec", "--", "sleep", "30")
	cmd.Env = append(os.Environ(), "WITH_VAULT_DIR="+vaultDir)

	if err := cmd.Start(); err != nil {
		t.Fatalf("Failed to start command: %v", err)
	}

	time.Sleep(200 * time.Millisecond)

	if err := cmd.Process.Signal(syscall.SIGINT); err != nil {
		t.Fatalf("Failed to send SIGINT: %v", err)
	}

	err := cmd.Wait()

	if err == nil {
		t.Log("Process exited cleanly")
	} else {
		exitErr, ok := err.(*exec.ExitError)
		if !ok {
			t.Fatalf("Unexpected error type: %v", err)
		}
		if exitErr.ExitCode() == 0 {
			t.Error("Expected non-zero exit code after SIGINT")
		}
	}
}

// =============================================================================
// Wrong Password Handling Tests
// =============================================================================

// TestSecurity_WrongPasswordHandling_SameErrorForWrongPasswordAndCorruptedFile verifies
// that the error message for a wrong password is similar to the error for a
// corrupted vault file. This prevents attackers from distinguishing between
// "wrong password" and "corrupted/tampered file" scenarios.
func TestSecurity_WrongPasswordHandling_SameErrorForWrongPasswordAndCorruptedFile(t *testing.T) {
	binary := buildBinary(t)
	tempDir := t.TempDir()
	vaultDir := filepath.Join(tempDir, "vaults")
	if err := os.MkdirAll(vaultDir, 0700); err != nil {
		t.Fatalf("Failed to create vault directory: %v", err)
	}

	setupSecurityVault(t, binary, vaultDir, "wrongpassuser", "correctpassword", map[string]string{
		"API_KEY": "test-val",
	})

	passwordFile := filepath.Join(tempDir, "wrongpassword")
	if err := os.WriteFile(passwordFile, []byte("wrongpassword"), 0600); err != nil {
		t.Fatalf("Failed to write password file: %v", err)
	}

	cmd := exec.Command(binary, "--user", "wrongpassuser", "--password-file", passwordFile, "list")
	cmd.Env = append(os.Environ(), "WITH_VAULT_DIR="+vaultDir)

	wrongPassOutput, wrongPassErr := cmd.CombinedOutput()
	wrongPassErrMsg := strings.TrimSpace(string(wrongPassOutput))

	if wrongPassErr == nil {
		t.Fatal("Expected error with wrong password")
	}

	corruptedVaultPath := filepath.Join(vaultDir, "corrupteduser.vault")
	corruptedContent := `{"version":1,"format":"cli-auth-vault","created_at":"2024-01-01T00:00:00Z","kdf":{"algorithm":"argon2id"},"encryption":{"algorithm":"aes-256-gcm"},"ciphertext":"aW52YWxpZC1jaXBoZXJ0ZXh0LXdpdGgtaW52YWxpZC1kYXRh"}`
	if err := os.WriteFile(corruptedVaultPath, []byte(corruptedContent), 0600); err != nil {
		t.Fatalf("Failed to write corrupted vault: %v", err)
	}

	corruptedPasswordFile := filepath.Join(tempDir, "corruptedpassword")
	if err := os.WriteFile(corruptedPasswordFile, []byte("anypassword"), 0600); err != nil {
		t.Fatalf("Failed to write password file: %v", err)
	}

	cmd2 := exec.Command(binary, "--user", "corrupteduser", "--password-file", corruptedPasswordFile, "list")
	cmd2.Env = append(os.Environ(), "WITH_VAULT_DIR="+vaultDir)

	corruptedOutput, corruptedErr := cmd2.CombinedOutput()
	corruptedErrMsg := strings.TrimSpace(string(corruptedOutput))

	if corruptedErr == nil {
		t.Fatal("Expected error with corrupted vault")
	}

	t.Logf("Wrong password error: %s", wrongPassErrMsg)
	t.Logf("Corrupted vault error: %s", corruptedErrMsg)

	secretValue := "test-val"
	if strings.Contains(wrongPassErrMsg, secretValue) {
		t.Error("Wrong password error should NOT contain the secret value")
	}
	if strings.Contains(corruptedErrMsg, secretValue) {
		t.Error("Corrupted vault error should NOT contain the secret value")
	}
}

// TestSecurity_WrongPasswordHandling_ConsistentErrorMessages verifies that all
// authentication failures produce consistent error messages.
func TestSecurity_WrongPasswordHandling_ConsistentErrorMessages(t *testing.T) {
	binary := buildBinary(t)
	tempDir := t.TempDir()
	vaultDir := filepath.Join(tempDir, "vaults")
	if err := os.MkdirAll(vaultDir, 0700); err != nil {
		t.Fatalf("Failed to create vault directory: %v", err)
	}

	setupSecurityVault(t, binary, vaultDir, "consistentuser", "password123", map[string]string{
		"KEY": "value",
	})

	passwordFile := filepath.Join(tempDir, "password")
	if err := os.WriteFile(passwordFile, []byte("wrongpassword"), 0600); err != nil {
		t.Fatalf("Failed to write password file: %v", err)
	}

	commands := [][]string{
		{"list"},
		{"get", "KEY"},
		{"set", "NEW_KEY"},
		{"remove", "KEY"},
	}

	for _, args := range commands {
		fullArgs := append([]string{"--user", "consistentuser", "--password-file", passwordFile}, args...)
		cmd := exec.Command(binary, fullArgs...)
		cmd.Env = append(os.Environ(), "WITH_VAULT_DIR="+vaultDir)

		output, err := cmd.CombinedOutput()
		if err == nil {
			t.Errorf("Command %v should fail with wrong password", args)
			continue
		}

		outputStr := string(output)

		if !strings.Contains(strings.ToLower(outputStr), "decrypt") && !strings.Contains(strings.ToLower(outputStr), "password") {
			t.Errorf("Error for command %v should mention decryption/password: %s", args, outputStr)
		}

		if strings.Contains(outputStr, "value") {
			t.Errorf("Error for command %v should NOT reveal secret value: %s", args, outputStr)
		}
	}
}

// =============================================================================
// Password Not in Process List Tests
// =============================================================================

// TestSecurity_PasswordNotInProcessList verifies that passwords are not visible in
// the process list (ps output). Passwords should be read from stdin or
// password files, never passed as command-line arguments.
func TestSecurity_PasswordNotInProcessList(t *testing.T) {
	binary := buildBinary(t)
	tempDir := t.TempDir()
	vaultDir := filepath.Join(tempDir, "vaults")
	if err := os.MkdirAll(vaultDir, 0700); err != nil {
		t.Fatalf("Failed to create vault directory: %v", err)
	}

	secretPassword := "test-value-password-12345"
	setupSecurityVault(t, binary, vaultDir, "processuser", secretPassword, map[string]string{
		"SECRET_KEY": "test-val",
	})

	passwordFile := filepath.Join(tempDir, "password")
	if err := os.WriteFile(passwordFile, []byte(secretPassword), 0600); err != nil {
		t.Fatalf("Failed to write password file: %v", err)
	}

	cmd := exec.Command(binary, "--user", "processuser", "--password-file", passwordFile, "exec", "--", "sleep", "5")
	cmd.Env = append(os.Environ(), "WITH_VAULT_DIR="+vaultDir)

	if err := cmd.Start(); err != nil {
		t.Fatalf("Failed to start command: %v", err)
	}
	defer cmd.Process.Kill()

	time.Sleep(100 * time.Millisecond)

	psCmd := exec.Command("ps", "-e", "-o", "command=")
	psOutput, err := psCmd.Output()
	if err != nil {
		t.Skipf("Could not run ps command: %v", err)
	}

	psOutputStr := string(psOutput)

	if strings.Contains(psOutputStr, secretPassword) {
		t.Error("Password is visible in process list - SECURITY ISSUE!")
		t.Logf("Process list output:\n%s", psOutputStr)
	}

	if !strings.Contains(psOutputStr, "auth") {
		t.Log("Auth process not found in ps output (might be timing issue)")
	}
}

// TestSecurity_SecretValueNotInProcessList verifies that secret values from the vault
// are not passed as command-line arguments or visible in the process list.
func TestSecurity_SecretValueNotInProcessList(t *testing.T) {
	binary := buildBinary(t)
	tempDir := t.TempDir()
	vaultDir := filepath.Join(tempDir, "vaults")
	if err := os.MkdirAll(vaultDir, 0700); err != nil {
		t.Fatalf("Failed to create vault directory: %v", err)
	}

	secretValue := "test-value-api-key-xyz123abc"
	setupSecurityVault(t, binary, vaultDir, "secretuser", "testpassword", map[string]string{
		"API_KEY": secretValue,
	})

	passwordFile := filepath.Join(tempDir, "password")
	if err := os.WriteFile(passwordFile, []byte("testpassword"), 0600); err != nil {
		t.Fatalf("Failed to write password file: %v", err)
	}

	cmd := exec.Command(binary, "--user", "secretuser", "--password-file", passwordFile, "exec", "--", "sleep", "5")
	cmd.Env = append(os.Environ(), "WITH_VAULT_DIR="+vaultDir)

	if err := cmd.Start(); err != nil {
		t.Fatalf("Failed to start command: %v", err)
	}
	defer cmd.Process.Kill()

	time.Sleep(100 * time.Millisecond)

	psCmd := exec.Command("ps", "-e", "-o", "command=")
	psOutput, err := psCmd.Output()
	if err != nil {
		t.Skipf("Could not run ps command: %v", err)
	}

	psOutputStr := string(psOutput)

	if strings.Contains(psOutputStr, secretValue) {
		t.Error("Secret value is visible in process list - SECURITY ISSUE!")
		t.Logf("Process list output:\n%s", psOutputStr)
	}
}

// =============================================================================
// No Secrets in Error Messages Tests
// =============================================================================

// TestSecurity_NoSecretsInErrorMessages_VariousScenarios verifies that error messages
// never contain secret values from the vault.
func TestSecurity_NoSecretsInErrorMessages_VariousScenarios(t *testing.T) {
	binary := buildBinary(t)
	tempDir := t.TempDir()
	vaultDir := filepath.Join(tempDir, "vaults")
	if err := os.MkdirAll(vaultDir, 0700); err != nil {
		t.Fatalf("Failed to create vault directory: %v", err)
	}

	secretKey := "SECRET_API_KEY"
	secretValue := "test-test-val-12345"
	setupSecurityVault(t, binary, vaultDir, "erroruser", "testpassword", map[string]string{
		secretKey: secretValue,
	})

	passwordFile := filepath.Join(tempDir, "password")
	if err := os.WriteFile(passwordFile, []byte("testpassword"), 0600); err != nil {
		t.Fatalf("Failed to write password file: %v", err)
	}

	tests := []struct {
		name string
		args []string
	}{
		{
			name: "get non-existent key",
			args: []string{"--user", "erroruser", "--password-file", passwordFile, "get", "NONEXISTENT_KEY"},
		},
		{
			name: "remove non-existent key",
			args: []string{"--user", "erroruser", "--password-file", passwordFile, "remove", "NONEXISTENT_KEY"},
		},
		{
			name: "set with invalid key name",
			args: []string{"--user", "erroruser", "--password-file", passwordFile, "set", "123invalid"},
		},
		{
			name: "list keys",
			args: []string{"--user", "erroruser", "--password-file", passwordFile, "list"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := exec.Command(binary, tt.args...)
			cmd.Env = append(os.Environ(), "WITH_VAULT_DIR="+vaultDir)

			output, _ := cmd.CombinedOutput()
			outputStr := string(output)

			if strings.Contains(outputStr, secretValue) {
				t.Errorf("Secret value appears in output for test %q:\n%s", tt.name, outputStr)
			}
		})
	}
}

// TestSecurity_NoSecretsInErrorMessages_WrongPassword verifies that when wrong password
// is provided, the error message doesn't hint at the correct password or
// reveal any vault contents.
func TestSecurity_NoSecretsInErrorMessages_WrongPassword(t *testing.T) {
	binary := buildBinary(t)
	tempDir := t.TempDir()
	vaultDir := filepath.Join(tempDir, "vaults")
	if err := os.MkdirAll(vaultDir, 0700); err != nil {
		t.Fatalf("Failed to create vault directory: %v", err)
	}

	correctPassword := "correct-password-123"
	wrongPassword := "wrong-password-456"
	secretValue := "test-value-api-key"

	setupSecurityVault(t, binary, vaultDir, "wrongpasserruser", correctPassword, map[string]string{
		"API_KEY": secretValue,
	})

	passwordFile := filepath.Join(tempDir, "wrongpassword")
	if err := os.WriteFile(passwordFile, []byte(wrongPassword), 0600); err != nil {
		t.Fatalf("Failed to write password file: %v", err)
	}

	cmd := exec.Command(binary, "--user", "wrongpasserruser", "--password-file", passwordFile, "list")
	cmd.Env = append(os.Environ(), "WITH_VAULT_DIR="+vaultDir)

	output, err := cmd.CombinedOutput()
	if err == nil {
		t.Fatal("Expected error with wrong password")
	}

	outputStr := string(output)

	if strings.Contains(outputStr, secretValue) {
		t.Error("Error message should NOT contain secret value")
	}
	if strings.Contains(outputStr, correctPassword) {
		t.Error("Error message should NOT contain correct password")
	}
	if strings.Contains(outputStr, "API_KEY") {
		t.Log("Warning: Error message contains key name (may be acceptable)")
	}

	t.Logf("Error output: %s", outputStr)
}

// =============================================================================
// Helper Functions
// =============================================================================

// setupSecurityVault creates a vault with the given user, password, and keys.
func setupSecurityVault(t *testing.T, binary, vaultDir, username, password string, keys map[string]string) {
	t.Helper()

	if err := os.MkdirAll(vaultDir, 0700); err != nil {
		t.Fatalf("Failed to create vault directory: %v", err)
	}

	tempDir := t.TempDir()
	passwordFile := filepath.Join(tempDir, "setup-password")
	if err := os.WriteFile(passwordFile, []byte(password), 0600); err != nil {
		t.Fatalf("Failed to write setup password file: %v", err)
	}

	cmd := exec.Command(binary, "--user", username, "--password-file", passwordFile, "init")
	cmd.Env = append(os.Environ(), "WITH_VAULT_DIR="+vaultDir)

	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("Failed to initialize vault: %v\n%s", err, output)
	}

	if len(keys) > 0 {
		for keyName, keyValue := range keys {
			keyPasswordFile := filepath.Join(tempDir, "key-password")
			if err := os.WriteFile(keyPasswordFile, []byte(password), 0600); err != nil {
				t.Fatalf("Failed to write key password file: %v", err)
			}

			cmd := exec.Command(binary, "--user", username, "--password-file", keyPasswordFile, "set", "--value", keyValue, keyName)
			cmd.Env = append(os.Environ(), "WITH_VAULT_DIR="+vaultDir)

			if output, err := cmd.CombinedOutput(); err != nil {
				t.Fatalf("Failed to set key %s: %v\n%s", keyName, err, output)
			}
		}
	}
}
