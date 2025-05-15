## Project Structure

```
simplewal/
├── cmd/                    # Command-line entry points (if building an executable)
│   └── simplewal/          # Main application binary
│       └── main.go         # Entry point for the CLI or server
├── internal/               # Private code (not importable by other projects)
│   ├── wal/                # Core WAL implementation
│   │   ├── wal.go          # Main WAL logic (e.g., append, read, recover)
│   │   ├── segment.go      # Segment file management (if using segmented logs)
│   │   ├── entry.go        # Log entry structure and serialization
│   │   └── wal_test.go     # Unit tests for WAL logic
│   └── storage/            # Low-level storage abstractions
│       ├── storage.go      # File I/O operations
│       └── storage_test.go # Tests for storage layer
├── pkg/                    # Public, reusable code (if exposing a library)
│   └── api/                # Public API for WAL
│       ├── api.go          # Public interfaces and methods
│       └── api_test.go     # Tests for public API
├── proto/                  # Protocol Buffers (optional, for serialization)
│   └── wal.proto           # Protobuf definitions for log entries
├── scripts/                # Utility scripts
│   └── generate.sh         # Script for generating protos or other code
├── examples/               # Example usage of the WAL
│   └── basic/              # Simple example
│       └── main.go         # Example code using the WAL
├── BUILD                   # Root Bazel build file
├── README.md               # Project documentation
├── LICENSE                 # License file (e.g., MIT, Apache-2.0)
└── .gitignore              # Git ignore file
```

## Usage

Build all
```
bazel build //...
```

Run cmd tool
```
bazel run //:wal -- 2 3
```

Run tests
```
bazel test //...
```
