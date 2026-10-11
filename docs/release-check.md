# Release verification — 11 October 2026

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
Native Linux and Windows release binaries were not exercised.

## Outstanding release gate

GitHub Actions passed for commit 00420a14cfcfbfafe2760540232f65f1711e9bab:
https://github.com/MrMoneyInTheBank/jobit/actions/runs/38132425136

That run predates the latest local database-location and release changes.
The final release commit must be published and pass the configured workflow
before section 7 can be marked fully complete. This audit did not push commits.
