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

func (zerologLoggerFactory) NewLogger(p *RootContextParams) LevelLogger {
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
		logger,
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
		Logger: childLogger,
	}
}

var _ LoggerFactory = zerologLoggerFactory{}

type zerologLevelLogger struct {
	zerolog.Logger
}

type zerologLogLevelEvent struct {
	*zerolog.Event
}

type zerologLogData struct {
	*zerolog.Event
}

var _ MsgData = &zerologLogData{}

func (l *zerologLevelLogger) Error() LogLevelEvent {
	return zerologLogLevelEvent{Event: l.Logger.Error()}
}

func (l *zerologLevelLogger) Warn() LogLevelEvent {
	return &zerologLogLevelEvent{Event: l.Logger.Warn()}
}

func (l *zerologLevelLogger) Info() LogLevelEvent {
	return &zerologLogLevelEvent{Event: l.Logger.Info()}
}

func (l *zerologLevelLogger) Debug() LogLevelEvent {
	return &zerologLogLevelEvent{Event: l.Logger.Debug()}
}

func (l *zerologLevelLogger) Trace() LogLevelEvent {
	return &zerologLogLevelEvent{Event: l.Logger.Trace()}
}

func (l *zerologLevelLogger) WithLevel(level LogLevel) LogLevelEvent {
	zerologLevel, err := zerolog.ParseLevel(level.String())
	if err != nil {
		zerologLevel = zerolog.DebugLevel
		l.Logger.Warn().Err(err).Msgf("Invalid log level: %s. Will use %s", level, zerologLevel)
	}

	return &zerologLogLevelEvent{Event: l.Logger.WithLevel(zerologLevel)}
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
