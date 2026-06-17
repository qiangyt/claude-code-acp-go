# claude-code-acp-go

A Go implementation of the [Agent Client Protocol (ACP)](https://agentclientprotocol.com), inspired by [claude-code-acp](https://github.com/zed-industries/claude-code-acp).

## Overview

This project provides a Go implementation of the ACP protocol, enabling AI agents to communicate with clients using a standardized JSON-RPC 2.0 based protocol over stdio.

## Features

- **ACP Protocol Implementation**: Full support for the ACP 1.0 protocol
- **Session Management**: Create, load, and cancel sessions
- **MCP Integration**: Model Context Protocol server for tool execution
- **Permission System**: Configurable permission checking for tool operations
- **Multi-modal Prompts**: Support for text, image, and audio content
- **Streaming Updates**: Real-time session updates via notifications

## Project Structure

```
claude-code-acp-go/
├── cmd/claude-code-acp/     # CLI entry point
├── internal/
│   ├── acp/                 # ACP protocol implementation
│   ├── golden/              # Golden file testing framework
│   ├── mcp/                 # MCP server and tools
│   ├── settings/            # Settings management
│   ├── tools/               # Tool type definitions
│   ├── transport/           # NDJSON transport layer
│   └── utils/               # Utility functions
├── pkg/api/                 # Public API
├── golden/                  # Golden test files
└── e2e/                     # End-to-end tests
```

## Installation

```bash
go get github.com/anthropics/claude-code-acp-go
```

## Quick Start

```go
package main

import (
    "context"
    "os"

    api "github.com/anthropics/claude-code-acp-go/pkg/api"
    acpsdk "github.com/coder/acp-go-sdk"
)

func main() {
    // Create a new client
    client := api.NewClient()

    // Start from stdio
    ctx := context.Background()
    conn, err := client.StartFromStdio(ctx, os.Stdout, os.Stdin)
    if err != nil {
        panic(err)
    }
    defer conn.Done()

    // Initialize
    initResp, err := client.Initialize(ctx, acpsdk.InitializeRequest{
        ProtocolVersion: 1,
        ClientCapabilities: acpsdk.ClientCapabilities{
            Terminal: true,
        },
    })
    if err != nil {
        panic(err)
    }

    // Create a session
    sessionResp, err := client.NewSession(ctx, acpsdk.NewSessionRequest{
        Cwd: "/path/to/project",
    })
    if err != nil {
        panic(err)
    }

    // Use session...
    _ = sessionResp.SessionId
}
```

## Configuration

### Options

```go
client := api.NewClientWithOptions(
    api.NewOptions().
        WithName("my-agent").
        WithWorkingDir("/work/directory"),
)
```

### Settings Files

Settings are loaded from the following locations (in order of priority):

1. `$CWD/.claude/settings.local.json` (local, highest priority)
2. `$CWD/.claude/settings.json` (project)
3. `~/.claude/settings.json` (user)

Example settings file:

```json
{
  "model": "claude-3-opus",
  "permissions": {
    "allow": ["Read", "Bash(ls:*)"],
    "deny": ["Bash(rm:*)"]
  },
  "env": {
    "ANTHROPIC_API_KEY": "your-api-key"
  }
}
```

## Testing

```bash
# Run all tests
go test ./...

# Run with coverage
go test -cover ./...

# Run with race detection
go test -race ./...

# Update golden files
go test -update-golden ./...
```

## Test Coverage

| Package | Coverage |
|---------|----------|
| internal/acp | 97.7% |
| internal/mcp | 96.9% |
| internal/settings | 100% |
| internal/transport | 100% |
| internal/utils | 100% |
| pkg/api | 96.9% |

## Protocol Support

### Implemented Methods

- `initialize` - Initialize the agent
- `session/new` - Create a new session
- `session/load` - Load an existing session
- `session/cancel` - Cancel a session
- `session/prompt` - Send a user prompt
- `session/setMode` - Set session mode

### Implemented Notifications

- `session/update` - Session state updates
- `session/requestPermission` - Permission requests

## License

MIT License

## References

- [ACP Specification](https://agentclientprotocol.com)
- [Original TypeScript Implementation](https://github.com/zed-industries/claude-code-acp)
- [ACP Go SDK](https://github.com/coder/acp-go-sdk)
- [MCP Go SDK](https://github.com/modelcontextprotocol/go-sdk)
