//go:generate sh -c "go run ./cmd/generate-zerolog/... > logger_zerolog_generated.go"

package diag

import (
	"fmt"
	"io"
	"os"
	"time"

	"github.com/rs/zerolog"
)

func init() {
	zerolog.TimeFieldFormat = time.RFC3339Nano
	zerolog.MessageFieldName = "msg"
}

type zerologLoggerFactory struct{}

func (zerologLoggerFactory) NewLogger(p *rootContextParams) LevelLogger {
	var logger zerolog.Logger

	var out io.Writer = os.Stderr
	if p.Out != nil {
		out = p.Out
	}

	if p.Pretty {
		logger = zerolog.New(
			zerolog.ConsoleWriter{Out: out},
		)
	} else {
		logger = zerolog.New(out)
	}

	zerologLevel, err := zerolog.ParseLevel(p.LogLevel.String())
	if err != nil {
		panic(fmt.Errorf("invalid log level %s: %w", p.LogLevel, err))
	}

	logger = logger.
		With().
		Timestamp().
		Logger().
		Level(zerologLevel)

	contextData := zerolog.Dict().
		Str("correlationId", p.DiagData.CorrelationID)
	for k, v := range p.DiagData.Entries {
		contextData = contextData.Str(k, v)
	}

	logger = logger.With().
		Dict("context", contextData).
		Logger()

	return &zerologLevelLogger{
		Logger:               logger,
		cloudPlatformAdapter: p.cloudPlatformAdapter,
	}
}

func (zerologLoggerFactory) ChildLogger(logger LevelLogger, diagOpts DiagOpts) LevelLogger {
	zerologLogger, ok := logger.(*zerologLevelLogger)
	if !ok {
		panic(fmt.Errorf("zerologLoggerFactory.ForkLogger: logger is not a *zerologLevelLogger"))
	}

	diagData := diagOpts.DiagData

	contextData := zerolog.Dict().
		Str("correlationId", diagData.CorrelationID)
	for k, v := range diagData.Entries {
		contextData = contextData.Str(k, v)
	}
	childLogger := zerologLogger.Logger.With().
		Dict("context", contextData).
		Logger()

	if diagOpts.Level != nil {
		logLevel := diagOpts.Level.String()
		zerologLevel, err := zerolog.ParseLevel(logLevel)
		if err != nil {
			childLogger.Warn().Err(err).Msgf("unexpected log level: %s", logLevel)
		} else {
			childLogger = childLogger.Level(zerologLevel)
		}
	}

	return &zerologLevelLogger{
		Logger:               childLogger,
		cloudPlatformAdapter: zerologLogger.cloudPlatformAdapter,
	}
}

var _ LoggerFactory = zerologLoggerFactory{}

type zerologLevelLogger struct {
	zerolog.Logger
	cloudPlatformAdapter
}

func (l *zerologLevelLogger) appendCloudPlatformLevelData(level LogLevel, evt *zerolog.Event) {
	if l.cloudPlatformAdapter != nil {
		l.cloudPlatformAdapter.appendLevelData(level, zerologLogFieldAppender{Event: evt})
	}
}

type zerologLogLevelEvent struct {
	*zerolog.Event
}

type zerologLogFieldAppender struct {
	*zerolog.Event
}

func (l zerologLogFieldAppender) Str(key, val string) {
	l.Event.Str(key, val)
}

type zerologLogData struct {
	*zerolog.Event
}

var _ MsgData = &zerologLogData{}

func (l *zerologLevelLogger) Error() LogLevelEvent {
	evt := l.Logger.Error()
	l.appendCloudPlatformLevelData(LogLevelErrorValue, evt)
	return zerologLogLevelEvent{Event: evt}
}

func (l *zerologLevelLogger) Warn() LogLevelEvent {
	evt := l.Logger.Warn()
	l.appendCloudPlatformLevelData(LogLevelWarnValue, evt)
	return &zerologLogLevelEvent{Event: evt}
}

func (l *zerologLevelLogger) Info() LogLevelEvent {
	evt := l.Logger.Info()
	l.appendCloudPlatformLevelData(LogLevelInfoValue, evt)
	return &zerologLogLevelEvent{Event: evt}
}

func (l *zerologLevelLogger) Debug() LogLevelEvent {
	evt := l.Logger.Debug()
	l.appendCloudPlatformLevelData(LogLevelDebugValue, evt)
	return &zerologLogLevelEvent{Event: evt}
}

func (l *zerologLevelLogger) Trace() LogLevelEvent {
	evt := l.Logger.Trace()
	l.appendCloudPlatformLevelData(LogLevelTraceValue, evt)
	return &zerologLogLevelEvent{Event: evt}
}

func (l *zerologLevelLogger) WithLevel(level LogLevel) LogLevelEvent {
	zerologLevel, err := zerolog.ParseLevel(level.String())
	if err != nil {
		zerologLevel = zerolog.DebugLevel
		l.Logger.Warn().Err(err).Msgf("Invalid log level: %s. Will use %s", level, zerologLevel)
	}

	evt := l.Logger.WithLevel(zerologLevel)
	l.appendCloudPlatformLevelData(level, evt)
	return &zerologLogLevelEvent{Event: evt}
}

func (l *zerologLevelLogger) NewData() MsgData {
	return &zerologLogData{Event: zerolog.Dict()}
}

func (l zerologLogLevelEvent) WithDataFn(dataFn func(data MsgData)) LogLevelEvent {
	evt := &zerologLogData{Event: zerolog.Dict()}
	dataFn(evt)
	return &zerologLogLevelEvent{
		Event: l.Event.Dict("data", evt.Event),
	}
}

func (e zerologLogLevelEvent) WithData(data MsgData) LogLevelEvent {
	zerologData, ok := data.(*zerologLogData)
	if !ok {
		panic(fmt.Errorf("zerologLogLevelEvent.WithData: data is not a *zerologLogData"))
	}

	return &zerologLogLevelEvent{
		Event: e.Event.Dict("data", zerologData.Event),
	}
}

func (e zerologLogLevelEvent) WithError(err error) LogLevelEvent {
	return &zerologLogLevelEvent{Event: e.Event.Err(err)}
}

func (d *zerologLogData) Dict(key string, data MsgData) MsgData {
	zerologData, ok := data.(*zerologLogData)
	if !ok {
		panic(fmt.Errorf("MsgData instance is not zerolog data"))
	}
	return &zerologLogData{Event: d.Event.Dict(key, zerologData.Event)}
}
