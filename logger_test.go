package diag

import (
	"bufio"
	"bytes"
	"encoding/json"
	"testing"

	"github.com/gofrs/uuid"
	"github.com/jaswdr/faker"
	"github.com/stretchr/testify/assert"
)

// TestContext used for testing purposes only
type TestContext struct {
	CorrelationID string `json:"correlationId"`
}

// TestLogMessage used for testing purposes only
type TestLogMessage struct {
	Level   string      `json:"level"`
	Msg     string      `json:"msg"`
	Time    string      `json:"time"`
	Context TestContext `json:"context"`
}

func TestLoggerLevel(t *testing.T) {
	fake := faker.New()

	var output bytes.Buffer
	outputWriter := bufio.NewWriter(&output)

	ctx := RootContext(
		NewRootContextParams().
			WithLogLevel(LogLevelTraceValue).
			WithOutput(outputWriter).
			WithCorrelationID(uuid.Must(uuid.NewV4()).String()),
	)
	log := Log(ctx)

	type testCase struct {
		log       LogLevelEvent
		wantLevel string
	}

	tests := []testCase{
		{log: log.Error(), wantLevel: "error"},
		{log: log.Warn(), wantLevel: "warn"},
		{log: log.Info(), wantLevel: "info"},
		{log: log.Debug(), wantLevel: "debug"},
		{log: log.Trace(), wantLevel: "trace"},
	}

	for _, tt := range tests {
		t.Run(tt.wantLevel, func(t *testing.T) {
			output.Reset()
			msg := fake.Lorem().Sentence(3)
			tt.log.Msg(msg)
			outputWriter.Flush()

			var logMessage TestLogMessage
			assert.NoError(t, json.Unmarshal(output.Bytes(), &logMessage))

			assert.Equal(t, TestLogMessage{
				Level: tt.wantLevel,
				Msg:   msg,
				Time:  logMessage.Time,
				Context: TestContext{
					CorrelationID: logMessage.Context.CorrelationID,
				},
			}, logMessage)
		})
	}
}
