package hint

import (
	"testing"
)

func TestAnalyzeRedundantCd(t *testing.T) {
	tests := []struct {
		name       string
		command    string
		workingDir string
		wantHint   bool
		wantMsg    string
	}{
		{
			name:       "cd to current dir with && is redundant",
			command:    "cd /home/user/project && ls",
			workingDir: "/home/user/project",
			wantHint:   true,
			wantMsg:    "ls",
		},
		{
			name:       "cd to current dir with semicolon is redundant",
			command:    "cd /home/user/project; ls",
			workingDir: "/home/user/project",
			wantHint:   true,
			wantMsg:    "ls",
		},
		{
			name:       "cd to different dir is not redundant",
			command:    "cd /other/dir && ls",
			workingDir: "/home/user/project",
			wantHint:   false,
		},
		{
			name:       "cd alone is not flagged",
			command:    "cd /home/user/project",
			workingDir: "/home/user/project",
			wantHint:   false,
		},
		{
			name:       "no cd at all",
			command:    "ls -la",
			workingDir: "/home/user/project",
			wantHint:   false,
		},
		{
			name:       "cd with trailing slash matches",
			command:    "cd /home/user/project/ && ls",
			workingDir: "/home/user/project",
			wantHint:   true,
			wantMsg:    "ls",
		},
		{
			name:       "multiple commands after cd",
			command:    "cd /home/user/project && ls && echo hello",
			workingDir: "/home/user/project",
			wantHint:   true,
			wantMsg:    "ls && echo hello",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hints := Analyze(tt.command, tt.workingDir)
			found := false
			for _, h := range hints {
				if h.Type == RedundantCd {
					found = true
					if tt.wantMsg != "" && !contains(h.Message, tt.wantMsg) {
						t.Errorf("expected hint message to contain %q, got %q", tt.wantMsg, h.Message)
					}
				}
			}
			if tt.wantHint && !found {
				t.Errorf("expected redundant cd hint, got none")
			}
			if !tt.wantHint && found {
				t.Errorf("did not expect redundant cd hint, but got one")
			}
		})
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && searchSubstring(s, substr)
}

func searchSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
