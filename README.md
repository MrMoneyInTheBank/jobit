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

Implemented: local persistence, responsive application list, sorting and filtering,
full details, validated creation and editing, status updates, and confirmed deletion.
Release verification is tracked in docs/release-check.md.

## Installation

Install Go 1.25.6 or newer from the [official Go installer](https://go.dev/doc/install/),
a C compiler for your operating system, and use a terminal to run Jobit.

| Platform | C compiler setup |
| --- | --- |
| macOS | Run `xcode-select --install` to install Apple's Command Line Tools (Clang). |
| Ubuntu / Debian | Run `sudo apt-get update && sudo apt-get install -y build-essential`. |
| Windows | Install a MinGW-w64 GCC toolchain matching your Go architecture and put its compiler directory on PATH. Check `gcc --version` in the shell used for installation. |

Jobit's [SQLite driver requires CGO and a C compiler](https://github.com/mattn/go-sqlite3#installation)
when building from source. The default build bundles SQLite; you do not need
to install a SQLite server, SQLite CLI, or separate SQLite development library.

Install a published version without cloning this repository (macOS/Linux):

```sh
CGO_ENABLED=1 go install github.com/MrMoneyInTheBank/jobit/cmd/jobit@latest
jobit --help
jobit
```

On Windows, use PowerShell:

```powershell
$env:CGO_ENABLED = "1"
go install github.com/MrMoneyInTheBank/jobit/cmd/jobit@latest
jobit --help
jobit
```

Add Go's binary installation directory to PATH if `jobit` is not found:
`go env GOBIN`, or `bin` under `go env GOPATH` when GOBIN is empty
(normally `~/go/bin` on macOS/Linux or `%USERPROFILE%\go\bin` on Windows).
The compiler is needed during installation; subsequent launches use the built
executable and the local data file. No repository checkout is needed at runtime.
`@latest` installs published source, so unreleased local changes require a build
from the working tree.

## Data location

Jobit uses one persistent database per user, independent of where it is launched:

| Platform | Default database |
| --- | --- |
| macOS | `~/Library/Application Support/jobit/jobit.db` |
| Linux / other Unix | `$XDG_DATA_HOME/jobit/jobit.db`, or `~/.local/share/jobit/jobit.db` when XDG_DATA_HOME is absent or relative |
| Windows | `%LOCALAPPDATA%\jobit\jobit.db` |

Run `jobit --help` (or `jobit -h`) to see the resolved default and effective
database locations. Help does not create directories or open the database.
Normal startup creates the default directory and database, then initializes
the embedded schema. Reinstalling the executable does not remove your data.

`JOBIT_DB_DSN` overrides the default with a SQLite path or DSN. Relative
overrides are relative to your working directory; custom parent directories
must already exist. For example:

```sh
JOBIT_DB_DSN="$HOME/my-applications.db" jobit
```

Earlier development builds used `./jobit.db`. Existing files are not moved or
merged automatically. To continue using one, set JOBIT_DB_DSN to its absolute
path. To adopt the new default, close all Jobit instances, back up the old file,
create the directory shown by `jobit --help`, and copy the database there.
Do not overwrite an existing destination database. If SQLite sidecar files
remain, use a SQLite backup procedure rather than copying a live database.

## Run and develop

Cloning requires Git; Task v3 is optional. The tasks and check script set
CGO_ENABLED=1 explicitly.

From the repository root:

```sh
CGO_ENABLED=1 go run ./cmd/jobit
```

Direct launches and production use the per-user location described above.
The development task defaults to ./dev.db to isolate sample data.
Use `JOBIT_DB_DSN=./dev.db CGO_ENABLED=1 go run ./cmd/jobit` for a separate
development file. An explicit nonempty DSN passed to store.OpenDB takes
precedence over the environment. Tests use isolated temporary databases.

With Task v3 installed, `task run-dev` runs with ./dev.db (unless JOBIT_DB_DSN is set), and
`task run-prod` builds `bin/jobit` and runs it with the same database rules.
`task seed` (or `CGO_ENABLED=1 go run ./cmd/seed`) appends 50 sample applications
to ./dev.db by default. JOBIT_DB_DSN explicitly overrides both development and
seed commands; setting it to your personal database will modify that database.

The startup path is: cmd/jobit opens and initializes SQLite, reads applications
through the store mapper, constructs table rows, and starts Bubble Tea.
cmd/seed uses the same database selection and schema, then inserts sample rows.
The schema stores pay in separate nullable min/max/currency columns.

## Keyboard contract

These are the intended bindings across screens. The list, scrollable details,
and a complete add form are implemented, along with sorting and text/status
filtering. Existing applications can be edited and deleted.

| Action | Keys | Screen / behavior |
| --- | --- | --- |
| Move selection | Up/Down or k/j | Lists and selection menus |
| Page navigation | PageUp/PageDown | Lists |
| First/last row | Home/End or g/G | Lists and details |
| Sort | s | Open sort selection; Enter applies |
| Filter | / | Open filter input; Enter applies |
| Clear filter | Esc | List with an active filter |
| Details | Enter | Open selected application |
| Add | a | Open the full application form from the list |
| Edit | e | Edit the selected application from list or details |
| Update status | u | Open editing directly on the status selector |
| Delete | d | Show confirmation; Enter deletes, Esc cancels |
| Next / previous field | Tab / Shift+Tab | Forms |
| Submit form | Ctrl+S | Save; Enter stays available for fields |
| Cancel / back | Esc | Discard form input, close a menu, or return to the list |
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

Press a to add an application, including from the empty list. The form includes
company, position, location, ranking, application date, status, referral, work
mode, minimum and maximum pay, currency, notes, and both links.

Company and position are required and trimmed. The date defaults to today in
UTC and must be a valid YYYY-MM-DD date. Status defaults to submitted; referral
defaults to no. Use Left/Right to cycle status, referral, and work mode. These
are controlled selectors: typing, pasting, and character deletion cannot create
invalid values. Ctrl+U resets a selector to its default. Work mode defaults to
unspecified.

Ranking is blank or 0 for unranked, otherwise 1–5. Pay amounts are optional,
nonnegative whole numbers without commas, and maximum must be at least minimum
when both are supplied. A blank amount stays absent; an explicit 0 is retained.
Invalid ranking and pay characters are rejected while typing or pasting.
Complete invalid dates are also rejected; partial dates remain editable.
Currency is optional and accepts three letters, normalized to uppercase. Pay
can contain just one amount or currency alone. Notes and links are optional;
nonempty links must be complete HTTP or HTTPS URLs. Required text, completed
dates, currency length, and the pay range are validated again before saving.

Tab/Shift+Tab move between fields; Enter moves to the next field except in
notes, where it inserts a newline. Left/Right move the text cursor in other
fields; Home/End (or Ctrl+A/Ctrl+E) jump to its ends. Backspace/Delete edit text,
and Ctrl+U clears the field. Long values and notes scroll to keep the cursor
visible. Each field shows its input rule above the form.

Ctrl+S validates and saves. Invalid input stays in the form with the invalid
field focused and an error displayed; database failures also retain the draft.
Saving disables further edits and repeat submissions until the result arrives.
Esc immediately discards the unsaved draft and returns to the list without
writing a record; Ctrl+C quits the app and also abandons unsaved input. Successful
saves select the new row if visible under the active filters. Entries persist
in the selected SQLite database across launches.


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


The details view includes every application field and the database record ID.
Dates use UTC calendar days; missing dates and optional text display a dash,
and missing pay displays Unknown. Notes wrap at word boundaries and preserve
paragraph breaks; URLs wrap without dropping characters. Use g/Home and G/End
to jump to the top or bottom. Resizing preserves the relative reading position;
opening another details view starts at the top. Esc restores the selected
application and the active list sort and filters.


## Editing and deletion

From the list or details, e opens all application fields prefilled for editing.
Use u to jump directly to the status selector, change it with Left/Right, then
Ctrl+S to save. Editing retains the database ID. Ctrl+U clears an optional text
or pay field; clearing both amounts and currency removes pay information.
Required fields cannot be saved empty. Esc discards the draft and returns to
the screen where editing began.

Press d to request deletion. The confirmation identifies the company and
position; Enter permanently deletes the record, while Esc cancels. Failed
updates retain the draft, and failed deletions retain the record and allow a
retry. Sort and filter settings survive both operations; an updated record
hidden by the active filter is reported in the list's status line.
