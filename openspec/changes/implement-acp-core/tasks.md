# Tasks: ACP 核心实现

## 1. 项目基础设施

- [ ] 1.1 初始化项目结构
  - 创建 `cmd/claude-code-acp/main.go`
  - 创建 `internal/` 子目录结构
  - 创建 `pkg/api/` 目录
  - 创建 `e2e/` 测试目录
  - 创建 `golden/` 黄金测试目录
  - 创建 `testdata/` 测试数据目录

- [ ] 1.2 初始化 Go 模块
  - 创建 `go.mod` 并配置依赖
  - 配置 `schlunsen/claude-agent-sdk-go`
  - 配置 `modelcontextprotocol/go-sdk`
  - 配置 `stretchr/testify`
  - 配置 go-comm 本地依赖

- [ ] 1.3 创建 Makefile
  - build、test、test-coverage 目标
  - test-race、test-e2e 目标
  - lint、fmt、clean 目标
  - generate、record-golden 目标

## 2. 传输层实现 (internal/transport/)

- [ ] 2.1 实现 NDJSON 编解码器 (ndjson.go)
  - 编写 Encoder 测试用例
  - 实现 Encoder
  - 编写 Decoder 测试用例
  - 实现 Decoder
  - 确认 100% 覆盖率

- [ ] 2.2 实现 stdio 传输层 (stdio.go)
  - 编写 Transport 接口测试
  - 实现 stdio Transport
  - 处理并发读写
  - 确认 100% 覆盖率

## 3. 协议类型定义 (internal/acp/protocol.go)

- [ ] 3.1 定义基础请求/响应类型
  - 编写类型定义测试
  - 定义 InitializeRequest/Response
  - 定义 ClientCapabilities/AgentCapabilities
  - 定义 ContentBlock 及其子类型
  - 确认 JSON 序列化正确

- [ ] 3.2 定义会话相关类型
  - 定义 NewSessionRequest/Response
  - 定义 LoadSessionRequest/Response
  - 定义 CancelSessionRequest
  - 定义 SessionUpdate 及其变体
  - 确认 100% 覆盖率

- [ ] 3.3 定义工具相关类型
  - 定义 ToolInfo 结构
  - 定义 ToolKind 枚举
  - 定义 ToolCallContent/Location
  - 确认 100% 覆盖率

## 4. 通用工具实现 (internal/utils/)

- [ ] 4.1 实现 Pushable 流 (pushable.go)
  - 编写 Pushable 测试用例
  - 实现 Push 方法
  - 实现 End 方法
  - 实现 Channel/Iter 方法
  - 处理并发安全
  - 确认 100% 覆盖率

- [ ] 4.2 实现路径编码工具 (encoding.go)
  - 编写路径编码测试
  - 实现相对路径解析
  - 实现 URI 编码/解码
  - 确认 100% 覆盖率

## 5. 会话管理实现 (internal/acp/session.go)

- [ ] 5.1 实现 Session 结构
  - 编写 Session 测试用例
  - 实现 Session 创建
  - 实现会话状态管理
  - 实现会话取消
  - 确认 100% 覆盖率

- [ ] 5.2 实现终端管理
  - 编写 Terminal 测试用例
  - 实现 Terminal 创建和状态追踪
  - 实现终端输出获取
  - 实现终端终止
  - 确认 100% 覆盖率

## 6. 设置管理实现 (internal/settings/)

- [ ] 6.1 实现设置管理器 (manager.go)
  - 编写 Manager 测试用例
  - 实现多来源加载
  - 实现设置合并逻辑
  - 实现设置访问 API
  - 确认 100% 覆盖率

- [ ] 6.2 实现权限规则 (permissions.go)
  - 编写权限规则测试
  - 实现规则解析
  - 实现规则匹配逻辑
  - 实现 CheckPermission 方法
  - 确认 100% 覆盖率

- [ ] 6.3 实现设置来源 (sources.go)
  - 编写来源检测测试
  - 实现各来源路径解析
  - 实现文件加载和解析
  - 确认 100% 覆盖率

## 7. 工具转换实现 (internal/tools/)

- [ ] 7.1 实现工具转换器 (converter.go)
  - 编写转换器测试用例
  - 实现 ToolInfoFromToolUse 函数
  - 实现 Read 工具转换
  - 实现 Write 工具转换
  - 实现 Edit 工具转换
  - 实现 Bash 工具转换
  - 实现其他工具转换
  - 确认 100% 覆盖率

- [ ] 7.2 实现工具类型定义 (types.go)
  - 编写类型测试
  - 定义 ToolInfo 结构
  - 定义 ToolKind 枚举
  - 定义辅助类型
  - 确认 100% 覆盖率

## 8. 权限控制实现 (internal/acp/permissions.go)

- [ ] 8.1 实现权限检查
  - 编写权限检查测试
  - 实现 PermissionManager
  - 实现规则匹配
  - 实现敏感操作检测
  - 确认 100% 覆盖率

- [ ] 8.2 实现权限请求通知
  - 编写通知测试
  - 实现权限请求发送
  - 实现响应处理
  - 实现决策缓存
  - 确认 100% 覆盖率

## 9. Prompt 处理实现 (internal/acp/prompts.go)

- [ ] 9.1 实现 Prompt 转换
  - 编写 Prompt 转换测试
  - 实现文本 Prompt 处理
  - 实现图片 Prompt 处理
  - 实现音频 Prompt 处理
  - 实现嵌入式上下文处理
  - 确认 100% 覆盖率

- [ ] 9.2 实现流式响应处理
  - 编写流式响应测试
  - 实现响应流迭代
  - 实现内容块累积
  - 实现 Stop reasons 处理
  - 确认 100% 覆盖率

## 10. MCP 服务器实现 (internal/mcp/)

- [ ] 10.1 实现 MCP 服务器 (server.go)
  - 编写服务器测试
  - 实现服务器初始化
  - 实现工具注册
  - 实现会话上下文管理
  - 确认 100% 覆盖率

- [ ] 10.2 实现 MCP 工具处理 (tools.go)
  - 编写工具处理测试
  - 实现 Read 工具处理
  - 实现 Write 工具处理
  - 实现 Edit 工具处理
  - 实现 Bash 工具处理
  - 实现 BashOutput 工具处理
  - 实现 KillShell 工具处理
  - 确认 100% 覆盖率

## 11. ACP Agent 核心实现 (internal/acp/agent.go)

- [ ] 11.1 实现 Agent 结构
  - 编写 Agent 测试用例
  - 实现 ClaudeAcpAgent 结构
  - 实现 NewClaudeAcpAgent 构造函数
  - 实现配置选项模式
  - 确认 100% 覆盖率

- [ ] 11.2 实现 Initialize 方法
  - 编写 Initialize 测试
  - 实现版本协商
  - 实现能力声明
  - 实现客户端能力记录
  - 确认 100% 覆盖率

- [ ] 11.3 实现 NewSession 方法
  - 编写 NewSession 测试
  - 实现会话创建
  - 实现 Prompt 处理启动
  - 实现流式响应发送
  - 确认 100% 覆盖率

- [ ] 11.4 实现 LoadSession 方法
  - 编写 LoadSession 测试
  - 实现会话恢复
  - 实现状态重建
  - 确认 100% 覆盖率

- [ ] 11.5 实现 CancelSession 方法
  - 编写 CancelSession 测试
  - 实现会话取消
  - 实现资源清理
  - 确认 100% 覆盖率

- [ ] 11.6 实现 Run 方法
  - 编写 Run 测试
  - 实现主事件循环
  - 实现消息分发
  - 实现优雅关闭
  - 确认 100% 覆盖率

## 12. 通知实现 (internal/acp/notifications.go)

- [ ] 12.1 实现通知发送
  - 编写通知测试
  - 实现 SessionUpdate 发送
  - 实现权限请求发送
  - 实现错误通知发送
  - 确认 100% 覆盖率

## 13. CLI 入口实现 (cmd/claude-code-acp/main.go)

- [ ] 13.1 实现 CLI 入口
  - 编写 CLI 测试
  - 实现命令行参数解析
  - 实现日志配置
  - 实现 Agent 启动
  - 实现信号处理
  - 确认 100% 覆盖率

## 14. 公开 API 实现 (pkg/api/)

- [ ] 14.1 实现客户端 API (client.go)
  - 编写 API 测试
  - 实现 Client 结构
  - 实现便捷方法
  - 确认 100% 覆盖率

- [ ] 14.2 实现配置选项 (options.go)
  - 编写选项测试
  - 实现选项模式
  - 实现默认配置
  - 确认 100% 覆盖率

## 15. 黄金文件测试

- [ ] 15.1 创建黄金测试框架
  - 实现黄金文件读取
  - 实现黄金文件比较
  - 实现更新模式

- [ ] 15.2 录制黄金文件
  - initialize 基本流程
  - session/new 基本流程
  - session/new 带多模态 prompt
  - session/cancel 流程
  - 工具调用流程

## 16. 端到端测试 (e2e/)

- [ ] 16.1 实现兼容性测试 (e2e/compat/)
  - 与 TypeScript 版本输出对比
  - 协议字段完整性验证
  - 边界情况处理

- [ ] 16.2 实现协议合规测试 (e2e/conformance/)
  - ACP 规范验证
  - 错误处理验证
  - 性能基准

- [ ] 16.3 实现 Zed 集成测试 (e2e/zed/)
  - 启动和初始化
  - 基本对话流程
  - 工具调用流程
  - 错误恢复流程

## 17. 文档和收尾

- [ ] 17.1 更新项目文档
  - 更新 README.md
  - 更新 doc/README.md 项目状态
  - 添加使用示例

- [ ] 17.2 代码质量检查
  - 运行 golangci-lint
  - 修复所有警告
  - 确认 100% 测试覆盖率
  - 确认无竞态条件

---

## 依赖关系说明

- 任务 1-4 可并行执行（基础设施层）
- 任务 5 依赖任务 3、4
- 任务 6 可独立执行
- 任务 7 依赖任务 3
- 任务 8 依赖任务 6
- 任务 9 依赖任务 3、4、5
- 任务 10 依赖任务 5、7、8
- 任务 11 依赖任务 3、5、6、7、8、9、10
- 任务 12 依赖任务 3
- 任务 13 依赖任务 11
- 任务 14 依赖任务 11
- 任务 15 依赖任务 11
- 任务 16 依赖任务 13、15
- 任务 17 依赖所有前置任务

## 验证检查点

每个模块完成后必须确认：
- [ ] 所有测试通过
- [ ] 行覆盖率 = 100%
- [ ] 分支覆盖率 = 100%
- [ ] 代码已通过 golangci-lint
- [ ] 无竞态条件 (`go test -race`)
