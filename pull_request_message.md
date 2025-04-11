## Fix Go modules compatibility issues in validator and runner packages

Updated the code to work properly with Go modules by:

- Fixed CallExpr handling in the validator.go file to correctly work with the mvdan.cc/sh/v3/syntax package
- Updated the interp.Config usage in runner.go to match the latest API
- Improved code quality by fixing various linting issues:
  - Replaced magic numbers with named constants
  - Fixed printf-like formatting function names (LogError → LogErrorf and LogInfo → LogInfof)
  - Fixed exitAfterDefer issues by restructuring code to avoid os.Exit() in the middle of functions
  - Fixed unused parameter warnings

All tests are now passing, and the code successfully builds and passes the linter.