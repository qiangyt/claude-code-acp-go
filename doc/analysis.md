# TypeScript 项目分析

> 本文档详细分析原始 TypeScript 实现 `zed-industries/claude-code-acp` 的架构和组件。

## 项目概述

**项目名称**: `@zed-industries/claude-code-acp`

**用途**: 一个适配器，将 ACP (Agent Client Protocol) 桥接到 Claude Agent SDK，使 ACP 兼容客户端（如 Zed 编辑器）能够与 Claude Code 通信。

**仓库**: https://github.com/zed-industries/claude-code-acp

**许可证**: Apache-2.0

**版本**: 0.16.1

---

## 项目结构

```
claude-code-acp/
├── src/
│   ├── index.ts              # CLI 入口点
│   ├── lib.ts                # 库导出
│   ├── acp-agent.ts          # 主 ACP 代理实现 (1633 行)
│   ├── tools.ts              # 工具转换工具 (832 行)
│   ├── mcp-server.ts         # ACP 工具的 MCP 服务器 (912 行)
│   ├── settings.ts           # 设置管理 (528 行)
│   ├── utils.ts              # 实用函数 (187 行)
│   └── tests/
│       ├── acp-agent.test.ts
│       ├── tools.test.ts
│       ├── settings.test.ts
│       └── ...
├── package.json
├── tsconfig.json
└── README.md
```

**总代码量**: ~4000+ 行 TypeScript

---

## ACP 协议解释

### 什么是 ACP?

Agent Client Protocol (ACP) 是 AI 编码代理和编辑器之间的标准化通信协议。

### 协议特性

1. **NDJSON 流式传输**: 通过 stdin/stdout 进行换行分隔 JSON 通信
2. **请求-响应模式**: 客户端发送请求，代理响应
3. **会话基础**: 所有交互都限定在会话范围内
4. **双向通信**: 客户端和代理都可以发起更新

### 协议流程

```
Client (e.g., Zed)                    Agent (claude-code-acp)
      |                                      |
      |------ initialize() ----------------->|
      |<----- InitializeResponse ------------|
      |                                      |
      |------ newSession() ----------------->|
      |<----- NewSessionResponse ------------|
      |                                      |
      |------ prompt() --------------------->|
      |<----- sessionUpdate (streaming) -----|
      |<----- sessionUpdate (streaming) -----|
      |<----- PromptResponse ----------------|
      |                                      |
      |------ setSessionMode() ------------->|
      |<----- SetSessionModeResponse --------|
```

### ACP 类型 (来自 `@agentclientprotocol/sdk`)

- `Agent` - 代理必须实现的接口
- `AgentSideConnection` - 代理端连接处理器
- `ClientCapabilities` - 客户端支持的能力
- `SessionNotification` - 会话期间发送的更新
- `NewSessionRequest/Response` - 会话创建
- `PromptRequest/Response` - 用户提示
- `LoadSessionRequest/Response` - 会话恢复
- `ListSessionsRequest/Response` - 会话列表

---

## 核心组件

### 1. `index.ts` - 入口点

**职责**:
1. 加载托管设置并应用环境变量
2. 将控制台输出重定向到 stderr (stdout 用于 ACP)
3. 调用 `runAcp()` 启动代理
4. 通过 `stdin.resume()` 保持进程活跃

```typescript
// 入口点加载平台特定路径的托管设置
// 将所有控制台输出重定向到 stderr
// stdout 保留给 ACP 协议
// 处理未处理的拒绝
// 启动 ACP 代理
```

### 2. `acp-agent.ts` - 核心代理实现

**类**: `ClaudeAcpAgent` (实现 `Agent` 接口)

**关键属性**:

```typescript
sessions: { [key: string]: Session }  // 活跃会话
client: AgentSideConnection           // 到客户端的连接
toolUseCache: ToolUseCache            // 工具使用跟踪缓存
backgroundTerminals: {...}            // 后台终端管理
clientCapabilities?: ClientCapabilities
logger: Logger
```

**关键方法**:

| 方法 | 用途 |
|------|------|
| `initialize()` | 返回代理能力、认证方法 |
| `newSession()` | 创建新的 Claude Code 会话 |
| `prompt()` | 处理用户提示，流式响应 |
| `cancel()` | 取消正在进行的操作 |
| `setSessionMode()` | 更改权限模式 |
| `canUseTool()` | 工具执行的权限回调 |
| `loadSession()` | 从磁盘加载现有会话 |
| `unstable_listSessions()` | 列出可用会话 |
| `unstable_forkSession()` | 分支会话 |
| `unstable_resumeSession()` | 恢复之前的会话 |

**Session 类型**:

```typescript
type Session = {
  query: Query;                        // Claude Agent SDK 查询对象
  input: Pushable<SDKUserMessage>;     // 推送式输入流
  cancelled: boolean;
  permissionMode: PermissionMode;
  settingsManager: SettingsManager;
};
```

### 3. `tools.ts` - 工具转换工具

**用途**: 在 Claude Code 工具格式和 ACP 工具格式之间转换

**关键函数**:

1. **`toolInfoFromToolUse()`** - 将 Claude 工具使用转换为 ACP `ToolInfo`:

```typescript
interface ToolInfo {
  title: string;              // 人类可读标题
  kind: ToolKind;             // "read" | "edit" | "execute" | "search" | "think" | "fetch" | "switch_mode" | "other"
  content: ToolCallContent[]; // 内容块
  locations?: ToolCallLocation[]; // 文件位置
}
```

2. **`toolUpdateFromToolResult()`** - 将工具结果转换为 ACP 更新

3. **`planEntries()`** - 将 TodoWrite todos 转换为 ACP plan 条目

4. **Hook 回调** 用于 `PreToolUse` 和 `PostToolUse` 事件

**工具名称映射**:

```typescript
const acpToolNames = {
  read: "mcp__acp__Read",
  edit: "mcp__acp__Edit",
  write: "mcp__acp__Write",
  bash: "mcp__acp__Bash",
  killShell: "mcp__acp__KillShell",
  bashOutput: "mcp__acp__BashOutput",
};
```

### 4. `mcp-server.ts` - MCP 服务器实现

**用途**: 创建一个 MCP (Model Context Protocol) 服务器，提供 ACP 特定工具

**提供的工具**:

| 工具 | 描述 |
|------|------|
| `Read` | 通过 ACP 客户端读取文件 |
| `Write` | 通过 ACP 客户端写入文件 |
| `Edit` | 通过 ACP 客户端编辑文件 |
| `Bash` | 通过 ACP 客户端执行终端命令 |
| `BashOutput` | 获取后台 shell 的输出 |
| `KillShell` | 终止后台进程 |

**为什么使用 MCP?**: MCP 允许 Claude Code 使用委托给 ACP 客户端而不是直接执行的工具。这实现了:
- 客户端文件访问（遵守项目边界）
- 通过客户端进行终端管理
- 一致的工具接口

### 5. `settings.ts` - 设置管理

**类**: `SettingsManager`

**用途**: 从多个来源加载和管理 Claude Code 设置，具有正确的优先级

**设置来源（按优先级顺序）**:

1. 用户设置 (`~/.claude/settings.json`)
2. 项目设置 (`<cwd>/.claude/settings.json`)
3. 本地项目设置 (`<cwd>/.claude/settings.local.json`)
4. 企业托管设置（平台特定）

**设置结构**:

```typescript
interface ClaudeCodeSettings {
  permissions?: PermissionSettings;
  env?: Record<string, string>;
  model?: string;
}

interface PermissionSettings {
  allow?: string[];   // 自动允许的工具
  deny?: string[];    // 自动拒绝的工具
  ask?: string[];     // 总是询问的工具
  additionalDirectories?: string[];
  defaultMode?: string;
}
```

**权限规则语法**:

```
"Read"              - 所有 Read 操作
"Read(./.env)"      - 特定文件
"Read(./secrets/**)" - glob 模式
"Bash(npm run lint)" - 精确命令
"Bash(npm run:*)"   - 带通配符的命令前缀
```

### 6. `utils.ts` - 工具

**关键工具**:

1. **`Pushable<T>`** - 允许推送式消费的异步可迭代对象:

```typescript
class Pushable<T> implements AsyncIterable<T> {
  push(item: T): void;
  end(): void;
  [Symbol.asyncIterator](): AsyncIterator<T>;
}
```

2. **流转换器**:

```typescript
nodeToWebWritable(nodeStream: Writable): WritableStream<Uint8Array>
nodeToWebReadable(nodeStream: Readable): ReadableStream<Uint8Array>
```

3. **`encodeProjectPath()`** - 为 Claude 会话存储编码路径

4. **`extractLinesWithByteLimit()`** - 带大小限制提取文件内容

---

## 数据结构和类型

### 核心类型

```typescript
// 会话跟踪
type Session = {
  query: Query;
  input: Pushable<SDKUserMessage>;
  cancelled: boolean;
  permissionMode: PermissionMode;
  settingsManager: SettingsManager;
};

// 后台终端状态
type BackgroundTerminal =
  | { handle: TerminalHandle; status: "started"; lastOutput: TerminalOutputResponse | null }
  | { status: "aborted" | "exited" | "killed" | "timedOut"; pendingOutput: TerminalOutputResponse };

// 工具使用缓存用于结果关联
type ToolUseCache = {
  [key: string]: {
    type: "tool_use" | "server_tool_use" | "mcp_tool_use";
    id: string;
    name: string;
    input: unknown;
  };
};

// 新会话元数据
type NewSessionMeta = {
  claudeCode?: {
    options?: Options;  // Claude Agent SDK 选项
  };
};

// 工具更新元数据
type ToolUpdateMeta = {
  claudeCode?: {
    toolName: string;
    toolResponse?: unknown;
  };
};
```

### ACP 通知类型

代理发送各种 `SessionNotification` 更新:

```typescript
// 消息块
{ sessionUpdate: "agent_message_chunk", content: { type: "text", text: "..." } }
{ sessionUpdate: "user_message_chunk", content: { type: "text", text: "..." } }

// 思考
{ sessionUpdate: "agent_thought_chunk", content: { type: "text", text: "..." } }

// 工具调用
{ sessionUpdate: "tool_call", toolCallId: "...", kind: "...", title: "...", rawInput: {...} }
{ sessionUpdate: "tool_call_update", toolCallId: "...", status: "completed" | "failed", rawOutput: {...} }

// Plan 更新
{ sessionUpdate: "plan", entries: [{ content: "...", status: "pending" | "in_progress" | "completed", priority: "medium" }] }

// 模式变更
{ sessionUpdate: "current_mode_update", currentModeId: "..." }

// 命令
{ sessionUpdate: "available_commands_update", availableCommands: [...] }
```

---

## 消息传递 / 通信流程

### 初始化流程

```
1. 客户端通过 stdin/stdout NDJSON 流连接
2. 客户端发送带能力的 initialize 请求
3. 代理响应:
   - protocolVersion: 1
   - agentCapabilities (promptCapabilities, mcpCapabilities, loadSession, sessionCapabilities)
   - agentInfo (name, title, version)
   - authMethods
```

### 会话创建流程

```
1. 客户端发送 newSession({ cwd, mcpServers, _meta })
2. 代理:
   - 为 cwd 创建 SettingsManager
   - 创建 Pushable 输入流
   - 配置 MCP 服务器（包括 ACP 服务器）
   - 设置 hooks (PreToolUse, PostToolUse)
   - 使用 Claude Agent SDK 创建 Query
   - 返回 sessionId, models 和 modes
3. 代理发送 available_commands_update
```

### Prompt 处理流程

```
1. 客户端发送 prompt({ sessionId, prompt: [...] })
2. 代理:
   - 通过 promptToClaude() 将 ACP prompt 转换为 SDK 格式
   - 推送到会话的输入流
   - 迭代 query.next() 结果:
     a. "system" 消息 -> 跳过/记录
     b. "result" 消息 -> 返回 stopReason
     c. "stream_event" 消息 -> 转换为 ACP 通知
     d. "user"/"assistant" 消息 -> 转换为 ACP 通知
   - 为每个通知发送 sessionUpdate
3. 返回带 stopReason 的 PromptResponse
```

### 权限流程

```
1. Claude 想要使用工具
2. SDK 调用 canUseTool 回调
3. 代理检查 permissionMode:
   - "bypassPermissions" -> 允许
   - "acceptEdits" + 编辑工具 -> 允许
   - 否则 -> client.requestPermission()
4. 客户端显示权限对话框
5. 用户选择选项
6. 代理返回 { behavior: "allow" | "deny", ... }
```

### 流式响应转换

```typescript
// 从流事件到 ACP 通知
streamEventToAcpNotifications(message, sessionId, toolUseCache, client, logger)

// 处理:
// - content_block_start -> 新工具使用、文本、思考
// - content_block_delta -> text_delta, thinking_delta
// - message_start/delta/stop -> 无内容
// - content_block_stop -> 无内容
```

---

## 依赖

### 生产依赖

| 包 | 版本 | 用途 |
|---|------|------|
| `@agentclientprotocol/sdk` | 0.14.1 | ACP 协议类型和连接处理 |
| `@anthropic-ai/claude-agent-sdk` | 0.2.38 | Claude Code SDK 用于 AI 交互 |
| `@modelcontextprotocol/sdk` | 1.26.0 | MCP 服务器实现 |
| `diff` | 8.0.3 | 文件编辑的 diff 生成 |
| `minimatch` | 10.1.2 | 权限的 glob 模式匹配 |

### 开发依赖

| 包 | 版本 | 用途 |
|---|------|------|
| `@anthropic-ai/sdk` | 0.74.0 | Anthropic API 类型 |
| `@types/node` | 25.2.3 | Node.js 类型定义 |
| `typescript` | 5.9.3 | TypeScript 编译器 |
| `vitest` | 4.0.18 | 测试框架 |
| `eslint` + 插件 | various | 代码检查 |
| `prettier` | 3.8.1 | 代码格式化 |

---

## 支持的功能

1. **上下文 @-mentions** - 在提示中引用文件
2. **图片** - 提示和响应中的图片支持
3. **工具调用** - 带权限请求
4. **Following** - 跟踪代理的文件位置
5. **编辑审查** - 应用前审查文件更改
6. **TODO 列表** - 通过 TodoWrite 的 Plan 管理
7. **交互式终端** - 前台和后台
8. **斜杠命令** - 来自 MCP 服务器的自定义命令
9. **客户端 MCP 服务器** - 来自客户端的额外工具
10. **会话管理** - 列表、加载、分支、恢复会话
11. **模型选择** - 在 Claude 模型之间切换
12. **权限模式** - default, acceptEdits, bypassPermissions, dontAsk, plan

---

## 配置

### 环境变量

- `ANTHROPIC_API_KEY` - 认证 API 密钥
- `CLAUDE_CONFIG_DIR` - 覆盖配置目录位置
- `CLAUDE_CODE_EXECUTABLE` - 覆盖 Claude Code 可执行文件路径
- `MAX_THINKING_TOKENS` - 配置思考令牌预算
- `IS_SANDBOX` - 在 root 模式下启用绕过权限

### 设置文件

- `~/.claude/settings.json` - 用户设置
- `<project>/.claude/settings.json` - 项目设置
- `<project>/.claude/settings.local.json` - 本地项目设置 (git-ignored)
- 平台特定的企业托管设置

---

## 复杂度分析

| 模块 | 代码行数 | 复杂度 | Go 实现难度 |
|------|---------|--------|-------------|
| acp-agent.ts | 1633 | 高 | ⭐⭐⭐⭐ |
| tools.ts | 832 | 中 | ⭐⭐⭐ |
| mcp-server.ts | 912 | 中高 | ⭐⭐⭐⭐ |
| settings.ts | 528 | 中 | ⭐⭐ |
| utils.ts | 187 | 低 | ⭐ |

---

*文档生成日期: 2026-02-13*