package limiter

import (
	"bytes"
	"fmt"
	"strings"
	"testing"

	"github.com/alecthomas/assert/v2"
)

// TestOutputLimiterBasic tests basic writing functionality.
func TestOutputLimiterBasic(t *testing.T) {
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
		// Expected message now includes the remaining bytes info
		expectedPrefix := string(toWrite1) + string(toWrite2[:50])
		expected := expectedPrefix + fmt.Sprintf("\n\n[Output truncated, exceeded %d bytes limit. %d bytes remaining]\n"+
			"If you need to view the complete output, consider using commands like tail or modifying your command to ensure the output stays within the limits.",
			limiter.MaxBytes, 50)
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
}

// TestOutputLimiterEdgeCases tests edge cases like exact size matches.
func TestOutputLimiterEdgeCases(t *testing.T) {
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

// TestOutputLimiterRemainingBytes tests the remaining bytes functionality.
func TestOutputLimiterRemainingBytes(t *testing.T) {
	t.Run("Should track total input bytes and remaining bytes", func(t *testing.T) {
		buf := &bytes.Buffer{}
		limiter := NewOutputLimiter(buf, 100)

		// First write under limit
		toWrite1 := make([]byte, 60)
		for i := range toWrite1 {
			toWrite1[i] = 'a'
		}

		n1, err := limiter.Write(toWrite1)
		assert.NoError(t, err)
		assert.Equal(t, 60, n1)
		assert.Equal(t, 60, limiter.TotalInputBytes)
		assert.Equal(t, 0, limiter.GetRemainingBytes())

		// Second write exceeding limit
		toWrite2 := make([]byte, 80)
		for i := range toWrite2 {
			toWrite2[i] = 'b'
		}

		n2, err := limiter.Write(toWrite2)
		assert.NoError(t, err)
		assert.Equal(t, 80, n2)
		assert.Equal(t, 140, limiter.TotalInputBytes) // 60 + 80
		assert.True(t, limiter.WasTruncated())
		assert.Equal(t, 40, limiter.GetRemainingBytes()) // 40 remaining bytes that couldn't be written

		// Third write that should be ignored but still counted in total
		toWrite3 := make([]byte, 50)
		for i := range toWrite3 {
			toWrite3[i] = 'c'
		}

		n3, err := limiter.Write(toWrite3)
		assert.NoError(t, err)
		assert.Equal(t, 50, n3)
		assert.Equal(t, 190, limiter.TotalInputBytes)    // 60 + 80 + 50
		assert.Equal(t, 90, limiter.GetRemainingBytes()) // 40 + 50 remaining bytes
	})

	t.Run("Should include remaining bytes in truncation message", func(t *testing.T) {
		buf := &bytes.Buffer{}
		limiter := NewOutputLimiter(buf, 20)

		// Write enough to truncate
		toWrite := make([]byte, 50)
		for i := range toWrite {
			toWrite[i] = 'a'
		}

		_, err := limiter.Write(toWrite)
		assert.NoError(t, err)
		assert.True(t, limiter.WasTruncated())

		// Check that the truncation message contains the remaining bytes info
		expectedMessage := "[Output truncated, exceeded 20 bytes limit. 30 bytes remaining]"
		assert.True(t, strings.Contains(buf.String(), expectedMessage))
	})
}
