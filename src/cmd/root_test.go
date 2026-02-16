package cmd_test

import (
	"testing"

	alias "github.com/howmanysmall/npm-registry-mcp/src/cmd"
)

func TestExecute_NoArgs_RunsMCP(t *testing.T) {
	t.Parallel()
	t.Helper()

	ranMCP := false

	if err := alias.Execute(nil, func() {
		ranMCP = true
	}); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if !ranMCP {
		t.Fatal("expected MCP mode to run when no args provided")
	}
}

func TestExecute_OnlyRootFlag_RunsMCP(t *testing.T) {
	t.Parallel()
	t.Helper()

	ranMCP := false

	if err := alias.Execute([]string{"--json"}, func() {
		ranMCP = true
	}); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if !ranMCP {
		t.Fatal("expected MCP mode to run when only root flags are provided")
	}
}

func TestExecute_HelpFlag_DoesNotRunMCP(t *testing.T) {
	t.Parallel()
	t.Helper()

	ranMCP := false

	if err := alias.Execute([]string{"--help"}, func() {
		ranMCP = true
	}); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if ranMCP {
		t.Fatal("did not expect MCP mode to run for help flag")
	}
}
