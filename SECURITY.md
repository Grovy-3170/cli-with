# Security Policy

## Supported Versions

The following versions of `with` are currently supported with security updates:

| Version | Supported          |
| ------- | ------------------ |
| 0.1.1+  | :white_check_mark: |

## Reporting a Vulnerability

We take security vulnerabilities seriously. If you discover a security issue, please report it responsibly.

### How to Report

1. **DO NOT** create a public GitHub issue for security vulnerabilities
2. Email security reports to [amanak3170@gmail.com](mailto:amanak3170@gmail.com)
3. Include the following information:
   - Description of the vulnerability
   - Steps to reproduce the issue
   - Potential impact assessment
   - Any suggested fixes (optional)

### What to Expect

- **Acknowledgment**: We will acknowledge your report within 48 hours
- **Initial Assessment**: We will provide an initial assessment within 7 days
- **Regular Updates**: We will provide updates on the progress of fixing the vulnerability
- **Disclosure**: We will coordinate disclosure with you before any public announcement

### Response Timeline

| Phase                      | Timeline            |
| -------------------------- | ------------------- |
| Initial Response           | Within 48 hours    |
| Severity Assessment        | Within 7 days      |
| Fix Development            | Based on severity  |
| Security Advisory Release  | After fix is ready |

### Severity Classification

- **Critical**: Patch released within 24-48 hours
- **High**: Patch released within 7 days
- **Medium**: Patch released within 30 days
- **Low**: Patch released in next scheduled release

## Security Features

- Argon2id key derivation (64MiB memory, t=3, p=1)
- AES-256-GCM encryption with random nonces
- OS keychain integration for credential storage
- Memory locking to prevent swapping
- File permissions enforcement (0600)
- Subprocess-only environment variable injection

## Security Best Practices

### Password Management

- **Use strong master passwords**: At least 12 characters with mixed case, numbers, and symbols
- **Never share master passwords**: Each user should have their own vault
- **Store master passwords securely**: Use a password manager like 1Password, Bitwarden, or KeePassXC

### File Security

- **Verify file permissions**: Vault files should have `0600` permissions, directories `0700`
- **Never commit vault files**: Add `*.vault` files to `.gitignore`
- **Avoid password files in production**: If using `--password-file`, set permissions to `0600` and delete after use

### Key Management

- **Regularly rotate API keys**: Update keys periodically to limit exposure from potential leaks
- **Remove unused keys**: Periodically review and delete deprecated keys
- **Use minimal key permissions**: Request only the permissions your application needs

### Environment Safety

- **Prefer `with exec`**: This isolates keys to the subprocess environment
- **Avoid `with get` in scripts**: This prints keys to stdout which may be logged
- **Be aware of process listing**: Other users on shared systems may see running processes

### Shared Server Security

- Each user has a separate vault file with restricted permissions
- File permissions prevent cross-user access
- Consider OS keychain integration for additional protection
- Monitor for unauthorized access attempts
