package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestMain(m *testing.M) {
	// Create a temporary directory for testing
	tempDir := os.TempDir()
	testDir := filepath.Join(tempDir, "blog-test")
	
	// Change to test directory to avoid conflicts
	originalDir, _ := os.Getwd()
	os.Chdir(testDir)
	
	// Create the test directory if it doesn't exist
	os.MkdirAll(testDir, 0755)
	
	// Run tests
	exitCode := m.Run()
	
	// Cleanup
	os.Chdir(originalDir)
	os.RemoveAll(testDir)
	
	os.Exit(exitCode)
}

func TestCLICommands(t *testing.T) {
	// This would test the CLI commands, but since they're cobra commands
	// we might need to test them differently
	
	// For now, just verify our command structure is valid
	if rootCmd == nil {
		t.Error("Root command should not be nil")
	}
	
	// Check that all expected commands are registered
	commands := rootCmd.Commands()
	if len(commands) == 0 {
		t.Error("Expected commands to be registered")
	}
	
	// Verify specific commands exist
	cmdMap := make(map[string]bool)
	for _, cmd := range commands {
		cmdMap[cmd.Use] = true
	}
	
	expectedCommands := []string{"build", "preview", "deploy", "validate", "clean"}
	for _, expected := range expectedCommands {
		if !cmdMap[expected] {
			t.Errorf("Missing command: %s", expected)
		}
	}
}