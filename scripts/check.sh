#!/bin/sh
set -eu
cd "$(dirname "$0")/.."
export CGO_ENABLED=1
format_check() {
    unformatted=$(gofmt -l ./cmd ./internal)
    if [ -n "$unformatted" ]; then
        printf 'Run gofmt on:\n%s\n' "$unformatted" >&2
        return 1
    fi
}
case "${1:-all}" in
    fmt) format_check ;;
    lint) go vet ./... ;;
    test) go test ./... ;;
    all)
        result=0
        format_check || result=1
        go vet ./... || result=1
        go test ./... || result=1
        exit "$result"
        ;;
    *) echo "Usage: $0 [fmt|lint|test|all]" >&2; exit 2 ;;
esac
