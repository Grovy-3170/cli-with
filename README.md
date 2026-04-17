# with

[![Go Version](https://img.shields.io/badge/go-1.23+-00ADD8?style=flat&logo=go)](https://golang.org/dl/)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

**Run any command *with* your secrets. No leaks. No drama.**

```bash
with exec -- curl -H "Authorization: Bearer $OPENAI_API_KEY" https://api.openai.com/v1/models
```

Your API keys stay encrypted, never touch your shell history, and vanish after the command runs.

## Why?

- 🔐 **Your secrets, your vault** — encrypted locally with a master password
- 🚀 **Zero config exports** — no more `export API_KEY=...` in your bashrc
- 👥 **Shared server friendly** — each user gets their own isolated vault
- 🧹 **Clean environment** — secrets exist only in the subprocess, nowhere else

## Features

- **Encryption**: Argon2id + AES-256-GCM (industry standard)
- **Per-user vaults**: Perfect for shared machines — your keys, your vault
- **Minimal env inheritance**: Only essential vars (PATH, HOME, etc.) pass through
- **Cross-platform**: Linux, macOS, Windows

## Installation

### Pre-built Binary (recommended)

No Go toolchain required. Pick the one-liner for your platform:

**macOS (Apple Silicon):**
```bash
curl -L https://github.com/Grovy-3170/cli-with/releases/latest/download/with_Darwin_arm64.tar.gz | tar -xz with && sudo mv with /usr/local/bin/
```

**macOS (Intel):**
```bash
curl -L https://github.com/Grovy-3170/cli-with/releases/latest/download/with_Darwin_amd64.tar.gz | tar -xz with && sudo mv with /usr/local/bin/
```

**Linux (x86_64):**
```bash
curl -L https://github.com/Grovy-3170/cli-with/releases/latest/download/with_Linux_amd64.tar.gz | tar -xz with && sudo mv with /usr/local/bin/
```

**Linux (ARM64):**
```bash
curl -L https://github.com/Grovy-3170/cli-with/releases/latest/download/with_Linux_arm64.tar.gz | tar -xz with && sudo mv with /usr/local/bin/
```

**Windows:** Download the `.zip` for your architecture from the [Releases page](https://github.com/Grovy-3170/cli-with/releases), extract `with.exe`, and add its directory to your `PATH`.

Verify with `with version`.

On macOS, the first run may trigger a Gatekeeper warning — right-click the binary and choose "Open" once, or run `xattr -d com.apple.quarantine /usr/local/bin/with`.

### Go Install

If you already have Go installed:

```bash
go install github.com/Grovy-3170/cli-with/cmd/with@latest
```

> **Note:** This installs to `$(go env GOPATH)/bin`. Add it to your PATH:
> ```bash
> # Add to ~/.zshrc or ~/.bashrc
> export PATH="$PATH:$(go env GOPATH)/bin"
> ```
> Then run `source ~/.zshrc` or restart your terminal.

### Build from Source

For contributors or anyone who wants the tip of `main`:

```bash
git clone https://github.com/Grovy-3170/cli-with.git
cd cli-with
make build
```

The binary will be created at `./with`. Install it system-wide with:

```bash
make install
```

### Updating

Re-run the same install command you used the first time. For binary installs, re-running the `curl … | tar … && sudo mv …` one-liner replaces the old binary in place.

### Uninstalling

```bash
# If installed via pre-built binary
sudo rm /usr/local/bin/with

# If installed via go install
rm $(go env GOPATH)/bin/with

# If built from source
make uninstall
```

## Quick Start

### 1. Initialize Your Vault

**Interactive mode** (prompts for username):
```bash
with init
```

**Explicit username**:
```bash
with init --user alice
```

You'll be prompted to create a master password. This password encrypts all your API keys.

### 2. Store an API Key

```bash
with set --user alice OPENAI_API_KEY
```

or 

```bash
with set OPENAI_API_KEY
```

You'll be prompted to enter the key value securely (hidden input).

Alternatively, provide the value directly:

```bash
with set --user alice ANTHROPIC_API_KEY --value "your-api-key-here"
```

### 3. List Your Keys

```bash
with list
```

or

```bash
with list --user alice
```

This shows the names of all stored keys (not their values).

### 4. Use Keys in Commands

```bash
with exec -- curl -H "Authorization: Bearer $OPENAI_API_KEY" https://api.openai.com/v1/models
```

With explicit user:
```bash
with exec --user alice -- curl -H "Authorization: Bearer $OPENAI_API_KEY" https://api.openai.com/v1/models
```

The `$OPENAI_API_KEY` environment variable is available inside the command, but never exposed to your shell history or parent process.

### 5. Get a Key Value

```bash
with get --user alice OPENAI_API_KEY
```

### 6. Remove a Key or Vault

Remove a specific key:

```bash
with remove --user alice OLD_API_KEY
```

Remove entire vault:

```bash
with remove --user alice
```

## Aliases

Save frequently-used `with exec` invocations as named shortcuts, then activate
them as native shell commands with one line in your shell config.

### Save an alias

```bash
with alias add <name> [--user <user>] [--password <pw>] [--password-file <path>] -- <command> [args...]
```

The global flags (`--user`, `--password`, `--password-file`) are captured at
save time. Everything after `--` is the command the alias will run.

### List and remove

```bash
with alias list
with alias remove <name>
```

### Activate in your shell

Add to `~/.bashrc` or `~/.zshrc`:

```bash
eval "$(with alias shell)"
```

For fish, add to `~/.config/fish/config.fish`:

```bash
with alias shell --shell fish | source
```

After reloading your shell, each saved alias runs as a native command. The
shell is auto-detected from `$SHELL`; override with `--shell bash|zsh|fish`.

### Storage

Aliases are stored in plain JSON at `~/.config/cli-with/aliases.json`, or at
the path set in `WITH_ALIAS_FILE` if defined.

## Usage

### Global Flags

| Flag | Description |
|------|-------------|
| `--user string` | Username for the vault (prompts interactively if not provided) |
| `--password string` | Vault password (use for scripting; takes precedence over `--password-file`) |
| `--password-file string` | Path to file containing the vault password |

### Commands

#### `with init`

Initialize a new user vault.

```bash
with init              # Interactive - prompts for username
with init --user <username>
```

Creates an encrypted vault file at `~/.config/cli-with/users/<username>.vault`.

#### `with set`

Add or update an API key.

```bash
with set --user <username> <KEY_NAME>
with set --user <username> <KEY_NAME> --value "secret-value"
```

Key names must start with a letter or underscore, followed by letters, digits, or underscores.

#### `with get`

Retrieve the value of a specific key.

```bash
with get --user <username> <KEY_NAME>
```

#### `with list`

List all key names stored in the vault.

```bash
with list --user <username>
```

#### `with exec`

Execute a command with all keys available as environment variables.

```bash
with exec -- <command> [args...]
with exec --user <username> -- <command> [args...]
```

The `--` separator is required to distinguish command arguments from flags.

Examples:

```bash
# Run a script with your secrets
with exec -- python my_script.py

# With explicit user
with exec --user alice -- python my_script.py

# Use with curl
with exec -- curl -H "X-API-Key: $MY_API_KEY" https://api.example.com

# Chain commands
with exec -- sh -c 'echo $OPENAI_API_KEY | wc -c'
```

#### `with remove`

Remove a specific key or the entire vault.

```bash
# Remove a specific key
with remove --user <username> <KEY_NAME>

# Remove entire vault (prompts for confirmation)
with remove --user <username>
```

#### `with alias`

Save, list, and remove shortcuts for `with exec`, and emit shell lines that
activate them as native aliases.

```bash
with alias add <name> [--user <user>] [--password <pw>] [--password-file <path>] -- <command> [args...]
with alias list
with alias remove <name>
with alias shell [--shell bash|zsh|fish]
```

See the [Aliases](#aliases) section for details.

#### `with version`

Print the version number.

```bash
with version
```

### Automation with Passwords

For CI/CD or automation, avoid interactive prompts using either `--password` or `--password-file`.

**Inline password** (simplest):

```bash
with --user alice --password "my-secure-password" list
```

**Password file** (preferred for shared environments):

```bash
# Store password securely in a temp file (restricted permissions)
echo "my-secure-password" > /tmp/vault-password
chmod 600 /tmp/vault-password

# Use in commands
with --user alice --password-file /tmp/vault-password list

# Clean up
rm /tmp/vault-password
```

`--password` takes precedence over `--password-file` if both are provided.

**Warning**: Password files should have strict permissions (0600) and be deleted after use. Prefer `--password-file` over `--password` when the password may appear in process listings.

### Custom Vault Location

Set the `WITH_VAULT_DIR` environment variable to change where vaults are stored:

```bash
export WITH_VAULT_DIR=/secure/vault-location
with init --user alice
```

## Security Considerations

### Password Best Practices

1. **Use strong master passwords**: At least 12 characters with mixed case, numbers, and symbols
2. **Never share master passwords**: Each user should have their own vault
3. **Avoid password files in production**: Use only in secure, automated environments with proper access controls
4. **Consider a password manager**: Store your master password in a password manager like 1Password or Bitwarden

### File Permissions

The vault directory (`~/.config/cli-with/users/`) is created with `0700` permissions. Vault files have `0600` permissions. Verify these permissions:

```bash
ls -la ~/.config/cli-with/users/
```

If permissions are incorrect, fix them:

```bash
chmod 700 ~/.config/cli-with/users/
chmod 600 ~/.config/cli-with/users/*.vault
```

### Environment Isolation

When using `with exec` or `with <command>`:

- Keys are injected only into the subprocess environment
- The parent shell never sees the key values
- Keys don't appear in shell history
- Only minimal environment variables (`PATH`, `HOME`, `USER`, `SHELL`, `TMPDIR`) are inherited

This prevents accidental key exposure through:
- Shell history logging
- Process listing (`ps eww`)
- Environment dumps
- Debug output

### Key Rotation

Regularly rotate your API keys:

```bash
# Update a key
with set --user alice EXISTING_KEY --value "new-secret-value"

# Remove old keys you no longer need
with remove --user alice DEPRECATED_KEY
```

### Shared Server Considerations

On shared servers:

- Each user has a separate vault file
- File permissions prevent users from reading each other's vaults
- The master password is required to decrypt any vault
- Consider using OS keychain integration for additional security

## Troubleshooting

### "vault for user 'X' already exists"

You're trying to initialize a vault that already exists. Either use a different username or remove the existing vault:

```bash
with remove --user X
with init --user X
```

### "vault for user 'X' does not exist"

Run `with init --user X` first to create the vault.

### "passwords do not match"

During `with init`, the password and confirmation must match exactly. Try again.

### "incorrect password" or "decrypting vault: ..."

The master password you entered is incorrect. Try again carefully. If you've forgotten your password, there's no recovery option. You must delete the vault and reinitialize:

```bash
with remove --user alice
with init --user alice
```

### "keyring unavailable"

On Linux, the OS keychain requires a running secret service (like GNOME Keyring or KDE Wallet). If unavailable, the vault falls back to file-based encrypted storage.

### Key names with special characters

Key names must follow these rules:
- Start with a letter (a-z, A-Z) or underscore (_)
- Followed by letters, digits (0-9), or underscores

Valid: `API_KEY`, `_secret`, `openai2`
Invalid: `2FAST`, `my-key`, `api.key`

### Permission denied errors

Ensure you have write access to the vault directory:

```bash
# Check directory permissions
ls -la ~/.config/cli-with/

# Fix if needed
chmod 700 ~/.config/cli-with
chmod 700 ~/.config/cli-with/users
```

### Command not found in exec

When using `with exec`, the command must be in `PATH`. Use the full path if needed:

```bash
with exec --user alice -- /usr/bin/curl https://example.com
```

### Environment variables not expanded in shell

When using shell variables in commands, wrap with `sh -c`:

```bash
# This won't work - $VAR is expanded by parent shell
with exec --user alice -- echo $MY_KEY

# This works - variable is expanded in subprocess
with exec --user alice -- sh -c 'echo $MY_KEY'
```

## Contributing

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/my-feature`)
3. Make your changes
4. Run tests (`make test`)
5. Run linters (`make lint`)
6. Commit your changes
7. Push to the branch
8. Open a Pull Request

### Development Setup

```bash
git clone https://github.com/Grovy-3170/cli-with.git
cd cli-with
go mod download
make test
```

## License

MIT License - see [LICENSE](LICENSE) file for details.
