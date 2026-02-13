# Project Context

## Purpose

**claude-code-acp-go** 是一个将 TypeScript 版本的 [claude-code-acp](https://github.com/zed-industries/claude-code-acp) 移植到 Go 语言的实现。

**核心目标:**
- 实现 **Agent Client Protocol (ACP)** - 标准化的 AI 编码代理与代码编辑器/IDE 通信协议
- 与 Zed 编辑器等 ACP 客户端无缝集成
- 与 TypeScript 版本功能完全对等
- 提供单一静态二进制文件，无运行时依赖

## Tech Stack

| 组件 | 技术 | 说明 |
|------|------|------|
| **语言** | Go 1.24+ | 主要开发语言 |
| **Claude Agent SDK** | `schlunsen/claude-agent-sdk-go` | 非官方社区移植，生产就绪 |
| **MCP SDK** | `modelcontextprotocol/go-sdk` | 官方 Go SDK |
| **通用库** | `go-comm` (`/data1/baton/go-comm`) | 项目无关的 Go 通用工具库 |
| **日志** | `phuslu/log` (via go-comm) | 结构化日志 |
| **配置** | YAML/TOML (via go-comm) | 配置管理 |
| **测试** | `testify` | 测试框架 |

## Project Conventions

### Code Style

#### 命名约定 (遵循 go-comm 规范)
- **类型**: 使用 `T` 后缀（如 `LoggerT`、`ConfigT`）
- **类型别名**: 使用 `type Type = *TypeT`（如 `type Logger = *LoggerT`）
- **函数**: 驼峰命名，Panic 版本使用 `P` 后缀（如 `NewLoggerP`、`RequiredStringP`）
- **接口**: 无 `T` 后缀（如 `Plugin`、`PluginLoader`）

#### 错误处理
- **双函数模式**: 每个可能出错的函数提供两个版本
  - 普通版本：返回 `(result, error)`
  - Panic 版本：以 `P` 结尾，出错时 panic
- 优先使用 Panic 版本简化错误处理，仅在最外层捕获 panic

#### 类型转换
- **Required 模式**: 字段必需，缺失时返回错误/panic
- **Optional 模式**: 字段可选，返回默认值

```go
// 必需字段
name := comm.RequiredStringP("config", "name", m)

// 可选字段（带默认值）
port, has := comm.OptionalStringP("config", "port", m, "8080")
```

#### 本地化
- 所有用户可见的错误消息必须使用 i18n (`comm.T()` 或 `comm.LocalizeError()`)
- 除 UI 界面外，代码注释和文档使用中文

### Architecture Patterns

#### 分层架构
```
cmd/claude-code-acp/          # CLI 入口
internal/
├── acp/                      # ACP 协议层（核心）
│   ├── agent.go              # 核心代理实现
│   ├── session.go            # 会话管理
│   ├── protocol.go           # ACP 类型定义
│   └── ...
├── tools/                    # 工具格式转换
├── mcp/                      # MCP 服务器实现
├── settings/                 # 设置管理
├── transport/                # NDJSON 传输层
└── utils/                    # 通用工具
pkg/api/                      # 公开 API
e2e/                          # 端到端测试
golden/                       # 黄金测试文件
```

#### 关键模式
- **Pushable Stream**: 可推送的异步可迭代对象（泛型实现）
- **会话管理**: 每个 ACP 会话独立的 Session 对象
- **权限控制**: 基于 allow/deny/ask 规则的权限决策
- **工具转换**: SDK 工具 → ACP ToolInfo 格式转换

### Testing Strategy

#### TDD 强制规范 (Red-Green-Refactor)

1. **红灯阶段**: 先编写失败的测试用例，禁止编写生产代码
2. **绿灯阶段**: 编写最小化生产代码使测试通过，禁止重构
3. **重构阶段**: 在测试保护下优化代码结构

#### 覆盖率要求
- **行覆盖率**: 必须 = 100%
- **分支覆盖率**: 必须 = 100%

#### 测试类型
- **单元测试**: `internal/` 各模块的 `*_test.go` 文件
- **黄金文件测试**: `golden/` 目录，用于协议兼容性验证
- **端到端测试**: `e2e/` 目录
  - `e2e/compat/` - 与 TypeScript 版本的兼容性测试
  - `e2e/conformance/` - ACP 协议合规测试
  - `e2e/zed/` - Zed 编辑器集成测试

#### 测试子代理
- 使用不同的 sub agent 来编写测试和实现代码

### Git Workflow

- **主分支**: `master`
- **提交信息**: 中文描述，清晰说明变更内容
- **代码审查**: 通过 Pull Request 流程

## Domain Context

### ACP (Agent Client Protocol)

ACP 是一个标准化协议，类似于 LSP (Language Server Protocol)：
- **LSP** 标准化了语言服务器集成
- **ACP** 标准化了代理-编辑器通信

### 通信方式
- **传输层**: NDJSON over stdin/stdout
- **消息格式**: JSON-RPC 风格的请求/响应

### 核心功能
1. **会话管理**: `initialize`, `session/new`, `session/load`, `session/cancel`
2. **Prompt 处理**: 多模态内容（文本、图片、音频）
3. **工具系统**: Read, Write, Edit, Bash, BashOutput, KillShell
4. **权限控制**: 基于规则的 allow/deny/ask 决策

## Important Constraints

### 技术约束
- Go 版本必须 >= 1.24
- 必须与 go-comm 库保持依赖版本一致
- 新依赖需评估是否应添加到 go-comm

### 业务约束
- 协议兼容性: 与 ACP 规范 100% 兼容
- 客户端兼容: 与 Zed 编辑器等 ACP 客户端无缝集成
- 功能对等: 与 TypeScript 版本功能完全一致

### 质量约束
- TDD 流程不可跳过
- 100% 测试覆盖率必须达成
- 禁止先写生产代码后补测试

## External Dependencies

### 核心依赖

| 依赖 | 用途 | 来源 |
|------|------|------|
| `go-comm` | 通用工具库 | `/data1/baton/go-comm` |
| `claude-agent-sdk-go` | Claude Agent SDK | github.com/schlunsen/claude-agent-sdk-go |
| `go-sdk` | MCP 官方 SDK | github.com/modelcontextprotocol/go-sdk |
| `testify` | 测试框架 | github.com/stretchr/testify |

### go-comm 核心模块

| 模块 | 文件 | 功能 |
|------|------|------|
| 日志系统 | `logger.go` | 结构化日志、文件轮转、追踪 ID |
| 配置管理 | `config.go` | YAML/TOML 解析、环境变量 |
| 国际化 | `i18n.go` | 多语言支持 (12种语言) |
| 文件操作 | `fileops.go` | 文件复制、缓存、回退 |
| Shell 执行 | `gosh.go` | Shell 脚本解析执行 |
| 插件系统 | `plugin.go` | 插件加载、生命周期管理 |
| 类型转换 | `string.go`, `int.go`, etc. | Required/Optional 模式 |

### 参考资源

- [ACP 官方文档](https://agentclientprotocol.com)
- [原始 TypeScript 实现](https://github.com/zed-industries/claude-code-acp)
- [Claude Agent SDK Go](https://github.com/schlunsen/claude-agent-sdk-go)
- [MCP Go SDK 官方](https://github.com/modelcontextprotocol/go-sdk)
