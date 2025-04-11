# Coding Rules for Go Tool Development

Please follow these coding rules when developing tools in Go. These guidelines are designed to ensure that the code is clean, maintainable, and adheres to Go's best practices.

+ Simplicity: Write clear, minimal code that avoids unnecessary features. Ensure the tool is easy to understand and maintain, following Go's philosophy of simplicity.

+ Modularity: Design the tool with independent components, each handling a single responsibility. Use packages to separate concerns for easy updates and scalability.
+ Concurrency: Leverage Go's goroutines and channels to handle multiple tasks efficiently. Ensure safe, concurrent execution for improved performance.
+ Explicit Error Handling: Implement robust error handling using Go's multi-value returns. Check and handle errors explicitly to ensure reliability.
+ Reusability: Create reusable functions and packages to avoid code duplication. Follow DRY principles to enhance maintainability and flexibility.
