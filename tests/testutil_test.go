package tests

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

// buildBinary builds the with CLI binary and returns its path.
// The binary is built once per test process and cached for reuse.
func buildBinary(t *testing.T) string {
	t.Helper()

	if cachedBinary != "" {
		if _, err := os.Stat(cachedBinary); err == nil {
			return cachedBinary
		}
	}

	tempDir, err := os.MkdirTemp("", "cli-with-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}

	binaryPath := filepath.Join(tempDir, "with")

	cmd := exec.Command("go", "build", "-buildvcs=false", "-o", binaryPath, "./cmd/with")
	cmd.Dir = projectRoot

	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Failed to build binary: %v\n%s", err, output)
	}

	if _, err := os.Stat(binaryPath); err != nil {
		t.Fatalf("Binary not created at %s: %v", binaryPath, err)
	}

	cachedBinary = ""

	return binaryPath
}

var cachedBinary string

var projectRoot string

func init() {
	wd, err := os.Getwd()
	if err != nil {
		panic("Failed to get working directory: " + err.Error())
	}

	dir := wd
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			projectRoot = dir
			return
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			panic("Could not find project root (go.mod)")
		}
		dir = parent
	}
}
