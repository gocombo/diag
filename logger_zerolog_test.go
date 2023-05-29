package diag

import (
	"bufio"
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestZerolog_LoggerFactory(t *testing.T) {
	factory := zerologLoggerFactory{}
	t.Run("NewLogger", func(t *testing.T) {
		t.Run("returns a new logger", func(t *testing.T) {
			var output bytes.Buffer
			outputWriter := bufio.NewWriter(&output)
			logger := factory.NewLogger(&RootContextParams{
				LogLevel: LogLevelInfoValue,
				Out:      outputWriter,
			})
			assert.IsType(t, &zerologLevelLogger{}, logger)

			msg := fake.Lorem().Sentence(3)
			logger.Info().Msg(msg)

			outputWriter.Flush()
			outputStr := output.String()
			assert.Contains(t, outputStr, msg)
			assert.Contains(t, outputStr, `"msg":"`+msg+`"`)
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
}
