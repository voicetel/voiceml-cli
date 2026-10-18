# Changelog

## [0.2.1] - 2026-10-18

### Fixed
- **Credential leak in REPL history**: `login` / `set api-key` lines were
  saved to `~/.voiceml/history` in plaintext (mode 0644). History now
  records only recognized read-only commands (list/get/health/openapi
  variants); credential entry, mutations, payment data, unknown commands
  and malformed input are never persisted. The history and config files
  are written with 0600/0700 permissions.
- **`--api-key` and `VOICEML_API_KEY` values are no longer persisted**;
  they apply for the session only, as documented.
- **Dropped JSON filters**: unknown or misspelled filter keys (including
  `StartTime>` / `DateSent<` operator keys) are now rejected with an
  error instead of being silently dropped, so `calls list` can no longer
  return unfiltered results by accident.
- **Piped JSON is never colorized**: color codes are emitted only on a
  TTY, so `voiceml-cli ... | jq` and redirects produce valid JSON.
- **`-h` exits 0** (was 1).
- `diagnostics health` / `openapi` no longer require credentials (the
  endpoints are unauthenticated) and no longer send them.
- The `VOICEMEL_*` env-var spelling (historical typo alias) still works
  and is documented; `VOICEML_*` is canonical and wins when both are set.

### Security
- Added `.github/workflows/security-scan.yml` (workspace Go baseline):
  gosec advisory SARIF artifact + govulncheck fatal on reachable CVEs.
- Release binaries now build with a stable Go toolchain (`release.yml`
  pinned go1.23, which is EOL; the v0.2.0 binary shipped 57 known
  stdlib vulnerabilities).
- Removed the repo-wide gosec excludes (G101/G204/G304) and every
  per-line `//nolint` — underlying findings are fixed, not silenced.

### Docs
- Fictional `555-0100`-range example numbers throughout; README wording
  corrected (VoiceML issues its own credentials; Twilio-issued
  credentials do not work).

### Pending release action (not fixable in-repo)
- `go install github.com/voicetel/voiceml-cli@latest` currently resolves
  to the stale v0.6.6 tag because pre-rename tags outrank the v0.2.x
  line. The next release must be a version above v0.6.6 (with retractions
  for v0.6.3/v0.6.4/v0.6.6) to reclaim `@latest`; until then, install
  via `make install` or a release binary.
