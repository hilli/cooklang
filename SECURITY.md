# Security Policy

## Reporting a Vulnerability

We take security vulnerabilities seriously. If you discover a security issue, please report it responsibly.

### How to Report

**Preferred method:** Use GitHub's private vulnerability reporting feature:

1. Go to the [Security tab](../../security) of this repository
2. Click "Report a vulnerability"
3. Fill out the form with details about the vulnerability

**Alternative:** If private vulnerability reporting is not available, please [open a security issue](../../issues/new?template=security_vulnerability.yml) and avoid including sensitive details publicly. We will follow up to establish a secure communication channel.

### What to Include

When reporting a vulnerability, please include:

- Description of the vulnerability
- Steps to reproduce the issue
- Potential impact assessment
- Any suggested fixes (if applicable)

### What to Expect

- **Acknowledgment:** We will acknowledge receipt of your report within one week
- **Assessment:** We will investigate and assess the severity of the issue
- **Timeline:** We will provide an estimated timeline for a fix after assessment
- **Credit:** We will credit reporters in the release notes (unless you prefer to remain anonymous)

### Please Do Not

- Disclose the vulnerability publicly before it has been addressed
- Exploit the vulnerability beyond what is necessary to demonstrate it
- Access or modify other users' data

## Supported Versions

| Version | Supported          |
| ------- | ------------------ |
| 0.4.x   | :white_check_mark: |
| < 0.4   | :x:                |

## Security Best Practices

When using this library:

- Keep your dependencies up to date
- Validate and sanitize any user-provided recipe input before parsing
- Be cautious when rendering recipe content to HTML (use appropriate escaping)
