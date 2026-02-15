## Context

MCP (Model Context Protocol) 是用于工具集成的标准协议。本项目需要实现 MCP 服务器以支持 ACP 协议中的工具调用功能。

**约束条件：**
- 必须使用 `schlunsen/claude-agent-sdk-go` 提供的 MCP 工厂函数
- 工具行为必须与 TypeScript 版本完全对齐
- 必须遵循 TDD 流程，达到 100% 测试覆盖率
- 必须遵循 go-comm 的命名和错误处理规范

**相关方：**
- ACP Agent 核心层
- 权限管理层
- 会话管理层

## Goals / Non-Goals

**Goals:**
- 使用 claude-agent-sdk-go 实现 MCP 服务器
- 实现 Read, Write, Edit, Bash, BashOutput, KillShell 六个核心工具
- 与 ACP 权限系统完整集成
- 支持工具调用前后的钩子函数
- 100% 测试覆盖率

**Non-Goals:**
- 不实现自定义工具扩展机制（后续迭代）
- 不实现 MCP 资源功能（非 ACP 需要）
- 不实现 MCP Prompts 功能（非 ACP 需要）

## Decisions

### Decision 1: 使用 claude-agent-sdk-go MCP 工厂函数

**选择：** 使用 `schlunsen/claude-agent-sdk-go` 提供的 MCP 工厂函数

**原因：**
- SDK 已提供生产就绪的 MCP 实现
- 内置权限回调和钩子支持
- 与 Claude API 集成良好
- 减少自定义代码量

**替代方案：**
1. 使用 `modelcontextprotocol/go-sdk` 从头实现 - 工作量大，需要手动集成
2. 完全自定义实现 - 不必要，SDK 已提供所需功能

### Decision 2: 工具适配层设计

**选择：** 创建独立的工具适配层 (`internal/mcp/tools.go`)

**原因：**
- 分离 MCP 协议处理和工具实现逻辑
- 便于测试单个工具
- 支持未来添加新工具
- 隔离外部 SDK 变更影响

**架构：**
```
ACP ToolCall → MCP Server → ToolAdapter → Concrete Tool → Result
                    ↑
            Permission Check
```

### Decision 3: 权限集成策略

**选择：** 在 MCP 服务器层通过 PreToolUse 钩子集成权限检查

**原因：**
- 所有工具调用都经过统一检查点
- 可以在工具执行前拒绝敏感操作
- 支持 "ask" 模式向用户请求许可

**实现：**
```go
func (s *ServerT) preToolUseHook(ctx context.Context, toolName string, params map[string]any) error {
    decision := s.permissionChecker.Check(ctx, toolName, params)
    switch decision {
    case DecisionAllow:
        return nil
    case DecisionDeny:
        return ErrPermissionDenied
    case DecisionAsk:
        return s.requestPermission(ctx, toolName, params)
    }
}
```

### Decision 4: 会话-终端关联

**选择：** 在 SessionContext 中维护终端 ID 到进程的映射

**原因：**
- BashOutput 和 KillShell 需要追踪执行中的进程
- 支持多个并发终端
- 便于会话清理时终止所有子进程

**数据结构：**
```go
type SessionContextT struct {
    SessionID   string
    WorkingDir  string
    Terminals   map[string]*TerminalInfo
    mu          sync.RWMutex
}

type TerminalInfo struct {
    ID       string
    Cmd      *exec.Cmd
    StartTime time.Time
    Status   TerminalStatus
}
```

## Risks / Trade-offs

### Risk 1: SDK 兼容性
- **风险：** `schlunsen/claude-agent-sdk-go` 是非官方 SDK，可能有 API 变更
- **缓解：** 通过适配层隔离 SDK 变更；锁定依赖版本

### Risk 2: 工具行为差异
- **风险：** Go 实现与 TypeScript 版本行为可能存在细微差异
- **缓解：** 使用黄金文件测试验证输出一致性；编写对比测试

### Risk 3: 并发安全
- **风险：** 多会话并发访问终端映射可能导致竞态
- **缓解：** 使用 sync.RWMutex 保护共享状态；启用 -race 检测

## Migration Plan

本次为新增功能，无需迁移。

## Open Questions

1. ~~是否需要支持自定义工具扩展？~~ → Non-Goal，后续迭代
2. ~~Bash 命令的超时默认值是多少？~~ → 参考 TypeScript 版本，默认 120 秒
3. 是否需要限制单个会话的最大终端数量？→ 待定，可后续添加
