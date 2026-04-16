package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Grovy-3170/cli-with/internal/aliases"
)

// aliasTestSetup points the alias store at a fresh temp file for the duration of t.
func aliasTestSetup(t *testing.T) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "aliases.json")
	t.Setenv("WITH_ALIAS_FILE", path)
	return path
}

// captureStdout runs fn with os.Stdout redirected, returning captured output.
func captureStdout(t *testing.T, fn func()) string {
	t.Helper()
	old := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("os.Pipe: %v", err)
	}
	os.Stdout = w
	defer func() { os.Stdout = old }()

	fn()

	w.Close()
	var buf bytes.Buffer
	if _, err := buf.ReadFrom(r); err != nil {
		t.Fatalf("ReadFrom: %v", err)
	}
	return buf.String()
}

func TestAliasAdd_SavesCorrectly(t *testing.T) {
	path := aliasTestSetup(t)

	rootCmd := newRootCmd()
	rootCmd.SetArgs([]string{
		"alias", "add", "clippy",
		"--user", "alice",
		"--password", "",
		"--", "cargo", "clippy",
	})
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("alias add failed: %v", err)
	}

	store, err := aliases.Load(path)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}
	a, ok := store.Aliases["clippy"]
	if !ok {
		t.Fatal("alias 'clippy' was not saved")
	}
	if a.User != "alice" {
		t.Errorf("expected user 'alice', got %q", a.User)
	}
	if !a.PasswordSet {
		t.Error("expected PasswordSet=true after --password was used")
	}
	if a.Password != "" {
		t.Errorf("expected empty password, got %q", a.Password)
	}
	if len(a.Command) != 2 || a.Command[0] != "cargo" || a.Command[1] != "clippy" {
		t.Errorf("expected [cargo clippy], got %v", a.Command)
	}
}

func TestAliasAdd_InvalidName(t *testing.T) {
	aliasTestSetup(t)

	rootCmd := newRootCmd()
	rootCmd.SetArgs([]string{"alias", "add", "123bad", "--", "echo", "hi"})
	err := rootCmd.Execute()
	if err == nil {
		t.Fatal("expected error for invalid alias name")
	}
	if !strings.Contains(err.Error(), "invalid key name") {
		t.Errorf("expected 'invalid key name' error, got: %v", err)
	}
}

func TestAliasAdd_MissingCommand(t *testing.T) {
	aliasTestSetup(t)

	rootCmd := newRootCmd()
	rootCmd.SetArgs([]string{"alias", "add", "nocmd"})
	if err := rootCmd.Execute(); err == nil {
		t.Fatal("expected error when no command is supplied")
	}
}

func TestAliasAdd_PasswordNotFlaggedStaysUnset(t *testing.T) {
	path := aliasTestSetup(t)

	rootCmd := newRootCmd()
	rootCmd.SetArgs([]string{"alias", "add", "foo", "--user", "alice", "--", "ls", "-la"})
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("alias add failed: %v", err)
	}

	store, err := aliases.Load(path)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}
	a := store.Aliases["foo"]
	if a.PasswordSet {
		t.Error("expected PasswordSet=false when --password is not used")
	}
}

func TestAliasList_EmptyStore(t *testing.T) {
	aliasTestSetup(t)

	out := captureStdout(t, func() {
		rootCmd := newRootCmd()
		rootCmd.SetArgs([]string{"alias", "list"})
		if err := rootCmd.Execute(); err != nil {
			t.Fatalf("alias list failed: %v", err)
		}
	})

	if !strings.Contains(out, "No aliases saved") {
		t.Errorf("expected 'No aliases saved', got: %s", out)
	}
}

func TestAliasList_ShowsAliases(t *testing.T) {
	aliasTestSetup(t)

	add := newRootCmd()
	add.SetArgs([]string{"alias", "add", "clippy", "--user", "alice", "--", "cargo", "clippy"})
	if err := add.Execute(); err != nil {
		t.Fatalf("add failed: %v", err)
	}

	out := captureStdout(t, func() {
		list := newRootCmd()
		list.SetArgs([]string{"alias", "list"})
		if err := list.Execute(); err != nil {
			t.Fatalf("list failed: %v", err)
		}
	})

	if !strings.Contains(out, "clippy") {
		t.Errorf("missing alias name in output: %s", out)
	}
	if !strings.Contains(out, "cargo clippy") {
		t.Errorf("missing command in output: %s", out)
	}
	if !strings.Contains(out, "alice") {
		t.Errorf("missing user in output: %s", out)
	}
}

func TestAliasRemove_Works(t *testing.T) {
	path := aliasTestSetup(t)

	add := newRootCmd()
	add.SetArgs([]string{"alias", "add", "foo", "--", "ls"})
	if err := add.Execute(); err != nil {
		t.Fatalf("add failed: %v", err)
	}

	rm := newRootCmd()
	rm.SetArgs([]string{"alias", "remove", "foo"})
	if err := rm.Execute(); err != nil {
		t.Fatalf("remove failed: %v", err)
	}

	store, err := aliases.Load(path)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}
	if _, ok := store.Aliases["foo"]; ok {
		t.Error("alias 'foo' should have been removed")
	}
}

func TestAliasRemove_Missing(t *testing.T) {
	aliasTestSetup(t)

	rm := newRootCmd()
	rm.SetArgs([]string{"alias", "remove", "nonexistent"})
	err := rm.Execute()
	if err == nil {
		t.Fatal("expected error for missing alias")
	}
	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("expected 'not found' error, got: %v", err)
	}
}

func TestAliasShell_BashOutput(t *testing.T) {
	aliasTestSetup(t)

	add := newRootCmd()
	add.SetArgs([]string{"alias", "add", "clippy", "--user", "alice", "--", "cargo", "clippy"})
	if err := add.Execute(); err != nil {
		t.Fatalf("add failed: %v", err)
	}

	out := captureStdout(t, func() {
		shell := newRootCmd()
		shell.SetArgs([]string{"alias", "shell", "--shell", "bash"})
		if err := shell.Execute(); err != nil {
			t.Fatalf("shell failed: %v", err)
		}
	})

	want := `alias clippy='with exec --user alice -- cargo clippy'`
	if !strings.Contains(out, want) {
		t.Errorf("missing expected line\nwant substring: %s\ngot:            %s", want, out)
	}
}

func TestAliasShell_FishOutput(t *testing.T) {
	aliasTestSetup(t)

	add := newRootCmd()
	add.SetArgs([]string{"alias", "add", "ll", "--", "ls", "-la"})
	if err := add.Execute(); err != nil {
		t.Fatalf("add failed: %v", err)
	}

	out := captureStdout(t, func() {
		shell := newRootCmd()
		shell.SetArgs([]string{"alias", "shell", "--shell", "fish"})
		if err := shell.Execute(); err != nil {
			t.Fatalf("shell failed: %v", err)
		}
	})

	want := `alias ll 'with exec -- ls -la'`
	if !strings.Contains(out, want) {
		t.Errorf("missing expected line\nwant substring: %s\ngot:            %s", want, out)
	}
}

func TestAliasShell_EmptyStoreSilentSuccess(t *testing.T) {
	aliasTestSetup(t)

	out := captureStdout(t, func() {
		shell := newRootCmd()
		shell.SetArgs([]string{"alias", "shell", "--shell", "bash"})
		if err := shell.Execute(); err != nil {
			t.Fatalf("shell failed: %v", err)
		}
	})

	if strings.TrimSpace(out) != "" {
		t.Errorf("expected empty output for empty store, got: %q", out)
	}
}
