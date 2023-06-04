package diag

import (
	"bufio"
	"bytes"
	"encoding/json"
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
