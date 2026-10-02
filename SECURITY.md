# Security Policy

bluefin-cli is a command-line tool for managing bluefin systems. It interfaces with system-level operations and manages user environments. Security vulnerabilities could impact system stability and user security.

## Reporting a vulnerability

**Do not open a public issue for security vulnerabilities.** Instead, use GitHub's private vulnerability reporting:

1. Go to the [bluefin-cli security advisories page](https://github.com/tuna-os/bluefin-cli/security/advisories/new).
2. Click **Report a vulnerability** and draft a security advisory.
3. Include:
   - Affected component (command, subcommand, or feature)
   - Minimal steps to reproduce
   - Impact and severity (privilege escalation, system state corruption, data leakage, etc.)
   - Suggested fix, if you have one

Alternatively, email the maintainers through a GitHub Security Advisory draft if the affected version is not yet released.

## What to report

Report vulnerabilities in:

- **Privilege escalation**: Commands executing with unintended elevated privileges or allowing unprivileged users to perform privileged operations
- **Command injection**: Unsafe handling of arguments, environment variables, or user input that could lead to arbitrary command execution
- **Unsafe system operations**: Improper error handling, unsafe file operations, or race conditions that could corrupt system state
- **Credential leakage**: Insecure storage or transmission of authentication tokens, API keys, or credentials
- **Dependency vulnerabilities**: Vulnerable dependencies or unsafe dependency chains
- **Supply chain issues**: Unsigned artifacts, unpinned dependencies, or tampered releases

## Out of scope

Do not report vulnerabilities in:

- **Fedora Silverblue/Fedora CoreOS**: The base operating system, ostree, or system services. Report to [Fedora Security](https://docs.fedoraproject.org/en-US/security/).
- **systemd**: The init system and system manager. Report to [systemd project](https://github.com/systemd/systemd/security/advisories).
- **Container runtimes**: Podman, containers, or OCI runtimes. Report to the relevant upstream project.
- **Third-party tools**: Utilities or services that bluefin-cli integrates with. Report to the tool's maintainers.

## Response timeline

We aim to:

- **Acknowledge** your report within **5 business days**
- **Triage** the vulnerability and confirm reproducibility within **10 days**
- **Fix and release** a patch through the normal release pipeline, usually within **30 days** of triage
- **Coordinate disclosure**: we prefer coordinated disclosure. Please give us a reasonable window (at least 30 days after we confirm a fix is ready) before publishing details publicly

## Disclosure

Once a fix is released, we will:

1. Publish a GitHub Security Advisory with CVE details (if applicable)
2. Add an entry to the CHANGELOG with the version and fix summary
3. Notify users through release notes and distribution channels

If you discovered the vulnerability responsibly and wish to be credited, tell us in your report.

## Security boundaries

bluefin-cli operates at the user level but can invoke system-level operations:

- **bluefin-cli handles**: Command-line parsing, user configuration, state management
- **System handles**: Privilege escalation via sudo, system package managers, ostree operations
- **User handles**: Secure credential storage, careful command execution, system permissions

Always use `sudo` explicitly when system-level operations require elevation. Never run bluefin-cli commands with `sudo` by default.

## Questions or feedback

If you have questions about this policy or feedback on bluefin-cli's security posture, open a public issue in the [bluefin-cli tracker](https://github.com/tuna-os/bluefin-cli/issues) or reach out to the maintainers.
