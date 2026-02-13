# Design: ACP 核心实现架构

## Context

### 背景
将 TypeScript 版本的 `claude-code-acp` 移植到 Go 语言。原始项目使用 TypeScript 实现了 Agent Client Protocol (ACP)，用于标准化 AI 编码代理与代码编辑器/IDE 之间的通信。

### 约束
- Go 版本必须 >= 1.24
- 必须与 go-comm 库保持依赖版本一致
- 协议兼容性: 与 ACP 规范 100% 兼容
- 客户端兼容: 与 Zed 编辑器等 ACP 客户端无缝集成
- 功能对等: 与 TypeScript 版本功能完全一致
- TDD 流程不可跳过，100% 测试覆盖率

### 利益相关者
- 终端用户：使用 Zed 编辑器的开发者
- ACP 客户端开发者：集成此代理的编辑器/IDE 开发者
- Claude Code 用户：希望获得原生 Go 二进制体验的用户

## Goals / Non-Goals

### Goals
1. 实现完整的 ACP 协议支持
2. 与 TypeScript 版本功能完全对等
3. 提供单一静态二进制文件，无运行时依赖
4. 100% 测试覆盖率
5. 清晰的代码结构和良好的可维护性
6. 遵循 go-comm 库的代码规范

### Non-Goals
1. 不实现 ACP 规范之外的扩展功能
2. 不支持 TypeScript 版本已废弃的功能
3. 不提供图形用户界面
4. 不支持 Windows 平台 (初期版本)

## Decisions

### 1. 分层架构设计

**决定**: 采用清晰的分层架构，从上到下依次为：
1. CLI 入口层 (cmd/)
2. ACP 协议层 (internal/acp/)
3. 工具转换层 (internal/tools/)
4. MCP 服务器层 (internal/mcp/)
5. 传输层 (internal/transport/)
6. 通用工具层 (internal/utils/)

**理由**:
- 关注点分离，每层职责明确
- 便于独立测试和验证
- 与 TypeScript 版本架构对应，便于对照实现

**替代方案**:
- 单一模块架构：不利于测试和维护
- 微服务架构：增加复杂度，不适合单一二进制目标

### 2. 传输层选择

**决定**: 使用 NDJSON (Newline-Delimited JSON) over stdin/stdout

**理由**:
- ACP 规范要求
- 简单可靠，易于调试
- 与进程通信模型匹配

**替代方案**:
- WebSocket: 增加复杂度，需要额外服务
- HTTP: 不适合双向流式通信

### 3. SDK 选择

**决定**:
- Claude Agent SDK: `github.com/schlunsen/claude-agent-sdk-go` (社区移植)
- MCP SDK: `github.com/modelcontextprotocol/go-sdk` (官方)

**理由**:
- 社区 SDK 已生产就绪，API 与 TypeScript 版本接近
- 官方 MCP SDK 提供最佳兼容性

**替代方案**:
- 自己实现 SDK: 工作量大，难以保证兼容性
- 使用其他社区 SDK: 成熟度不足

### 4. 并发模型

**决定**: 使用 Go 原生 goroutine 和 channel 实现 Pushable 流

**理由**:
- 符合 Go 惯例
- 与 TypeScript 版本的异步迭代器模型对应
- channel 提供天然的同步机制

**替代方案**:
- 使用第三方流库: 增加依赖
- 使用回调模式: 不符合 Go 风格

### 5. 权限模型

**决定**: 基于 allow/deny/ask 三级规则的权限控制

**理由**:
- 与 TypeScript 版本一致
- 灵活且易于配置
- 支持细粒度控制

**规则格式**:
```
Read                      # 允许所有 Read 操作
Read(./src/**)            # 允许读取 src 目录
Bash(npm run:*)           # 允许 npm run 相关命令
Bash(rm:*)                # 拒绝 rm 命令 (在 deny 列表中)
```

### 6. 设置来源优先级

**决定**: 按以下优先级合并设置（低到高）：
1. 用户设置 (`~/.claude/settings.json`)
2. 项目设置 (`$CWD/.claude/settings.json`)
3. 本地项目设置 (`$CWD/.claude/settings.local.json`)
4. 企业托管设置 (系统级配置)

**理由**:
- 与 TypeScript 版本一致
- 允许企业统一管控
- 支持项目和用户级定制

### 7. 错误处理模式

**决定**: 采用双函数模式

**模式**:
```go
// 普通版本：返回 (result, error)
func (m *Manager) Load(path string) (*Config, error)

// Panic 版本：以 P 结尾
func (m *Manager) LoadP(path string) *Config
```

**理由**:
- 遵循 go-comm 规范
- 简化内部错误处理
- 明确区分可恢复和不可恢复错误

### 8. 命名约定

**决定**: 遵循 go-comm 命名规范

**规则**:
- 类型使用 `T` 后缀：`LoggerT`, `ConfigT`
- 类型别名使用 `type Type = *TypeT`
- Panic 版本使用 `P` 后缀：`NewLoggerP`, `RequiredStringP`
- 接口无 `T` 后缀：`Plugin`, `PluginLoader`

## Risks / Trade-offs

### 风险 1: SDK 兼容性
- **风险**: 社区 SDK 可能与官方 SDK 存在差异
- **缓解**: 使用黄金文件测试验证行为一致性；关注社区 SDK 更新

### 风险 2: 协议变更
- **风险**: ACP 规范可能更新
- **缓解**: 设计灵活的类型系统；预留 `_meta` 字段扩展

### 风险 3: 性能问题
- **风险**: 流式响应可能出现延迟或阻塞
- **缓解**: 使用带缓冲的 channel；实现背压机制

### 风险 4: 测试覆盖率
- **风险**: 达到 100% 覆盖率可能耗时较长
- **缓解**: 使用不同的子代理编写测试和实现；采用黄金文件测试减少手动编写

### 风险 5: 跨平台兼容
- **风险**: 不同操作系统行为可能不一致
- **缓解**: 初期聚焦 Linux/macOS；使用 CI 测试多平台

## Migration Plan

本提案为初始实现，无需迁移。

### 部署步骤
1. 构建静态二进制：`go build -o bin/claude-code-acp ./cmd/claude-code-acp`
2. 配置环境变量（如 API 密钥）
3. 配置 ACP 客户端（如 Zed）使用此二进制

### 回滚计划
如遇严重问题，可切换回 TypeScript 版本。

## Open Questions

1. **日志格式**: 是否需要支持结构化日志输出到文件？
2. **国际化范围**: 除了 UI 消息，错误消息是否需要国际化？
3. **性能基准**: 是否需要定义具体的性能指标（如响应延迟）？
4. **版本兼容**: 是否需要支持多版本的 ACP 协议？

## Appendix

### A. 目录结构

```
claude-code-acp-go/
├── cmd/
│   └── claude-code-acp/
│       └── main.go              # CLI 入口点
│
├── internal/
│   ├── acp/                     # ACP 协议层（核心）
│   │   ├── agent.go             # 核心代理实现
│   │   ├── session.go           # 会话管理
│   │   ├── permissions.go       # 权限控制
│   │   ├── protocol.go          # ACP 类型定义
│   │   ├── prompts.go           # Prompt 处理
│   │   └── notifications.go     # 通知发送
│   │
│   ├── tools/                   # 工具格式转换
│   │   ├── converter.go         # 工具转换器
│   │   ├── types.go             # 工具类型
│   │   ├── read.go              # Read 工具
│   │   ├── write.go             # Write 工具
│   │   ├── edit.go              # Edit 工具
│   │   └── bash.go              # Bash 工具
│   │
│   ├── mcp/                     # MCP 服务器
│   │   ├── server.go            # MCP 服务器实现
│   │   └── tools.go             # MCP 工具注册
│   │
│   ├── settings/                # 设置管理
│   │   ├── manager.go           # 设置管理器
│   │   ├── permissions.go       # 权限规则
│   │   └── sources.go           # 设置来源
│   │
│   ├── transport/               # 传输层
│   │   ├── ndjson.go            # NDJSON 编解码
│   │   └── stdio.go             # stdio 传输
│   │
│   └── utils/                   # 通用工具
│       ├── pushable.go          # Pushable 流
│       └── encoding.go          # 路径编码
│
├── pkg/
│   └── api/                     # 公开 API
│       ├── client.go
│       └── options.go
│
├── e2e/                         # 端到端测试
│   ├── compat/                  # 兼容性测试
│   ├── conformance/             # 协议合规测试
│   └── zed/                     # Zed 集成测试
│
└── golden/                      # 黄金测试文件
    ├── initialize-basic.jsonl
    ├── session-new.jsonl
    └── ...
```

### B. 核心接口

```go
// Agent 接口
type Agent interface {
    Initialize(ctx context.Context, req InitializeRequest) (*InitializeResponse, error)
    NewSession(ctx context.Context, req NewSessionRequest) (*NewSessionResponse, error)
    LoadSession(ctx context.Context, req LoadSessionRequest) (*LoadSessionResponse, error)
    CancelSession(ctx context.Context, req CancelSessionRequest) error
    Run(ctx context.Context) error
}

// Transport 接口
type Transport interface {
    Send(msg any) error
    Receive() (map[string]any, error)
    Close() error
}

// Session 接口
type Session interface {
    ID() string
    Cwd() string
    Cancel()
    IsCancelled() bool
}
```

### C. 参考资源

- [ACP 官方文档](https://agentclientprotocol.com)
- [原始 TypeScript 实现](https://github.com/zed-industries/claude-code-acp)
- [Claude Agent SDK Go](https://github.com/schlunsen/claude-agent-sdk-go)
- [MCP Go SDK 官方](https://github.com/modelcontextprotocol/go-sdk)
- [go-comm 通用库](/data1/baton/go-comm)
