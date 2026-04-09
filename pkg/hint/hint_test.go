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

func TestAnalyzeAbsolutePath(t *testing.T) {
	tests := []struct {
		name       string
		command    string
		workingDir string
		wantHint   bool
		wantMsg    string
	}{
		{
			name:       "absolute path as command can be relative",
			command:    "/home/user/project/bin/main",
			workingDir: "/home/user/project",
			wantHint:   true,
			wantMsg:    "./bin/main",
		},
		{
			name:       "absolute path as argument can be relative",
			command:    "cat /home/user/project/README.md",
			workingDir: "/home/user/project",
			wantHint:   true,
			wantMsg:    "./README.md",
		},
		{
			name:       "absolute path outside working dir is not flagged",
			command:    "cat /etc/hosts",
			workingDir: "/home/user/project",
			wantHint:   false,
		},
		{
			name:       "relative path is not flagged",
			command:    "cat ./README.md",
			workingDir: "/home/user/project",
			wantHint:   false,
		},
		{
			name:       "working dir path itself is flagged as .",
			command:    "ls /home/user/project",
			workingDir: "/home/user/project",
			wantHint:   true,
			wantMsg:    ".",
		},
		{
			name:       "multiple absolute paths generate hints",
			command:    "cp /home/user/project/a.txt /home/user/project/b.txt",
			workingDir: "/home/user/project",
			wantHint:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hints := Analyze(tt.command, tt.workingDir)
			found := false
			for _, h := range hints {
				if h.Type == AbsolutePathConvertible {
					found = true
					if tt.wantMsg != "" && !contains(h.Message, tt.wantMsg) {
						t.Errorf("expected hint message to contain %q, got %q", tt.wantMsg, h.Message)
					}
				}
			}
			if tt.wantHint && !found {
				t.Errorf("expected absolute path hint, got none")
			}
			if !tt.wantHint && found {
				t.Errorf("did not expect absolute path hint, but got one")
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
