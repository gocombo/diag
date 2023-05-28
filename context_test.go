package diag

import (
	"io"
	"testing"

	"github.com/jaswdr/faker"
	"github.com/stretchr/testify/assert"
)

func TestContext_RootContext(t *testing.T) {
	fake := faker.New()

	t.Run("initializes a new diag context", func(t *testing.T) {
		ctx := RootContext(NewRootContextParams())
		log := Log(ctx)
		assert.NotNil(t, log)
		diagData := DiagData(ctx)
		assert.NotEmpty(t, diagData.CorrelationID)
		assert.Empty(t, diagData.Entries)
	})
	t.Run("initializes a new diag context with custom data", func(t *testing.T) {
		wantCorrelationID := fake.UUID().V4()
		ctx := RootContext(
			NewRootContextParams().
				WithLogLevel(LogLevelInfoValue).
				WithLoggerFactory(zerologLoggerFactory{}).
				WithOutput(io.Discard).
				WithPretty(fake.Bool()).
				WithCorrelationID(wantCorrelationID),
		)
		log := Log(ctx)
		assert.NotNil(t, log)
		diagData := DiagData(ctx)
		assert.Equal(t, wantCorrelationID, diagData.CorrelationID)
	})
}
