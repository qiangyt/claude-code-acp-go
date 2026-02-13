# 如何实现 100% 正确

> 本文档定义了确保 Go 实现与 ACP 协议和原始 TypeScript 实现完全一致的策略和方法。

## 核心原则

**100% 正确 = 协议兼容 + 功能对等 + 行为一致**

```
┌─────────────────────────────────────────────────────────────┐
│                    100% 正确性保证                           │
├─────────────────────────────────────────────────────────────┤
│                                                             │
│   ┌─────────────┐   ┌─────────────┐   ┌─────────────┐      │
│   │  协议兼容   │ + │  功能对等   │ + │  行为一致   │      │
│   └─────────────┘   └─────────────┘   └─────────────┘      │
│          │                 │                 │              │
│          └─────────────────┼─────────────────┘              │
│                            ▼                                │
│                   ┌───────────────┐                         │
│                   │  100% 正确    │                         │
│                   └───────────────┘                         │
│                                                             │
└─────────────────────────────────────────────────────────────┘
```

---

## 策略一: 黄金文件测试 (Golden File Testing)

### 概念

录制 TypeScript 实现的真实请求/响应，Go 实现必须产生完全相同的输出。

### 流程

```
                    原始 TypeScript 实现
                            │
                            ▼
              ┌─────────────────────────────┐
              │   录制真实请求/响应到文件    │
              │   (golden/*.jsonl)          │
              └─────────────────────────────┘
                            │
                            ▼
              ┌─────────────────────────────┐
              │   Go 实现: 读取相同输入     │
              │   对比输出是否完全一致      │
              └─────────────────────────────┘
```

### 实现

#### 1. 录制器 (TypeScript)

```typescript
// scripts/record-golden.ts
import { ClaudeAcpAgent } from "./src/acp-agent";
import * as fs from "fs";

async function recordGolden(name: string, scenario: (agent: ClaudeAcpAgent) => Promise<void>) {
    const recordings: { input: any; output: any }[] = [];

    // 包装 agent 方法以录制
    const agent = new ClaudeAcpAgent({
        onMessage: (input, output) => {
            recordings.push({ input, output });
        }
    });

    await scenario(agent);

    // 保存到黄金文件
    fs.writeFileSync(
        `golden/${name}.jsonl`,
        recordings.map(r => JSON.stringify(r)).join("\n")
    );
}

// 录制各种场景
recordGolden("initialize-basic", async (agent) => {
    await agent.initialize({ protocolVersion: 1, clientCapabilities: {} });
});

recordGolden("prompt-simple", async (agent) => {
    await agent.initialize({ protocolVersion: 1, clientCapabilities: {} });
    const session = await agent.newSession({ cwd: "/tmp" });
    await agent.prompt({ sessionId: session.sessionId, prompt: [{ type: "text", text: "Hello" }] });
});

// ... 更多场景
```

#### 2. 验证器 (Go)

```go
// internal/golden/golden_test.go
package golden

import (
    "bufio"
    "encoding/json"
    "os"
    "testing"

    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
)

type GoldenRecord struct {
    Input  json.RawMessage `json:"input"`
    Output json.RawMessage `json:"output"`
}

func TestGolden_Initialize(t *testing.T) {
    records := loadGoldenFile(t, "initialize-basic.jsonl")

    agent := NewClaudeAcpAgent()

    for i, record := range records {
        t.Run(fmt.Sprintf("record_%d", i), func(t *testing.T) {
            var input InitializeRequest
            require.NoError(t, json.Unmarshal(record.Input, &input))

            // Go 实现处理相同输入
            output := agent.Initialize(input)

            // 必须完全一致
            gotJSON, _ := json.Marshal(output)
            assert.JSONEq(t, string(record.Output), string(gotJSON))
        })
    }
}

func loadGoldenFile(t *testing.T, name string) []GoldenRecord {
    file, err := os.Open(filepath.Join("golden", name))
    require.NoError(t, err)
    defer file.Close()

    var records []GoldenRecord
    scanner := bufio.NewScanner(file)
    for scanner.Scan() {
        var record GoldenRecord
        require.NoError(t, json.Unmarshal(scanner.Bytes(), &record))
        records = append(records, record)
    }

    return records
}
```

### 黄金文件列表

```
golden/
├── initialize-basic.jsonl
├── initialize-with-capabilities.jsonl
├── session-new.jsonl
├── session-load.jsonl
├── prompt-simple.jsonl
├── prompt-with-resource.jsonl
├── prompt-with-image.jsonl
├── tool-read.jsonl
├── tool-write.jsonl
├── tool-edit.jsonl
├── tool-bash.jsonl
├── permission-request.jsonl
├── mode-switch.jsonl
├── plan-update.jsonl
└── cancel-prompt.jsonl
```

---

## 策略二: 双向协议兼容性测试

### 概念

同时运行 TypeScript 和 Go 实现，使用相同的测试客户端，确保两者行为完全一致。

### 架构

```
┌──────────────────┐                    ┌──────────────────┐
│  TypeScript ACP  │  ◄── NDJSON ──►    │    Go ACP        │
│    (参考实现)     │                    │   (新实现)        │
└──────────────────┘                    └──────────────────┘
        │                                       │
        │         使用相同测试客户端             │
        └───────────────────────────────────────┘
                            │
                            ▼
                  ┌─────────────────┐
                  │   比较结果       │
                  │   必须完全一致   │
                  └─────────────────┘
```

### 实现

```go
// e2e/compat/compat_test.go
package compat

import (
    "context"
    "os/exec"
    "testing"
    "time"

    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
)

type AgentProcess struct {
    cmd    *exec.Cmd
    stdin  io.WriteCloser
    stdout io.Reader
}

func StartTypeScriptAgent(t *testing.T) *AgentProcess {
    cmd := exec.Command("npm", "run", "--silent", "dev")
    cmd.Dir = "../refer/claude-code-acp"
    stdin, _ := cmd.StdinPipe()
    stdout, _ := cmd.StdoutPipe()
    cmd.Start()

    return &AgentProcess{cmd: cmd, stdin: stdin, stdout: stdout}
}

func StartGoAgent(t *testing.T) *AgentProcess {
    cmd := exec.Command("go", "run", "./cmd/claude-code-acp")
    stdin, _ := cmd.StdinPipe()
    stdout, _ := cmd.StdoutPipe()
    cmd.Start()

    return &AgentProcess{cmd: cmd, stdin: stdin, stdout: stdout}
}

func (a *AgentProcess) Send(request interface{}) (map[string]interface{}, error) {
    data, _ := json.Marshal(request)
    a.stdin.Write(append(data, '\n'))

    scanner := bufio.NewScanner(a.stdout)
    scanner.Scan()

    var response map[string]interface{}
    json.Unmarshal(scanner.Bytes(), &response)
    return response, nil
}

func (a *AgentProcess) Stop() {
    a.cmd.Process.Kill()
}

func TestCompat_Initialize(t *testing.T) {
    tests := []struct {
        name    string
        request map[string]interface{}
    }{
        {
            name: "basic",
            request: map[string]interface{}{
                "jsonrpc": "2.0",
                "id":      0,
                "method":  "initialize",
                "params": map[string]interface{}{
                    "protocolVersion": 1,
                    "clientCapabilities": map[string]interface{}{
                        "fs": map[string]interface{}{
                            "readTextFile":  true,
                            "writeTextFile": true,
                        },
                    },
                },
            },
        },
        // ... 更多测试用例
    }

    for _, tc := range tests {
        t.Run(tc.name, func(t *testing.T) {
            // 启动两个实现
            tsAgent := StartTypeScriptAgent(t)
            defer tsAgent.Stop()

            goAgent := StartGoAgent(t)
            defer goAgent.Stop()

            // 等待启动
            time.Sleep(500 * time.Millisecond)

            // 发送相同请求
            tsResult, err := tsAgent.Send(tc.request)
            require.NoError(t, err)

            goResult, err := goAgent.Send(tc.request)
            require.NoError(t, err)

            // 关键: 两个实现必须产生相同的输出
            assert.Equal(t, tsResult["result"].(map[string]interface{})["protocolVersion"],
                         goResult["result"].(map[string]interface{})["protocolVersion"])

            // 深度比较能力
            assert.Equal(t,
                tsResult["result"].(map[string]interface{})["agentCapabilities"],
                goResult["result"].(map[string]interface{})["agentCapabilities"])
        })
    }
}

func TestCompat_FullSession(t *testing.T) {
    scenarios := []struct {
        name     string
        commands []map[string]interface{}
    }{
        {
            name: "simple conversation",
            commands: []map[string]interface{}{
                { /* initialize */ },
                { /* session/new */ },
                { /* session/prompt */ },
            },
        },
        {
            name: "tool usage",
            commands: []map[string]interface{}{
                { /* initialize */ },
                { /* session/new */ },
                { /* session/prompt with tool */ },
                { /* permission response */ },
            },
        },
    }

    for _, scenario := range scenarios {
        t.Run(scenario.name, func(t *testing.T) {
            tsAgent := StartTypeScriptAgent(t)
            defer tsAgent.Stop()

            goAgent := StartGoAgent(t)
            defer goAgent.Stop()

            time.Sleep(500 * time.Millisecond)

            for _, cmd := range scenario.commands {
                tsResult, _ := tsAgent.Send(cmd)
                goResult, _ := goAgent.Send(cmd)

                // 每个步骤都必须一致
                assertResultsEqual(t, tsResult, goResult)
            }
        })
    }
}
```

---

## 策略三: 类型级 1:1 映射

### 概念

从 ACP Schema 自动生成 Go 类型，确保与 TypeScript 类型完全匹配。

### 实现

#### 1. 使用 go-jsonschema

```bash
# 安装生成器
go install github.com/atombender/go-jsonschema/cmd/gojsonschema@latest

# 从 ACP Schema 生成
gojsonschema -p acp https://agentclientprotocol.com/protocol/schema.json -o internal/acp/types.gen.go
```

#### 2. 手动验证类型映射

```go
// internal/acp/types.go

// 从 ACP Schema 精确映射
// 确保每个字段都有正确的 JSON 标签

type InitializeRequest struct {
    ProtocolVersion    int                `json:"protocolVersion"`
    ClientCapabilities ClientCapabilities `json:"clientCapabilities"`
    ClientInfo         *Implementation    `json:"clientInfo,omitempty"`
    Meta               map[string]any     `json:"_meta,omitempty"`
}

// TypeScript (参考):
// interface InitializeRequest {
//   protocolVersion: ProtocolVersion;
//   clientCapabilities: ClientCapabilities;
//   clientInfo?: Implementation;
//   _meta?: { [key: string]: unknown };
// }

type InitializeResponse struct {
    ProtocolVersion    int               `json:"protocolVersion"`
    AgentCapabilities  AgentCapabilities `json:"agentCapabilities"`
    AgentInfo          *Implementation   `json:"agentInfo,omitempty"`
    AuthMethods        []AuthMethod      `json:"authMethods,omitempty"`
    Meta               map[string]any    `json:"_meta,omitempty"`
}

// 验证函数
func ValidateTypeMapping() error {
    // 使用 TypeScript 测试数据验证 Go 类型
    tsData := `{"protocolVersion":1,"clientCapabilities":{"fs":{"readTextFile":true}}}`

    var req InitializeRequest
    if err := json.Unmarshal([]byte(tsData), &req); err != nil {
        return fmt.Errorf("type mismatch: %w", err)
    }

    if req.ProtocolVersion != 1 {
        return fmt.Errorf("protocol version not parsed correctly")
    }

    if !req.ClientCapabilities.FS.ReadTextFile {
        return fmt.Errorf("capabilities not parsed correctly")
    }

    return nil
}
```

#### 3. 类型对比测试

```go
// internal/acp/types_test.go
func TestTypeMapping_InitializeRequest(t *testing.T) {
    // 从 TypeScript 测试中提取的测试用例
    testCases := []struct {
        name     string
        tsJSON   string
        expected InitializeRequest
    }{
        {
            name:   "basic",
            tsJSON: `{"protocolVersion":1,"clientCapabilities":{"fs":{"readTextFile":false,"writeTextFile":false},"terminal":false}}`,
            expected: InitializeRequest{
                ProtocolVersion: 1,
                ClientCapabilities: ClientCapabilities{
                    FS: FileSystemCapability{
                        ReadTextFile:  false,
                        WriteTextFile: false,
                    },
                    Terminal: false,
                },
            },
        },
    }

    for _, tc := range testCases {
        t.Run(tc.name, func(t *testing.T) {
            var got InitializeRequest
            require.NoError(t, json.Unmarshal([]byte(tc.tsJSON), &got))
            assert.Equal(t, tc.expected, got)
        })
    }
}
```

---

## 策略四: 测试优先级矩阵

| 优先级 | 测试类型 | 覆盖目标 | 测试数量 | 方法 | 验收标准 |
|--------|---------|----------|----------|------|----------|
| **P0** | 协议兼容性 | 100% | ~20 | Golden 文件 | 所有黄金文件通过 |
| **P0** | 工具转换 | 100% | ~30 | 单元测试 + TS 对比 | 与 TS 输出完全一致 |
| **P1** | 会话管理 | 100% | ~15 | 状态机测试 | 无状态泄漏 |
| **P1** | Prompt 处理 | 100% | ~20 | 集成测试 | 正确的 stop reasons |
| **P1** | 权限系统 | 100% | ~15 | 模拟客户端 | 正确的权限流程 |
| **P2** | 边界情况 | 80% | ~25 | 模糊测试 | 无崩溃 |
| **P2** | 性能 | 关键路径 | ~10 | 基准测试 | 响应时间 < 100ms |
| **P3** | E2E | 主流程 | ~5 | 真实客户端 | Zed 集成成功 |

---

## 策略五: 持续验证流水线

### GitHub Actions 配置

```yaml
# .github/workflows/compatibility.yml
name: Compatibility Check

on:
  push:
    branches: [main, develop]
  pull_request:
    branches: [main]

jobs:
  protocol-compat:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - name: Setup Node.js
        uses: actions/setup-node@v4
        with:
          node-version: '20'

      - name: Setup Go
        uses: actions/setup-go@v5
        with:
          go-version: '1.24'

      - name: Setup TypeScript Reference
        run: |
          cd refer/claude-code-acp
          npm install
          npm run build

      - name: Build Go Implementation
        run: |
          go mod download
          go build ./...

      - name: Run Golden File Tests
        run: |
          go test ./internal/golden/... -v

      - name: Run Compatibility Tests
        run: |
          go test ./e2e/compat/... -v -timeout 30m
        env:
          ANTHROPIC_API_KEY: ${{ secrets.ANTHROPIC_API_KEY }}

      - name: Run Protocol Conformance
        run: |
          go test ./e2e/conformance/... -v

      - name: Generate Coverage Report
        run: |
          go test -coverprofile=coverage.out ./...
          go tool cover -func=coverage.out

      - name: Upload Coverage
        uses: codecov/codecov-action@v4
        with:
          file: ./coverage.out

  zed-integration:
    runs-on: ubuntu-latest
    if: github.event_name == 'push' && github.ref == 'refs/heads/main'
    steps:
      - uses: actions/checkout@v4

      - name: Setup Go
        uses: actions/setup-go@v5
        with:
          go-version: '1.24'

      - name: Build Binary
        run: |
          go build -o claude-code-acp-go ./cmd/claude-code-acp

      - name: Run Zed Integration Tests
        run: |
          go test ./e2e/zed/... -v
        env:
          RUN_ZED_TESTS: "1"
          ZED_PATH: /usr/bin/zed

  race-detection:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - name: Setup Go
        uses: actions/setup-go@v5
        with:
          go-version: '1.24'

      - name: Run Race Detection
        run: |
          go test -race ./...
```

---

## 策略六: 关键实现检查点

### 检查点清单

```go
// internal/acp/checkpoints.go

// CP1: 消息顺序验证
func VerifyMessageOrder(t *testing.T, updates []SessionUpdate) {
    toolCalls := make(map[string]int) // toolCallID -> index of first appearance

    for i, update := range updates {
        switch u := update.(type) {
        case ToolCall:
            // tool_call 必须是第一次出现
            if _, exists := toolCalls[u.ToolCallID]; exists {
                t.Errorf("tool_call %s appeared more than once", u.ToolCallID)
            }
            toolCalls[u.ToolCallID] = i

        case ToolCallUpdate:
            // tool_call_update 必须在 tool_call 之后
            if firstIdx, exists := toolCalls[u.ToolCallID]; !exists {
                t.Errorf("tool_call_update for %s before tool_call", u.ToolCallID)
            } else if firstIdx > i {
                t.Errorf("tool_call_update at %d before tool_call at %d", i, firstIdx)
            }
        }
    }
}

// CP2: ID 关联验证
func VerifyToolCallIds(t *testing.T, call ToolCall, update ToolCallUpdate) {
    assert.Equal(t, call.ToolCallID, update.ToolCallID,
        "tool_call and tool_call_update must have matching ToolCallId")
}

// CP3: 能力合规验证
func VerifyCapabilityCompliance(t *testing.T, caps AgentCapabilities, behavior interface{}) {
    if caps.PromptCapabilities.Image {
        // 如果声明支持图片，必须能处理图片
        _, canHandle := behavior.(ImageHandler)
        assert.True(t, canHandle, "agent claims image support but doesn't implement ImageHandler")
    }

    if caps.MCPCapabilities.HTTP {
        // 如果声明支持 HTTP MCP，必须有 HTTP 传输
        _, hasHTTP := behavior.(HTTPTransporter)
        assert.True(t, hasHTTP, "agent claims HTTP MCP support but doesn't implement HTTPTransporter")
    }
}

// CP4: 状态转换验证
func VerifyStateTransitions(t *testing.T, transitions []ToolStatus) {
    validTransitions := map[ToolStatus][]ToolStatus{
        "pending":    {"in_progress", "cancelled"},
        "in_progress": {"completed", "failed", "cancelled"},
        "completed":  {}, // terminal
        "failed":     {}, // terminal
        "cancelled":  {}, // terminal
    }

    for i := 1; i < len(transitions); i++ {
        prev := transitions[i-1]
        curr := transitions[i]

        valid := false
        for _, allowed := range validTransitions[prev] {
            if allowed == curr {
                valid = true
                break
            }
        }

        if !valid {
            t.Errorf("invalid state transition: %s -> %s", prev, curr)
        }
    }
}

// CP5: 会话隔离验证
func VerifySessionIsolation(t *testing.T, sessions map[string]*Session) {
    for id1, s1 := range sessions {
        for id2, s2 := range sessions {
            if id1 == id2 {
                continue
            }

            // 会话必须完全隔离
            assert.NotEqual(t, s1.Query, s2.Query, "sessions share query object")
            assert.NotEqual(t, s1.Input, s2.Input, "sessions share input stream")
        }
    }
}
```

---

## 验收标准

### 必须通过的测试

| 测试类别 | 标准 | 验证方法 |
|----------|------|----------|
| 黄金文件测试 | 100% 通过 | `go test ./internal/golden/...` |
| 兼容性测试 | 100% 通过 | `go test ./e2e/compat/...` |
| 协议合规 | 100% 通过 | `go test ./e2e/conformance/...` |
| 单元测试 | 100% 通过 | `go test ./...` |
| 竞态检测 | 0 问题 | `go test -race ./...` |
| 覆盖率 | >= 80% | `go test -cover ./...` |

### 功能验收清单

```markdown
- [ ] 初始化握手与 TS 版本输出一致
- [ ] 会话创建返回有效的 sessionId
- [ ] 会话加载正确重放历史
- [ ] 纯文本 prompt 正确处理
- [ ] 带资源 prompt 正确处理
- [ ] 带图片 prompt 正确处理
- [ ] Read 工具正确转换
- [ ] Write 工具正确转换
- [ ] Edit 工具正确转换
- [ ] Bash 工具正确转换
- [ ] 权限请求包含正确结构
- [ ] acceptEdits 模式正确工作
- [ ] bypassPermissions 模式正确工作
- [ ] Plan 更新正确发送
- [ ] 取消操作正确响应
- [ ] 斜杠命令正确处理
- [ ] 与 Zed 编辑器集成成功
```

---

## 总结

```
┌─────────────────────────────────────────────────────────────┐
│                  100% 正确性保证策略                         │
├─────────────────────────────────────────────────────────────┤
│                                                             │
│  1. 黄金文件测试 - 录制/回放真实请求响应                     │
│  2. 双向对比测试 - 与 TypeScript 并行运行                   │
│  3. 类型精确映射 - 从 Schema 生成，JSON 兼容                │
│  4. 优先级矩阵 - 重点关注协议兼容性                         │
│  5. CI/CD 流水线 - 每次提交验证                             │
│  6. 检查点验证 - 关键路径运行时验证                         │
│                                                             │
│  结果: 与 TypeScript 版本行为完全一致的 Go 实现              │
│                                                             │
└─────────────────────────────────────────────────────────────┘
```

---

*文档生成日期: 2026-02-13*