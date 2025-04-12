package runner

import (
	"io"
	"testing"

	"github.com/alecthomas/assert/v2"

	"github.com/shimizu1995/secure-shell-server/pkg/config"
	"github.com/shimizu1995/secure-shell-server/pkg/logger"
	"github.com/shimizu1995/secure-shell-server/pkg/validator"
)

func setupCustomConfig() *config.ShellCommandConfig {
	return &config.ShellCommandConfig{
		AllowedDirectories: []string{"/home", "/tmp"},
		AllowCommands: []config.AllowCommand{
			{Command: "ls"},
			{Command: "cat"},
			{Command: "echo"},
			{Command: "grep"},
			{Command: "find"},
			{Command: "git", SubCommands: []string{"status", "log", "diff"}, DenySubCommands: []string{"push", "commit"}},
		},
		DenyCommands: []config.DenyCommand{
			{Command: "rm", Message: "Remove command is not allowed"},
			{Command: "sudo", Message: "Sudo is not allowed for security reasons"},
		},
		DefaultErrorMessage: "Command not allowed by security policy",
		WorkingDir:          "",
		MaxExecutionTime:    config.DefaultExecutionTimeout,
	}
}

func TestSafeRunner_CommandValidation(t *testing.T) {
	cfg := setupCustomConfig()
	log := logger.New()
	validatorObj := validator.New(cfg, log)
	safeRunner := New(cfg, validatorObj, log)

	// Set up the runner but don't capture output for validation tests
	// This avoids data races with concurrent command execution
	safeRunner.SetOutputs(io.Discard, io.Discard)

	// 基本的な許可されたコマンド
	t.Run("BasicAllowedCommand", func(t *testing.T) {
		ctx := t.Context()
		err := safeRunner.RunCommand(ctx, "echo hello")
		assert.NoError(t, err)
	})

	// 複数行の許可されたコマンド
	t.Run("MultilineAllowedCommands", func(t *testing.T) {
		ctx := t.Context()
		err := safeRunner.RunCommand(ctx, "echo hello\nls -l")
		assert.NoError(t, err)
	})

	// 明示的に拒否されたコマンド
	t.Run("ExplicitlyDeniedCommand", func(t *testing.T) {
		ctx := t.Context()
		err := safeRunner.RunCommand(ctx, "rm -rf /tmp/test")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "command execution error: command \"rm\" is denied: Remove command is not allowed")
	})

	// 許可リストにないコマンド
	t.Run("CommandNotInAllowList", func(t *testing.T) {
		ctx := t.Context()
		err := safeRunner.RunCommand(ctx, "chmod 777 file.txt")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "command execution error: command \"chmod\" is not permitted: Command not allowed by security policy")
	})

	// 許可されたサブコマンドを持つコマンド
	t.Run("CommandWithAllowedSubcommand", func(t *testing.T) {
		ctx := t.Context()
		err := safeRunner.RunCommand(ctx, "git status")
		assert.NoError(t, err)
	})

	// 拒否されたサブコマンドを持つコマンド
	t.Run("CommandWithDeniedSubcommand", func(t *testing.T) {
		ctx := t.Context()
		err := safeRunner.RunCommand(ctx, "git push")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "command execution error: subcommand \"push\" is not allowed for command \"git\"")
	})

	// 許可リストにないサブコマンドを持つコマンド
	t.Run("CommandWithSubcommandNotInAllowList", func(t *testing.T) {
		ctx := t.Context()
		err := safeRunner.RunCommand(ctx, "git clone https://github.com/example/repo.git")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "command execution error: subcommand \"clone\" is not allowed for command \"git\"")
	})

	// コマンドの構文エラー
	t.Run("SyntaxErrorInCommand", func(t *testing.T) {
		ctx := t.Context()
		err := safeRunner.RunCommand(ctx, "echo 'unclosed string")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "command execution error: parse error: ")
	})

	// リダイレクションを持つコマンド
	t.Run("CommandWithRedirection", func(t *testing.T) {
		ctx := t.Context()
		err := safeRunner.RunCommand(ctx, "echo hello > /tmp/test.txt")
		assert.NoError(t, err)
	})

	// 空のコマンド
	t.Run("EmptyCommand", func(t *testing.T) {
		ctx := t.Context()
		err := safeRunner.RunCommand(ctx, "")
		assert.NoError(t, err)
	})
}

func TestSafeRunner_PipelineValidation(t *testing.T) {
	cfg := setupCustomConfig()
	// パイプラインテスト用にprintf コマンドを許可リストに追加
	cfg.AllowCommands = append(cfg.AllowCommands, config.AllowCommand{Command: "printf"})

	log := logger.New()
	validatorObj := validator.New(cfg, log)
	safeRunner := New(cfg, validatorObj, log)

	// Set up the runner but don't capture output for validation tests
	safeRunner.SetOutputs(io.Discard, io.Discard)

	// すべて許可されたコマンドのパイプライン
	t.Run("AllAllowedCommands", func(t *testing.T) {
		ctx := t.Context()
		err := safeRunner.RunCommand(ctx, "echo 'hello' | grep hello")
		assert.NoError(t, err)
	})

	// 1つの拒否されたコマンドを含むパイプライン
	t.Run("OneDisallowedCommand", func(t *testing.T) {
		ctx := t.Context()
		err := safeRunner.RunCommand(ctx, "echo 'hello world' | grep hello | sudo cat")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "command \"sudo\" is denied")
	})

	// 中間に拒否されたコマンドを含む複雑なパイプライン
	t.Run("ComplexPipelineWithDisallowedCommand", func(t *testing.T) {
		ctx := t.Context()
		err := safeRunner.RunCommand(ctx, "echo 'test' | sudo grep test | cat")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "command \"sudo\" is denied")
	})

	// 許可リストにないコマンドを含むパイプライン
	t.Run("CommandNotInAllowlist", func(t *testing.T) {
		ctx := t.Context()
		err := safeRunner.RunCommand(ctx, "echo 'test' | grep test | awk '{print $1}'")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "command \"awk\" is not permitted")
	})

	// シンプルな許可されたコマンド
	t.Run("SimpleAllowedCommand", func(t *testing.T) {
		ctx := t.Context()
		err := safeRunner.RunCommand(ctx, "echo 'single command'")
		assert.NoError(t, err)
	})
}
