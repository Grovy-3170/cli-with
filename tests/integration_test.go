package tests

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"testing"
)

func TestFullWorkflow(t *testing.T) {
	binary := buildBinary(t)
	tempDir := t.TempDir()
	vaultDir := filepath.Join(tempDir, "vaults")

	password := "testpassword123"
	username := "workflowuser"

	t.Run("init", func(t *testing.T) {
		passwordFile := filepath.Join(tempDir, "init-password")
		if err := os.WriteFile(passwordFile, []byte(password), 0600); err != nil {
			t.Fatalf("Failed to write password file: %v", err)
		}

		cmd := exec.Command(binary, "--user", username, "--password-file", passwordFile, "init")
		cmd.Env = append(os.Environ(), "WITH_VAULT_DIR="+vaultDir)

		output, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("Failed to initialize vault: %v\n%s", err, output)
		}

		if !strings.Contains(string(output), "Vault initialized successfully") {
			t.Errorf("Expected success message, got: %s", output)
		}

		vaultPath := filepath.Join(vaultDir, username+".vault")
		if _, err := os.Stat(vaultPath); os.IsNotExist(err) {
			t.Error("Vault file was not created")
		}
	})

	t.Run("set_first_key", func(t *testing.T) {
		passwordFile := filepath.Join(tempDir, "set-password-1")
		if err := os.WriteFile(passwordFile, []byte(password), 0600); err != nil {
			t.Fatalf("Failed to write password file: %v", err)
		}

		cmd := exec.Command(binary, "--user", username, "--password-file", passwordFile, "set", "--value", "my-secret-api-key", "API_KEY")
		cmd.Env = append(os.Environ(), "WITH_VAULT_DIR="+vaultDir)

		output, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("Failed to set key: %v\n%s", err, output)
		}

		if !strings.Contains(string(output), "Key 'API_KEY' set successfully") {
			t.Errorf("Expected success message, got: %s", output)
		}
	})

	t.Run("set_second_key", func(t *testing.T) {
		passwordFile := filepath.Join(tempDir, "set-password-2")
		if err := os.WriteFile(passwordFile, []byte(password), 0600); err != nil {
			t.Fatalf("Failed to write password file: %v", err)
		}

		cmd := exec.Command(binary, "--user", username, "--password-file", passwordFile, "set", "--value", "my-database-url", "DATABASE_URL")
		cmd.Env = append(os.Environ(), "WITH_VAULT_DIR="+vaultDir)

		output, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("Failed to set key: %v\n%s", err, output)
		}

		if !strings.Contains(string(output), "Key 'DATABASE_URL' set successfully") {
			t.Errorf("Expected success message, got: %s", output)
		}
	})

	t.Run("list", func(t *testing.T) {
		passwordFile := filepath.Join(tempDir, "list-password")
		if err := os.WriteFile(passwordFile, []byte(password), 0600); err != nil {
			t.Fatalf("Failed to write password file: %v", err)
		}

		cmd := exec.Command(binary, "--user", username, "--password-file", passwordFile, "list")
		cmd.Env = append(os.Environ(), "WITH_VAULT_DIR="+vaultDir)

		output, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("Failed to list keys: %v\n%s", err, output)
		}

		if !strings.Contains(string(output), "API_KEY") {
			t.Errorf("Expected API_KEY in list, got: %s", output)
		}
		if !strings.Contains(string(output), "DATABASE_URL") {
			t.Errorf("Expected DATABASE_URL in list, got: %s", output)
		}
	})

	t.Run("get", func(t *testing.T) {
		passwordFile := filepath.Join(tempDir, "get-password")
		if err := os.WriteFile(passwordFile, []byte(password), 0600); err != nil {
			t.Fatalf("Failed to write password file: %v", err)
		}

		cmd := exec.Command(binary, "--user", username, "--password-file", passwordFile, "get", "API_KEY")
		cmd.Env = append(os.Environ(), "WITH_VAULT_DIR="+vaultDir)

		output, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("Failed to get key: %v\n%s", err, output)
		}

		if strings.TrimSpace(string(output)) != "my-secret-api-key" {
			t.Errorf("Expected 'my-secret-api-key', got: %s", output)
		}
	})

	t.Run("remove_key", func(t *testing.T) {
		passwordFile := filepath.Join(tempDir, "remove-password")
		if err := os.WriteFile(passwordFile, []byte(password), 0600); err != nil {
			t.Fatalf("Failed to write password file: %v", err)
		}

		cmd := exec.Command(binary, "--user", username, "--password-file", passwordFile, "remove", "API_KEY")
		cmd.Env = append(os.Environ(), "WITH_VAULT_DIR="+vaultDir)

		output, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("Failed to remove key: %v\n%s", err, output)
		}

		if !strings.Contains(string(output), "Key 'API_KEY' removed successfully") {
			t.Errorf("Expected success message, got: %s", output)
		}
	})

	t.Run("list_after_remove", func(t *testing.T) {
		passwordFile := filepath.Join(tempDir, "list-after-password")
		if err := os.WriteFile(passwordFile, []byte(password), 0600); err != nil {
			t.Fatalf("Failed to write password file: %v", err)
		}

		cmd := exec.Command(binary, "--user", username, "--password-file", passwordFile, "list")
		cmd.Env = append(os.Environ(), "WITH_VAULT_DIR="+vaultDir)

		output, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("Failed to list keys: %v\n%s", err, output)
		}

		if !strings.Contains(string(output), "DATABASE_URL") {
			t.Errorf("Expected DATABASE_URL in list, got: %s", output)
		}
		if strings.Contains(string(output), "API_KEY") {
			t.Errorf("API_KEY should not be in list after removal, got: %s", output)
		}
	})

	t.Run("delete_vault", func(t *testing.T) {
		passwordFile := filepath.Join(tempDir, "delete-password")
		if err := os.WriteFile(passwordFile, []byte(password), 0600); err != nil {
			t.Fatalf("Failed to write password file: %v", err)
		}

		cmd := exec.Command(binary, "--user", username, "--password-file", passwordFile, "remove")
		cmd.Env = append(os.Environ(), "WITH_VAULT_DIR="+vaultDir)
		cmd.Stdin = strings.NewReader("yes\n")
		output, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("Failed to delete vault: %v\n%s", err, output)
		}

		if !strings.Contains(string(output), "deleted successfully") {
			t.Errorf("Expected success message, got: %s", output)
		}
	})
}

func TestConcurrentAccess(t *testing.T) {
	binary := buildBinary(t)
	tempDir := t.TempDir()
	vaultDir := filepath.Join(tempDir, "vaults")

	password := "concurrentpass123"
	username := "concurrentuser"

	passwordFile := filepath.Join(tempDir, "init-password")
	if err := os.WriteFile(passwordFile, []byte(password), 0600); err != nil {
		t.Fatalf("Failed to write password file: %v", err)
	}

	cmd := exec.Command(binary, "--user", username, "--password-file", passwordFile, "init")
	cmd.Env = append(os.Environ(), "WITH_VAULT_DIR="+vaultDir)
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("Failed to initialize vault: %v\n%s", err, output)
	}

	for i := 0; i < 5; i++ {
		keyName := "INIT_KEY_" + string(rune('A'+i))
		keyValue := "initial-value-" + string(rune('A'+i))

		passwordFile := filepath.Join(tempDir, "set-password-"+string(rune('A'+i)))
		if err := os.WriteFile(passwordFile, []byte(password), 0600); err != nil {
			t.Fatalf("Failed to write password file: %v", err)
		}

		cmd := exec.Command(binary, "--user", username, "--password-file", passwordFile, "set", "--value", keyValue, keyName)
		cmd.Env = append(os.Environ(), "WITH_VAULT_DIR="+vaultDir)
		if output, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("Failed to set initial key: %v\n%s", err, output)
		}
	}

	var wg sync.WaitGroup
	errChan := make(chan error, 30)

	for i := 0; i < 10; i++ {
		wg.Add(2)

		go func(id int) {
			defer wg.Done()
			passwordFile := filepath.Join(tempDir, "read-password-"+string(rune('0'+id)))
			if err := os.WriteFile(passwordFile, []byte(password), 0600); err != nil {
				errChan <- err
				return
			}

			cmd := exec.Command(binary, "--user", username, "--password-file", passwordFile, "list")
			cmd.Env = append(os.Environ(), "WITH_VAULT_DIR="+vaultDir)
			if _, err := cmd.CombinedOutput(); err != nil {
				errChan <- err
			}
		}(i)

		go func(id int) {
			defer wg.Done()
			keyName := "INIT_KEY_" + string(rune('A'+(id%5)))

			passwordFile := filepath.Join(tempDir, "get-password-"+string(rune('0'+id)))
			if err := os.WriteFile(passwordFile, []byte(password), 0600); err != nil {
				errChan <- err
				return
			}

			cmd := exec.Command(binary, "--user", username, "--password-file", passwordFile, "get", keyName)
			cmd.Env = append(os.Environ(), "WITH_VAULT_DIR="+vaultDir)
			if _, err := cmd.CombinedOutput(); err != nil {
				errChan <- err
			}
		}(i)
	}

	wg.Wait()
	close(errChan)

	var errors []error
	for err := range errChan {
		errors = append(errors, err)
	}

	if len(errors) > 0 {
		t.Errorf("Concurrent read access had %d errors:", len(errors))
		for i, err := range errors {
			if i < 5 {
				t.Errorf("  Error %d: %v", i, err)
			}
		}
	}

	passwordFile = filepath.Join(tempDir, "final-list-password")
	if err := os.WriteFile(passwordFile, []byte(password), 0600); err != nil {
		t.Fatalf("Failed to write password file: %v", err)
	}

	cmd = exec.Command(binary, "--user", username, "--password-file", passwordFile, "list")
	cmd.Env = append(os.Environ(), "WITH_VAULT_DIR="+vaultDir)
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Failed to list keys after concurrent access: %v\n%s", err, output)
	}

	for i := 0; i < 5; i++ {
		keyName := "INIT_KEY_" + string(rune('A'+i))
		if !strings.Contains(string(output), keyName) {
			t.Errorf("Key %s not found after concurrent access", keyName)
		}
	}
}

func TestMultipleKeys(t *testing.T) {
	binary := buildBinary(t)
	tempDir := t.TempDir()
	vaultDir := filepath.Join(tempDir, "vaults")

	password := "multikeypass123"
	username := "multikeyuser"

	passwordFile := filepath.Join(tempDir, "init-password")
	if err := os.WriteFile(passwordFile, []byte(password), 0600); err != nil {
		t.Fatalf("Failed to write password file: %v", err)
	}

	cmd := exec.Command(binary, "--user", username, "--password-file", passwordFile, "init")
	cmd.Env = append(os.Environ(), "WITH_VAULT_DIR="+vaultDir)
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("Failed to initialize vault: %v\n%s", err, output)
	}

	keys := map[string]string{
		"API_KEY":        "test-api-key-12345",
		"DATABASE_URL":   "localhost:5432",
		"SECRET_TOKEN":   "test-token-xyz",
		"AWS_ACCESS_KEY": "test-aws-access-key",
		"AWS_SECRET_KEY": "test-aws-secret-key",
	}

	for keyName, keyValue := range keys {
		passwordFile := filepath.Join(tempDir, "set-password-"+keyName)
		if err := os.WriteFile(passwordFile, []byte(password), 0600); err != nil {
			t.Fatalf("Failed to write password file: %v", err)
		}

		cmd := exec.Command(binary, "--user", username, "--password-file", passwordFile, "set", "--value", keyValue, keyName)
		cmd.Env = append(os.Environ(), "WITH_VAULT_DIR="+vaultDir)
		if output, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("Failed to set key %s: %v\n%s", keyName, err, output)
		}
	}

	passwordFile = filepath.Join(tempDir, "list-password")
	if err := os.WriteFile(passwordFile, []byte(password), 0600); err != nil {
		t.Fatalf("Failed to write password file: %v", err)
	}

	cmd = exec.Command(binary, "--user", username, "--password-file", passwordFile, "list")
	cmd.Env = append(os.Environ(), "WITH_VAULT_DIR="+vaultDir)
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Failed to list keys: %v\n%s", err, output)
	}

	outputStr := string(output)
	for keyName := range keys {
		if !strings.Contains(outputStr, keyName) {
			t.Errorf("Key %s not found in list output", keyName)
		}
	}

	for keyName, expectedValue := range keys {
		passwordFile := filepath.Join(tempDir, "get-password-"+keyName)
		if err := os.WriteFile(passwordFile, []byte(password), 0600); err != nil {
			t.Fatalf("Failed to write password file: %v", err)
		}

		cmd := exec.Command(binary, "--user", username, "--password-file", passwordFile, "get", keyName)
		cmd.Env = append(os.Environ(), "WITH_VAULT_DIR="+vaultDir)
		output, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("Failed to get key %s: %v\n%s", keyName, err, output)
		}

		actualValue := strings.TrimSpace(string(output))
		if actualValue != expectedValue {
			t.Errorf("Key %s: expected %q, got %q", keyName, expectedValue, actualValue)
		}
	}

	passwordFile = filepath.Join(tempDir, "delete-password")
	if err := os.WriteFile(passwordFile, []byte(password), 0600); err != nil {
		t.Fatalf("Failed to write password file: %v", err)
	}

	cmd = exec.Command(binary, "--user", username, "--password-file", passwordFile, "remove")
	cmd.Env = append(os.Environ(), "WITH_VAULT_DIR="+vaultDir)
	cmd.Stdin = strings.NewReader("yes\n")
	output, err = cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Failed to delete vault: %v\n%s", err, output)
	}
}

func TestExecCommand(t *testing.T) {
	binary := buildBinary(t)
	tempDir := t.TempDir()
	vaultDir := filepath.Join(tempDir, "vaults")

	password := "execpass123"
	username := "execuser"

	passwordFile := filepath.Join(tempDir, "init-password")
	if err := os.WriteFile(passwordFile, []byte(password), 0600); err != nil {
		t.Fatalf("Failed to write password file: %v", err)
	}

	cmd := exec.Command(binary, "--user", username, "--password-file", passwordFile, "init")
	cmd.Env = append(os.Environ(), "WITH_VAULT_DIR="+vaultDir)
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("Failed to initialize vault: %v\n%s", err, output)
	}

	keys := map[string]string{
		"MY_API_KEY":  "test-api-key-12345",
		"MY_DATABASE": "test-database-url",
	}

	for keyName, keyValue := range keys {
		passwordFile := filepath.Join(tempDir, "set-password-"+keyName)
		if err := os.WriteFile(passwordFile, []byte(password), 0600); err != nil {
			t.Fatalf("Failed to write password file: %v", err)
		}

		cmd := exec.Command(binary, "--user", username, "--password-file", passwordFile, "set", "--value", keyValue, keyName)
		cmd.Env = append(os.Environ(), "WITH_VAULT_DIR="+vaultDir)
		if output, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("Failed to set key %s: %v\n%s", keyName, err, output)
		}
	}

	passwordFile = filepath.Join(tempDir, "exec-password")
	if err := os.WriteFile(passwordFile, []byte(password), 0600); err != nil {
		t.Fatalf("Failed to write password file: %v", err)
	}

	cmd = exec.Command(binary, "--user", username, "--password-file", passwordFile, "exec", "--", "env")
	cmd.Env = append(os.Environ(), "WITH_VAULT_DIR="+vaultDir)
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Failed to exec command: %v\n%s", err, output)
	}

	outputStr := string(output)
	for keyName, expectedValue := range keys {
		expected := keyName + "=" + expectedValue
		if !strings.Contains(outputStr, expected) {
			t.Errorf("Environment variable %s not found in exec output", expected)
		}
	}
}

func TestVaultNotFound(t *testing.T) {
	binary := buildBinary(t)
	tempDir := t.TempDir()
	vaultDir := filepath.Join(tempDir, "vaults")

	passwordFile := filepath.Join(tempDir, "password")
	if err := os.WriteFile(passwordFile, []byte("testpassword"), 0600); err != nil {
		t.Fatalf("Failed to write password file: %v", err)
	}

	commands := []struct {
		name string
		args []string
	}{
		{"list", []string{"list"}},
		{"get", []string{"get", "SOME_KEY"}},
		{"set", []string{"set", "--value", "test", "SOME_KEY"}},
		{"remove", []string{"remove", "SOME_KEY"}},
	}

	for _, tc := range commands {
		t.Run(tc.name, func(t *testing.T) {
			fullArgs := append([]string{"--user", "nonexistentuser", "--password-file", passwordFile}, tc.args...)
			cmd := exec.Command(binary, fullArgs...)
			cmd.Env = append(os.Environ(), "WITH_VAULT_DIR="+vaultDir)

			output, err := cmd.CombinedOutput()
			if err == nil {
				t.Errorf("Expected error when vault doesn't exist for command %s", tc.name)
			}

			outputStr := string(output)
			if !strings.Contains(outputStr, "does not exist") && !strings.Contains(outputStr, "no such file") {
				t.Errorf("Expected 'does not exist' or 'no such file' error, got: %s", output)
			}
		})
	}
}

func TestKeyNotFound(t *testing.T) {
	binary := buildBinary(t)
	tempDir := t.TempDir()
	vaultDir := filepath.Join(tempDir, "vaults")

	password := "testpassword123"
	username := "keynotfounduser"

	passwordFile := filepath.Join(tempDir, "init-password")
	if err := os.WriteFile(passwordFile, []byte(password), 0600); err != nil {
		t.Fatalf("Failed to write password file: %v", err)
	}

	cmd := exec.Command(binary, "--user", username, "--password-file", passwordFile, "init")
	cmd.Env = append(os.Environ(), "WITH_VAULT_DIR="+vaultDir)
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("Failed to initialize vault: %v\n%s", err, output)
	}

	passwordFile = filepath.Join(tempDir, "get-password")
	if err := os.WriteFile(passwordFile, []byte(password), 0600); err != nil {
		t.Fatalf("Failed to write password file: %v", err)
	}

	cmd = exec.Command(binary, "--user", username, "--password-file", passwordFile, "get", "NONEXISTENT_KEY")
	cmd.Env = append(os.Environ(), "WITH_VAULT_DIR="+vaultDir)
	output, err := cmd.CombinedOutput()
	if err == nil {
		t.Error("Expected error when getting non-existent key")
	}

	if !strings.Contains(string(output), "not found") {
		t.Errorf("Expected 'not found' error, got: %s", output)
	}
}

func TestInvalidKeyName(t *testing.T) {
	binary := buildBinary(t)
	tempDir := t.TempDir()
	vaultDir := filepath.Join(tempDir, "vaults")

	password := "testpassword123"
	username := "invalidkeyuser"

	passwordFile := filepath.Join(tempDir, "init-password")
	if err := os.WriteFile(passwordFile, []byte(password), 0600); err != nil {
		t.Fatalf("Failed to write password file: %v", err)
	}

	cmd := exec.Command(binary, "--user", username, "--password-file", passwordFile, "init")
	cmd.Env = append(os.Environ(), "WITH_VAULT_DIR="+vaultDir)
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("Failed to initialize vault: %v\n%s", err, output)
	}

	invalidKeyNames := []string{
		"123invalid",
		"invalid-key",
		"invalid.key",
		"invalid key",
	}

	for _, keyName := range invalidKeyNames {
		t.Run(keyName, func(t *testing.T) {
			passwordFile := filepath.Join(tempDir, "set-password")
			if err := os.WriteFile(passwordFile, []byte(password), 0600); err != nil {
				t.Fatalf("Failed to write password file: %v", err)
			}

			cmd := exec.Command(binary, "--user", username, "--password-file", passwordFile, "set", "--value", "somevalue", keyName)
			cmd.Env = append(os.Environ(), "WITH_VAULT_DIR="+vaultDir)
			output, err := cmd.CombinedOutput()
			if err == nil {
				t.Errorf("Expected error for invalid key name %q", keyName)
			}

			if !strings.Contains(string(output), "invalid key name") {
				t.Errorf("Expected 'invalid key name' error for %q, got: %s", keyName, output)
			}
		})
	}
}

func TestWrongPassword(t *testing.T) {
	binary := buildBinary(t)
	tempDir := t.TempDir()
	vaultDir := filepath.Join(tempDir, "vaults")

	correctPassword := "correctpassword123"
	username := "wrongpassuser"

	passwordFile := filepath.Join(tempDir, "init-password")
	if err := os.WriteFile(passwordFile, []byte(correctPassword), 0600); err != nil {
		t.Fatalf("Failed to write password file: %v", err)
	}

	cmd := exec.Command(binary, "--user", username, "--password-file", passwordFile, "init")
	cmd.Env = append(os.Environ(), "WITH_VAULT_DIR="+vaultDir)
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("Failed to initialize vault: %v\n%s", err, output)
	}

	passwordFile = filepath.Join(tempDir, "set-password")
	if err := os.WriteFile(passwordFile, []byte(correctPassword), 0600); err != nil {
		t.Fatalf("Failed to write password file: %v", err)
	}

	cmd = exec.Command(binary, "--user", username, "--password-file", passwordFile, "set", "--value", "testvalue", "TEST_KEY")
	cmd.Env = append(os.Environ(), "WITH_VAULT_DIR="+vaultDir)
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("Failed to set key: %v\n%s", err, output)
	}

	wrongPasswordFile := filepath.Join(tempDir, "wrong-password")
	if err := os.WriteFile(wrongPasswordFile, []byte("wrongpassword"), 0600); err != nil {
		t.Fatalf("Failed to write password file: %v", err)
	}

	commands := []struct {
		name string
		args []string
	}{
		{"list", []string{"list"}},
		{"get", []string{"get", "TEST_KEY"}},
		{"set", []string{"set", "--value", "newvalue", "ANOTHER_KEY"}},
		{"remove", []string{"remove", "TEST_KEY"}},
	}

	for _, tc := range commands {
		t.Run(tc.name, func(t *testing.T) {
			fullArgs := append([]string{"--user", username, "--password-file", wrongPasswordFile}, tc.args...)
			cmd := exec.Command(binary, fullArgs...)
			cmd.Env = append(os.Environ(), "WITH_VAULT_DIR="+vaultDir)

			output, err := cmd.CombinedOutput()
			if err == nil {
				t.Errorf("Expected error with wrong password for command %s", tc.name)
			}

			outputStr := strings.ToLower(string(output))
			if !strings.Contains(outputStr, "decrypt") && !strings.Contains(outputStr, "password") {
				t.Errorf("Expected decryption/password error, got: %s", output)
			}
		})
	}
}

func TestDeleteEntireVault(t *testing.T) {
	binary := buildBinary(t)
	tempDir := t.TempDir()
	vaultDir := filepath.Join(tempDir, "vaults")

	password := "deletepass123"
	username := "deleteuser"

	passwordFile := filepath.Join(tempDir, "init-password")
	if err := os.WriteFile(passwordFile, []byte(password), 0600); err != nil {
		t.Fatalf("Failed to write password file: %v", err)
	}

	cmd := exec.Command(binary, "--user", username, "--password-file", passwordFile, "init")
	cmd.Env = append(os.Environ(), "WITH_VAULT_DIR="+vaultDir)
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("Failed to initialize vault: %v\n%s", err, output)
	}

	passwordFile = filepath.Join(tempDir, "set-password")
	if err := os.WriteFile(passwordFile, []byte(password), 0600); err != nil {
		t.Fatalf("Failed to write password file: %v", err)
	}

	cmd = exec.Command(binary, "--user", username, "--password-file", passwordFile, "set", "--value", "testvalue", "TEST_KEY")
	cmd.Env = append(os.Environ(), "WITH_VAULT_DIR="+vaultDir)
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("Failed to set key: %v\n%s", err, output)
	}

	vaultPath := filepath.Join(vaultDir, username+".vault")
	if _, err := os.Stat(vaultPath); os.IsNotExist(err) {
		t.Fatal("Vault file was not created")
	}

	passwordFile = filepath.Join(tempDir, "delete-password")
	if err := os.WriteFile(passwordFile, []byte(password), 0600); err != nil {
		t.Fatalf("Failed to write password file: %v", err)
	}

	cmd = exec.Command(binary, "--user", username, "--password-file", passwordFile, "remove")
	cmd.Env = append(os.Environ(), "WITH_VAULT_DIR="+vaultDir)
	cmd.Stdin = strings.NewReader("yes\n")
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Failed to delete vault: %v\n%s", err, output)
	}

	if !strings.Contains(string(output), "deleted successfully") {
		t.Errorf("Expected success message, got: %s", output)
	}

	if _, err := os.Stat(vaultPath); !os.IsNotExist(err) {
		t.Error("Vault file should have been deleted")
	}
}
