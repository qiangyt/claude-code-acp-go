# 测试验证集设计

> 本文档定义了确保 Go 实现与 ACP 协议和原始 TypeScript 实现完全兼容的完整测试套件。

## 测试架构概览

```
┌─────────────────────────────────────────────────────────────────┐
│                        测试金字塔                                │
├─────────────────────────────────────────────────────────────────┤
│                                                                 │
│                        ┌─────────┐                              │
│                        │   E2E   │  第九层: 端到端集成           │
│                      ┌─┴─────────┴─┐                            │
│                      │   设置管理   │  第八层: 设置               │
│                    ┌─┴─────────────┴─┐                          │
│                    │    MCP 集成      │  第七层: MCP             │
│                  ┌─┴─────────────────┴─┐                        │
│                  │      模式切换        │  第六层: 模式           │
│                ┌─┴─────────────────────┴─┐                      │
│                │         Plan            │  第五层: Plan         │
│              ┌─┴─────────────────────────┴─┐                    │
│              │          工具调用            │  第四层: 工具       │
│            ┌─┴─────────────────────────────┴─┐                  │
│            │           Prompt 处理           │  第三层: Prompt   │
│          ┌─┴─────────────────────────────────┴─┐                │
│          │             会话管理                │  第二层: 会话   │
│        ┌─┴─────────────────────────────────────┴─┐              │
│        │               协议层                     │  第一层: 协议 │
│        └─────────────────────────────────────────┘              │
│                                                                 │
└─────────────────────────────────────────────────────────────────┘
```

---

## 第一层: 协议层测试

**文件**: `internal/acp/protocol_test.go`

### T1.1 初始化握手

```go
func TestInitialize_Handshake(t *testing.T) {
    tests := []struct {
        name     string
        request  InitializeRequest
        expected InitializeResponse
    }{
        {
            name: "basic initialization",
            request: InitializeRequest{
                ProtocolVersion: 1,
                ClientCapabilities: ClientCapabilities{
                    FS: FileSystemCapability{
                        ReadTextFile:  true,
                        WriteTextFile: true,
                    },
                    Terminal: true,
                },
            },
            expected: InitializeResponse{
                ProtocolVersion: 1,
                AgentCapabilities: AgentCapabilities{
                    LoadSession: true,
                    PromptCapabilities: PromptCapabilities{
                        Image:            true,
                        Audio:            true,
                        EmbeddedContext:  true,
                    },
                    MCPCapabilities: MCPCapabilities{
                        HTTP: true,
                        SSE:  true,
                    },
                },
            },
        },
    }

    for _, tc := range tests {
        t.Run(tc.name, func(t *testing.T) {
            agent := NewClaudeAcpAgent()
            result := agent.Initialize(tc.request)

            assert.Equal(t, tc.expected.ProtocolVersion, result.ProtocolVersion)
            assert.Equal(t, tc.expected.AgentCapabilities.LoadSession, result.AgentCapabilities.LoadSession)
        })
    }
}
```

### T1.2 版本协商

```go
func TestInitialize_VersionNegotiation(t *testing.T) {
    tests := []struct {
        name            string
        clientVersion   int
        expectedVersion int
    }{
        {"client supports latest", 1, 1},
        {"client supports older", 0, 1}, // Agent 返回其支持的最高版本
    }

    for _, tc := range tests {
        t.Run(tc.name, func(t *testing.T) {
            agent := NewClaudeAcpAgent()
            result := agent.Initialize(InitializeRequest{
                ProtocolVersion: tc.clientVersion,
            })

            assert.Equal(t, tc.expectedVersion, result.ProtocolVersion)
        })
    }
}
```

### T1.3 能力交换

```go
func TestInitialize_CapabilitiesExchange(t *testing.T) {
    agent := NewClaudeAcpAgent()

    // 测试所有能力正确传递
    result := agent.Initialize(InitializeRequest{
        ClientCapabilities: ClientCapabilities{
            FS: FileSystemCapability{
                ReadTextFile:  true,
                WriteTextFile: true,
            },
            Terminal: true,
        },
    })

    // 验证代理能力
    assert.True(t, result.AgentCapabilities.LoadSession)
    assert.NotNil(t, result.AgentCapabilities.PromptCapabilities)
    assert.NotNil(t, result.AgentCapabilities.MCPCapabilities)
}
```

### T1.4 NDJSON 编解码

```go
func TestNDJSON_EncodeDecode(t *testing.T) {
    tests := []struct {
        name     string
        input    string
        expected []map[string]interface{}
        hasError bool
    }{
        {
            name:  "single line",
            input: `{"jsonrpc":"2.0","id":0,"method":"initialize","params":{}}`,
            expected: []map[string]interface{}{
                {"jsonrpc": "2.0", "id": float64(0), "method": "initialize", "params": map[string]interface{}{}},
            },
        },
        {
            name: "multiple lines",
            input: `{"jsonrpc":"2.0","id":0,"method":"initialize"}
{"jsonrpc":"2.0","id":1,"method":"session/new"}`,
            expected: []map[string]interface{}{
                {"jsonrpc": "2.0", "id": float64(0), "method": "initialize"},
                {"jsonrpc": "2.0", "id": float64(1), "method": "session/new"},
            },
        },
        {
            name:  "empty lines ignored",
            input: `{"jsonrpc":"2.0","id":0}\n\n{"jsonrpc":"2.0","id":1}`,
            expected: []map[string]interface{}{
                {"jsonrpc": "2.0", "id": float64(0)},
                {"jsonrpc": "2.0", "id": float64(1)},
            },
        },
        {
            name:     "invalid json",
            input:    `{invalid}`,
            hasError: true,
        },
    }

    for _, tc := range tests {
        t.Run(tc.name, func(t *testing.T) {
            reader := strings.NewReader(tc.input)
            decoder := NewNDJSONDecoder(reader)

            var results []map[string]interface{}
            for {
                var msg map[string]interface{}
                err := decoder.Decode(&msg)
                if err == io.EOF {
                    break
                }
                if err != nil {
                    if tc.hasError {
                        return // 预期错误
                    }
                    t.Fatalf("unexpected error: %v", err)
                }
                results = append(results, msg)
            }

            assert.Equal(t, len(tc.expected), len(results))
        })
    }
}
```

---

## 第二层: 会话管理测试

**文件**: `internal/acp/session_test.go`

### T2.1 创建会话

```go
func TestSession_New(t *testing.T) {
    agent := NewClaudeAcpAgent()
    agent.Initialize(defaultInitializeRequest())

    cwd := "/home/user/project"
    result, err := agent.NewSession(NewSessionRequest{
        Cwd:        cwd,
        MCPServers: []MCPServer{},
    })

    require.NoError(t, err)
    assert.NotEmpty(t, result.SessionID)
    assert.NotNil(t, result.Modes)
    assert.NotNil(t, result.Models)
}
```

### T2.2 加载会话

```go
func TestSession_Load(t *testing.T) {
    agent := NewClaudeAcpAgent()
    agent.Initialize(defaultInitializeRequest())

    // 先创建一个会话
    createResult, _ := agent.NewSession(NewSessionRequest{
        Cwd: "/home/user/project",
    })

    // 发送一些消息
    agent.Prompt(PromptRequest{
        SessionID: createResult.SessionID,
        Prompt: []ContentBlock{
            {Type: "text", Text: "Hello"},
        },
    })

    // 加载会话
    var updates []SessionUpdate
    result, err := agent.LoadSession(LoadSessionRequest{
        SessionID:  createResult.SessionID,
        Cwd:        "/home/user/project",
    }, func(update SessionUpdate) {
        updates = append(updates, update)
    })

    require.NoError(t, err)
    assert.NotEmpty(t, updates) // 应该重放对话历史
}
```

### T2.3 列出会话

```go
func TestSession_List(t *testing.T) {
    agent := NewClaudeAcpAgent()
    agent.Initialize(defaultInitializeRequest())

    // 创建多个会话
    for i := 0; i < 3; i++ {
        agent.NewSession(NewSessionRequest{
            Cwd: fmt.Sprintf("/home/user/project%d", i),
        })
    }

    result, err := agent.ListSessions(ListSessionsRequest{})

    require.NoError(t, err)
    assert.GreaterOrEqual(t, len(result.Sessions), 3)
}
```

### T2.4 分支会话

```go
func TestSession_Fork(t *testing.T) {
    agent := NewClaudeAcpAgent()
    agent.Initialize(defaultInitializeRequest())

    // 创建原会话
    original, _ := agent.NewSession(NewSessionRequest{
        Cwd: "/home/user/project",
    })

    // 分支
    forked, err := agent.ForkSession(ForkSessionRequest{
        SessionID: original.SessionID,
    })

    require.NoError(t, err)
    assert.NotEqual(t, original.SessionID, forked.SessionID)
    // 分支的会话应该有相同的历史
}
```

### T2.5 并发会话

```go
func TestSession_ConcurrentAccess(t *testing.T) {
    agent := NewClaudeAcpAgent()
    agent.Initialize(defaultInitializeRequest())

    var wg sync.WaitGroup
    errors := make(chan error, 10)

    for i := 0; i < 10; i++ {
        wg.Add(1)
        go func(id int) {
            defer wg.Done()
            _, err := agent.NewSession(NewSessionRequest{
                Cwd: fmt.Sprintf("/home/user/project%d", id),
            })
            if err != nil {
                errors <- err
            }
        }(i)
    }

    wg.Wait()
    close(errors)

    for err := range errors {
        t.Errorf("concurrent session creation failed: %v", err)
    }
}
```

---

## 第三层: Prompt 处理测试

**文件**: `internal/acp/prompt_test.go`

### T3.1 纯文本 Prompt

```go
func TestPrompt_TextOnly(t *testing.T) {
    agent := setupAgentWithSession(t)

    var messageChunks []string
    result, err := agent.Prompt(PromptRequest{
        SessionID: agent.DefaultSessionID(),
        Prompt: []ContentBlock{
            {Type: "text", Text: "What is 2 + 2?"},
        },
    }, func(notification SessionNotification) {
        if notification.Update.SessionUpdate == "agent_message_chunk" {
            if text, ok := notification.Update.Content.(TextContent); ok {
                messageChunks = append(messageChunks, text.Text)
            }
        }
    })

    require.NoError(t, err)
    assert.Equal(t, "end_turn", result.StopReason)
    assert.NotEmpty(t, messageChunks)
}
```

### T3.2 带 Context 的 Prompt

```go
func TestPrompt_WithResource(t *testing.T) {
    agent := setupAgentWithSession(t)

    result, err := agent.Prompt(PromptRequest{
        SessionID: agent.DefaultSessionID(),
        Prompt: []ContentBlock{
            {Type: "text", Text: "Analyze this code:"},
            {
                Type: "resource",
                Resource: EmbeddedResource{
                    URI:      "file:///home/user/main.go",
                    MimeType: "text/x-go",
                    Text:     "package main\n\nfunc main() {}",
                },
            },
        },
    })

    require.NoError(t, err)
    assert.Equal(t, "end_turn", result.StopReason)
}
```

### T3.3 带 Image 的 Prompt

```go
func TestPrompt_WithImage(t *testing.T) {
    agent := setupAgentWithSession(t)

    // 读取测试图片
    imageData, _ := os.ReadFile("testdata/test.png")

    result, err := agent.Prompt(PromptRequest{
        SessionID: agent.DefaultSessionID(),
        Prompt: []ContentBlock{
            {Type: "text", Text: "What's in this image?"},
            {
                Type: "image",
                Image: ImageContent{
                    Source: ImageSource{
                        Type: "base64",
                        MediaType: "image/png",
                        Data: base64.StdEncoding.EncodeToString(imageData),
                    },
                },
            },
        },
    })

    require.NoError(t, err)
    assert.Equal(t, "end_turn", result.StopReason)
}
```

### T3.4 Stop Reasons

```go
func TestPrompt_StopReasons(t *testing.T) {
    tests := []struct {
        name           string
        setup          func(*ClaudeAcpAgent)
        expectedReason string
    }{
        {
            name:           "end_turn",
            setup:          func(a *ClaudeAcpAgent) {},
            expectedReason: "end_turn",
        },
        {
            name: "cancelled",
            setup: func(a *ClaudeAcpAgent) {
                go func() {
                    time.Sleep(100 * time.Millisecond)
                    a.Cancel(CancelNotification{
                        SessionID: a.DefaultSessionID(),
                    })
                }()
            },
            expectedReason: "cancelled",
        },
    }

    for _, tc := range tests {
        t.Run(tc.name, func(t *testing.T) {
            agent := setupAgentWithSession(t)
            tc.setup(agent)

            result, _ := agent.Prompt(PromptRequest{
                SessionID: agent.DefaultSessionID(),
                Prompt:    []ContentBlock{{Type: "text", Text: "Hello"}},
            })

            assert.Equal(t, tc.expectedReason, result.StopReason)
        })
    }
}
```

### T3.5 取消 Prompt

```go
func TestPrompt_Cancel(t *testing.T) {
    agent := setupAgentWithSession(t)

    // 启动长时间运行的 prompt
    done := make(chan struct{})
    go func() {
        defer close(done)
        agent.Prompt(PromptRequest{
            SessionID: agent.DefaultSessionID(),
            Prompt:    []ContentBlock{{Type: "text", Text: "Count to 1000000"}},
        })
    }()

    // 等待一小段时间后取消
    time.Sleep(100 * time.Millisecond)
    agent.Cancel(CancelNotification{
        SessionID: agent.DefaultSessionID(),
    })

    <-done
    // 验证返回了 cancelled stop reason
}
```

---

## 第四层: 工具调用测试

**文件**: `internal/acp/tools_test.go`

### T4.1 Read 工具

```go
func TestTool_Read(t *testing.T) {
    agent := setupAgentWithSession(t)
    mockClient := NewMockClient()
    mockClient.Files["/test/file.txt"] = "Hello, World!"

    var toolCalls []ToolCall
    var toolUpdates []ToolCallUpdate

    agent.Prompt(PromptRequest{
        SessionID: agent.DefaultSessionID(),
        Prompt:    []ContentBlock{{Type: "text", Text: "Read /test/file.txt"}},
    }, func(notification SessionNotification) {
        switch notification.Update.SessionUpdate {
        case "tool_call":
            toolCalls = append(toolCalls, notification.Update.(ToolCall))
        case "tool_call_update":
            toolUpdates = append(toolUpdates, notification.Update.(ToolCallUpdate))
        }
    })

    // 验证 tool_call 在 tool_call_update 之前
    require.GreaterOrEqual(t, len(toolCalls), 1)
    require.GreaterOrEqual(t, len(toolUpdates), 1)

    // 验证 locations 正确设置
    readCall := toolCalls[0]
    assert.Equal(t, "read", readCall.Kind)
    assert.NotEmpty(t, readCall.Locations)
    assert.Equal(t, "/test/file.txt", readCall.Locations[0].Path)
}
```

### T4.2 Write 工具

```go
func TestTool_Write(t *testing.T) {
    agent := setupAgentWithSession(t)
    mockClient := NewMockClient()

    agent.Prompt(PromptRequest{
        SessionID: agent.DefaultSessionID(),
        Prompt:    []ContentBlock{{Type: "text", Text: "Write 'test content' to /test/new.txt"}},
    })

    // 验证文件被写入
    assert.Equal(t, "test content", mockClient.Files["/test/new.txt"])
}
```

### T4.3 Edit 工具

```go
func TestTool_Edit(t *testing.T) {
    tests := []struct {
        name          string
        initial       string
        oldString     string
        newString     string
        expected      string
        expectError   bool
    }{
        {
            name:        "successful edit",
            initial:     "Hello World",
            oldString:   "World",
            newString:   "Go",
            expected:    "Hello Go",
            expectError: false,
        },
        {
            name:        "old_string not found",
            initial:     "Hello World",
            oldString:   "NotPresent",
            newString:   "Go",
            expectError: true,
        },
    }

    for _, tc := range tests {
        t.Run(tc.name, func(t *testing.T) {
            agent := setupAgentWithSession(t)
            mockClient := NewMockClient()
            mockClient.Files["/test/file.txt"] = tc.initial

            err := agent.EditFile("/test/file.txt", tc.oldString, tc.newString)

            if tc.expectError {
                assert.Error(t, err)
            } else {
                assert.NoError(t, err)
                assert.Equal(t, tc.expected, mockClient.Files["/test/file.txt"])
            }
        })
    }
}
```

### T4.4 Bash 工具

```go
func TestTool_Bash(t *testing.T) {
    agent := setupAgentWithSession(t)
    mockClient := NewMockClient()

    var terminalCalls []TerminalCall
    mockClient.OnTerminalCreate = func(req CreateTerminalRequest) {
        terminalCalls = append(terminalCalls, TerminalCall{
            Command: req.Command,
            Args:    req.Args,
            Cwd:     req.Cwd,
        })
    }

    agent.Prompt(PromptRequest{
        SessionID: agent.DefaultSessionID(),
        Prompt:    []ContentBlock{{Type: "text", Text: "Run 'ls -la'"}},
    })

    require.GreaterOrEqual(t, len(terminalCalls), 1)
    assert.Equal(t, "ls", terminalCalls[0].Command)
    assert.Contains(t, terminalCalls[0].Args, "-la")
}
```

### T4.5 工具权限请求

```go
func TestTool_PermissionRequest(t *testing.T) {
    agent := setupAgentWithSession(t)
    mockClient := NewMockClient()

    var permissionRequests []RequestPermissionRequest
    mockClient.OnRequestPermission = func(req RequestPermissionRequest) RequestPermissionResponse {
        permissionRequests = append(permissionRequests, req)
        return RequestPermissionResponse{
            Outcome: SelectedPermissionOutcome{
                OptionID: "allow_once",
            },
        }
    }

    agent.Prompt(PromptRequest{
        SessionID: agent.DefaultSessionID(),
        Prompt:    []ContentBlock{{Type: "text", Text: "Delete all files"}},
    })

    // 验证请求了权限
    require.GreaterOrEqual(t, len(permissionRequests), 1)

    // 验证结构包含 title
    req := permissionRequests[0]
    assert.NotEmpty(t, req.ToolCall.Title)
    assert.NotEmpty(t, req.ToolCall.ToolCallID)
    assert.NotNil(t, req.ToolCall.RawInput)

    // 验证选项
    optionKinds := make(map[string]bool)
    for _, opt := range req.Options {
        optionKinds[opt.Kind] = true
    }
    assert.True(t, optionKinds["allow_once"])
    assert.True(t, optionKinds["allow_always"])
    assert.True(t, optionKinds["deny_once"])
    assert.True(t, optionKinds["deny_always"])
}
```

### T4.6 后台进程

```go
func TestTool_BackgroundBash(t *testing.T) {
    agent := setupAgentWithSession(t)
    mockClient := NewMockClient()

    // 启动后台进程
    terminalID, _ := agent.CreateTerminal(CreateTerminalRequest{
        SessionID: agent.DefaultSessionID(),
        Command:   "sleep",
        Args:      []string{"100"},
    })

    // 获取输出
    output, _ := agent.TerminalOutput(TerminalOutputRequest{
        SessionID:  agent.DefaultSessionID(),
        TerminalID: terminalID,
    })
    assert.Empty(t, output.Output) // 还没有输出

    // 终止进程
    agent.KillTerminal(KillTerminalCommandRequest{
        SessionID:  agent.DefaultSessionID(),
        TerminalID: terminalID,
    })

    // 释放终端
    agent.ReleaseTerminal(ReleaseTerminalRequest{
        SessionID:  agent.DefaultSessionID(),
        TerminalID: terminalID,
    })
}
```

---

## 第五层: Plan 测试

**文件**: `internal/acp/plan_test.go`

### T5.1 Plan 创建

```go
func TestPlan_Creation(t *testing.T) {
    agent := setupAgentWithSession(t)

    var planUpdates []Plan
    agent.Prompt(PromptRequest{
        SessionID: agent.DefaultSessionID(),
        Prompt:    []ContentBlock{{Type: "text", Text: "Create a plan to refactor this code"}},
    }, func(notification SessionNotification) {
        if notification.Update.SessionUpdate == "plan" {
            planUpdates = append(planUpdates, notification.Update.(Plan))
        }
    })

    require.GreaterOrEqual(t, len(planUpdates), 1)

    // 验证 priority 映射
    plan := planUpdates[0]
    for _, entry := range plan.Entries {
        assert.Contains(t, []string{"high", "medium", "low"}, entry.Priority)
        assert.Contains(t, []string{"pending", "in_progress", "completed"}, entry.Status)
    }
}
```

### T5.2 Plan 状态更新

```go
func TestPlan_StatusUpdate(t *testing.T) {
    agent := setupAgentWithSession(t)

    var planUpdates []Plan
    agent.Prompt(PromptRequest{
        SessionID: agent.DefaultSessionID(),
        Prompt:    []ContentBlock{{Type: "text", Text: "Implement a feature with multiple steps"}},
    }, func(notification SessionNotification) {
        if notification.Update.SessionUpdate == "plan" {
            planUpdates = append(planUpdates, notification.Update.(Plan))
        }
    })

    // 验证状态转换: pending -> in_progress -> completed
    if len(planUpdates) >= 3 {
        // 检查第一个条目的状态变化
        firstEntry := planUpdates[0].Entries[0]
        lastEntry := planUpdates[len(planUpdates)-1].Entries[0]

        // 初始应该是 pending 或 in_progress
        assert.Contains(t, []string{"pending", "in_progress"}, firstEntry.Status)

        // 最终应该是 completed
        assert.Equal(t, "completed", lastEntry.Status)
    }
}
```

---

## 第六层: 模式切换测试

**文件**: `internal/acp/mode_test.go`

### T6.1 默认模式

```go
func TestMode_Default(t *testing.T) {
    agent := setupAgentWithSession(t)

    // 默认模式应该要求权限
    mockClient := NewMockClient()
    permissionRequested := false
    mockClient.OnRequestPermission = func(req RequestPermissionRequest) RequestPermissionResponse {
        permissionRequested = true
        return RequestPermissionResponse{
            Outcome: SelectedPermissionOutcome{OptionID: "allow_once"},
        }
    }

    agent.Prompt(PromptRequest{
        SessionID: agent.DefaultSessionID(),
        Prompt:    []ContentBlock{{Type: "text", Text: "Write a file"}},
    })

    assert.True(t, permissionRequested)
}
```

### T6.2 acceptEdits 模式

```go
func TestMode_AcceptEdits(t *testing.T) {
    agent := setupAgentWithSession(t)

    // 设置为 acceptEdits 模式
    agent.SetSessionMode(SetSessionModeRequest{
        SessionID: agent.DefaultSessionID(),
        ModeID:    "acceptEdits",
    })

    mockClient := NewMockClient()
    permissionRequested := false
    mockClient.OnRequestPermission = func(req RequestPermissionRequest) RequestPermissionResponse {
        permissionRequested = true
        return RequestPermissionResponse{
            Outcome: SelectedPermissionOutcome{OptionID: "allow_once"},
        }
    }

    agent.Prompt(PromptRequest{
        SessionID: agent.DefaultSessionID(),
        Prompt:    []ContentBlock{{Type: "text", Text: "Write a file"}},
    })

    // 在 acceptEdits 模式下，编辑工具应该自动允许
    assert.False(t, permissionRequested)
}
```

### T6.3 bypassPermissions 模式

```go
func TestMode_BypassPermissions(t *testing.T) {
    agent := setupAgentWithSession(t)

    // 设置为 bypassPermissions 模式
    agent.SetSessionMode(SetSessionModeRequest{
        SessionID: agent.DefaultSessionID(),
        ModeID:    "bypassPermissions",
    })

    mockClient := NewMockClient()
    permissionRequested := false
    mockClient.OnRequestPermission = func(req RequestPermissionRequest) RequestPermissionResponse {
        permissionRequested = true
        return RequestPermissionResponse{
            Outcome: SelectedPermissionOutcome{OptionID: "allow_once"},
        }
    }

    agent.Prompt(PromptRequest{
        SessionID: agent.DefaultSessionID(),
        Prompt:    []ContentBlock{{Type: "text", Text: "Do anything"}},
    })

    // 在 bypassPermissions 模式下，所有工具应该自动允许
    assert.False(t, permissionRequested)
}
```

### T6.4 模式切换通知

```go
func TestMode_SwitchNotification(t *testing.T) {
    agent := setupAgentWithSession(t)

    var modeUpdates []CurrentModeUpdate
    agent.SetSessionMode(SetSessionModeRequest{
        SessionID: agent.DefaultSessionID(),
        ModeID:    "bypassPermissions",
    }, func(notification SessionNotification) {
        if notification.Update.SessionUpdate == "current_mode_update" {
            modeUpdates = append(modeUpdates, notification.Update.(CurrentModeUpdate))
        }
    })

    require.GreaterOrEqual(t, len(modeUpdates), 1)
    assert.Equal(t, "bypassPermissions", modeUpdates[0].CurrentModeID)
}
```

---

## 第七层: MCP 集成测试

**文件**: `internal/mcp/server_test.go`

### T7.1 Stdio MCP 服务器

```go
func TestMCP_StdioTransport(t *testing.T) {
    server := NewACPMCPServer()

    // 使用 stdio 传输
    transport := &StdioTransport{}
    err := server.Start(transport)

    require.NoError(t, err)
    defer server.Stop()

    // 测试工具调用
    result, err := server.CallTool("Read", map[string]interface{}{
        "file_path": "/test/file.txt",
    })

    require.NoError(t, err)
    assert.NotNil(t, result)
}
```

### T7.2 HTTP MCP 服务器

```go
func TestMCP_HttpTransport(t *testing.T) {
    server := NewACPMCPServer()

    // 使用 HTTP 传输
    transport := &HTTPTransport{
        URL: "http://localhost:8080/mcp",
    }
    err := server.Start(transport)

    require.NoError(t, err)
    defer server.Stop()

    // 测试连接
    assert.True(t, server.IsConnected())
}
```

### T7.3 SSE MCP 服务器

```go
func TestMCP_SseTransport(t *testing.T) {
    server := NewACPMCPServer()

    // 使用 SSE 传输
    transport := &SSETransport{
        URL: "http://localhost:8080/sse",
    }
    err := server.Start(transport)

    require.NoError(t, err)
    defer server.Stop()
}
```

### T7.4 ACP 工具映射

```go
func TestMCP_AcpToolsMapping(t *testing.T) {
    server := NewACPMCPServer()

    // 验证所有 ACP 工具正确注册
    expectedTools := []string{
        "mcp__acp__Read",
        "mcp__acp__Write",
        "mcp__acp__Edit",
        "mcp__acp__Bash",
        "mcp__acp__BashOutput",
        "mcp__acp__KillShell",
    }

    tools := server.ListTools()
    for _, expected := range expectedTools {
        found := false
        for _, tool := range tools {
            if tool.Name == expected {
                found = true
                break
            }
        }
        assert.True(t, found, "Tool %s not found", expected)
    }
}
```

---

## 第八层: 设置管理测试

**文件**: `internal/settings/manager_test.go`

### T8.1 设置合并优先级

```go
func TestSettings_MergePrecedence(t *testing.T) {
    tests := []struct {
        name     string
        sources  []SettingSource
        expected PermissionSettings
    }{
        {
            name: "project overrides user",
            sources: []SettingSource{
                {Type: "user", Settings: ClaudeCodeSettings{
                    Permissions: &PermissionSettings{Allow: []string{"Read"}},
                }},
                {Type: "project", Settings: ClaudeCodeSettings{
                    Permissions: &PermissionSettings{Allow: []string{"Read", "Write"}},
                }},
            },
            expected: PermissionSettings{Allow: []string{"Read", "Write"}},
        },
        {
            name: "local overrides project",
            sources: []SettingSource{
                {Type: "project", Settings: ClaudeCodeSettings{
                    Permissions: &PermissionSettings{Allow: []string{"Read"}},
                }},
                {Type: "local", Settings: ClaudeCodeSettings{
                    Permissions: &PermissionSettings{Allow: []string{"Read", "Bash"}},
                }},
            },
            expected: PermissionSettings{Allow: []string{"Read", "Bash"}},
        },
    }

    for _, tc := range tests {
        t.Run(tc.name, func(t *testing.T) {
            manager := NewSettingsManager("/home/user/project")
            for _, source := range tc.sources {
                manager.Load(source.Type, source.Settings)
            }

            result := manager.GetPermissions()
            assert.Equal(t, tc.expected.Allow, result.Allow)
        })
    }
}
```

### T8.2 权限规则匹配

```go
func TestSettings_PermissionRules(t *testing.T) {
    tests := []struct {
        name        string
        rule        string
        toolName    string
        toolInput   map[string]interface{}
        shouldMatch bool
    }{
        {
            name:        "match all reads",
            rule:        "Read",
            toolName:    "Read",
            toolInput:   map[string]interface{}{"file_path": "/any/file.txt"},
            shouldMatch: true,
        },
        {
            name:        "match specific file",
            rule:        "Read(./.env)",
            toolName:    "Read",
            toolInput:   map[string]interface{}{"file_path": "/home/user/project/.env"},
            shouldMatch: true,
        },
        {
            name:        "match glob pattern",
            rule:        "Read(./secrets/**)",
            toolName:    "Read",
            toolInput:   map[string]interface{}{"file_path": "/home/user/project/secrets/api_key.txt"},
            shouldMatch: true,
        },
        {
            name:        "match command prefix",
            rule:        "Bash(npm run:*)",
            toolName:    "Bash",
            toolInput:   map[string]interface{}{"command": "npm run test"},
            shouldMatch: true,
        },
        {
            name:        "no match different command",
            rule:        "Bash(npm run:*)",
            toolName:    "Bash",
            toolInput:   map[string]interface{}{"command": "rm -rf /"},
            shouldMatch: false,
        },
    }

    for _, tc := range tests {
        t.Run(tc.name, func(t *testing.T) {
            manager := NewSettingsManager("/home/user/project")
            manager.AddPermissionRule(tc.rule)

            matches := manager.MatchesRule(tc.toolName, tc.toolInput)
            assert.Equal(t, tc.shouldMatch, matches)
        })
    }
}
```

### T8.3 环境变量传递

```go
func TestSettings_EnvironmentVariables(t *testing.T) {
    manager := NewSettingsManager("/home/user/project")
    manager.Load("project", ClaudeCodeSettings{
        Env: map[string]string{
            "API_KEY":    "secret123",
            "DEBUG_MODE": "true",
        },
    })

    env := manager.GetEnvironment()

    assert.Equal(t, "secret123", env["API_KEY"])
    assert.Equal(t, "true", env["DEBUG_MODE"])
}
```

---

## 第九层: 端到端集成测试

**文件**: `e2e/integration_test.go`

### T9.1 完整对话流程

```go
func TestE2E_FullConversation(t *testing.T) {
    if testing.Short() {
        t.Skip("Skipping E2E test in short mode")
    }

    // 启动 Go ACP 代理
    agent := StartAgent(t)
    defer agent.Stop()

    // 初始化
    initResult := agent.Initialize(InitializeRequest{
        ProtocolVersion: 1,
        ClientCapabilities: ClientCapabilities{
            FS: FileSystemCapability{
                ReadTextFile:  true,
                WriteTextFile: true,
            },
            Terminal: true,
        },
    })
    assert.Equal(t, 1, initResult.ProtocolVersion)

    // 创建会话
    sessionResult := agent.NewSession(NewSessionRequest{
        Cwd: t.TempDir(),
    })
    assert.NotEmpty(t, sessionResult.SessionID)

    // 发送 prompt
    var responses []string
    promptResult := agent.Prompt(PromptRequest{
        SessionID: sessionResult.SessionID,
        Prompt:    []ContentBlock{{Type: "text", Text: "Hello, Claude!"}},
    }, func(notification SessionNotification) {
        if notification.Update.SessionUpdate == "agent_message_chunk" {
            if text, ok := notification.Update.Content.(TextContent); ok {
                responses = append(responses, text.Text)
            }
        }
    })

    assert.Equal(t, "end_turn", promptResult.StopReason)
    assert.NotEmpty(t, responses)
}
```

### T9.2 与 Zed 编辑器集成

```go
func TestE2E_ZedCompatibility(t *testing.T) {
    if os.Getenv("RUN_ZED_TESTS") == "" {
        t.Skip("Set RUN_ZED_TESTS=1 to run Zed compatibility tests")
    }

    // 使用模拟的 Zed 客户端
    zedClient := NewMockZedClient()

    // 连接到 Go ACP 代理
    err := zedClient.Connect("claude-code-acp-go")
    require.NoError(t, err)
    defer zedClient.Disconnect()

    // 执行完整的编辑器工作流
    // 1. 打开项目
    session := zedClient.OpenProject("/home/user/project")

    // 2. 发送请求
    response := session.Prompt("Refactor this function")

    // 3. 验证响应格式
    assert.NotEmpty(t, response.Text)
    assert.NotNil(t, response.ToolCalls)

    // 4. 批准工具
    for _, call := range response.ToolCalls {
        session.ApproveTool(call.ToolCallID)
    }

    // 5. 验证完成
    assert.Equal(t, "end_turn", response.StopReason)
}
```

### T9.3 斜杠命令

```go
func TestE2E_SlashCommands(t *testing.T) {
    agent := StartAgent(t)
    defer agent.Stop()

    session := agent.CreateSession(t.TempDir())

    // 等待命令可用
    commands := agent.WaitForAvailableCommands(session.SessionID, 5*time.Second)
    assert.GreaterOrEqual(t, len(commands), 1)

    // 测试 /compact 命令
    agent.Prompt(session.SessionID, "Tell me a long story")
    result := agent.Prompt(session.SessionID, "/compact")

    assert.Equal(t, "end_turn", result.StopReason)

    // 测试自定义命令
    customResult := agent.Prompt(session.SessionID, "/quick-math")
    assert.Contains(t, customResult.Text, "30")
}
```

---

## 测试优先级矩阵

| 优先级 | 测试类型 | 覆盖目标 | 测试数量 | 方法 |
|--------|---------|----------|----------|------|
| **P0** | 协议兼容性 | 100% | ~20 | Golden 文件 + 双向对比 |
| **P0** | 工具转换 | 100% | ~30 | 单元测试 + TypeScript 对比 |
| **P1** | 会话管理 | 100% | ~15 | 状态机测试 |
| **P1** | Prompt 处理 | 100% | ~20 | 集成测试 |
| **P1** | 权限系统 | 100% | ~15 | 模拟客户端 |
| **P2** | 边界情况 | 80% | ~25 | 模糊测试 |
| **P2** | 性能 | 关键路径 | ~10 | 基准测试 |
| **P3** | E2E | 主流程 | ~5 | 真实客户端 |

---

## 运行测试

```bash
# 运行所有单元测试
go test ./...

# 运行特定层级的测试
go test ./internal/acp/... -v

# 运行集成测试
go test -tags=integration ./e2e/... -v

# 运行 E2E 测试 (需要环境变量)
RUN_ZED_TESTS=1 go test ./e2e/... -v

# 运行基准测试
go test -bench=. ./...

# 运行竞态检测
go test -race ./...

# 生成覆盖率报告
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

---

*文档生成日期: 2026-02-13*