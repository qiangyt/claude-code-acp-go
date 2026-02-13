# SDK 对比分析

> 本文档对比 Claude Agent SDK 和 MCP Go SDK 的不同实现选项。

## 1. Claude Agent SDK Go 版本

### 发现: 非官方社区移植

**仓库**: [schlunsen/claude-agent-sdk-go](https://github.com/schlunsen/claude-agent-sdk-go)

| 指标 | 详情 |
|------|------|
| 状态 | ✅ **生产就绪** (v0.1.0) |
| 代码量 | ~9,800 行生产代码 + 2,100 行测试 |
| 测试覆盖率 | 60%+ |
| Go 版本 | 1.24+ |
| 功能完整度 | 100% (8个阶段全部完成) |

### 功能对照表

| 功能 | Python SDK | TypeScript SDK | Go SDK |
|------|-----------|----------------|--------|
| One-shot queries | ✅ | ✅ | ✅ |
| Interactive client | ✅ | ✅ | ✅ |
| Tool permissions | ✅ | ✅ | ✅ |
| Hook system | ✅ | ✅ | ✅ |
| MCP servers | ✅ | ✅ | ✅ |
| Streaming | ✅ | ✅ | ✅ |
| CLI discovery | ✅ | ✅ | ✅ |
| Error types | ✅ | ✅ | ✅ |

### 关键架构差异

| 方面 | TypeScript/Python | Go |
|------|------------------|-----|
| 并发模型 | async/await + Promise | goroutine + channel |
| Context | 隐式 | 显式 `context.Context` |
| 选项构建 | dataclass/object | 流式 API (Builder 模式) |
| 消息迭代 | async generator | channel |

### 使用示例

```go
package main

import (
    "context"
    "fmt"

    sdk "github.com/schlunsen/claude-agent-sdk-go"
)

func main() {
    ctx := context.Background()

    // 简单查询
    messages, err := sdk.Query(ctx, "What is 2 + 2?", nil)
    if err != nil {
        panic(err)
    }

    for msg := range messages {
        fmt.Println(msg)
    }
}
```

### 交互式客户端

```go
package main

import (
    "context"
    "fmt"

    sdk "github.com/schlunsen/claude-agent-sdk-go"
)

func main() {
    ctx := context.Background()

    options := sdk.NewClaudeAgentOptions().
        WithModel("claude-opus-4-20250514").
        WithAllowedTools("Bash", "Write", "Read").
        WithPermissionCallback(func(ctx context.Context, toolName string, input interface{}) (bool, error) {
            fmt.Printf("Tool %s requested. Allow? (y/n): ", toolName)
            return true, nil
        })

    client, err := sdk.NewClient(options)
    if err != nil {
        panic(err)
    }

    if err := client.Connect(ctx); err != nil {
        panic(err)
    }
    defer client.Close(ctx)

    if err := client.Query(ctx, "List the files in the current directory"); err != nil {
        panic(err)
    }

    for msg := range client.ReceiveResponse(ctx) {
        fmt.Println(msg)
    }
}
```

### 认证方式

```bash
# 方式 1: API Key (按量付费)
export CLAUDE_API_KEY=your-api-key-here

# 方式 2: OAuth Token (Max 订阅)
export CLAUDE_CODE_OAUTH_TOKEN=your-oauth-token-here
```

### 状态总结

```
✅ Phase 1: Foundation & Types (100%)
✅ Phase 2: Transport Layer (100%)
✅ Phase 3: Message Parsing (100%)
✅ Phase 4: Control Protocol (100%)
✅ Phase 5: Public API (100%)
✅ Phase 6: Testing & Validation (100%)
✅ Phase 7: Documentation & Examples (100%)
✅ Phase 8: Polish & Release (100%)
```

### 免责声明

> ⚠️ 这是**非官方社区移植**，不隶属于 Anthropic。使用风险自负。

---

## 2. MCP Go SDK 对比

### 选项 A: mark3labs/mcp-go (社区)

**仓库**: [mark3labs/mcp-go](https://github.com/mark3labs/mcp-go)

| 维度 | 详情 |
|------|------|
| 维护方 | 社区 (Ed Zynda 原作者) |
| 成熟度 | 较早发布，功能丰富 |
| 传输支持 | stdio, SSE, HTTP, in-process |
| OAuth | ✅ 完整支持 |
| 测试 | 丰富的测试套件 |
| 文档 | 独立文档站 |

### 项目结构

```
mark3labs/mcp-go/
├── client/                 # MCP 客户端实现
│   ├── client.go
│   ├── elicitation.go
│   ├── sampling.go
│   ├── sse.go
│   ├── stdio.go
│   └── transport/
├── server/                 # MCP 服务器实现
│   ├── server.go
│   ├── session.go
│   ├── sse.go
│   └── stdio.go
├── mcp/                    # 核心类型定义
│   ├── tools.go
│   ├── prompts.go
│   ├── resources.go
│   └── types.go
├── mcptest/                # 测试工具
└── examples/               # 示例代码
```

### 选项 B: modelcontextprotocol/go-sdk (官方)

**仓库**: [modelcontextprotocol/go-sdk](https://github.com/modelcontextprotocol/go-sdk)

| 维度 | 详情 |
|------|------|
| 维护方 | 官方 (MCP 团队 + Google Go 团队) |
| 成熟度 | 2025年正式发布 |
| 传输支持 | stdio, SSE, HTTP, streamable HTTP |
| OAuth | ✅ 完整支持 + OAuth 扩展 |
| 测试 | 包含一致性测试 |
| 文档 | pkg.go.dev + docs/ 目录 |

### 项目结构

```
modelcontextprotocol/go-sdk/
├── mcp/                    # 核心 API
│   ├── client.go
│   ├── server.go
│   ├── tool.go
│   ├── resource.go
│   ├── prompt.go
│   └── transport.go
├── jsonrpc/                # JSON-RPC 实现
├── auth/                   # OAuth 认证
├── oauthex/                # OAuth 扩展
├── internal/               # 内部工具
│   └── jsonrpc2/
└── examples/               # 示例代码
```

### 使用示例

```go
package main

import (
    "context"
    "log"

    "github.com/modelcontextprotocol/go-sdk/mcp"
)

type Input struct {
    Name string `json:"name" jsonschema:"the name of the person to greet"`
}

type Output struct {
    Greeting string `json:"greeting" jsonschema:"the greeting to tell to the user"`
}

func SayHi(ctx context.Context, req *mcp.CallToolRequest, input Input) (
    *mcp.CallToolResult,
    Output,
    error,
) {
    return nil, Output{Greeting: "Hi " + input.Name}, nil
}

func main() {
    server := mcp.NewServer(&mcp.Implementation{Name: "greeter", Version: "v1.0.0"}, nil)
    mcp.AddTool(server, &mcp.Tool{Name: "greet", Description: "say hi"}, SayHi)

    if err := server.Run(context.Background(), &mcp.StdioTransport{}); err != nil {
        log.Fatal(err)
    }
}
```

---

## 详细对比表

| 维度 | mark3labs/mcp-go | modelcontextprotocol/go-sdk |
|------|------------------|------------------------------|
| **维护方** | 社区 | 官方 (MCP 团队 + Google Go 团队) |
| **发布时间** | 较早 | 2025年 |
| **传输支持** | | |
| - stdio | ✅ | ✅ |
| - SSE | ✅ | ✅ |
| - HTTP | ✅ | ✅ |
| - in-process | ✅ | ✅ |
| - streamable HTTP | ❓ | ✅ |
| **OAuth** | ✅ 完整 | ✅ 完整 + 扩展 |
| **Sampling** | ✅ | ✅ |
| **Elicitation** | ✅ | ✅ |
| **测试** | 丰富 | 包含一致性测试 |
| **文档** | 独立文档站 | pkg.go.dev + docs/ |
| **GitHub 使用** | 历史使用 | **已迁移** (GitHub MCP Server) |

---

## 关键发现

### GitHub 官方迁移

> **重要**: GitHub 官方 MCP Server 已从 `mark3labs/mcp-go` 迁移到 `modelcontextprotocol/go-sdk`

来源: [GitHub Changelog - 2025-12-10](https://github.blog/changelog/2025-12-10-the-github-mcp-server-adds-support-for-tool-specific-configuration-and-more/)

> Both the local and remote GitHub MCP Server have been fully migrated from mark3labs/mcp-go to the official Go SDK for the Model Context Protocol.

这验证了官方 SDK 的成熟度和推荐地位。

---

## 推荐结论

### Claude Agent SDK

**推荐**: `schlunsen/claude-agent-sdk-go`

理由:
- 生产就绪 (v0.1.0)
- 功能完整 (100%)
- 活跃维护
- 完整测试覆盖

### MCP SDK

**推荐**: `modelcontextprotocol/go-sdk`

理由:
- 官方维护，长期支持有保障
- MCP 团队和 Google Go 团队协作
- GitHub 官方已迁移至此
- 包含一致性测试

---

## 依赖声明示例

```go
// go.mod
module github.com/your-org/claude-code-acp-go

go 1.24

require (
    // Claude Agent SDK (非官方社区移植)
    github.com/schlunsen/claude-agent-sdk-go v0.1.0

    // MCP SDK (官方)
    github.com/modelcontextprotocol/go-sdk v0.2.0
)
```

---

## 替代方案

如果需要其他选项，以下是社区替代品:

| SDK | 仓库 | 说明 |
|-----|------|------|
| mcp-golang | [metoro-io/mcp-golang](https://github.com/metoro-io/mcp-golang) | 另一个社区实现 |
| go-mcp | [ThinkInAIXYZ/go-mcp](https://github.com/ThinkInAIXYZ/go-mcp) | 另一个社区实现 |

官方 SDK 的 README 中感谢了这些替代方案的启发。

---

*文档生成日期: 2026-02-13*