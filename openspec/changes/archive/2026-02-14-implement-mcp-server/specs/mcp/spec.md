## ADDED Requirements

### Requirement: MCP Server Creation

系统 SHALL 使用 `schlunsen/claude-agent-sdk-go` 提供的 MCP 工厂函数创建 MCP 服务器实例。

#### Scenario: 创建 MCP 服务器成功
- **WHEN** 调用 NewServer 工厂函数
- **THEN** 返回配置完成的 MCP Server 实例
- **AND** 服务器处于未启动状态

#### Scenario: 创建 MCP 服务器带配置选项
- **WHEN** 调用 NewServer 并传入配置选项
- **THEN** 服务器使用指定配置初始化
- **AND** 权限回调和钩子函数已注册

### Requirement: Tool Registration

系统 SHALL 支持动态注册 MCP 工具。

#### Scenario: 注册单个工具
- **WHEN** 调用 RegisterTool 方法
- **THEN** 工具被添加到服务器的工具列表
- **AND** 工具可通过名称调用

#### Scenario: 注册多个工具
- **WHEN** 调用 RegisterTools 方法传入多个工具
- **THEN** 所有工具被批量注册
- **AND** 每个工具可独立调用

### Requirement: Server Lifecycle

系统 SHALL 提供 MCP 服务器的生命周期管理。

#### Scenario: 启动服务器
- **WHEN** 调用 Start 方法
- **THEN** 服务器开始监听工具调用请求
- **AND** 返回 nil 表示启动成功

#### Scenario: 停止服务器
- **WHEN** 调用 Stop 方法
- **THEN** 服务器停止接收新请求
- **AND** 等待进行中的请求完成
- **AND** 释放所有资源

#### Scenario: 优雅关闭
- **WHEN** 收到关闭信号
- **THEN** 服务器等待所有活跃会话完成
- **AND** 终止所有子进程
- **AND** 清理临时资源

### Requirement: Read Tool

系统 SHALL 提供 Read 工具以读取文件内容。

#### Scenario: 读取存在的文件
- **WHEN** 调用 Read 工具并传入有效文件路径
- **THEN** 返回文件的完整内容
- **AND** 内容以字符串形式返回

#### Scenario: 读取不存在的文件
- **WHEN** 调用 Read 工具并传入不存在的路径
- **THEN** 返回错误信息
- **AND** 错误表明文件不存在

#### Scenario: 读取目录路径
- **WHEN** 调用 Read 工具并传入目录路径
- **THEN** 返回目录列表信息
- **AND** 列表包含目录内的文件和子目录

### Requirement: Write Tool

系统 SHALL 提供 Write 工具以写入文件内容。

#### Scenario: 写入新文件
- **WHEN** 调用 Write 工具并传入新文件路径和内容
- **THEN** 创建文件并写入内容
- **AND** 自动创建父目录（如果不存在）

#### Scenario: 覆盖已存在的文件
- **WHEN** 调用 Write 工具并传入已存在文件的路径
- **THEN** 文件内容被完全覆盖
- **AND** 不保留原有内容

#### Scenario: 写入无权限的路径
- **WHEN** 调用 Write 工具并传入无写入权限的路径
- **THEN** 返回权限错误

### Requirement: Edit Tool

系统 SHALL 提供 Edit 工具以编辑文件内容。

#### Scenario: 替换存在的字符串
- **WHEN** 调用 Edit 工具并传入文件路径、旧字符串和新字符串
- **THEN** 文件中的旧字符串被替换为新字符串
- **AND** 返回替换结果信息

#### Scenario: 替换不存在的字符串
- **WHEN** 调用 Edit 工具并传入文件中不存在的字符串
- **THEN** 返回错误表明未找到匹配
- **AND** 文件内容保持不变

#### Scenario: 多次替换
- **WHEN** 文件中存在多个匹配的字符串
- **THEN** 所有匹配都被替换

### Requirement: Bash Tool

系统 SHALL 提供 Bash 工具以执行 Shell 命令。

#### Scenario: 执行简单命令
- **WHEN** 调用 Bash 工具并传入有效命令
- **THEN** 在 Shell 中执行命令
- **AND** 返回命令的标准输出
- **AND** 返回命令的退出码

#### Scenario: 执行失败命令
- **WHEN** 调用 Bash 工具并执行会失败的命令
- **THEN** 返回标准错误输出
- **AND** 返回非零退出码

#### Scenario: 命令超时
- **WHEN** 命令执行超过配置的超时时间
- **THEN** 终止命令执行
- **AND** 返回超时错误

#### Scenario: 长时间运行命令
- **WHEN** 调用 Bash 工具并指定为后台运行
- **THEN** 立即返回终端 ID
- **AND** 命令在后台继续执行

### Requirement: BashOutput Tool

系统 SHALL 提供 BashOutput 工具以获取后台命令的输出。

#### Scenario: 获取存在的终端输出
- **WHEN** 调用 BashOutput 工具并传入有效终端 ID
- **THEN** 返回该终端的累积输出
- **AND** 输出包含标准输出和标准错误

#### Scenario: 获取不存在的终端输出
- **WHEN** 调用 BashOutput 工具并传入不存在的终端 ID
- **THEN** 返回错误表明终端不存在

#### Scenario: 获取已完成命令的输出
- **WHEN** 终端对应的命令已执行完成
- **THEN** 返回完整的输出
- **AND** 包含最终的退出状态

### Requirement: KillShell Tool

系统 SHALL 提供 KillShell 工具以终止后台命令。

#### Scenario: 终止存在的终端
- **WHEN** 调用 KillShell 工具并传入有效终端 ID
- **THEN** 发送终止信号给对应进程
- **AND** 返回终止成功

#### Scenario: 终止不存在的终端
- **WHEN** 调用 KillShell 工具并传入不存在的终端 ID
- **THEN** 返回错误表明终端不存在

#### Scenario: 终止已完成的终端
- **WHEN** 终端对应的命令已执行完成
- **THEN** 返回表明进程已结束

### Requirement: Permission Integration

系统 SHALL 在工具执行前进行权限检查。

#### Scenario: 允许的工具调用
- **WHEN** 工具调用匹配 allow 规则
- **THEN** 直接执行工具
- **AND** 不发送权限请求

#### Scenario: 拒绝的工具调用
- **WHEN** 工具调用匹配 deny 规则
- **THEN** 拒绝执行工具
- **AND** 返回权限被拒绝错误

#### Scenario: 需要询问的工具调用
- **WHEN** 工具调用匹配 ask 规则
- **THEN** 发送权限请求通知给客户端
- **AND** 等待用户决策
- **AND** 根据用户决策执行或拒绝

### Requirement: Session Context Management

系统 SHALL 为每个会话维护独立的 MCP 上下文。

#### Scenario: 创建会话上下文
- **WHEN** 创建新的 ACP 会话
- **THEN** 创建对应的 MCP 会话上下文
- **AND** 上下文包含独立的终端追踪

#### Scenario: 清理会话上下文
- **WHEN** 会话结束或取消
- **THEN** 终止所有关联的后台进程
- **AND** 释放会话资源

#### Scenario: 并发会话隔离
- **WHEN** 多个会话同时使用 MCP 工具
- **THEN** 每个会话的终端相互独立
- **AND** 一个会话的操作不影响其他会话
