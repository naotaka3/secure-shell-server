# Coding Rules for Go Tool Development

## Rules

When developing tools in Go, it's essential to adhere to the language's conventions and best practices. Below are the coding rules to follow:

Please follow these coding rules when developing tools in Go. These guidelines are designed to ensure that the code is clean, maintainable, and adheres to Go's best practices.

+ Simplicity: Write clear, minimal code that avoids unnecessary features. Ensure the tool is easy to understand and maintain, following Go's philosophy of simplicity.

+ Modularity: Design the tool with independent components, each handling a single responsibility. Use packages to separate concerns for easy updates and scalability.
+ Concurrency: Leverage Go's goroutines and channels to handle multiple tasks efficiently. Ensure safe, concurrent execution for improved performance.
+ Explicit Error Handling: Implement robust error handling using Go's multi-value returns. Check and handle errors explicitly to ensure reliability.
+ Reusability: Create reusable functions and packages to avoid code duplication. Follow DRY principles to enhance maintainability and flexibility.

## Folder Structure

Please generate a Go project structure following best practices. The structure should include:

+ `cmd/`: Application entry point(s). Each subfolder represents a separate binary.
+ `internal/`: Application-specific packages not meant to be imported by other projects.
+ `pkg/`: Reusable packages that can be imported by external applications.
+ `api/`: Protocol definitions (e.g., OpenAPI specs or Protobuf files).
+ `web/`: Static web assets or frontend code (if applicable).
+ `scripts/`: Utility scripts for building, testing, or deployment.
+ `test/`: Integration or end-to-end tests outside the core module.

Also include a basic `go.mod` file and a `README.md` placeholder.

Assume the application name is `myapp`. Please scaffold the folders with sample files (e.g., `main.go`, `config.go`, etc.) and basic content or comments.
