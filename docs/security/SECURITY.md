# Security Baseline

- Secrets must never be committed to Git.
- Production credentials are injected by the deployment platform.
- Cloudflare terminates the public edge and provides WAF/rate-limit controls.
- API endpoints use explicit validation and bounded request bodies.
- Database roles follow least privilege.
- Administrative APIs are separated from public APIs.
- User-generated reports/comments require abuse controls.
- Backups and restore procedures must be tested, not merely configured.
- Logs must avoid unnecessary personal data.
- Security-sensitive dependencies are pinned and updated through reviewed pull requests.
