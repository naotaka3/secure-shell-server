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

// checkAbsolutePaths finds absolute paths in the command that are under workingDir
// and suggests shorter relative alternatives.
func checkAbsolutePaths(command string, workingDir string) []Hint {
	if workingDir == "" {
		return nil
	}

	cleanWorking := filepath.Clean(workingDir)
	// Ensure workingDir ends with separator for prefix matching
	prefix := cleanWorking + string(filepath.Separator)

	var hints []Hint
	seen := make(map[string]bool)

	// Split command into shell tokens (whitespace-separated).
	// This is a simple heuristic — does not handle quoting, but covers common cases.
	tokens := strings.Fields(command)
	for _, token := range tokens {
		if !filepath.IsAbs(token) {
			continue
		}

		cleanToken := filepath.Clean(token)

		if seen[cleanToken] {
			continue
		}
		seen[cleanToken] = true

		var relPath string
		switch {
		case cleanToken == cleanWorking:
			relPath = "."
		case strings.HasPrefix(cleanToken, prefix):
			relPath = "./" + cleanToken[len(prefix):]
		default:
			continue
		}

		msg := fmt.Sprintf(
			"[Hint] %q can be shortened to %q (relative to current directory). "+
				"This saves tokens.",
			token, relPath,
		)
		hints = append(hints, Hint{Type: AbsolutePathConvertible, Message: msg})
	}

	return hints
}
