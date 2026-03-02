package cmd

import (
	"bytes"
	"testing"
)

func TestExecute(t *testing.T) {
	// Test basic root command execution (which does nothing without subcommands)
	rootCmd.SetArgs([]string{"--help"})
	if err := rootCmd.Execute(); err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
}

func TestReportCommand_MissingArgs(t *testing.T) {
	// Should fail because --type is required
	rootCmd.SetArgs([]string{"report"})

	// Capture output to not clutter test logs
	b := new(bytes.Buffer)
	rootCmd.SetOut(b)
	rootCmd.SetErr(b)

	err := rootCmd.Execute()
	if err == nil {
		t.Errorf("Expected error for missing --type flag, got nil")
	}
}

func TestApiPushCommand_Help(t *testing.T) {
	rootCmd.SetArgs([]string{"api-push", "--help"})

	b := new(bytes.Buffer)
	rootCmd.SetOut(b)
	rootCmd.SetErr(b)

	if err := rootCmd.Execute(); err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
}

func TestReportCommand_InvalidDB(t *testing.T) {
	// Set an invalid DB_URL to force the db connection to fail and return safely instead of panicking
	rootCmd.SetArgs([]string{"report", "--type=sales-trend"})

	b := new(bytes.Buffer)
	rootCmd.SetOut(b)
	rootCmd.SetErr(b)

	// It shouldn't crash, it should just log error and return
	err := rootCmd.Execute()
	if err != nil {
		t.Errorf("Expected no hard error from Execute, got %v", err)
	}
}

func TestApiPushCommand_InvalidDB(t *testing.T) {
	// Set an invalid DB_URL to force the db connection to fail and return safely instead of panicking
	rootCmd.SetArgs([]string{"api-push"})

	b := new(bytes.Buffer)
	rootCmd.SetOut(b)
	rootCmd.SetErr(b)

	// It shouldn't crash, it should just log error and return
	err := rootCmd.Execute()
	if err != nil {
		t.Errorf("Expected no hard error from Execute, got %v", err)
	}
}
