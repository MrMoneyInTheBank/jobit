# Release verification — 12 October 2026

## Current local checks

- `scripts/check.sh all`, `go build ./cmd/jobit ./cmd/seed`, and
  `go test -race ./...` pass on macOS with `GOCACHE=/tmp/jobit-go-cache`.
- Regression tests cover rejection without data loss for unversioned databases,
  preservation of records during the version-1 column rename, and CSV export
  against a legacy database.
- GitHub Actions now checks macOS and Linux. The updated workflow needs a run
  on the release commit before release verification is complete.

## Passed locally

- Formatting, go vet ./..., and go test ./... through scripts/check.sh.
- CGO_ENABLED=1 go test -race ./...; no races reported. macOS linker
  emitted LC_DYSYMTAB warnings, but all packages passed.
- Existing regression coverage includes SQLite round trips and errors,
  transaction rollback, complete form persistence, edit/delete recovery,
  keyboard navigation, responsive layouts, sorting, filters, and details.
- A staged Go fixture passed the pre-commit hook when formatted, and failed
  when unformatted despite an unstaged formatting fix.
- A real terminal session launched against a fresh temporary SQLite database,
  added three applications with different locations and rankings, navigated,
  sorted, filtered, opened details, changed status, cancelled deletion, quit,
  and reopened the persisted records.
- Terminal dimensions changed while list, form, sort, filter, details, and
  deletion confirmation were open (including 30 columns by 12 lines).
- Normal launch works without seed data.
- Development and seeding now default to ./dev.db. Production retains the
  per-user database. JOBIT_DB_DSN remains an explicit override.
- Git ignores databases, SQLite WAL/SHM/journal sidecars, logs, and bin/.

Environment: macOS, Go 1.25.6, CGO enabled, Apple Command Line Tools.
The sandbox required GOCACHE=/tmp/jobit-go-cache.
Native Linux release binaries were not exercised.

## Outstanding release gate

GitHub Actions passed for commit 00420a14cfcfbfafe2760540232f65f1711e9bab:
https://github.com/MrMoneyInTheBank/jobit/actions/runs/38132425136

That run predates the current database and release changes.
The final release commit must be published and pass the configured workflow
before section 7 can be marked fully complete. This audit did not push commits.
