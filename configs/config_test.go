package configs_test

import (
	"data-automation-service/configs"
	"os"
	"testing"
)

func TestLoadConfig(t *testing.T) {
	os.Setenv("DB_HOST", "test_host")
	os.Setenv("DB_PORT", "1234")
	os.Setenv("DB_USER", "test_user")
	os.Setenv("DB_PASSWORD", "test_pass")
	os.Setenv("DB_NAME", "test_db")
	os.Setenv("DB_SSLMODE", "require")
	os.Setenv("BI_API_URL", "http://test.com")
	os.Setenv("BI_API_TOKEN", "test_token")
	os.Setenv("DB_MAX_OPEN_CONNS", "50")
	os.Setenv("DB_MAX_IDLE_CONNS", "10")

	// Clean up after test
	defer func() {
		os.Unsetenv("DB_HOST")
		os.Unsetenv("DB_PORT")
		os.Unsetenv("DB_USER")
		os.Unsetenv("DB_PASSWORD")
		os.Unsetenv("DB_NAME")
		os.Unsetenv("DB_SSLMODE")
		os.Unsetenv("BI_API_URL")
		os.Unsetenv("BI_API_TOKEN")
		os.Unsetenv("DB_MAX_OPEN_CONNS")
		os.Unsetenv("DB_MAX_IDLE_CONNS")
	}()

	cfg := configs.LoadConfig()

	if cfg.Host != "test_host" {
		t.Errorf("Expected Host 'test_host', got '%s'", cfg.Host)
	}
	if cfg.Port != "1234" {
		t.Errorf("Expected Port '1234', got '%s'", cfg.Port)
	}
	if cfg.User != "test_user" {
		t.Errorf("Expected User 'test_user', got '%s'", cfg.User)
	}
	if cfg.Password != "test_pass" {
		t.Errorf("Expected Password 'test_pass', got '%s'", cfg.Password)
	}
	if cfg.Name != "test_db" {
		t.Errorf("Expected Name 'test_db', got '%s'", cfg.Name)
	}
	if cfg.SSLMode != "require" {
		t.Errorf("Expected SSLMode 'require', got '%s'", cfg.SSLMode)
	}
	if cfg.URL != "http://test.com" {
		t.Errorf("Expected URL 'http://test.com', got '%s'", cfg.URL)
	}
	if cfg.Token != "test_token" {
		t.Errorf("Expected Token 'test_token', got '%s'", cfg.Token)
	}
	if cfg.MaxOpenConns != 50 {
		t.Errorf("Expected MaxOpenConns 50, got %d", cfg.MaxOpenConns)
	}
	if cfg.MaxIdleConns != 10 {
		t.Errorf("Expected MaxIdleConns 10, got %d", cfg.MaxIdleConns)
	}

	dsn := cfg.DSN()
	expectedDSN := "postgres://test_user:test_pass@test_host:1234/test_db?sslmode=require"
	if dsn != expectedDSN {
		t.Errorf("Expected DSN '%s', got '%s'", expectedDSN, dsn)
	}
}
