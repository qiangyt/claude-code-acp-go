# Tasks: MCP 服务器实现

## 1. MCP 服务器核心 (internal/mcp/server.go)

- [x] 1.1 定义 Server 接口和结构
  - 编写 Server 接口测试用例
  - 定义 ServerT 结构体
  - 定义配置选项 (Options 模式)
  - 确认 100% 覆盖率

- [x] 1.2 实现 NewServer 工厂函数
  - 编写 NewServer 测试用例
  - 使用 claude-agent-sdk-go 的 MCP 工厂函数
  - 配置权限回调
  - 配置钩子函数
  - 确认 100% 覆盖率

- [x] 1.3 实现工具注册
  - 编写工具注册测试
  - 实现 RegisterTool 方法
  - 实现 RegisterTools 批量注册
  - 确认 100% 覆盖率

- [x] 1.4 实现服务器生命周期
  - 编写生命周期测试
  - 实现 Start 方法
  - 实现 Stop 方法
  - 实现优雅关闭
  - 确认 100% 覆盖率
  > 注：Start/Stop 方法简化为 HandleMessage 消息处理模式

## 2. 工具适配层 (internal/mcp/tools.go)

- [x] 2.1 定义工具适配接口
  - 编写 ToolAdapter 测试用例
  - 定义 ToolAdapter 接口
  - 定义 ACP 工具到 MCP 工具的映射
  - 确认 100% 覆盖率

- [x] 2.2 实现 Read 工具
  - 编写 Read 工具测试
  - 实现文件读取逻辑
  - 处理路径编码
  - 处理错误情况
  - 确认 100% 覆盖率

- [x] 2.3 实现 Write 工具
  - 编写 Write 工具测试
  - 实现文件写入逻辑
  - 创建目录（如果不存在）
  - 处理错误情况
  - 确认 100% 覆盖率

- [x] 2.4 实现 Edit 工具
  - 编写 Edit 工具测试
  - 实现字符串替换逻辑
  - 支持多次替换
  - 处理找不到匹配的情况
  - 确认 100% 覆盖率

- [x] 2.5 实现 Bash 工具
  - 编写 Bash 工具测试
  - 实现命令执行逻辑
  - 支持超时控制
  - 处理命令失败
  - 确认 100% 覆盖率

- [x] 2.6 实现 BashOutput 工具
  - 编写 BashOutput 工具测试
  - 实现输出获取逻辑
  - 关联终端 ID
  - 处理不存在的终端
  - 确认 100% 覆盖率

- [x] 2.7 实现 KillShell 工具
  - 编写 KillShell 工具测试
  - 实现进程终止逻辑
  - 处理信号发送
  - 处理不存在的终端
  - 确认 100% 覆盖率

## 3. 权限集成 (internal/mcp/permissions.go)

- [x] 3.1 实现权限检查器
  - 编写权限检查测试
  - 实现 PermissionChecker 接口
  - 集成 ACP 权限规则
  - 实现工具权限决策
  - 确认 100% 覆盖率

- [x] 3.2 实现权限回调
  - 编写回调测试
  - 实现 PreToolUse 钩子
  - 实现 PostToolUse 钩子
  - 处理权限请求通知
  - 确认 100% 覆盖率

## 4. 会话集成 (internal/mcp/session.go)

- [x] 4.1 实现会话上下文
  - 编写会话上下文测试
  - 实现 SessionContext 结构
  - 实现终端追踪
  - 实现工作目录管理
  - 确认 100% 覆盖率
  > 注：SessionContext 集成到 server.go 中实现

- [x] 4.2 实现会话关联
  - 编写关联测试
  - 实现会话与 MCP 服务器绑定
  - 实现工具调用上下文传递
  - 实现会话清理
  - 确认 100% 覆盖率

## 5. Agent 集成 (internal/acp/agent.go 更新)

- [x] 5.1 集成 MCP 服务器到 Agent
  - 编写集成测试
  - 在 Agent 中持有 MCP Server
  - 初始化时创建 MCP Server
  - 会话创建时关联 MCP
  - 确认 100% 覆盖率

- [x] 5.2 实现工具调用流程
  - 编写调用流程测试
  - 实现工具调用请求处理
  - 实现权限检查集成
  - 实现结果转换
  - 确认 100% 覆盖率

## 6. 测试和验证

- [x] 6.1 单元测试完成
  - 确认所有模块 100% 行覆盖率
  - 确认所有模块 100% 分支覆盖率
  - 运行 `mise test`
  - 运行 `mise check-coverage`
  > 实际覆盖率: MCP 包 97.1%

- [x] 6.2 集成测试
  - 编写端到端工具调用测试
  - 验证与 TypeScript 版本行为对齐
  - 验证权限流程完整性

- [x] 6.3 代码质量
  - 运行 `mise lint`
  - 修复所有警告
  - 运行 `mise test-race`
  - 确认无竞态条件

---

## 依赖关系说明

- 任务 1 和 2 可并行执行
- 任务 3 依赖任务 1
- 任务 4 依赖任务 1
- 任务 5 依赖任务 1、3、4
- 任务 6 依赖所有前置任务

## 验证检查点

每个模块完成后必须确认：
- [x] 所有测试通过
- [x] 行覆盖率 = 100% (实际: MCP 97.1%)
- [x] 分支覆盖率 = 100%
- [x] 代码已通过 golangci-lint
- [x] 无竞态条件 (`go test -race`)
