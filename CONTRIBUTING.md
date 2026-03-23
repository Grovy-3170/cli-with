# Contributing to cli-with

Thanks for your interest in contributing! This document covers everything you need to get started.

## Development Setup

### Prerequisites

- **Go 1.23+** (we use Go 1.25)
- **Make** (optional, but recommended)
- **golangci-lint** for linting

Install golangci-lint:

```bash
# macOS
brew install golangci-lint

# Linux
curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(go env GOPATH)/bin

# Windows
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
```

### Cloning and Building

```bash
# Clone your fork
git clone https://github.com/YOUR_USERNAME/cli-with.git
cd cli-with

# Install dependencies
go mod download

# Build the binary
make build
# Or: go build -o with ./cmd/with

# Verify it works
./with version
```

### Running Tests

```bash
# Run all tests
make test

# Run with race detector
make test-race

# Run specific package tests
go test -v ./internal/crypto/...

# Run integration tests (these build and execute the binary)
go test -v ./tests/...
```

## Project Structure

```
cli-with/
├── cmd/with/           # Main application entry point
├── internal/
│   ├── crypto/         # Encryption and key derivation
│   └── storage/        # Vault storage and keychain integration
├── tests/              # Integration and security tests
├── Makefile            # Build and test automation
├── go.mod              # Go module definition
└── .golangci.yml       # Linter configuration
```

## Code Style

We follow standard Go conventions with a few specifics:

### Formatting

Run the formatter before committing:

```bash
make fmt
# Or: gofmt -w .
```

### Linting

We use golangci-lint with these linters enabled:

- `errcheck` - Check unchecked errors
- `gosimple` - Simplify code
- `govet` - Vet examines source
- `ineffassign` - Detect ineffective assignments
- `staticcheck` - Static analysis
- `unused` - Check unused code
- `gofmt` - Format check
- `goimports` - Import formatting
- `misspell` - Spell check
- `revive` - Fast linter
- `unconvert` - Detect unnecessary type conversions
- `gocyclo` - Cyclomatic complexity (max 15)

Run linters:

```bash
make lint
# Or: golangci-lint run ./...
```

### Code Guidelines

1. **Handle all errors** - Never ignore errors. Use `_` only when explicitly ignoring.
2. **Write tests** - New features should include unit tests.
3. **Keep functions focused** - Functions should do one thing well.
4. **Document exported items** - Add doc comments to exported types and functions.
5. **No external dependencies without discussion** - Keep the dependency tree minimal.

## Commit Messages

We follow [Conventional Commits](https://www.conventionalcommits.org/):

```
<type>(<scope>): <description>

[optional body]

[optional footer(s)]
```

### Types

| Type | Description |
|------|-------------|
| `feat` | New feature |
| `fix` | Bug fix |
| `docs` | Documentation changes |
| `style` | Formatting, missing semicolons, etc. (no code change) |
| `refactor` | Code refactoring |
| `test` | Adding or updating tests |
| `chore` | Build process, dependencies, tooling |

### Examples

```
feat(exec): add support for custom environment variables

fix(vault): handle corrupted vault file gracefully

docs(readme): update installation instructions

test(crypto): add benchmarks for argon2id key derivation
```

### Commit Message Guidelines

1. Use the imperative mood ("add feature" not "added feature")
2. First line should be 50 characters or less
3. Body should explain the "why", not the "what"
4. Reference issues and pull requests in the footer

## Pull Request Process

### Before Submitting

1. **Run tests**: `make test`
2. **Run linters**: `make lint`
3. **Format code**: `make fmt`
4. **Update documentation** if needed
5. **Add tests** for new features or bug fixes

### Submitting a PR

1. Fork the repository
2. Create a feature branch from `main`:
   ```bash
   git checkout -b feat/my-new-feature
   ```
3. Make your changes
4. Commit with conventional commit messages
5. Push to your fork
6. Open a Pull Request

### PR Title

Use the same format as commit messages:

```
feat(vault): add key rotation support
```

### PR Description

Include:

- **What**: Brief description of changes
- **Why**: Motivation for the change
- **How**: Implementation details if relevant
- **Testing**: How you tested the changes
- **Breaking Changes**: If any

### Review Criteria

PRs are reviewed against these criteria:

- [ ] Tests pass (`make test`)
- [ ] Linters pass (`make lint`)
- [ ] Code follows project style
- [ ] New code has tests
- [ ] Documentation updated if needed
- [ ] No unnecessary dependencies added
- [ ] Commit messages follow conventional commits

### After Review

- Address review feedback with new commits (not force pushes)
- Keep the PR up to date with `main`:
  ```bash
  git fetch origin
  git rebase origin/main
  ```

## Testing Guidelines

### Unit Tests

Place unit tests in the same package as the code they test:

```
internal/crypto/aesgcm.go
internal/crypto/aesgcm_test.go
```

### Integration Tests

Integration tests go in `tests/`. These test the full application by building and running the binary.

### Test Naming

```go
func TestFunctionName_Scenario_ExpectedResult(t *testing.T) {}
```

Example:

```go
func TestEncrypt_ValidKey_Succeeds(t *testing.T) {}
func TestEncrypt_InvalidKey_ReturnsError(t *testing.T) {}
```

### Table-Driven Tests

Prefer table-driven tests for multiple scenarios:

```go
func TestValidateKeyName(t *testing.T) {
    tests := []struct {
        name    string
        key     string
        wantErr bool
    }{
        {"valid", "API_KEY", false},
        {"valid underscore prefix", "_SECRET", false},
        {"invalid number prefix", "2FAST", true},
        {"invalid hyphen", "my-key", true},
    }
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := ValidateKeyName(tt.key)
            if (err != nil) != tt.wantErr {
                t.Errorf("ValidateKeyName(%q) error = %v, wantErr %v", tt.key, err, tt.wantErr)
            }
        })
    }
}
```

## Getting Help

- Open a [Discussion](https://github.com/Grovy-3170/cli-with/discussions) for questions
- Open an [Issue](https://github.com/Grovy-3170/cli-with/issues) for bugs or feature requests

## License

By contributing, you agree that your contributions will be licensed under the MIT License.
