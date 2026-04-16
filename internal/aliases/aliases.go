// Package aliases stores named shortcuts for `with exec` invocations and
// generates the shell lines that activate them.
package aliases

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// Alias is one saved `with exec` shortcut.
type Alias struct {
	Command      []string `json:"command"`
	User         string   `json:"user,omitempty"`
	Password     string   `json:"password,omitempty"`
	PasswordSet  bool     `json:"password_set,omitempty"`
	PasswordFile string   `json:"password_file,omitempty"`
}

// Store is the on-disk collection of aliases.
type Store struct {
	Aliases map[string]Alias `json:"aliases"`
}

// Load reads the store from path. A missing file returns an empty store.
func Load(path string) (*Store, error) {
	data, err := os.ReadFile(path) // #nosec G304 -- path from config, not arbitrary input
	if err != nil {
		if os.IsNotExist(err) {
			return &Store{Aliases: make(map[string]Alias)}, nil
		}
		return nil, fmt.Errorf("reading alias file: %w", err)
	}
	var s Store
	if err := json.Unmarshal(data, &s); err != nil {
		return nil, fmt.Errorf("parsing alias file: %w", err)
	}
	if s.Aliases == nil {
		s.Aliases = make(map[string]Alias)
	}
	return &s, nil
}

// Save writes the store to path with 0600 permissions.
func (s *Store) Save(path string) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return fmt.Errorf("creating alias directory: %w", err)
	}
	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return fmt.Errorf("serializing aliases: %w", err)
	}
	f, err := os.OpenFile(path, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0600) // #nosec G304 -- path from config, not arbitrary input
	if err != nil {
		return fmt.Errorf("creating alias file: %w", err)
	}
	defer f.Close()
	if _, err := f.Write(data); err != nil {
		return fmt.Errorf("writing alias file: %w", err)
	}
	return nil
}

// Set adds or overwrites an alias.
func (s *Store) Set(name string, a Alias) {
	if s.Aliases == nil {
		s.Aliases = make(map[string]Alias)
	}
	s.Aliases[name] = a
}

// Remove deletes an alias. Returns true if the alias existed.
func (s *Store) Remove(name string) bool {
	if _, ok := s.Aliases[name]; !ok {
		return false
	}
	delete(s.Aliases, name)
	return true
}

// Names returns all alias names sorted alphabetically.
func (s *Store) Names() []string {
	names := make([]string, 0, len(s.Aliases))
	for n := range s.Aliases {
		names = append(names, n)
	}
	sort.Strings(names)
	return names
}

// ShellType selects the shell dialect for ShellLine.
type ShellType int

const (
	// ShellBash generates bash-compatible `alias name='...'` lines.
	ShellBash ShellType = iota
	// ShellZsh generates zsh-compatible lines (identical syntax to bash).
	ShellZsh
	// ShellFish generates fish-compatible `alias name '...'` lines.
	ShellFish
)

// ShellLine returns a shell command that defines the given alias.
func ShellLine(name string, a Alias, shell ShellType) string {
	parts := []string{"with", "exec"}
	if a.User != "" {
		parts = append(parts, "--user", quoteArg(a.User))
	}
	if a.PasswordSet {
		parts = append(parts, "--password", quoteArg(a.Password))
	}
	if a.PasswordFile != "" {
		parts = append(parts, "--password-file", quoteArg(a.PasswordFile))
	}
	parts = append(parts, "--")
	for _, c := range a.Command {
		parts = append(parts, quoteArg(c))
	}
	inner := strings.Join(parts, " ")

	switch shell {
	case ShellFish:
		return fmt.Sprintf("alias %s %s", name, outerQuote(inner))
	default:
		return fmt.Sprintf("alias %s=%s", name, outerQuote(inner))
	}
}

// quoteArg double-quotes s if it has shell-special chars or is empty.
// Inside the quotes, the chars ", $, \, and ` are backslash-escaped so the
// enclosing shell treats them literally.
func quoteArg(s string) string {
	if s == "" {
		return `""`
	}
	if isSafe(s) {
		return s
	}
	var b strings.Builder
	b.Grow(len(s) + 2)
	b.WriteByte('"')
	for _, r := range s {
		switch r {
		case '"', '$', '\\', '`':
			b.WriteByte('\\')
		}
		b.WriteRune(r)
	}
	b.WriteByte('"')
	return b.String()
}

// isSafe reports whether s contains only chars that never need shell quoting.
func isSafe(s string) bool {
	for _, r := range s {
		switch {
		case r >= 'a' && r <= 'z':
		case r >= 'A' && r <= 'Z':
		case r >= '0' && r <= '9':
		case r == '_' || r == '-' || r == '/' || r == '.' || r == ':' || r == '=' || r == ',' || r == '@' || r == '+':
		default:
			return false
		}
	}
	return true
}

// outerQuote wraps s in single quotes, escaping any internal single quote
// with the POSIX idiom '\''.
func outerQuote(s string) string {
	return "'" + strings.ReplaceAll(s, "'", `'\''`) + "'"
}
