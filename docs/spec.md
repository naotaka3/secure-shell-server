# Secure Shell Specification

## File Structure

.
├── CHANGELOG.md
├── CODE_OF_CONDUCT.md
├── CONTRIBUTING.md
├── Dockerfile
├── Dockerfile.dev
├── IMPLEMENTATION.md
├── LICENSE
├── Makefile
├── README.md
├── cmd
│   └── secure-shell
│       ├── main.go
│       └── main_test.go
├── docker-compose.yml
├── docs
│   └── usage.md
├── examples
│   ├── README.md
│   ├── api_example.go
│   └── example.sh
├── go.mod
├── go.sum
├── main.go
├── main_test.go
└── pkg
    ├── config
    │   ├── config.go
    │   └── config_test.go
    ├── logger
    │   ├── logger.go
    │   └── logger_test.go
    ├── runner
    │   ├── runner.go
    │   └── runner_test.go
    └── validator
        ├── validator.go
        └── validator_test.go
