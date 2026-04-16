package aliases

import (
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestLoad_NonExistentReturnsEmptyStore(t *testing.T) {
	path := filepath.Join(t.TempDir(), "does-not-exist.json")
	store, err := Load(path)
	if err != nil {
		t.Fatalf("Load returned error for missing file: %v", err)
	}
	if store == nil {
		t.Fatal("expected non-nil store")
	}
	if store.Aliases == nil {
		t.Fatal("expected non-nil Aliases map")
	}
	if len(store.Aliases) != 0 {
		t.Errorf("expected empty store, got %d aliases", len(store.Aliases))
	}
}

func TestSet_AddAndOverwrite(t *testing.T) {
	s := &Store{}
	s.Set("foo", Alias{Command: []string{"echo", "1"}})
	s.Set("foo", Alias{Command: []string{"echo", "2"}})
	if len(s.Aliases) != 1 {
		t.Errorf("expected 1 alias after overwrite, got %d", len(s.Aliases))
	}
	if s.Aliases["foo"].Command[1] != "2" {
		t.Errorf("expected overwritten command [echo 2], got %v", s.Aliases["foo"].Command)
	}
}

func TestRemove(t *testing.T) {
	s := &Store{Aliases: map[string]Alias{
		"foo": {Command: []string{"x"}},
	}}
	if !s.Remove("foo") {
		t.Error("expected Remove to return true for existing alias")
	}
	if s.Remove("foo") {
		t.Error("expected Remove to return false for already-removed alias")
	}
	if _, ok := s.Aliases["foo"]; ok {
		t.Error("alias still present after removal")
	}
}

func TestNames_Sorted(t *testing.T) {
	s := &Store{Aliases: map[string]Alias{
		"charlie": {},
		"alpha":   {},
		"bravo":   {},
	}}
	names := s.Names()
	want := []string{"alpha", "bravo", "charlie"}
	if !reflect.DeepEqual(names, want) {
		t.Errorf("expected %v, got %v", want, names)
	}
}

func TestSaveLoad_RoundTrip(t *testing.T) {
	path := filepath.Join(t.TempDir(), "aliases.json")

	original := &Store{Aliases: map[string]Alias{
		"deploy": {
			Command:     []string{"bash", "-c", "echo deploying"},
			User:        "alice",
			Password:    "",
			PasswordSet: true,
		},
		"build": {
			Command:      []string{"make", "build"},
			PasswordFile: "/tmp/pw",
		},
	}}

	if err := original.Save(path); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	loaded, err := Load(path)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	if !reflect.DeepEqual(loaded.Aliases, original.Aliases) {
		t.Errorf("round-trip mismatch:\ngot:  %+v\nwant: %+v", loaded.Aliases, original.Aliases)
	}
}

func TestSave_FilePermissions(t *testing.T) {
	path := filepath.Join(t.TempDir(), "aliases.json")
	s := &Store{Aliases: map[string]Alias{"x": {Command: []string{"y"}}}}
	if err := s.Save(path); err != nil {
		t.Fatalf("Save failed: %v", err)
	}
	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("Stat failed: %v", err)
	}
	if perm := info.Mode().Perm(); perm != 0600 {
		t.Errorf("expected 0600 permissions, got %04o", perm)
	}
}

func TestSave_CreatesMissingDirectory(t *testing.T) {
	path := filepath.Join(t.TempDir(), "nested", "dir", "aliases.json")
	s := &Store{Aliases: map[string]Alias{"x": {Command: []string{"y"}}}}
	if err := s.Save(path); err != nil {
		t.Fatalf("Save failed when creating parent dirs: %v", err)
	}
	if _, err := os.Stat(path); err != nil {
		t.Errorf("expected file to exist, got: %v", err)
	}
}

func TestSave_JSONHasAliasesKey(t *testing.T) {
	path := filepath.Join(t.TempDir(), "aliases.json")
	s := &Store{Aliases: map[string]Alias{"foo": {Command: []string{"bar"}}}}
	if err := s.Save(path); err != nil {
		t.Fatalf("Save failed: %v", err)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile failed: %v", err)
	}
	var raw map[string]any
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if _, ok := raw["aliases"]; !ok {
		t.Errorf("expected top-level 'aliases' key, got: %s", data)
	}
}

func TestShellLine_BashAndZshIdentical(t *testing.T) {
	a := Alias{Command: []string{"cargo", "clippy"}, User: "alice"}
	bash := ShellLine("clippy", a, ShellBash)
	zsh := ShellLine("clippy", a, ShellZsh)
	want := `alias clippy='with exec --user alice -- cargo clippy'`
	if bash != want {
		t.Errorf("bash:\ngot:  %s\nwant: %s", bash, want)
	}
	if zsh != want {
		t.Errorf("zsh differed from bash:\ngot:  %s\nwant: %s", zsh, want)
	}
}

func TestShellLine_Fish(t *testing.T) {
	a := Alias{Command: []string{"cargo", "clippy"}, User: "alice"}
	got := ShellLine("clippy", a, ShellFish)
	want := `alias clippy 'with exec --user alice -- cargo clippy'`
	if got != want {
		t.Errorf("\ngot:  %s\nwant: %s", got, want)
	}
}

func TestShellLine_EmptyPassword(t *testing.T) {
	a := Alias{
		Command:     []string{"cargo", "clippy"},
		User:        "alice",
		Password:    "",
		PasswordSet: true,
	}
	got := ShellLine("clippy", a, ShellBash)
	want := `alias clippy='with exec --user alice --password "" -- cargo clippy'`
	if got != want {
		t.Errorf("\ngot:  %s\nwant: %s", got, want)
	}
}

func TestShellLine_PasswordFile(t *testing.T) {
	a := Alias{
		Command:      []string{"deploy.sh"},
		User:         "alice",
		PasswordFile: "/home/alice/.vault-pw",
	}
	got := ShellLine("deploy", a, ShellBash)
	want := `alias deploy='with exec --user alice --password-file /home/alice/.vault-pw -- deploy.sh'`
	if got != want {
		t.Errorf("\ngot:  %s\nwant: %s", got, want)
	}
}

func TestShellLine_NoCredentials(t *testing.T) {
	a := Alias{Command: []string{"ls", "-la"}}
	got := ShellLine("ll", a, ShellBash)
	want := `alias ll='with exec -- ls -la'`
	if got != want {
		t.Errorf("\ngot:  %s\nwant: %s", got, want)
	}
}

func TestShellLine_EscapesSpaces(t *testing.T) {
	a := Alias{Command: []string{"echo", "hello world"}}
	got := ShellLine("greet", a, ShellBash)
	want := `alias greet='with exec -- echo "hello world"'`
	if got != want {
		t.Errorf("\ngot:  %s\nwant: %s", got, want)
	}
}

func TestShellLine_EscapesSingleQuoteViaOuter(t *testing.T) {
	a := Alias{Command: []string{"echo", "it's"}}
	got := ShellLine("say", a, ShellBash)
	// The inner double-quoted "it's" contains ', which the outer
	// single-quote wrapper must escape as '\''.
	want := `alias say='with exec -- echo "it'\''s"'`
	if got != want {
		t.Errorf("\ngot:  %s\nwant: %s", got, want)
	}
}

func TestShellLine_EscapesDollarAndBacktick(t *testing.T) {
	a := Alias{Command: []string{"echo", "$HOME", "`date`"}}
	got := ShellLine("x", a, ShellBash)
	want := `alias x='with exec -- echo "\$HOME" "\` + "`" + `date\` + "`" + `"'`
	if got != want {
		t.Errorf("\ngot:  %s\nwant: %s", got, want)
	}
}
