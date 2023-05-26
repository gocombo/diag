package diag

import (
	"io"
	"testing"

	"github.com/gofrs/uuid"
)

func BenchmarkLogger(b *testing.B) {
	ctx := RootContext(
		NewRootContextParams().
			WithOutput(io.Discard).
			WithCorrelationID(uuid.Must(uuid.NewV4()).String()),
	)
	log := Log(ctx)
	b.Run("WithDataFn", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			log.Debug().
				WithDataFn(func(data MsgData) {
					data.Str("key", "value")
					data.Str("key1", "value1")
					data.Str("key2", "value2").
						Str("key3", "value3")
				}).
				Msg("Fibonacci is everywhere")
		}
	})
	b.Run("WithData", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			log.Debug().
				WithData(log.NewData().
					Str("key", "value").
					Str("key1", "value1").
					Str("key2", "value2").
					Str("key3", "value3")).
				Msg("Fibonacci is everywhere")
		}
	})
}
