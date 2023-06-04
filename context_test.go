package diag

import (
	"context"
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

func TestContext_Accessors(t *testing.T) {
	t.Run("Log", func(t *testing.T) {
		t.Run("returns the logger if one is set", func(t *testing.T) {
			ctx := RootContext(NewRootContextParams())
			assert.NotNil(t, Log(ctx))
			assert.Implements(t, (*LevelLogger)(nil), Log(ctx))
		})
		t.Run("panics if no logger", func(t *testing.T) {
			assert.PanicsWithError(
				t,
				"context does not contain a logger",
				func() { Log(context.Background()) })
		})
	})
	t.Run("DiagData", func(t *testing.T) {
		t.Run("returns the diag data if one is set", func(t *testing.T) {
			ctx := RootContext(NewRootContextParams())
			assert.NotNil(t, DiagData(ctx))
			assert.IsType(t, DiagData(ctx), ContextDiagData{})
		})
		t.Run("panics if no diag data", func(t *testing.T) {
			assert.PanicsWithError(
				t,
				"context does not contain diag data",
				func() { DiagData(context.Background()) })
		})
	})
	t.Run("LoggerFactory", func(t *testing.T) {
		t.Run("returns the logger factory if one is set", func(t *testing.T) {
			ctx := RootContext(NewRootContextParams())
			assert.NotNil(t, getLoggerFactory(ctx))
			assert.Implements(t, (*LoggerFactory)(nil), getLoggerFactory(ctx))
		})
		t.Run("panics if no logger factory", func(t *testing.T) {
			assert.PanicsWithError(
				t,
				"context does not contain a logger factory",
				func() { getLoggerFactory(context.Background()) })
		})
	})
}

func TestContext_DiagifyContext(t *testing.T) {
	t.Run("creates a child diag context with given values", func(t *testing.T) {
		diagContext := RootContext(NewRootContextParams())

		wantCorrelationID := fake.UUID().V4()

		wantEntries := map[string]string{
			fake.Lorem().Word(): fake.Lorem().Word(),
			fake.Lorem().Word(): fake.Lorem().Word(),
			fake.Lorem().Word(): fake.Lorem().Word(),
		}

		type testContextKey string
		var testContextKeyFoo testContextKey = testContextKey(fake.Lorem().Word())
		wantTestValue := fake.Lorem().Sentence(3)
		parentCtx := context.WithValue(context.Background(), testContextKeyFoo, wantTestValue)

		diagifiedCtx := DiagifyContext(
			parentCtx,
			diagContext,
			WithLogLevel(LogLevelInfoValue),
			WithCorrelationID(wantCorrelationID),
			WithAppendDiagEntries(wantEntries),
		)
		diagifiedLog := Log(diagifiedCtx)
		forkedDiagData := DiagData(diagifiedCtx)
		assert.NotNil(t, diagifiedLog)
		assert.Equal(t, wantCorrelationID, forkedDiagData.CorrelationID)
		assert.Equal(t, wantEntries, forkedDiagData.Entries)
		assert.Equal(t, wantTestValue, diagifiedCtx.Value(testContextKeyFoo))

		assert.NotNil(t, diagifiedCtx.Value(contextKeyLoggerFactory))
		assert.Equal(t, diagContext.Value(contextKeyLoggerFactory), diagifiedCtx.Value(contextKeyLoggerFactory))
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
			WithLogLevel(LogLevelInfoValue),
			WithCorrelationID(wantCorrelationID),
			WithAppendDiagEntries(wantEntries),
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
	t.Run("creates a copy but not child context", func(t *testing.T) {
		type testContextKey string
		var testContextKeyFoo testContextKey = testContextKey(fake.Lorem().Word())
		ctx, cancel := context.WithCancel(RootContext(NewRootContextParams()))
		ctx = context.WithValue(ctx, testContextKeyFoo, fake.Lorem().Sentence(3))

		cancel()

		forkedCtx := ForkContext(ctx)
		assert.Nil(t, forkedCtx.Value(testContextKeyFoo))
		assert.Nil(t, forkedCtx.Err())
	})
}
