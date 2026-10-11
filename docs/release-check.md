# Release verification

## Local checks — 12 October 2026

- `GOCACHE=/tmp/jobit-go-cache sh scripts/check.sh all` passed: formatting,
  `go vet ./...`, and `go test ./...`.
- `GOCACHE=/tmp/jobit-go-cache CGO_ENABLED=1 go test -race ./...` passed.
  Apple's linker printed `LC_DYSYMTAB` warnings; the tests still passed.
- The executable integration tests build Jobit and run its TUI in an isolated
  terminal. They cover first launch, opening existing applications, creating,
  editing, deleting, and reopening records, help and version flags, missing
  export filenames, and CSV export without changing the source database.
- Store and TUI tests cover persistence, validation, rollback, keyboard input,
  layouts, sorting, filtering, details, and pending writes.
- The local terminal walkthrough exercised adding records, navigation, sorting,
  filtering, details, status updates, cancellation, and restart persistence.

Environment: macOS, Go 1.25.6, CGO enabled, Apple Command Line Tools. The
temporary `GOCACHE` path is a local sandbox workaround.

## GitHub release gate

Push the release commit and wait for the **Go checks** workflow to pass on its
exact commit on both macOS and Linux. Tag that checked commit, then publish the
GitHub release for the tag. The workflow runs formatting, vet, tests, a build,
and race tests. A local macOS pass does not replace the Linux workflow result.
