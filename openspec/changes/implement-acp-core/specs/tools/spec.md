## ADDED Requirements

### Requirement: 工具信息转换

系统 SHALL 将 SDK 工具使用转换为 ACP ToolInfo 格式。

#### Scenario: Read 工具转换
- **WHEN** SDK 调用 Read 工具
- **THEN** 转换为 `kind: "read"` 的 ToolInfo，包含文件路径和行号

#### Scenario: Write 工具转换
- **WHEN** SDK 调用 Write 工具
- **THEN** 转换为 `kind: "edit"` 的 ToolInfo，包含 diff 内容

#### Scenario: Edit 工具转换
- **WHEN** SDK 调用 Edit 工具
- **THEN** 转换为 `kind: "edit"` 的 ToolInfo，包含 old_text/new_text diff

#### Scenario: Bash 工具转换
- **WHEN** SDK 调用 Bash 工具
- **THEN** 转换为 `kind: "execute"` 的 ToolInfo，显示命令

#### Scenario: Glob 工具转换
- **WHEN** SDK 调用 Glob 工具
- **THEN** 转换为 `kind: "search"` 的 ToolInfo

#### Scenario: Grep 工具转换
- **WHEN** SDK 调用 Grep 工具
- **THEN** 转换为 `kind: "search"` 的 ToolInfo

#### Scenario: Task 工具转换
- **WHEN** SDK 调用 Task 工具
- **THEN** 转换为 `kind: "other"` 的 ToolInfo，显示子代理信息

#### Scenario: 未知工具转换
- **WHEN** SDK 调用未识别的工具
- **THEN** 转换为 `kind: "other"` 的 ToolInfo，使用工具名称作为标题

### Requirement: 工具类型枚举

系统 SHALL 支持以下工具类型：
- `read` - 文件读取操作
- `edit` - 文件编辑操作
- `move` - 文件移动操作
- `search` - 搜索操作
- `execute` - 命令执行操作
- `think` - 思考/规划操作
- `fetch` - 网络获取操作
- `switch_mode` - 模式切换操作
- `other` - 其他操作

#### Scenario: 类型分类正确
- **WHEN** 转换工具信息
- **THEN** 根据工具名称正确分类工具类型

### Requirement: 工具位置信息

系统 SHALL 在 ToolInfo 中提供文件位置信息。

#### Scenario: 单文件位置
- **WHEN** 工具操作单个文件
- **THEN** Locations 包含该文件路径

#### Scenario: 多文件位置
- **WHEN** 工具操作多个文件
- **THEN** Locations 包含所有相关文件路径

#### Scenario: 行号信息
- **WHEN** 可用（如 Read 的 offset/limit）
- **THEN** Locations 包含行号范围

### Requirement: 工具内容差异

系统 SHALL 在 ToolInfo 中提供 diff 内容。

#### Scenario: Write 工具差异
- **WHEN** Write 工具创建新文件
- **THEN** Content 包含 `type: "diff"`，oldText 为空，newText 为文件内容

#### Scenario: Edit 工具差异
- **WHEN** Edit 工具修改文件
- **THEN** Content 包含 `type: "diff"`，包含 oldText 和 newText

### Requirement: MCP 工具注册

系统 SHALL 将 ACP 工具注册为 MCP 工具供 SDK 使用。

#### Scenario: 注册 Read 工具
- **WHEN** MCP 服务器启动
- **THEN** 注册 `Read` 工具，接受 file_path、offset、limit 参数

#### Scenario: 注册 Write 工具
- **WHEN** MCP 服务器启动
- **THEN** 注册 `Write` 工具，接受 file_path、content 参数

#### Scenario: 注册 Edit 工具
- **WHEN** MCP 服务器启动
- **THEN** 注册 `Edit` 工具，接受 file_path、old_string、new_string、replace_all 参数

#### Scenario: 注册 Bash 工具
- **WHEN** MCP 服务器启动
- **THEN** 注册 `Bash` 工具，接受 command、description、timeout、run_in_background 参数

#### Scenario: 注册 BashOutput 工具
- **WHEN** MCP 服务器启动
- **THEN** 注册 `BashOutput` 工具，接受 task_id 参数

#### Scenario: 注册 KillShell 工具
- **WHEN** MCP 服务器启动
- **THEN** 注册 `KillShell` 工具，接受 shell_id 参数

### Requirement: 工具调用代理

系统 SHALL 将 MCP 工具调用代理到 ACP 客户端。

#### Scenario: Read 工具代理
- **WHEN** SDK 调用 `mcp__acp__Read` 工具
- **THEN** 通过 ACP 客户端的 `readTextFile` 方法读取文件

#### Scenario: Write 工具代理
- **WHEN** SDK 调用 `mcp__acp__Write` 工具
- **THEN** 通过 ACP 客户端的 `writeTextFile` 方法写入文件

#### Scenario: Bash 工具代理
- **WHEN** SDK 调用 `mcp__acp__Bash` 工具
- **THEN** 通过 ACP 客户端的 `startShell` 和 `shellOutput` 方法执行命令

### Requirement: 工具缓存

系统 SHALL 缓存工具使用信息，避免重复转换。

#### Scenario: 缓存命中
- **WHEN** 相同的工具使用再次出现
- **THEN** 从缓存返回 ToolInfo

#### Scenario: 缓存更新
- **WHEN** 工具使用参数变化
- **THEN** 更新缓存条目
