package limiter

import (
	"fmt"
	"io"
)

// OutputLimiter wraps an io.Writer and limits the amount of data written.
// It also keeps track of whether the output was truncated.
type OutputLimiter struct {
	Writer            io.Writer
	MaxBytes          int
	BytesWritten      int
	Truncated         bool
	TruncationMessage string
}

// NewOutputLimiter creates a new OutputLimiter.
func NewOutputLimiter(writer io.Writer, maxBytes int) *OutputLimiter {
	return &OutputLimiter{
		Writer:            writer,
		MaxBytes:          maxBytes,
		BytesWritten:      0,
		Truncated:         false,
		TruncationMessage: fmt.Sprintf("\n\n[Output truncated, exceeded %d bytes limit]\n", maxBytes),
	}
}

// Write implements the io.Writer interface.
// It stops writing after MaxBytes and marks the output as truncated.
func (ol *OutputLimiter) Write(p []byte) (n int, err error) {
	// If we've already exceeded the limit, pretend we wrote all bytes
	// but don't actually write anything
	if ol.Truncated {
		return len(p), nil
	}

	remaining := ol.MaxBytes - ol.BytesWritten
	if remaining <= 0 {
		// We've reached the limit but haven't marked as truncated yet
		if !ol.Truncated {
			// Write the truncation message
			_, _ = ol.Writer.Write([]byte(ol.TruncationMessage))
			ol.Truncated = true
		}
		return len(p), nil
	}

	var writeLen int
	if len(p) > remaining {
		// Write only up to the limit
		writeLen = remaining
		written, writeErr := ol.Writer.Write(p[:writeLen])
		ol.BytesWritten += written
		err = writeErr

		// Mark as truncated and write the truncation message
		ol.Truncated = true
		_, _ = ol.Writer.Write([]byte(ol.TruncationMessage))

		// Pretend we wrote all bytes to not confuse the caller
		return len(p), err
	}

	// We can write all bytes
	written, err := ol.Writer.Write(p)
	ol.BytesWritten += written
	return written, err
}

// WasTruncated returns whether the output was truncated.
func (ol *OutputLimiter) WasTruncated() bool {
	return ol.Truncated
}
