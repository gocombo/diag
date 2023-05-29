package diag

import (
	"fmt"
	"io"
	"os"

	"github.com/rs/zerolog"
)

func mustParseZerologLevel(level LogLevel) zerolog.Level {
	zeroLogLevel, err := zerolog.ParseLevel(level.String())
	if err != nil {
		panic(fmt.Errorf("invalid log level %s: %w", level, err))
	}
	return zeroLogLevel
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

	logger = logger.
		With().
		Timestamp().
		Logger().
		Level(mustParseZerologLevel(p.LogLevel))

	logger = logger.With().
		Dict("context", zerolog.Dict().
			Str("correlationId", p.DiagData.CorrelationID),
		).
		Logger()

	return &zerologLevelLogger{
		logger,
	}
}

func (zerologLoggerFactory) ForkLogger(logger LevelLogger, opts DiagOpts) LevelLogger {
	zerologLogger, ok := logger.(*zerologLevelLogger)
	if !ok {
		panic("zerologLoggerFactory.ForkLogger: logger is not a *zerologLevelLogger")
	}
	// TODO: Set log level
	nextLogger := zerologLevelLogger{
		Logger: zerologLogger.Logger,
	}
	return &nextLogger
}

type zerologLevelLogger struct {
	zerolog.Logger
}

type zerologLogLevelEvent struct {
	*zerolog.Event
}

type zerologLogData struct {
	*zerolog.Event
}

func (d *zerologLogData) Str(key, value string) MsgData {
	return &zerologLogData{Event: d.Event.Str(key, value)}
}

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
	// TODO: It should emit warn event instead of potential panic
	return &zerologLogLevelEvent{Event: l.Logger.WithLevel(mustParseZerologLevel(level))}
}

func (l *zerologLevelLogger) NewData() MsgData {
	return &zerologLogData{Event: zerolog.Dict()}
}

func (l *zerologLevelLogger) NewLevelLogger(level LogLevel) LevelLogger {
	// TODO: It should emit warn event instead of potential panic
	return &zerologLevelLogger{Logger: l.Logger.Level(mustParseZerologLevel(level))}
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
		panic("zerologLogLevelEvent.WithData: data is not a *zerologLogData")
	}

	return &zerologLogLevelEvent{
		Event: e.Event.Dict("data", zerologData.Event),
	}
}

func (e zerologLogLevelEvent) WithError(err error) LogLevelEvent {
	return &zerologLogLevelEvent{Event: e.Event.Err(err)}
}
