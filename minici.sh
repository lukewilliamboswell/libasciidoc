#!/usr/bin/env bash
set -euo pipefail

red()   { printf "\033[1;31m%s\033[0m\n" "$*"; }
green() { printf "\033[1;32m%s\033[0m\n" "$*"; }
bold()  { printf "\033[1m%s\033[0m\n" "$*"; }

step() {
    bold "--- $1"
}

fail() {
    red "FAIL: $1"
    exit 1
}

pass() {
    green "OK"
    echo
}

# ── go mod tidy ──────────────────────────────────────────
step "Checking go mod tidy"
go mod tidy
git diff --exit-code go.mod go.sum || fail "go.mod/go.sum are not tidy — commit the changes"
pass

# ── build ────────────────────────────────────────────────
step "Building"
go build ./...
pass

# ── tests ────────────────────────────────────────────────
step "Running tests (with race detector)"
go test -race ./...
pass

# ── lint ─────────────────────────────────────────────────
step "Linting (golangci-lint)"
if command -v golangci-lint &>/dev/null; then
    golangci-lint run ./...
else
    go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@latest run ./...
fi
pass

# ── go generate ──────────────────────────────────────────
step "Checking generated code is up to date"
go generate ./...
git diff --exit-code || fail "generated code is out of date — commit the changes"
pass

# ── govulncheck ──────────────────────────────────────────
step "Checking for known vulnerabilities"
if command -v govulncheck &>/dev/null; then
    govulncheck ./...
else
    go run golang.org/x/vuln/cmd/govulncheck@latest ./...
fi
pass

# ── validate examples ────────────────────────────────────
step "Validating examples"
find examples -name '*.adoc' -print0 | while IFS= read -r -d '' f; do
    echo "  checking: $f"
    go run ./cmd/ascii2html -o - "$f" > /dev/null
    go run ./cmd/ascii2doc  -o - "$f" > /dev/null
done
pass

# ── build static site ───────────────────────────────────
step "Building static site"
go build -o ascii2html ./cmd/ascii2html/
./ascii2html --static-site --css style.css --base-path / -o _site www/
rm -f ascii2html
pass

# ── summary ──────────────────────────────────────────────
green "All checks passed!"
echo
bold "Serving site at http://localhost:8000"
echo "Press Ctrl-C to stop."
python3 -m http.server 8000 --directory _site
