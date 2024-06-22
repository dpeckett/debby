package boolean_test

import (
	"testing"

	"github.com/dpeckett/debby/internal/types/boolean"
	"github.com/stretchr/testify/require"
)

func TestBoolean(t *testing.T) {
	t.Run("MarshalText", func(t *testing.T) {
		t.Run("true", func(t *testing.T) {
			b := boolean.Boolean(true)

			text, err := b.MarshalText()
			require.NoError(t, err)

			require.Equal(t, "yes", string(text))
		})

		t.Run("false", func(t *testing.T) {
			b := boolean.Boolean(false)

			text, err := b.MarshalText()
			require.NoError(t, err)

			require.Equal(t, "no", string(text))
		})
	})

	t.Run("UnmarshalText", func(t *testing.T) {
		t.Run("yes", func(t *testing.T) {
			var b boolean.Boolean
			require.NoError(t, b.UnmarshalText([]byte("yes")))

			require.True(t, bool(b))
		})

		t.Run("no", func(t *testing.T) {
			var b boolean.Boolean
			require.NoError(t, b.UnmarshalText([]byte("no")))

			require.False(t, bool(b))
		})

		t.Run("true", func(t *testing.T) {
			var b boolean.Boolean
			require.NoError(t, b.UnmarshalText([]byte("true")))

			require.True(t, bool(b))
		})

		t.Run("false", func(t *testing.T) {
			var b boolean.Boolean
			require.NoError(t, b.UnmarshalText([]byte("false")))

			require.False(t, bool(b))
		})
	})
}
