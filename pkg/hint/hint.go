package hint

import (
	"fmt"
	"path/filepath"
	"strings"
)

// Type identifies the kind of token-saving hint.
type Type int

const (
	// RedundantCd indicates an unnecessary cd to the current working directory.
	RedundantCd Type = iota
	// AbsolutePathConvertible indicates an absolute path that could be relative.
	AbsolutePathConvertible
)

// Hint represents a single token-saving suggestion.
type Hint struct {
	Type    Type
	Message string
}

// Analyze inspects a command string against the current working directory
// and returns any token-saving hints.
func Analyze(command string, workingDir string) []Hint {
	var hints []Hint

	if h, ok := checkRedundantCd(command, workingDir); ok {
		hints = append(hints, h)
	}

	hints = append(hints, checkAbsolutePaths(command, workingDir)...)

	return hints
}

// checkRedundantCd detects patterns like "cd /current/dir && cmd" or "cd /current/dir; cmd"
// where the cd target matches the current working directory.
func checkRedundantCd(command string, workingDir string) (Hint, bool) {
	if !strings.HasPrefix(command, "cd ") {
		return Hint{}, false
	}

	// Try to find the separator (&& or ;) after the cd command
	rest := command[3:] // after "cd "

	var cdTarget, remainder string
	if idx := strings.Index(rest, "&&"); idx != -1 {
		cdTarget = strings.TrimSpace(rest[:idx])
		remainder = strings.TrimSpace(rest[idx+2:])
	} else if idx := strings.Index(rest, ";"); idx != -1 {
		cdTarget = strings.TrimSpace(rest[:idx])
		remainder = strings.TrimSpace(rest[idx+1:])
	} else {
		// cd alone with no following command — not worth hinting
		return Hint{}, false
	}

	if remainder == "" {
		return Hint{}, false
	}

	// Normalize paths for comparison
	cleanTarget := filepath.Clean(cdTarget)
	cleanWorking := filepath.Clean(workingDir)

	if cleanTarget == cleanWorking {
		msg := fmt.Sprintf(
			"[Hint] The cd to %q is unnecessary — you are already in that directory. "+
				"You can save tokens by running just: %s",
			cdTarget, remainder,
		)
		return Hint{Type: RedundantCd, Message: msg}, true
	}

	return Hint{}, false
}

// checkAbsolutePaths is a placeholder — implemented in Task 2.
func checkAbsolutePaths(_ string, _ string) []Hint {
	return nil
}
