package database_test

import (
	"data-automation-service/pkg/database"
	"testing"
)

func TestNewPostgresDB_EmptyURL(t *testing.T) {
	db, err := database.NewPostgresDB("", 25, 25)
	if err == nil {
		t.Errorf("Expected error for empty URL, got nil")
	}
	if db != nil {
		t.Errorf("Expected nil db for empty URL")
	}
}

func TestNewPostgresDB_InvalidURL(t *testing.T) {
	db, err := database.NewPostgresDB("postgres://invalid:invalid@invalid:5432/invalid", 25, 25)
	if err == nil {
		t.Errorf("Expected error for invalid URL, got nil")
	}
	if db != nil {
		t.Errorf("Expected nil db for invalid URL")
	}
}
