package diag

import (
	"bufio"
	"bytes"
	"encoding/hex"
	"encoding/json"
	"net"
	"testing"

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
