# Security Policy

## Reporting a Vulnerability

If you discover a security vulnerability, please report it privately by emailing [blank.xu@outlook.com] or opening a GitHub Security Advisory. We will respond as quickly as possible and keep your identity confidential.

## Supported Versions

We support the latest major version of the project. Older versions may not receive security updates.

## Dependency Security

We use automated tools to monitor for vulnerabilities in our Go dependencies. Our policies include:

- All dependencies are managed using Go modules.
- We scan dependencies using `govulncheck` (official Go tool) during CI runs.
- Dependencies with critical or high vulnerabilities must be updated or replaced before merging changes.

## Best Practices for Contributors

- Only use actively maintained and well-reviewed third-party packages.
- Run `go mod tidy` before submitting PRs to ensure module integrity.
- Run `govulncheck ./...` locally to check for known vulnerabilities.
