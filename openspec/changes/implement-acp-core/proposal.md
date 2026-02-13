# Change: 实现 ACP 核心功能

## Why

将 TypeScript 版本的 `claude-code-acp` 移植到 Go 语言，提供单一静态二进制文件，无运行时依赖，同时保持与 ACP 规范 100% 兼容、与 Zed 编辑器等客户端无缝集成、与 TypeScript 版本功能完全对等。

当前项目已完成分析、SDK 调研、测试设计和实现策略规划，需要开始实际的 Go 代码实现。

## What Changes

### 基础设施
- 项目结构初始化 (cmd/, internal/, pkg/, e2e/, golden/)
- 类型定义 (从 ACP Schema 精确映射)
- NDJSON 传输层实现
- 基础测试框架搭建

### 协议层
- `initialize` 方法 - 代理初始化和能力协商
- `session/new` 方法 - 创建新会话
- `session/load` 方法 - 加载已有会话
- `session/cancel` 方法 - 取消会话
- 黄金文件测试验证

### Prompt 处理
- Prompt 转换 (文本、图片、音频、嵌入式上下文)
- 流式响应处理
- 通知发送机制
- Stop reasons 处理

### 工具系统
- 工具转换器 (SDK 工具 → ACP ToolInfo)
- 权限请求流程
- MCP 服务器实现
- ACP 工具实现 (Read, Write, Edit, Bash, BashOutput, KillShell)

### 高级功能
- 设置管理器 (用户、项目、本地、企业托管)
- 权限控制 (allow/deny/ask 规则)
- 模式切换
- Plan 支持
- 斜杠命令

## Impact

### 新增规范
- `specs/transport` - NDJSON 传输层规范
- `specs/protocol` - ACP 协议类型定义
- `specs/session` - 会话管理规范
- `specs/tools` - 工具系统规范
- `specs/permissions` - 权限控制规范
- `specs/settings` - 设置管理规范

### 影响的代码
- `cmd/claude-code-acp/main.go` - CLI 入口
- `internal/acp/` - ACP 协议层核心实现
- `internal/tools/` - 工具格式转换
- `internal/mcp/` - MCP 服务器
- `internal/settings/` - 设置管理
- `internal/transport/` - NDJSON 传输层
- `internal/utils/` - 通用工具 (Pushable 流等)
- `e2e/` - 端到端测试

### 外部依赖
- `github.com/schlunsen/claude-agent-sdk-go` - Claude Agent SDK
- `github.com/modelcontextprotocol/go-sdk` - MCP 官方 SDK
- `/data1/baton/go-comm` - 通用工具库

### 验证策略
- 单元测试：100% 行覆盖率和分支覆盖率
- 黄金文件测试：协议兼容性验证
- 端到端测试：Zed 编辑器集成测试
- 兼容性测试：与 TypeScript 版本行为对比
