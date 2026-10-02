# Security Policy

bluefin-cli is a system management tool. We take security reports seriously.

## Reporting

**Do not open a public issue.** Use GitHub's private vulnerability reporting:

1. Go to [bluefin-cli security advisories](https://github.com/tuna-os/bluefin-cli/security/advisories/new).
2. Click **Report a vulnerability**.
3. Include:
   - Affected component
   - Steps to reproduce
   - Impact and severity
   - Suggested fix, if known

## In scope

- Privilege escalation
- Command injection
- Unsafe system operations
- Credential leakage
- Dependency vulnerabilities
- Supply chain issues

## Out of scope

- Fedora Silverblue — report to [Fedora Security](https://docs.fedoraproject.org/en-US/security/)
- systemd — report to [systemd](https://github.com/systemd/systemd/security/advisories)
- Container runtimes — report upstream
- Third-party tools — report to the tool maintainer

## Timeline

- Acknowledge: 5 business days
- Triage: 10 days
- Fix and release: 30 days
- Prefer coordinated disclosure (30 days before public details)

## After a fix

- Publish GitHub Security Advisory
- Update CHANGELOG
- Notify users via release notes
- Credit reporter if requested

## Boundaries

- bluefin-cli: command parsing and configuration
- System: privilege escalation and package management
- User: credential storage and permissions

Always use `sudo` explicitly. Do not run bluefin-cli commands with `sudo` by default.

## Questions

Open a [public issue](https://github.com/tuna-os/bluefin-cli/issues) or contact the maintainers.
