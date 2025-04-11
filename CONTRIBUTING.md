# Contributing to Secure Shell Server

Thank you for considering a contribution to the Secure Shell Server project! This document outlines the process for contributing to the project and the standards expected from contributors.

## Code of Conduct

Please note that this project adheres to a [Code of Conduct](CODE_OF_CONDUCT.md). By participating, you are expected to uphold this code.

## Getting Started

### Development Environment

1. Fork the repository on GitHub
2. Clone your fork locally
   ```bash
   git clone https://github.com/yourusername/secure-shell-server.git
   cd secure-shell-server
   ```
3. Install dependencies
   ```bash
   go mod tidy
   ```

### Development Workflow

1. Create a branch for your feature or bugfix
   ```bash
   git checkout -b feature/your-feature-name
   ```

2. Make your changes, following the coding standards outlined below

3. Add tests for your changes

4. Ensure all tests pass
   ```bash
   make test
   ```

5. Ensure code passes linting
   ```bash
   make lint
   ```

6. Commit your changes
   ```bash
   git commit -m "Add your meaningful commit message here"
   ```

7. Push to your fork
   ```bash
   git push origin feature/your-feature-name
   ```

8. Open a Pull Request on GitHub

## Coding Standards

### Go Code Style

- Follow Go's official [Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments)
- Use `gofmt` to format your code
- Follow the project's existing code style
- Keep lines under 100 characters where possible

### Project-Specific Rules

1. **Simplicity**: Write clear, minimal code that avoids unnecessary features. Ensure the tool is easy to understand and maintain, following Go's philosophy of simplicity.

2. **Modularity**: Design with independent components, each handling a single responsibility. Use packages to separate concerns for easy updates and scalability.

3. **Concurrency**: Leverage Go's goroutines and channels to handle multiple tasks efficiently. Ensure safe, concurrent execution for improved performance.

4. **Explicit Error Handling**: Implement robust error handling using Go's multi-value returns. Check and handle errors explicitly to ensure reliability.

5. **Reusability**: Create reusable functions and packages to avoid code duplication. Follow DRY principles to enhance maintainability and flexibility.

### Documentation

- All exported functions, types, and variables must have proper documentation
- Include usage examples for complex functionality
- Keep documentation up-to-date with code changes

### Testing

- Write tests for all new functionality
- Maintain or improve code coverage
- Include both unit tests and integration tests
- For security-related changes, include specific security tests

## Pull Request Process

1. Ensure your code follows the coding standards
2. Update any relevant documentation
3. Include tests for your changes
4. Ensure all tests and checks pass
5. Your PR will be reviewed by at least one maintainer
6. Address any feedback from reviewers
7. Once approved, a maintainer will merge your PR

## Security Considerations

Security is a primary concern for this project. When contributing, please consider:

1. **Command Validation**: Any changes to the validator must maintain or enhance security
2. **Input Sanitization**: All input must be properly validated
3. **Execution Environment**: Changes to the execution environment must maintain security constraints
4. **Error Handling**: Security-related errors must be properly logged and handled
5. **Disclosure**: For security vulnerabilities, please follow the security disclosure policy

## Reporting Security Vulnerabilities

If you discover a security vulnerability, please DO NOT open a public issue. Instead, please send an email to [security@example.com] with details about the vulnerability. We will respond promptly.

## License

By contributing to this project, you agree that your contributions will be licensed under the project's license. See [LICENSE](LICENSE) for details.

## Questions?

If you have any questions about contributing, please open an issue or contact the maintainers.