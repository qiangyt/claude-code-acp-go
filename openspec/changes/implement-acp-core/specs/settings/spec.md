## ADDED Requirements

### Requirement: 设置来源层次

系统 SHALL 从以下来源加载设置（按优先级从低到高）：
1. 用户设置 (`~/.claude/settings.json`)
2. 项目设置 (`$CWD/.claude/settings.json`)
3. 本地项目设置 (`$CWD/.claude/settings.local.json`)
4. 企业托管设置 (系统级配置)

#### Scenario: 加载用户设置
- **WHEN** 用户设置文件存在
- **THEN** 从 `~/.claude/settings.json` 加载设置

#### Scenario: 加载项目设置
- **WHEN** 项目设置文件存在
- **THEN** 从 `$CWD/.claude/settings.json` 加载设置

#### Scenario: 加载本地设置
- **WHEN** 本地设置文件存在
- **THEN** 从 `$CWD/.claude/settings.local.json` 加载设置（不应提交到版本控制）

#### Scenario: 设置合并
- **WHEN** 多个来源都有设置
- **THEN** 高优先级来源覆盖低优先级来源的同名设置

#### Scenario: 文件不存在
- **WHEN** 某个设置文件不存在
- **THEN** 跳过该来源，继续加载其他来源

### Requirement: 设置结构

系统 SHALL 支持以下设置字段：
- `permissions` - 权限配置
- `env` - 环境变量
- `model` - 模型选择

#### Scenario: 权限设置
- **WHEN** 设置包含 `permissions` 字段
- **THEN** 解析 allow、deny、ask、additionalDirectories 规则

#### Scenario: 环境变量设置
- **WHEN** 设置包含 `env` 字段
- **THEN** 将环境变量注入工具执行环境

#### Scenario: 模型设置
- **WHEN** 设置包含 `model` 字段
- **THEN** 使用指定模型创建会话

### Requirement: 设置验证

系统 SHALL 验证设置文件格式和内容。

#### Scenario: JSON 格式验证
- **WHEN** 设置文件包含无效 JSON
- **THEN** 返回解析错误并跳过该文件

#### Scenario: 字段类型验证
- **WHEN** 设置字段类型不正确
- **THEN** 返回验证错误

#### Scenario: 未知字段处理
- **WHEN** 设置包含未知字段
- **THEN** 忽略未知字段

### Requirement: 设置热重载

系统 SHALL 支持在运行时重新加载设置。

#### Scenario: 检测设置变更
- **WHEN** 设置文件被修改
- **THEN** 在下次会话创建时加载新设置

#### Scenario: 不影响运行中会话
- **WHEN** 设置被修改
- **THEN** 已运行的会话继续使用旧设置

### Requirement: 敏感信息保护

系统 SHALL 保护设置中的敏感信息。

#### Scenario: 本地设置不提交
- **WHEN** 用户创建本地设置
- **THEN** 提示添加到 `.gitignore`

#### Scenario: API 密钥不记录
- **WHEN** 设置包含 API 密钥
- **THEN** 不在日志中记录完整密钥

### Requirement: 设置访问 API

系统 SHALL 提供便捷的设置访问方法。

#### Scenario: 获取合并后设置
- **WHEN** 调用设置管理器的获取方法
- **THEN** 返回合并后的最终设置

#### Scenario: 获取特定来源设置
- **WHEN** 指定设置来源
- **THEN** 返回该来源的原始设置

#### Scenario: 检查权限
- **WHEN** 调用权限检查方法
- **THEN** 根据合并后的权限设置返回决策

### Requirement: 默认设置

系统 SHALL 为缺失的设置提供合理默认值。

#### Scenario: 无设置文件
- **WHEN** 没有任何设置文件存在
- **THEN** 使用内置默认设置

#### Scenario: 部分字段缺失
- **WHEN** 设置缺少某些字段
- **THEN** 使用字段的默认值
