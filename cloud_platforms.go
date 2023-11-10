package diag

type logFieldAppender interface {
	Str(key string, val string)
}

type cloudPlatformAdapter interface {
	appendLevelData(level LogLevel, target logFieldAppender)
}

type gcpAdapter struct{}

func (gcpAdapter) appendLevelData(level LogLevel, target logFieldAppender) {
}

var _ cloudPlatformAdapter = gcpAdapter{}
