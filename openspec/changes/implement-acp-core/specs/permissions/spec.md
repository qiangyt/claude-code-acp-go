## ADDED Requirements

### Requirement: 权限决策类型

系统 SHALL 支持三种权限决策：
- `allow` - 允许操作，无需用户确认
- `deny` - 拒绝操作
- `ask` - 需要用户确认

#### Scenario: 默认决策
- **WHEN** 没有匹配的规则
- **THEN** 默认返回 `ask` 决策

### Requirement: 权限规则格式

系统 SHALL 支持以下权限规则格式：
- `ToolName` - 匹配工具名称
- `ToolName(pattern)` - 匹配工具名称和参数模式
- `ToolName(arg:value)` - 匹配特定参数值

#### Scenario: 匹配工具名称
- **WHEN** 规则为 `Read`
- **THEN** 匹配所有 Read 工具调用

#### Scenario: 匹配路径模式
- **WHEN** 规则为 `Read(./src/**)`
- **THEN** 匹配读取 src 目录的 Read 调用

#### Scenario: 匹配命令前缀
- **WHEN** 规则为 `Bash(npm run:*)`
- **THEN** 匹配以 `npm run` 开头的 Bash 命令

#### Scenario: 通配符匹配
- **WHEN** 规则包含 `*`
- **THEN** `*` 匹配任意字符序列

### Requirement: 权限规则优先级

系统 SHALL 按以下优先级处理权限规则：
1. `deny` 规则（最高优先级）
2. `allow` 规则
3. `ask` 规则
4. 默认 `ask`（最低优先级）

#### Scenario: deny 优先于 allow
- **WHEN** 工具同时匹配 deny 和 allow 规则
- **THEN** 返回 `deny` 决策

#### Scenario: allow 覆盖 ask
- **WHEN** 工具匹配 allow 规则但不匹配 deny
- **THEN** 返回 `allow` 决策

### Requirement: 权限请求通知

系统 SHALL 在需要用户确认时发送权限请求通知。

#### Scenario: 发送权限请求
- **WHEN** 工具调用需要确认
- **THEN** 发送包含工具名称和参数的权限请求

#### Scenario: 接收权限响应
- **WHEN** 用户响应权限请求
- **THEN** 根据响应允许或拒绝工具调用

#### Scenario: 记住决策
- **WHEN** 用户选择"总是允许"
- **THEN** 添加规则到本地设置

### Requirement: 敏感操作检测

系统 SHALL 自动检测敏感操作并要求确认。

#### Scenario: 删除命令
- **WHEN** Bash 命令包含 `rm`
- **THEN** 自动要求确认

#### Scenario: 环境变量访问
- **WHEN** 尝试读取 `.env` 文件
- **THEN** 要求确认

#### Scenario: 网络操作
- **WHEN** 执行网络相关命令
- **THEN** 要求确认

### Requirement: 权限模式

系统 SHALL 支持不同的权限模式：
- `default` - 标准权限检查
- `plan` - 规划模式，限制实际执行

#### Scenario: 默认模式
- **WHEN** 权限模式为 `default`
- **THEN** 按规则检查权限

#### Scenario: 规划模式
- **WHEN** 权限模式为 `plan`
- **THEN** 限制修改性操作，仅允许读取和规划

### Requirement: 额外目录权限

系统 SHALL 支持配置额外的工作目录。

#### Scenario: 添加额外目录
- **WHEN** 设置 `additionalDirectories`
- **THEN** 允许在这些目录执行文件操作

#### Scenario: 目录边界检查
- **WHEN** 尝试在工作目录和额外目录之外操作
- **THEN** 拒绝操作
