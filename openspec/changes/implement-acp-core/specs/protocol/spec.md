## ADDED Requirements

### Requirement: 协议版本协商

系统 SHALL 在初始化时协商协议版本，确保客户端和代理使用兼容的协议版本。

#### Scenario: 版本兼容
- **WHEN** 客户端请求的协议版本在支持范围内
- **THEN** 返回协商后的协议版本

#### Scenario: 版本不兼容
- **WHEN** 客户端请求的协议版本不被支持
- **THEN** 返回错误响应说明版本不兼容

### Requirement: Initialize 请求处理

系统 SHALL 正确处理 `initialize` 请求，返回代理能力和支持的协议版本。

#### Scenario: 基本初始化
- **WHEN** 收到有效的 `initialize` 请求
- **THEN** 返回包含代理能力、协议版本和认证方法的响应

#### Scenario: 带客户端信息的初始化
- **WHEN** 客户端提供 `clientInfo`
- **THEN** 记录客户端信息用于日志和调试

#### Scenario: 客户端能力记录
- **WHEN** 客户端声明 `clientCapabilities`
- **THEN** 系统根据能力调整行为（如文件系统操作）

### Requirement: 协议类型定义

系统 SHALL 定义与 ACP Schema 精确映射的 Go 类型。

#### Scenario: 请求类型映射
- **WHEN** 解析 ACP 请求
- **THEN** 正确映射到对应的 Go 结构体

#### Scenario: 响应类型映射
- **WHEN** 生成 ACP 响应
- **THEN** Go 结构体正确序列化为 JSON

#### Scenario: 可选字段处理
- **WHEN** JSON 中缺少可选字段
- **THEN** 使用默认值或零值而不报错

### Requirement: 内容块处理

系统 SHALL 支持多种内容块类型：文本、图片、音频、嵌入式资源。

#### Scenario: 文本内容块
- **WHEN** 处理 `type: "text"` 的内容块
- **THEN** 正确提取文本内容

#### Scenario: 图片内容块
- **WHEN** 处理 `type: "image"` 的内容块
- **THEN** 正确解析图片的 MIME 类型、URL 或 base64 数据

#### Scenario: 音频内容块
- **WHEN** 处理 `type: "audio"` 的内容块
- **THEN** 正确解析音频数据和格式

#### Scenario: 嵌入式资源
- **WHEN** 处理 `type: "resource"` 的内容块
- **THEN** 正确解析嵌入式资源内容

### Requirement: 会话更新消息

系统 SHALL 支持发送各种类型的会话更新通知。

#### Scenario: 文本内容更新
- **WHEN** 代理生成文本响应
- **THEN** 发送 `contentBlock` 类型的会话更新

#### Scenario: 工具调用更新
- **WHEN** 代理调用工具
- **THEN** 发送包含工具信息的会话更新

#### Scenario: 状态更新
- **WHEN** 会话状态变化
- **THEN** 发送 `status` 类型的会话更新

### Requirement: 错误响应格式

系统 SHALL 使用标准格式返回错误响应。

#### Scenario: 请求错误
- **WHEN** 请求格式无效或缺少必需字段
- **THEN** 返回包含错误代码和描述的响应

#### Scenario: 内部错误
- **WHEN** 处理请求时发生内部错误
- **THEN** 返回通用错误响应，记录详细错误日志

### Requirement: 元数据支持

系统 SHALL 支持在请求和响应中携带 `_meta` 字段。

#### Scenario: 请求元数据透传
- **WHEN** 客户端在请求中包含 `_meta`
- **THEN** 系统保留元数据用于扩展功能

#### Scenario: 响应元数据添加
- **WHEN** 系统需要返回额外信息
- **THEN** 通过 `_meta` 字段添加非标准扩展
