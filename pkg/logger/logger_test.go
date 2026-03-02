package logger_test

import (
	"data-automation-service/pkg/logger"
	"testing"
)

func TestNewLogger(t *testing.T) {
	log := logger.NewLogger()
	if log == nil {
		t.Errorf("Expected non-nil logger")
	}
}
