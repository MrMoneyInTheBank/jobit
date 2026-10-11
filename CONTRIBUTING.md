# Contributing to Jobit

Thanks for helping improve Jobit. Keep changes focused and consistent with the
existing terminal-first, local-data design.

## Get started

Install Go 1.25.6 or newer and a C compiler. On macOS, use Apple's Command Line
Tools; on Ubuntu or Debian, install `build-essential`. SQLite is bundled by the
Go driver, so a SQLite server is not needed. The executable integration tests
also use the `script` terminal utility.

From the repository root:

```sh
task check             # formatting, vet, and tests
go test -race ./...     # race checks
task build             # executable at bin/jobit
```

Task v3 is optional; `sh scripts/check.sh all` runs the same checks as
`task check`. `task run` starts the app using your real per-user database, so
use the isolated test fixtures when developing data-changing behavior.

## Follow the existing design

- `internal/model` defines `JobApplication`, its enums, patches, and
  `StoredJobApplication`. Keep the model as the reference for application
  fields; the stored type adds the database ID.
- `internal/store` owns SQL, validation at the persistence boundary, and the
  fixed per-user database location. Keep SQL details out of the TUI.
- `internal/tui` owns input, views, and keyboard behavior. Keep database
  operations behind the callbacks provided by `cmd/jobit`.
- `cmd/jobit` wires the app together and handles flags and CSV export. Human
  facing CLI output uses Lipgloss; CSV remains plain machine-readable data.

The data contract matters: rank is required from 1 to 5; ordinary missing text
uses an empty string; `Pay` can be nil; a nil patch field means no change. Do
not add another source of truth for these rules. When changing the schema,
consider applications already stored by a released version and avoid silently
removing user data.

## Write and verify changes

Use `gofmt`. Prefer descriptive names, short functions, and blank lines between
logical steps. Return errors with useful context instead of hiding them. Add
comments when behavior is subtle, such as protecting a pending write or
preserving an existing database file.

Add tests for behavior and failure paths, especially persistence, terminal
input, and exports. Use temporary home directories and databases; tests must
not touch a contributor's Jobit data. For changes to startup or flags, exercise
the built executable as well as package functions. Run `task check` before
opening a PR, and run the race suite for concurrent or persistence changes.

Keep commits small and coherent. In a PR, explain the behavior changed, why it
changed, how it was checked, and any effect on stored data or CLI use.

## Commit messages

Use `type(scope): message` for the subject. Pick a type such as `feat`, `fix`,
`docs`, `test`, `refactor`, or `chore`, and name the area changed in the scope
(for example, `store`, `tui`, `cmd`, or `task`). Write the message in the
imperative, start it with a lowercase letter, omit the trailing period, and
keep the full subject under 72 characters. Examples:
`fix(store): preserve existing applications` and
`docs(readme): explain task workflows`.

Add a blank line and a body when the reason or tradeoff needs explanation.
Keep unrelated changes in separate commits.

(yeah I know I haven't been following this strictly either but this is one of
those "do as I say, not as I do moments")
