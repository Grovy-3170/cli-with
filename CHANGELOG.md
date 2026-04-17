# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [v0.3.2] - 2026-04-17

### Changed

- **Release archive filenames** no longer include the version (e.g., `with_Darwin_arm64.tar.gz` instead of `with_0.3.1_Darwin_arm64.tar.gz`). This enables stable `releases/latest/download/…` URLs, so the install one-liner in the README works without inline version detection. Older releases keep their original filenames.
- **README** reordered: pre-built binary install is now the primary path; `go install` and build-from-source are listed after.

## [v0.3.1] - 2026-04-17

### Added

- **Pre-built binaries**: Automated releases via GoReleaser. Every `v*` tag now produces archives for Linux, macOS, and Windows (amd64 + arm64) on the [Releases page](https://github.com/Grovy-3170/cli-with/releases). No Go toolchain required to install.

## [v0.3.0] - 2026-04-17

### Added

- **`with alias` command**: Save named shortcuts for `with exec` invocations with `with alias add`, list and remove them, and emit shell lines via `with alias shell` that activate all saved shortcuts as native aliases. Supports bash, zsh, and fish. Activate with a single `eval "$(with alias shell)"` line in your shell config.
- Aliases are stored at `~/.config/cli-with/aliases.json` (overridable via `WITH_ALIAS_FILE`).

## [v0.2.0] - 2026-04-15

### Added

- **`--password` flag**: Pass the vault password directly on the command line for scripting and automation. Takes precedence over `--password-file`.
- **Empty password support**: Vaults can now be initialized and used with an empty password, enabling passwordless local setups.

## [v0.1.1] - 2026-03-23

### Added

- **Secure Encryption**: Argon2id key derivation for password hashing
- **AES-256-GCM Encryption**: Industry-standard symmetric encryption for stored keys
- **Vault Management**: Initialize, set, get, list, and remove API keys
- **Secure Execution**: `exec` command to run commands with keys as environment variables
- **Environment Isolation**: Keys are only available in subprocess, not parent shell
- **OS Keychain Integration**: Uses system keychain (macOS Keychain, Windows Credential Manager, Linux Secret Service)
- **File-based Fallback**: Encrypted vault when keychain is unavailable
- **Password File Support**: Automation support via `--password-file` flag
- **Cross-platform Support**: Linux, macOS, and Windows

### Commands

- `with init` - Initialize a new user vault
- `with set` - Set or update an API key
- `with get` - Get the value of a specific key
- `with list` - List all keys for a user
- `with remove` - Remove a key or entire user vault
- `with exec` - Execute a command with keys as environment variables
- `with version` - Print the version number

### Security

- Minimal environment variable inheritance (only PATH, HOME, USER, SHELL, TMPDIR)
- Secure password input (hidden terminal input)
- File permissions: 0700 for vault directory, 0600 for vault files
- No key exposure to shell history or parent process
