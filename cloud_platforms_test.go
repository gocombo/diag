package diag

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGCPAdapter(t *testing.T) {
	t.Run("appendLevelData", func(t *testing.T) {
		tests := []struct {
			name       string
			level      LogLevel
			wantFields map[string]string
		}{
			{
				name:  "trace level",
				level: LogLevelTraceValue,
				wantFields: map[string]string{
					"severity": "DEBUG",
				},
			},
			{
				name:  "debug level",
				level: LogLevelDebugValue,
				wantFields: map[string]string{
					"severity": "DEBUG",
				},
			},
			{
				name:  "info level",
				level: LogLevelInfoValue,
				wantFields: map[string]string{
					"severity": "INFO",
				},
			},
			{
				name:  "warn level",
				level: LogLevelWarnValue,
				wantFields: map[string]string{
					"severity": "WARNING",
				},
			},
			{
				name:  "error level",
				level: LogLevelErrorValue,
				wantFields: map[string]string{
					"severity": "ERROR",
				},
			},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				fields := make(map[string]string)
				gcpAdapter{}.appendLevelData(tt.level, mockLogFieldAppender(fields))
				assert.Equal(t, tt.wantFields, fields)
			})
		}
	})
}

type mockLogFieldAppender map[string]string

func (m mockLogFieldAppender) Str(key, value string) {
	m[key] = value
}
