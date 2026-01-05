#!/usr/bin/env bash

set -euo pipefail

function run-check() {
	local name="$1"
	local cmd="$2"
	echo "🔍 Running ${name}..."
	if ! ${cmd}; then
		echo "❌ ${name} failed - blocking Claude"
		exit 2
	fi
	echo "✅ ${name} passed"
}

run-check "Format (auto-fix)" "go fmt ./..."
run-check "Lint" "golangci-lint run ./..."
run-check "Type check" "go vet ./..."
run-check "Tests" "go test -v -race ./..."

echo "🎉 All quality checks passed - codebase is clean!"
exit 0
