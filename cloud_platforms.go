package diag

type logFieldAppender interface {
	Str(key string, val string)
}

type cloudPlatformAdapter interface {
	appendLevelData(level LogLevel, target logFieldAppender)
}

type gcpAdapter struct{}

// appendLevelData appends GCP-specific log data to the given target.
// GCP log severity levels can be found here:
// https://cloud.google.com/logging/docs/reference/v2/rest/v2/LogEntry#logseverity
func (gcpAdapter) appendLevelData(level LogLevel, target logFieldAppender) {
	switch level {
	case LogLevelTraceValue:
		target.Str("severity", "DEBUG")
	case LogLevelDebugValue:
		target.Str("severity", "DEBUG")
	case LogLevelInfoValue:
		target.Str("severity", "INFO")
	case LogLevelWarnValue:
		target.Str("severity", "WARNING")
	case LogLevelErrorValue:
		target.Str("severity", "ERROR")
	default:
		target.Str("severity", "DEFAULT")
	}
}

var _ cloudPlatformAdapter = gcpAdapter{}
