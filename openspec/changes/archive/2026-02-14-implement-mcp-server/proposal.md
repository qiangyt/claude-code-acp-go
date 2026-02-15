# Change: 实现 MCP 服务器

## Why

当前项目需要实现 MCP (Model Context Protocol) 服务器，以支持工具调用功能。通过使用 `schlunsen/claude-agent-sdk-go` 提供的 MCP 工厂函数，可以快速实现一个生产就绪的 MCP 服务器，同时保持代码简洁和可维护性。

该 SDK 已提供：
- MCP 服务器工厂函数
- 权限回调支持 (Permission callbacks)
- 钩子系统 (PreToolUse, PostToolUse)
- 流式响应支持 (Go channels)

## What Changes

### MCP 服务器核心
- 使用 `schlunsen/claude-agent-sdk-go` 的 MCP 工厂函数创建服务器
- 实现 `internal/mcp/server.go` - MCP 服务器核心逻辑
- 实现 `internal/mcp/tools.go` - ACP 工具到 MCP 工具的适配

### 工具实现
- Read 工具 - 读取文件内容
- Write 工具 - 写入文件内容
- Edit 工具 - 编辑文件（字符串替换）
- Bash 工具 - 执行 Shell 命令
- BashOutput 工具 - 获取 Shell 输出
- KillShell 工具 - 终止 Shell 进程

### 权限集成
- 与 ACP 权限系统集成
- 工具调用前权限检查
- 敏感操作提示

### 会话管理
- MCP 服务器与会话关联
- 工具调用上下文管理
- 终端进程追踪

## Impact

### 新增规范
- `specs/mcp` - MCP 服务器实现规范

### 影响的代码
- `internal/mcp/server.go` - 新增 MCP 服务器实现
- `internal/mcp/tools.go` - 新增工具适配层
- `internal/acp/agent.go` - 集成 MCP 服务器
- `internal/acp/session.go` - 会话与 MCP 关联

### 外部依赖
- `github.com/schlunsen/claude-agent-sdk-go` - Claude Agent SDK (MCP 工厂函数)

### 验证策略
- 单元测试：100% 行覆盖率和分支覆盖率
- 工具行为与 TypeScript 版本对齐
- 权限检查完整性验证
