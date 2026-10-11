# jobit

Terminal-based job application tracker because I don't like GUIs.

## Overview

`jobit` is a personal, offline-first job application tracker built as a terminal user interface (TUI).

It’s designed to be:
- keyboard-driven
- simple
- fast
- flexible

The goal is to replace spreadsheets or Notion with a tool that lives entirely in the terminal and adapts to how *you* want to track applications.

## Core Ideas

- Local-first: data is stored locally
- Schema-driven: you define the fields you care about
- Terminal-native: no mouse, no browser, no GUI
- Opinionated simplicity over feature bloat

## Status

Early development.

## Run and develop

Use Go 1.25.6 (the version in go.mod), Git, and a C compiler. SQLite uses
github.com/mattn/go-sqlite3 with CGO; its default build bundles SQLite, so no
separate SQLite library installation is needed.

- macOS: install Xcode Command Line Tools with `xcode-select --install`.
- Ubuntu/Debian: install `build-essential`.
- Set `CGO_ENABLED=1` for Go builds, tests, and vet. The tasks and check script
  set this explicitly.

From the repository root:

```sh
CGO_ENABLED=1 go run ./cmd/jobit
```

The default database is `./jobit.db`, relative to the working directory.
`JOBIT_DB_DSN` overrides it, for example
`JOBIT_DB_DSN=./dev.db CGO_ENABLED=1 go run ./cmd/jobit`.
An explicit nonempty DSN passed to store.OpenDB takes precedence over the
environment. Tests explicitly use `:memory:` to avoid touching local data.

With Task v3 installed, `task run-dev` runs that same command and
`task run-prod` builds `bin/jobit` and runs it with the same database rules.
`task seed` (or `CGO_ENABLED=1 go run ./cmd/seed`) appends 50 sample applications
to the selected database. Use a separate development database when seeding.

The startup path is: cmd/jobit opens and initializes SQLite, reads applications
through the store mapper, constructs table rows, and starts Bubble Tea.
cmd/seed uses the same database selection and schema, then inserts sample rows.
The schema stores pay in separate nullable min/max/currency columns.

## Keyboard contract

These are the intended bindings across screens. The list, scrollable details,
and an essentials add form are implemented.
Sorting and text/status filtering are implemented. Editing and the full
application form remain planned.

| Action | Keys | Screen / behavior |
| --- | --- | --- |
| Move selection | Up/Down or k/j | Lists and selection menus |
| Page navigation | PageUp/PageDown | Lists |
| First/last row | Home/End or g/G | Lists |
| Sort | s | Open sort selection; Enter applies |
| Filter | / | Open filter input; Enter applies |
| Clear filter | Esc | List with an active filter |
| Details | Enter | Open selected application |
| Add / edit | a / e | Open form from list / details |
| Next / previous field | Tab / Shift+Tab | Forms |
| Submit form | Ctrl+S | Save; Enter stays available for fields |
| Cancel / back | Esc | Close input/menu or return to previous screen |
| Quit | q | List and details, outside text inputs |
| Quit anywhere | Ctrl+C | All screens |

At the root list Esc is a no-op when no filter is active. Printable keys,
including q, j, and k, enter text in input fields.

## Formatting, lint, and tests

The chosen lint command is `go vet ./...`. Formatting is checked with
`gofmt -l ./cmd ./internal`; any output fails the check.

```sh
sh scripts/check.sh fmt   # task fmt-check
sh scripts/check.sh lint  # task lint
sh scripts/check.sh test  # task test
sh scripts/check.sh all   # task check: runs all three, reports failure if any fail
gofmt -w ./cmd ./internal # task fmt: apply formatting
```

All checks share scripts/check.sh. GitHub Actions runs its `all` mode on every
push and pull request, using go.mod for Go setup, CGO_ENABLED=1, and
build-essential for the C compiler.

Enable the versioned local hook after each fresh clone:

```sh
git config --local core.hooksPath .githooks
# or: task hooks
```

This replaces any existing hooksPath setting for this clone. The pre-commit
hook runs the same formatting and lint checks when staged Go files, go.mod,
or go.sum change. It checks an isolated copy of the staged snapshot, so
unstaged fixes cannot conceal problems in a commit. It never reformats or
stages files. Run the full check command before pushing; CI also runs tests.

## Current verification baseline

The store and TUI use StoredJobApplication to carry database identity alongside
the application data. Rank 0 maps to SQL NULL (unranked). Pay remains structured
in nullable columns; patches can clear pay and optional strings.

Local verification on macOS used Go 1.25.6, CGO_ENABLED=1, /usr/bin/clang,
and the Command Line Tools MacOSX15.0 SDK. The sandbox disallowed the normal
Go build cache, so checks used `GOCACHE=/tmp/jobit-go-cache`. This is a sandbox
workaround, not a SQLite requirement. Formatting, tests, and vet pass using the shared check script.

Persistence reads are ordered by ascending database ID. Invalid stored dates
return an error identifying the row. Startup errors are written to stderr.
Persistence tests cover empty databases, DSN precedence, file reopen, optional
fields, invalid IDs, missing rows, schema failures, and seed rollback on insert
and commit failures.

## Application list and details

The list shows company, position, location, and ranking. Missing locations and
unranked applications display a dash. The selected row is highlighted; the
header shows its position in the list. Column widths and visible row count
adapt to the terminal, with long values shortened in the list.

Press Enter for every stored field, including pay, notes, and links. Details
wrap and scroll with j/k, arrows, or PageUp/PageDown. Esc returns to the same
selected row. Below 24 columns or 9 lines, a resize prompt replaces the view;
Ctrl+C remains available.

Press a to add an application, including from the empty list. The current
essentials form accepts company, position, location, and ranking. Company and
position are required; ranking is blank/0 for unranked or 1–5. Tab/Shift+Tab
move between fields, Backspace deletes, Ctrl+S saves, and Esc cancels. The
form defaults to submitted status and today's UTC date. The saved row is
selected, and errors keep the input available for correction. Remaining fields
and richer editing are planned for the full form in section 6.

## Sorting and filtering

Press s, choose with j/k or arrows, then Enter to apply. Esc cancels.
Sort by added order, company, position, location, or ranking in either
direction. Text sorts ignore case and surrounding whitespace. Empty values
and unranked applications stay last in both directions; ties use ascending
database ID. The default is oldest added first.

Press / to filter. Search is a case-insensitive substring across company,
position, location, and notes. Tab/Shift+Tab cycle the status constraint;
text and status must both match. Enter applies, Esc cancels the draft, and
Ctrl+U clears the draft text. From the list, Esc clears all active filters.
The toolbar shows the active sort and filter, and the header shows matching
and total counts.

Sorting and filtering only change the view. Selection follows the same
application ID when it remains visible; otherwise it resets to the first
match. Details and the add form retain these settings. A new matching
application becomes selected; if filters hide it, a saved confirmation explains
how to clear the filters. No-matches results have a separate empty state.
