package cmd

import (
	"testing"
)

func TestRootCommand_HasExpectedSubcommands(t *testing.T) {
	// Verify that the root command has exactly the commands we expect
	// after stripping TUI, schedule, daemon, and stubs.
	expected := map[string]bool{
		"config":   false,
		"export":   false,
		"favorite": false,
		"fetch":    false,
		"list":     false,
		"playlist": false,
		"rate":     false,
		"set":      false,
		"stats":    false,
	}

	// Cobra adds completion/help automatically — skip those
	builtins := map[string]bool{"completion": true, "help": true}
	for _, cmd := range rootCmd.Commands() {
		name := cmd.Name()
		if builtins[name] {
			continue
		}
		if _, ok := expected[name]; ok {
			expected[name] = true
		} else {
			t.Errorf("unexpected subcommand registered: %q", name)
		}
	}

	for name, found := range expected {
		if !found {
			t.Errorf("expected subcommand %q not found", name)
		}
	}
}

func TestRootCommand_VersionSet(t *testing.T) {
	if rootCmd.Version == "" {
		t.Error("root command should have a version set")
	}
}
