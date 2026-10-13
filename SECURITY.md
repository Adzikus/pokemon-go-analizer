# Security Policy

## Supported Versions

| Version | Supported |
|---------|-----------|
| Latest main branch | ✅ |

## Reporting a Vulnerability

Please report security vulnerabilities by opening a **private GitHub Security Advisory**:

1. Go to the repository's **Security** tab
2. Click **Report a vulnerability**
3. Fill in the details

Do NOT report security issues via public issues.

## Response Timeline

- **Acknowledgment**: Within 48 hours
- **Assessment**: Within 7 days
- **Fix & Release**: Depends on severity, typically within 30 days

## Scope

This is a local CLI tool that:
- Reads/writes local ODS files
- Downloads public data from GitHub (PvPoke gamemaster, rankings)
- Does NOT handle authentication, tokens, or user credentials
- Does NOT expose network services

## Out of Scope

- LibreOffice/ODS parser vulnerabilities (upstream)
- Go standard library vulnerabilities (upstream)
- User's local file permissions

## Disclosure

We follow responsible disclosure. Credit will be given to reporters in the fix commit (unless anonymity is requested).