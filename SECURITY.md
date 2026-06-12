—✅❌—––––————·# Security Policy

> Raaz is a platform for anonymous, emotionally honest conversations. Security and user privacy are foundational — not afterthoughts.

---

## Supported Versions

Only the latest release on the `main` branch receives security patches.

| Version | Supported |
|---------|----------|
| `main` (latest) | Actively supported |
| Older commits | No longer supported |

---

## Reporting a Vulnerability

**Do NOT open a public GitHub Issue for security vulnerabilities.** Public disclosure before a fix is available puts all users at risk.

### Preferred Channel: GitHub Private Security Advisory

1. Go to [Security Advisories](https://github.com/rish106-hub/Raaz/security/advisories)
2. Click **New draft security advisory**
3. Fill in: affected component, severity, steps to reproduce, and potential impact
4. Submit - only maintainers can see this

### Alternate Channel: Email

Send a detailed report to: **`arthakram@rishihood.edu.in`**

Include:
- Description of the vulnerability
- Steps to reproduce
- Potential impact and affected users
- Your suggested fix (optional)
- Whether you want public credit

---

## Response SLA

| Milestone | Target Time |
|-----------|------------|
| Acknowledgement | Within 48 hours |
| Triage and severity assessment | Within 5 business days |
| Patch released (critical) | Within 14 days |
| Patch released (high/medium) | Within 30 days |
| Public disclosure | After patch is live |

---

## Severity Classification

We follow the CVSS v3.1 scoring system:

| Severity | Score | Examples |
|----------|-------|---------|
| Critical | 9.0-10.0 | Auth bypass, mass data exposure, RCE |
| High | 7.0-8.9 | User PII leak, privilege escalation |
| Medium | 4.0-6.9 | Limited data exposure, CSRF |
| Low | 0.1-3.9 | Non-sensitive info disclosure |

---

## Scope

### In Scope
- Android app (`app/`)
- Go backend services (`server/`)
- Authentication and session management
- Matching engine and user pairing logic
- Data storage, encryption, and deletion flows
- Firebase / push notification pipeline
- CI/CD pipelines (`.github/workflows/`)

### Out of Scope
- Third-party services (Firebase, Cloudflare, etc.) - report to them directly
- Social engineering attacks
- Physical access attacks
- Denial of service (DoS) via resource exhaustion
- Known dependency vulnerabilities (check Dependabot alerts)

---

## Our Security Commitments

- TLS 1.3 enforced for all data in transit
- AES-256 encrypted storage at rest
- Daily key rotation for session secrets
- Zero ad-tracking SDKs - no external analytics
- DPDP Act 2023 compliant - explicit consent flows, data deletion on request
- 48-hour auto-deletion of conversation content
- No conversation data sold or shared with third parties
- Regular dependency audits via GitHub Dependabot

---

## Responsible Disclosure Policy

We follow coordinated disclosure:

1. You report privately
2. We acknowledge within 48 hours
3. We work on a fix
4. We release the fix
5. We publicly credit you (unless you prefer anonymity)
6. CVE is filed if applicable

We will not take legal action against researchers who:
- Report in good faith through the channels above
- Do not access or modify other users data
- Do not perform DoS attacks
- Do not publicly disclose before we patch

---

## Hall of Fame

Security researchers who responsibly disclose valid vulnerabilities will be credited here.

*(No entries yet - be the first!)*

---

*Raaz is built under [Arthakram](https://github.com/Arthakram) - Rishihood University*
