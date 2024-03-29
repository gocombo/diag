package diag

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseLogLevel(t *testing.T) {
	t.Run("valid", func(t *testing.T) {
		tests := []struct {
			levelStr string
			level    LogLevel
		}{
			{"trace", LogLevelTraceValue},
			{"debug", LogLevelDebugValue},
			{"info", LogLevelInfoValue},
			{"warn", LogLevelWarnValue},
			{"error", LogLevelErrorValue},
		}

		for _, test := range tests {
			t.Run(test.levelStr, func(t *testing.T) {
				level, ok := ParseLogLevel(test.levelStr)
				assert.True(t, ok)
				assert.Equal(t, test.level, level)
			})
		}
	})
	t.Run("invalid", func(t *testing.T) {
		tests := []string{
			"TRACE",
			"DEBUG",
			"INFO",
			"WARN",
			"ERROR",
			"invalid",
			fake.Lorem().Word(),
		}

		for _, test := range tests {
			t.Run(test, func(t *testing.T) {
				level, ok := ParseLogLevel(test)
				assert.False(t, ok)
				assert.Equal(t, LogLevelDebugValue, level)
			})
		}
	})
}
