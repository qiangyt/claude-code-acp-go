## ADDED Requirements

### Requirement: NDJSON 编码

系统 SHALL 实现 NDJSON (Newline-Delimited JSON) 编码器，将 JSON 对象转换为单行 JSON 字符串并以换行符分隔。

#### Scenario: 编码简单对象
- **WHEN** 调用编码器编码 `{"type": "test", "value": 123}`
- **THEN** 输出为 `{"type":"test","value":123}\n`

#### Scenario: 编码嵌套对象
- **WHEN** 调用编码器编码包含嵌套结构的对象
- **THEN** 正确序列化嵌套对象为紧凑 JSON 格式

#### Scenario: 编码包含特殊字符的字符串
- **WHEN** 调用编码器编码包含换行符或引号的字符串
- **THEN** 正确转义特殊字符

### Requirement: NDJSON 解码

系统 SHALL 实现 NDJSON 解码器，将 NDJSON 格式的输入流解析为 JSON 对象序列。

#### Scenario: 解码单行 JSON
- **WHEN** 输入为 `{"type":"response","id":1}\n`
- **THEN** 解码器返回 `{"type": "response", "id": 1}` 对象

#### Scenario: 解码多行 JSON
- **WHEN** 输入为多行 NDJSON
- **THEN** 解码器依次返回每行对应的 JSON 对象

#### Scenario: 处理空行
- **WHEN** 输入包含空行
- **THEN** 解码器跳过空行继续处理后续内容

#### Scenario: 处理格式错误
- **WHEN** 输入包含无效 JSON
- **THEN** 解码器返回解析错误

### Requirement: stdio 传输

系统 SHALL 实现 stdin/stdout 传输层，通过标准输入接收消息，通过标准输出发送消息。

#### Scenario: 从 stdin 读取消息
- **WHEN** ACP 客户端通过 stdin 发送 NDJSON 消息
- **THEN** 系统正确读取并解析消息

#### Scenario: 向 stdout 写入消息
- **WHEN** 系统需要发送响应或通知
- **THEN** 将消息编码为 NDJSON 并写入 stdout

#### Scenario: 并发读写
- **WHEN** 同时进行读写操作
- **THEN** 读写操作互不阻塞，正确处理并发

### Requirement: 消息分帧

系统 SHALL 正确处理消息分帧，确保每条消息完整传输。

#### Scenario: 完整消息传输
- **WHEN** 发送完整的 JSON 消息
- **THEN** 接收方收到完整且正确的消息

#### Scenario: 大消息处理
- **WHEN** 消息大小超过默认缓冲区
- **THEN** 系统正确处理大消息而不截断

### Requirement: 传输错误处理

系统 SHALL 正确处理传输层错误并返回有意义的错误信息。

#### Scenario: 连接断开
- **WHEN** stdin 或 stdout 意外关闭
- **THEN** 系统检测到断开并优雅终止

#### Scenario: 编码错误
- **WHEN** 尝试发送无法序列化的对象
- **THEN** 返回编码错误而不崩溃
