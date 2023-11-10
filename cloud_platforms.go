package diag

type logDataAppender interface {
	Str(key string, value string)
}

type cloudPlatformAdapter interface {
	appendLevelData(level LogLevel, target logDataAppender)
}

type gcpAdapter struct{}

func (gcpAdapter) appendLevelData(level LogLevel, target logDataAppender) {
}

var _ cloudPlatformAdapter = gcpAdapter{}
