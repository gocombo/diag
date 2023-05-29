package diag

import (
	"io"
	"testing"

	"github.com/jaswdr/faker"
	"github.com/stretchr/testify/assert"
)

var fake = faker.New()

func TestContext_RootContext(t *testing.T) {

	t.Run("initializes a new diag context", func(t *testing.T) {
		ctx := RootContext(NewRootContextParams())
		log := Log(ctx)
		assert.NotNil(t, log)
		diagData := DiagData(ctx)
		assert.NotEmpty(t, diagData.CorrelationID)
		assert.Empty(t, diagData.Entries)

		assert.NotNil(t, ctx.Value(contextKeyLogger))
		assert.NotNil(t, ctx.Value(contextKeyDiagData))
		assert.NotNil(t, ctx.Value(contextKeyLoggerFactory))
		assert.Implements(t, (*LoggerFactory)(nil), ctx.Value(contextKeyLoggerFactory))
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

func TestContext_ForkContext(t *testing.T) {
	t.Run("creates a new diag context with given values", func(t *testing.T) {
		ctx := RootContext(NewRootContextParams())
		rootLog := Log(ctx)
		rootDiagData := DiagData(ctx)

		wantCorrelationID := fake.UUID().V4()

		wantEntries := map[string]string{
			fake.Lorem().Word(): fake.Lorem().Word(),
			fake.Lorem().Word(): fake.Lorem().Word(),
			fake.Lorem().Word(): fake.Lorem().Word(),
		}

		forkedCtx := ForkContext(ctx,
			ForkWithLogLevel(LogLevelInfoValue),
			ForkWithCorrelationID(wantCorrelationID),
			ForkWithAppendDiagEntries(wantEntries),
		)
		forkedLog := Log(forkedCtx)
		if ok := assert.NotSame(t, rootLog, forkedLog); !ok {
			return
		}
		forkedDiagData := DiagData(forkedCtx)
		assert.NotEqual(t, rootDiagData, forkedDiagData)
		assert.Equal(t, wantCorrelationID, forkedDiagData.CorrelationID)
		assert.Equal(t, wantEntries, forkedDiagData.Entries)

		assert.NotNil(t, forkedCtx.Value(contextKeyLoggerFactory))
		assert.Equal(t, ctx.Value(contextKeyLoggerFactory), forkedCtx.Value(contextKeyLoggerFactory))
	})
}
