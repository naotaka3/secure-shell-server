package limiter

import (
	"fmt"
	"io"
)

// OutputLimiter wraps an io.Writer and limits the amount of data written.
// It also keeps track of whether the output was truncated and the total size of the original output.
type OutputLimiter struct {
	Writer            io.Writer
	MaxBytes          int
	BytesWritten      int
	TotalInputBytes   int
	Truncated         bool
	TruncationMessage string
}

// NewOutputLimiter creates a new OutputLimiter.
func NewOutputLimiter(writer io.Writer, maxBytes int) *OutputLimiter {
	return &OutputLimiter{
		Writer:            writer,
		MaxBytes:          maxBytes,
		BytesWritten:      0,
		TotalInputBytes:   0,
		Truncated:         false,
		TruncationMessage: fmt.Sprintf("\n\n[Output truncated, exceeded %d bytes limit]\n", maxBytes),
	}
}

// Write implements the io.Writer interface.
// It stops writing after MaxBytes and marks the output as truncated.
func (ol *OutputLimiter) Write(p []byte) (n int, err error) {
	// Always track the total input size
	ol.TotalInputBytes += len(p)

	// If we've already exceeded the limit, pretend we wrote all bytes
	// but don't actually write anything
	if ol.Truncated {
		return len(p), nil
	}

	remaining := ol.MaxBytes - ol.BytesWritten
	if remaining <= 0 {
		// We've reached the limit but haven't marked as truncated yet
		if !ol.Truncated {
			// Write the truncation message with remaining size info
			_, _ = ol.Writer.Write([]byte(ol.getTruncationMessage()))
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
		_, _ = ol.Writer.Write([]byte(ol.getTruncationMessage()))

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

// GetRemainingBytes returns the number of bytes that couldn't be written due to truncation.
func (ol *OutputLimiter) GetRemainingBytes() int {
	if !ol.Truncated {
		return 0
	}
	return ol.TotalInputBytes - ol.BytesWritten
}

// getTruncationMessage returns a formatted truncation message with information about
// the remaining output size.
func (ol *OutputLimiter) getTruncationMessage() string {
	remaining := ol.TotalInputBytes - ol.BytesWritten
	return fmt.Sprintf("\n\n[Output truncated, exceeded %d bytes limit. %d bytes remaining]\n",
		ol.MaxBytes, remaining)
}
