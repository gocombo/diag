package diag

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRootContext(t *testing.T) {
	t.Run("initializes a new diag context", func(t *testing.T) {
		ctx := RootContext(NewRootContextParams())
		log := Log(ctx)
		assert.NotNil(t, log)
		diagData := DiagData(ctx)
		assert.NotEmpty(t, diagData.CorrelationID)
		assert.Empty(t, diagData.Entries)
	})
}
