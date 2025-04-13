# Secure Shell Server

A Go-based MCP (Model Context Protocol) server designed to prevent Large Language Models (LLMs) from executing dangerous shell commands. The server provides a sandboxed environment that validates all commands against an allowlist, restricting operations to only those explicitly permitted.

## Important Security Notice

While this server implements multiple security measures, it cannot guarantee complete protection against all malicious commands or sophisticated attacks. Users should:

- Carefully configure the allowlist to include only necessary commands
- Regularly review logs for suspicious activity
- Use this as one layer in a comprehensive security strategy
- Not rely solely on this tool for high-security environments

## Features

- **Command Allowlisting**: Parses user-provided shell scripts and validates that only allowed commands are used.
- **Secure Execution**: Uses a custom runner to enforce the allowlist during command execution.
- **Directory Restrictions**: Limits file system access to only explicitly allowed directories.
- **Path Validation**: Verifies all path arguments to prevent access to unauthorized directories.
- **Timeout Enforcement**: Automatically terminates long-running commands to prevent resource exhaustion.
- **Detailed Logging**: Logs all command attempts and execution results for auditing and debugging.
- **MCP Server Mode**: Functions as a Model Context Protocol (MCP) server for secure command execution.

## Installation

```bash
go get /path/to/secure-shell-server
cd /path/to/secure-shell-server
make build
```

The binaries will be available in the `bin/` directory.

## Usage

### Starting the MCP Server

```bash
./bin/server -config=/path/to/config.json
```

### Command-Line Options for server

- `-config`: Path to configuration file
- `-stdio`: Use stdin/stdout for MCP communication
- `-port`: Port to listen on (default: 8080, when not using stdio)

## Claude Desktop Setup

To use secure-shell-server with Claude Desktop:

1. Edit your Claude Desktop configuration file located at:
   - macOS: `~/Library/Application Support/Claude/claude_desktop_config.json`
   - Windows: `%APPDATA%\Claude\claude_desktop_config.json`
2. Add the following to your configuration under the `tools` section:

```json
"shell": {
  "command": "/path/to/secure-shell-server/bin/server",
  "args": [
    "-config",
    "~/path/to/your/config.json"
  ]
}
```

3. Create a configuration file at a location of your choice (such as `~/.mcp_shell_config.json` on macOS or appropriate path on Windows) with your desired settings
4. Restart Claude Desktop to apply the changes

## Design and Implementation

The Secure Shell Server follows a modular design with the following components:

1. **Validator**: Parses and validates shell commands against an allowlist.
2. **Runner**: Executes validated commands using a secure custom runner.
3. **Config**: Manages allowlist configuration and runtime settings.
4. **Logger**: Provides detailed logging of all command attempts and results.
5. **Server**: MCP interface for secure shell execution service.

## Security Considerations

- Only explicitly allowlisted commands can be executed.
- File system access is restricted to specified directories.
- Command execution is constrained by a configurable timeout.
- Scripts are validated before execution to prevent dangerous operations.
- Special handling for commands like `find` and `xargs` that could execute other commands.
- Path arguments are validated to prevent access to restricted areas.

### Limitations

- Cannot prevent all sophisticated attacks or command chaining techniques.
- Command injection may still be possible in certain edge cases.
- Security effectiveness depends heavily on proper configuration.

## Development

### Prerequisites

- Go 1.20 or later
- `mvdan.cc/sh/v3` package

### Building

```bash
make build
```

### Testing

```bash
make test
```

### Linting

```bash
make lint
```

## License

This project is licensed under the terms found in the LICENSE file.

### Third-Party Licenses

This project uses the following third-party libraries:

- `mvdan.cc/sh/v3`: Licensed under the BSD-3-Clause license.
