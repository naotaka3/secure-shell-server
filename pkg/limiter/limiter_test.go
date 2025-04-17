package limiter

import (
	"bytes"
	"strings"
	"testing"

	"github.com/alecthomas/assert/v2"
)

func TestOutputLimiter(t *testing.T) {
	t.Run("Should write under limit", func(t *testing.T) {
		buf := &bytes.Buffer{}
		limiter := NewOutputLimiter(buf, 100)

		toWrite := make([]byte, 50)
		for i := range toWrite {
			toWrite[i] = 'a'
		}

		n, err := limiter.Write(toWrite)
		assert.NoError(t, err)
		assert.Equal(t, 50, n)
		assert.Equal(t, 50, limiter.BytesWritten)
		assert.False(t, limiter.WasTruncated())
		assert.Equal(t, string(toWrite), buf.String())
	})

	t.Run("Should truncate at limit", func(t *testing.T) {
		buf := &bytes.Buffer{}
		limiter := NewOutputLimiter(buf, 100)

		// First write under limit
		toWrite1 := make([]byte, 50)
		for i := range toWrite1 {
			toWrite1[i] = 'a'
		}

		n1, err := limiter.Write(toWrite1)
		assert.NoError(t, err)
		assert.Equal(t, 50, n1)

		// Second write exceeding limit
		toWrite2 := make([]byte, 100)
		for i := range toWrite2 {
			toWrite2[i] = 'b'
		}

		n2, err := limiter.Write(toWrite2)
		assert.NoError(t, err)
		assert.Equal(t, 100, n2) // Reports full length even though truncated
		assert.True(t, limiter.WasTruncated())

		// Should contain 50 'a's + 50 'b's + truncation message
		expected := string(toWrite1) + string(toWrite2[:50]) + limiter.TruncationMessage
		assert.Equal(t, expected, buf.String())
	})

	t.Run("Should ignore writes after truncation", func(t *testing.T) {
		buf := &bytes.Buffer{}
		limiter := NewOutputLimiter(buf, 10)

		// Write enough to truncate
		toWrite1 := make([]byte, 20)
		for i := range toWrite1 {
			toWrite1[i] = 'a'
		}

		_, err := limiter.Write(toWrite1)
		assert.NoError(t, err)
		assert.True(t, limiter.WasTruncated())
		bufContent := buf.String()

		// Try writing more
		toWrite2 := []byte("more data")
		_, err = limiter.Write(toWrite2)
		assert.NoError(t, err)

		// Buffer should not have changed
		assert.Equal(t, bufContent, buf.String())
	})

	t.Run("Should handle exact size match", func(t *testing.T) {
		buf := &bytes.Buffer{}
		limiter := NewOutputLimiter(buf, 10)

		// Write exactly at the limit
		toWrite := make([]byte, 10)
		for i := range toWrite {
			toWrite[i] = 'a'
		}

		n, err := limiter.Write(toWrite)
		assert.NoError(t, err)
		assert.Equal(t, 10, n)
		assert.Equal(t, 10, limiter.BytesWritten)
		assert.False(t, limiter.WasTruncated())

		// Another write should be truncated
		_, err = limiter.Write([]byte("more"))
		assert.NoError(t, err)
		assert.True(t, limiter.WasTruncated())
		assert.True(t, strings.HasPrefix(buf.String(), string(toWrite)))
		assert.True(t, strings.Contains(buf.String(), "truncated"))
	})
}
