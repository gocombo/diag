package diag

import (
	"io"
	"os"

	"github.com/rs/zerolog"
)

func newZerologLogger(p *RootContextParams) LevelLogger {
	var logger zerolog.Logger

	var out io.Writer = os.Stderr
	if p.out != nil {
		out = p.out
	}

	if p.pretty {
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
		Level(zerolog.DebugLevel)

	logger = logger.With().
		Dict("context", zerolog.Dict().
			Str("correlationId", p.correlationID),
		).
		Logger()

	return &zerologLevelLogger{
		logger,
	}
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
		panic("zerologLogLevelEvent.WithData: data is not a *zerologLogData")
	}

	return &zerologLogLevelEvent{
		Event: e.Event.Dict("data", zerologData.Event),
	}
}

func (e zerologLogLevelEvent) WithError(err error) LogLevelEvent {
	return &zerologLogLevelEvent{Event: e.Event.Err(err)}
}
