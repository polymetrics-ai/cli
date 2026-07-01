# Security Policy

## Reporting Vulnerabilities

Please do not open public issues for vulnerabilities.

Use GitHub private vulnerability reporting for this repository when available. If that is unavailable, email `security@polymetrics.ai` with:

- affected version or commit
- reproduction steps
- expected impact
- any known mitigations

We aim to acknowledge reports within 3 business days and provide a status update within 10 business days.

## Scope

Security-sensitive areas include credential storage, connector auth handling, local warehouse files, reverse-ETL write approval, generated docs, runtime orchestration, and CI/CD workflows.

## Supported Versions

Until the first stable release, security fixes target the `main` branch. After tagged releases begin, this file will list supported release lines.
