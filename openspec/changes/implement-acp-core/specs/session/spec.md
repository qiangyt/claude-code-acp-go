## ADDED Requirements

### Requirement: 会话创建

系统 SHALL 支持通过 `session/new` 方法创建新的 ACP 会话。

#### Scenario: 创建基本会话
- **WHEN** 收到有效的 `session/new` 请求
- **THEN** 创建新会话并返回会话 ID

#### Scenario: 带工作目录的会话
- **WHEN** 请求指定 `cwd` 参数
- **THEN** 会话在指定工作目录下运行

#### Scenario: 带初始 prompt 的会话
- **WHEN** 请求包含 `prompt` 参数
- **THEN** 会话立即开始处理 prompt

#### Scenario: 多模态 prompt 支持
- **WHEN** prompt 包含文本、图片、音频等多种内容
- **THEN** 正确解析并处理所有内容类型

### Requirement: 会话加载

系统 SHALL 支持通过 `session/load` 方法加载已存在的会话。

#### Scenario: 加载已存在会话
- **WHEN** 收到有效的 `session/load` 请求且会话存在
- **THEN** 恢复会话状态并继续处理

#### Scenario: 会话不存在
- **WHEN** 请求加载的会话不存在
- **THEN** 返回错误响应

#### Scenario: 检查能力声明
- **WHEN** 客户端请求加载会话
- **THEN** 仅当代理声明 `loadSession` 能力时才支持

### Requirement: 会话取消

系统 SHALL 支持通过 `session/cancel` 方法取消正在运行的会话。

#### Scenario: 取消运行中的会话
- **WHEN** 收到 `session/cancel` 请求
- **THEN** 停止当前操作并发送取消通知

#### Scenario: 取消已完成会话
- **WHEN** 尝试取消已完成的会话
- **THEN** 操作为无效果（幂等）

#### Scenario: 取消后清理资源
- **WHEN** 会话被取消
- **THEN** 释放相关资源（终端进程、临时文件等）

### Requirement: 会话状态管理

系统 SHALL 正确管理每个会话的独立状态。

#### Scenario: 会话隔离
- **WHEN** 多个会话同时运行
- **THEN** 各会话状态互不干扰

#### Scenario: 会话工作目录
- **WHEN** 会话指定工作目录
- **THEN** 工具操作基于该目录执行

#### Scenario: 会话权限模式
- **WHEN** 会话配置权限模式（plan/default）
- **THEN** 根据模式调整工具调用行为

### Requirement: 会话流式响应

系统 SHALL 支持会话的流式响应，逐步发送更新。

#### Scenario: 流式文本输出
- **WHEN** 代理生成文本响应
- **THEN** 通过会话更新逐步发送内容

#### Scenario: 流式工具调用
- **WHEN** 代理调用工具
- **THEN** 发送工具调用开始和结束通知

#### Scenario: 流式结束
- **WHEN** 会话处理完成
- **THEN** 发送最终状态更新

### Requirement: 会话输入流

系统 SHALL 支持 Pushable 输入流，允许在会话运行时添加新的用户消息。

#### Scenario: 推送用户消息
- **WHEN** 调用 Pushable 的 Push 方法
- **THEN** 消息被添加到会话输入流

#### Scenario: 结束输入流
- **WHEN** 调用 Pushable 的 End 方法
- **THEN** 输入流关闭，会话正常结束

#### Scenario: 并发推送
- **WHEN** 多个 goroutine 同时推送消息
- **THEN** 消息按顺序处理，无数据竞争

### Requirement: 会话终端管理

系统 SHALL 为每个会话管理独立的终端进程。

#### Scenario: 创建终端
- **WHEN** 执行 Bash 工具且需要在后台运行
- **THEN** 创建新的终端进程并返回 ID

#### Scenario: 获取终端输出
- **WHEN** 请求终端输出
- **THEN** 返回自上次请求以来的新输出

#### Scenario: 终止终端
- **WHEN** 调用 KillShell 或会话取消
- **THEN** 终止关联的终端进程

#### Scenario: 终端状态追踪
- **WHEN** 终端进程状态变化
- **THEN** 更新终端状态（started/exited/killed/timedOut/aborted）
