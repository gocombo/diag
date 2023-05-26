package diag

import (
	"context"
	"fmt"
	"time"

	"github.com/rs/zerolog"
)

func init() {
	zerolog.TimeFieldFormat = time.RFC3339Nano
	zerolog.MessageFieldName = "msg"
}

// MsgData defines a standard interface for adding data to a log message
type MsgData interface {
	Str(key string, value string) MsgData
	// Strs(key string, value []string) MsgData
	// Stringer(key string, value fmt.Stringer) MsgData
	// Bytes(key string, val []byte) MsgData
	// Hex(key string, val []byte) MsgData
	// RawJSON(key string, b []byte) MsgData
	// Bool(key string, b bool) MsgData
	// Bools(key string, b []bool) MsgData
	// Int(key string, i int) MsgData
	// Ints(key string, i []int) MsgData
	// Int8(key string, i int8) MsgData
	// Ints8(key string, i []int8) MsgData
	// Int16(key string, i int16) MsgData
	// Ints16(key string, i []int16) MsgData
	// Int32(key string, i int32) MsgData
	// Ints32(key string, i []int32) MsgData
	// Int64(key string, i int64) MsgData
	// Ints64(key string, i []int64) MsgData
	// UInt(key string, i uint) MsgData
	// UInts(key string, i []uint) MsgData
	// UInt8(key string, i uint8) MsgData
	// UInts8(key string, i []uint8) MsgData
	// UInt16(key string, i uint16) MsgData
	// UInts16(key string, i []uint16) MsgData
	// UInt32(key string, i uint32) MsgData
	// UInts32(key string, i []uint32) MsgData
	// UInt64(key string, i uint64) MsgData
	// UInts64(key string, i []uint64) MsgData
	// Float32(key string, i float32) MsgData
	// Floats32(key string, i []float32) MsgData
	// Float64(key string, i float64) MsgData
	// Floats64(key string, i []float64) MsgData
	// Time(key string, i time.Time) MsgData
	// Times(key string, i []time.Time) MsgData
	// Duration(key string, i time.Duration) MsgData
	// Durations(key string, i []time.Duration) MsgData
	// IPAddr(key string, ip net.IP) MsgData
	// IPPrefix(key string, pfx net.IPNet) MsgData
	// MACAddr(key string, ha net.HardwareAddr) MsgData

	// // Creates nested dictionary under a given key
	// Dict(key string, data MsgData) MsgData

	// // Interface is usually the slowest method of adding data to a log message
	// // Prefer using the above methods when possible
	// Interface(key string, i interface{}) MsgData
}

type LevelLogger interface {
	Error() LogLevelEvent
	Warn() LogLevelEvent
	Info() LogLevelEvent
	Debug() LogLevelEvent
	Trace() LogLevelEvent
	NewData() MsgData
}

type LogLevelEvent interface {
	WithError(err error) LogLevelEvent
	WithDataFn(func(data MsgData)) LogLevelEvent
	WithData(data MsgData) LogLevelEvent

	LogMessageEvent
}

type LogMessageEvent interface {
	Msg(msg string)
	Msgf(fmt string, v ...interface{})
}

func Log(ctx context.Context) LevelLogger {
	logger, ok := ctx.Value(contextKeyLogger).(LevelLogger)
	if !ok {
		// TODO: Better message
		panic(fmt.Errorf("context does not contain a logger"))
	}
	return logger
}
