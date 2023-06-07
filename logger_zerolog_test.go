package diag

import (
	"bufio"
	"bytes"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestZerolog_LoggerFactory(t *testing.T) {
	factory := zerologLoggerFactory{}
	t.Run("NewLogger", func(t *testing.T) {
		t.Run("returns a new logger", func(t *testing.T) {
			var output bytes.Buffer
			outputWriter := bufio.NewWriter(&output)
			wantCorrelationID := fake.UUID().V4()
			wantEntries := map[string]string{
				"key1": fake.Lorem().Word(),
				"key2": fake.Lorem().Word(),
			}
			logger := factory.NewLogger(&RootContextParams{
				DiagData: ContextDiagData{
					CorrelationID: wantCorrelationID,
					Entries:       wantEntries,
				},
				LogLevel: LogLevelInfoValue,
				Out:      outputWriter,
			})
			assert.IsType(t, &zerologLevelLogger{}, logger)

			msg := fake.Lorem().Sentence(3)
			logger.Info().Msg(msg)

			outputWriter.Flush()

			var logMessage TestLogMessage[map[string]string]
			json.Unmarshal(output.Bytes(), &logMessage)
			assert.Equal(t, TestLogMessage[map[string]string]{
				Level: "info",
				Msg:   msg,
				Time:  logMessage.Time,
				Context: map[string]string{
					"correlationId": wantCorrelationID,
					"key1":          wantEntries["key1"],
					"key2":          wantEntries["key2"],
				},
			}, logMessage)
		})

		t.Run("returns a new logger with pretty", func(t *testing.T) {
			var output bytes.Buffer
			outputWriter := bufio.NewWriter(&output)
			logger := factory.NewLogger(&RootContextParams{
				LogLevel: LogLevelInfoValue,
				Out:      outputWriter,
				Pretty:   true,
			})
			msg := fake.Lorem().Sentence(3)
			logger.Info().Msg(msg)
			outputWriter.Flush()
			outputStr := output.String()
			assert.Contains(t, outputStr, msg)
			assert.NotContains(t, outputStr, `"msg":"`+msg+`"`)
		})

		t.Run("panics if unknown log level", func(t *testing.T) {
			assert.Panics(t, func() {
				factory.NewLogger(&RootContextParams{
					LogLevel: LogLevel(fake.Lorem().Word()),
				})
			})
		})
	})

	t.Run("ChildLogger", func(t *testing.T) {
		t.Run("creates a derived logger", func(t *testing.T) {
			var output bytes.Buffer
			outputWriter := bufio.NewWriter(&output)

			rootDiagParams := ContextDiagData{
				CorrelationID: fake.UUID().V4(),
				Entries:       map[string]string{},
			}
			rootLogger := factory.NewLogger(&RootContextParams{
				LogLevel: LogLevelTraceValue,
				Out:      outputWriter,
				DiagData: rootDiagParams,
			})

			nextCorrelationId := fake.UUID().V4()
			wantChildEntries := map[string]string{
				"ch-key1": fake.Lorem().Word(),
				"ch-key2": fake.Lorem().Word(),
			}
			childLogger := factory.ChildLogger(rootLogger, DiagOpts{
				DiagData: ContextDiagData{
					CorrelationID: nextCorrelationId,
					Entries:       wantChildEntries,
				},
			})
			assert.IsType(t, &zerologLevelLogger{}, childLogger)

			msg := fake.Lorem().Sentence(3)
			childLogger.Info().Msg(msg)

			outputWriter.Flush()
			var logMessage TestLogMessage[map[string]string]
			json.Unmarshal(output.Bytes(), &logMessage)
			assert.Equal(t, TestLogMessage[map[string]string]{
				Level: "info",
				Msg:   msg,
				Time:  logMessage.Time,
				Context: map[string]string{
					"correlationId": nextCorrelationId,
					"ch-key1":       wantChildEntries["ch-key1"],
					"ch-key2":       wantChildEntries["ch-key2"],
				},
			}, logMessage)
		})
		t.Run("creates a derived logger with custom level", func(t *testing.T) {
			var output bytes.Buffer
			outputWriter := bufio.NewWriter(&output)

			rootDiagParams := ContextDiagData{
				CorrelationID: fake.UUID().V4(),
				Entries:       map[string]string{},
			}
			rootLogger := factory.NewLogger(&RootContextParams{
				LogLevel: LogLevelTraceValue,
				Out:      outputWriter,
				DiagData: rootDiagParams,
			})

			nextLogLevel := LogLevelWarnValue

			childLogger := factory.ChildLogger(rootLogger, DiagOpts{
				Level:    &nextLogLevel,
				DiagData: rootDiagParams,
			})
			assert.IsType(t, &zerologLevelLogger{}, childLogger)

			msg := fake.Lorem().Sentence(3)
			childLogger.Info().Msg(msg)

			outputWriter.Flush()
			assert.Empty(t, output.String())
		})
	})
}

type logDataFieldFn[TVal any] func(key string, value TVal) MsgData

func castLotDataFieldFn[TVal any](fn logDataFieldFn[TVal]) logDataFieldFn[any] {
	return func(key string, value any) MsgData {
		return fn(key, value.(TVal))
	}
}

func jsonify(data any) any {
	jsonData, err := json.Marshal(data)
	if err != nil {
		panic(err)
	}
	var result any
	if err := json.Unmarshal(jsonData, &result); err != nil {
		panic(err)
	}
	return result
}

func TestZerolog_LogData(t *testing.T) {
	factory := zerologLoggerFactory{}

	type testCase struct {
		name  string
		value any
		fn    logDataFieldFn[any]

		// expectedValue should be provided if serialization makes it different from value
		expectedValue any
	}

	type testCaseFn func(data MsgData) testCase

	tests := []testCaseFn{
		func(data MsgData) testCase {
			return testCase{
				name:  "Str",
				value: fake.Lorem().Sentence(3),
				fn:    castLotDataFieldFn(data.Str),
			}
		},
		func(data MsgData) testCase {
			return testCase{
				name:  "Strs",
				value: fake.Lorem().Words(3),
				fn:    castLotDataFieldFn(data.Strs),
			}
		},
		func(data MsgData) testCase {
			return testCase{
				name:  "Stringer",
				value: net.ParseIP(fake.Internet().Ipv4()),
				fn:    castLotDataFieldFn(data.Stringer),
			}
		},
		func(data MsgData) testCase {
			value := fake.Lorem().Bytes(10)
			return testCase{
				name:          "Bytes",
				value:         value,
				expectedValue: string(value),
				fn:            castLotDataFieldFn(data.Bytes),
			}
		},
		func(data MsgData) testCase {
			value := fake.Lorem().Bytes(10)
			return testCase{
				name:          "Hex",
				value:         value,
				expectedValue: hex.EncodeToString(value),
				fn:            castLotDataFieldFn(data.Hex),
			}
		},
		func(data MsgData) testCase {
			value := map[string]interface{}{
				"key1": fake.Lorem().Word(),
				"key2": fake.Lorem().Word(),
			}
			rawJSON, _ := json.Marshal(value)
			return testCase{
				name:          "RawJSON",
				value:         rawJSON,
				expectedValue: value,
				fn:            castLotDataFieldFn(data.RawJSON),
			}
		},
		func(data MsgData) testCase {
			return testCase{
				name:  "Bool",
				value: fake.Bool(),
				fn:    castLotDataFieldFn(data.Bool),
			}
		},
		func(data MsgData) testCase {
			return testCase{
				name:  "Bools",
				value: []bool{fake.Bool(), fake.Bool()},
				fn:    castLotDataFieldFn(data.Bools),
			}
		},
		func(data MsgData) testCase {
			return testCase{
				name:  "Int",
				value: fake.Int(),
				fn:    castLotDataFieldFn(data.Int),
			}
		},
		func(data MsgData) testCase {
			return testCase{
				name:  "Ints",
				value: []int{fake.Int(), fake.Int()},
				fn:    castLotDataFieldFn(data.Ints),
			}
		},
		func(data MsgData) testCase {
			return testCase{
				name:  "Int8",
				value: fake.Int8(),
				fn:    castLotDataFieldFn(data.Int8),
			}
		},
		func(data MsgData) testCase {
			return testCase{
				name:  "Ints8",
				value: []int8{fake.Int8(), fake.Int8()},
				fn:    castLotDataFieldFn(data.Ints8),
			}
		},
		func(data MsgData) testCase {
			return testCase{
				name:  "Int16",
				value: fake.Int16(),
				fn:    castLotDataFieldFn(data.Int16),
			}
		},
		func(data MsgData) testCase {
			return testCase{
				name:  "Ints16",
				value: []int16{fake.Int16(), fake.Int16()},
				fn:    castLotDataFieldFn(data.Ints16),
			}
		},
		func(data MsgData) testCase {
			return testCase{
				name:  "Int32",
				value: fake.Int32(),
				fn:    castLotDataFieldFn(data.Int32),
			}
		},
		func(data MsgData) testCase {
			return testCase{
				name:  "Ints32",
				value: []int32{fake.Int32(), fake.Int32()},
				fn:    castLotDataFieldFn(data.Ints32),
			}
		},
		func(data MsgData) testCase {
			return testCase{
				name:  "Int64",
				value: fake.Int64(),
				fn:    castLotDataFieldFn(data.Int64),
			}
		},
		func(data MsgData) testCase {
			return testCase{
				name:  "Ints64",
				value: []int64{fake.Int64(), fake.Int64()},
				fn:    castLotDataFieldFn(data.Ints64),
			}
		},
		func(data MsgData) testCase {
			return testCase{
				name:  "Uint",
				value: fake.UInt(),
				fn:    castLotDataFieldFn(data.Uint),
			}
		},
		func(data MsgData) testCase {
			return testCase{
				name:  "Uints",
				value: []uint{fake.UInt(), fake.UInt()},
				fn:    castLotDataFieldFn(data.Uints),
			}
		},
		func(data MsgData) testCase {
			return testCase{
				name:  "Uint8",
				value: fake.UInt8(),
				fn:    castLotDataFieldFn(data.Uint8),
			}
		},
		func(data MsgData) testCase {
			value := []uint8{fake.UInt8(), fake.UInt8()}
			return testCase{
				name:          "Uints8",
				value:         value,
				expectedValue: []interface{}{float64(value[0]), float64(value[1])},
				fn:            castLotDataFieldFn(data.Uints8),
			}
		},
		func(data MsgData) testCase {
			return testCase{
				name:  "Uint16",
				value: fake.UInt16(),
				fn:    castLotDataFieldFn(data.Uint16),
			}
		},
		func(data MsgData) testCase {
			return testCase{
				name:  "Uints16",
				value: []uint16{fake.UInt16(), fake.UInt16()},
				fn:    castLotDataFieldFn(data.Uints16),
			}
		},
		func(data MsgData) testCase {
			return testCase{
				name:  "Uint32",
				value: fake.UInt32(),
				fn:    castLotDataFieldFn(data.Uint32),
			}
		},
		func(data MsgData) testCase {
			return testCase{
				name:  "Uints32",
				value: []uint32{fake.UInt32(), fake.UInt32()},
				fn:    castLotDataFieldFn(data.Uints32),
			}
		},
		func(data MsgData) testCase {
			return testCase{
				name:  "Uint64",
				value: fake.UInt64(),
				fn:    castLotDataFieldFn(data.Uint64),
			}
		},
		func(data MsgData) testCase {
			return testCase{
				name:  "Uints64",
				value: []uint64{fake.UInt64(), fake.UInt64()},
				fn:    castLotDataFieldFn(data.Uints64),
			}
		},
		func(data MsgData) testCase {
			return testCase{
				name:  "Float32",
				value: fake.Float32(5, 10, 100000),
				fn:    castLotDataFieldFn(data.Float32),
			}
		},
		func(data MsgData) testCase {
			return testCase{
				name:  "Floats32",
				value: []float32{fake.Float32(5, 10, 100000), fake.Float32(5, 10, 100000)},
				fn:    castLotDataFieldFn(data.Floats32),
			}
		},
		func(data MsgData) testCase {
			return testCase{
				name:  "Float64",
				value: fake.Float64(5, 10, 100000),
				fn:    castLotDataFieldFn(data.Float64),
			}
		},
		func(data MsgData) testCase {
			return testCase{
				name:  "Floats64",
				value: []float64{fake.Float64(5, 10, 100000), fake.Float64(5, 10, 100000)},
				fn:    castLotDataFieldFn(data.Floats64),
			}
		},
		func(data MsgData) testCase {
			value := fake.Time().Time(time.Now())
			return testCase{
				name:          "Time",
				value:         value,
				expectedValue: value.Format(time.RFC3339Nano),
				fn:            castLotDataFieldFn(data.Time),
			}
		},
		func(data MsgData) testCase {
			value := []time.Time{
				fake.Time().Time(time.Now()),
				fake.Time().Time(time.Now()),
			}
			return testCase{
				name:  "Times",
				value: value,
				expectedValue: []interface{}{
					value[0].Format(time.RFC3339Nano),
					value[1].Format(time.RFC3339Nano),
				},
				fn: castLotDataFieldFn(data.Times),
			}
		},
		func(data MsgData) testCase {
			value := net.ParseIP(fake.Internet().Ipv4())
			return testCase{
				name:  "IPAddr",
				value: value,
				fn:    castLotDataFieldFn(data.IPAddr),
			}
		},
		func(data MsgData) testCase {
			ip := fake.Internet().Ipv4()
			mask := fake.IntBetween(8, 32)
			_, addr, err := net.ParseCIDR(fmt.Sprintf("%v/%v", ip, mask))
			if err != nil {
				panic(err)
			}
			return testCase{
				name:          "IPPrefix",
				value:         *addr,
				expectedValue: addr.String(),
				fn:            castLotDataFieldFn(data.IPPrefix),
			}
		},
		func(data MsgData) testCase {
			mac, err := net.ParseMAC(fake.Internet().MacAddress())
			if err != nil {
				panic(err)
			}
			return testCase{
				name:          "MACAddr",
				value:         mac,
				expectedValue: mac.String(),
				fn:            castLotDataFieldFn(data.MACAddr),
			}
		},
	}

	var output bytes.Buffer
	outputWriter := bufio.NewWriter(&output)
	logger := factory.NewLogger(&RootContextParams{
		Out: outputWriter, LogLevel: LogLevelDebugValue,
	})
	for _, test := range tests {
		data := logger.NewData()
		tt := test(data)
		t.Run(tt.name, func(t *testing.T) {
			output.Reset()
			wantKey := fake.Lorem().Word()
			wantValue := tt.value
			data := tt.fn(wantKey, wantValue)
			if tt.expectedValue != nil {
				wantValue = tt.expectedValue
			} else {
				wantValue = jsonify(tt.value)
			}
			logger.Info().WithData(data).Msg(fake.Lorem().Sentence(3))

			outputWriter.Flush()
			var logMessage map[string]interface{}
			json.Unmarshal(output.Bytes(), &logMessage)

			gotData := logMessage["data"].(map[string]interface{})
			assert.Equal(t, map[string]interface{}{
				wantKey: wantValue,
			}, gotData)
		})
	}
}
