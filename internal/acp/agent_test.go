package acp

import (
	"context"
	"io"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"claude-code-acp-go/internal/mcp"

	acpsdk "github.com/coder/acp-go-sdk"
)

func TestNewAgent(t *testing.T) {
	t.Run("创建新的 Agent", func(t *testing.T) {
		agent := NewAgent()
		assert.NotNil(t, agent)
	})
}

func TestNewAgentWithOptions(t *testing.T) {
	t.Run("使用选项创建 Agent", func(t *testing.T) {
		opts := NewAgentOptions().
			WithWorkingDir("/tmp").
			WithName("test-agent")

		agent := NewAgentWithOptions(opts)
		require.NotNil(t, agent)
		require.Equal(t, "/tmp", agent.workingDir)
		require.Equal(t, "test-agent", agent.name)
	})

	t.Run("使用默认选项创建 Agent", func(t *testing.T) {
		agent := NewAgentWithOptions(nil)
		require.NotNil(t, agent)
		require.Equal(t, ".", agent.workingDir)
	})

	t.Run("使用 PermissionChecker 创建 Agent", func(t *testing.T) {
		checker := mcp.NewDefaultPermissionChecker()
		opts := NewAgentOptions().
			WithName("test-with-checker").
			WithPermissionChecker(checker)

		agent := NewAgentWithOptions(opts)
		require.NotNil(t, agent)
		require.Equal(t, checker, agent.permissionChecker)
	})

	t.Run("创建 Agent 时 MCP 服务器失败会 panic", func(t *testing.T) {
		// 使用空名称会导致 MCP 服务器创建失败
		opts := NewAgentOptions().WithName("")

		defer func() {
			r := recover()
			require.NotNil(t, r)
			require.Contains(t, r.(error).Error(), "failed to create MCP server")
		}()

		NewAgentWithOptions(opts)
	})
}

func TestAgent_MCPServer(t *testing.T) {
	t.Run("Agent 包含 MCP 服务器", func(t *testing.T) {
		agent := NewAgent()
		require.NotNil(t, agent)
		require.NotNil(t, agent.mcpServer)
	})

	t.Run("创建会话后 MCP 服务器已注册工具", func(t *testing.T) {
		agent := NewAgent()
		_, err := agent.CreateSession("test-session")
		require.NoError(t, err)
		tools := agent.mcpServer.ListTools()
		require.GreaterOrEqual(t, len(tools), 6)
	})
}

func TestAgent_CreateSession(t *testing.T) {
	t.Run("创建会话", func(t *testing.T) {
		agent := NewAgent()
		session, err := agent.CreateSession("session-1")
		require.NoError(t, err)
		require.NotNil(t, session)
		require.Equal(t, "session-1", session.SessionID)
	})

	t.Run("创建会话时初始化 MCP 上下文", func(t *testing.T) {
		agent := NewAgent()
		session, err := agent.CreateSession("session-2")
		require.NoError(t, err)
		require.NotNil(t, session.MCPContext)
	})

	t.Run("创建重复会话失败", func(t *testing.T) {
		agent := NewAgent()
		_, err := agent.CreateSession("session-dup")
		require.NoError(t, err)

		// 再次创建同名的会话会失败，因为工具已经注册过了
		_, err = agent.CreateSession("session-dup")
		require.Error(t, err)
		require.Contains(t, err.Error(), "failed to register builtin tools")
	})
}

func TestAgent_Run(t *testing.T) {
	t.Run("Run 在 context 取消时返回", func(t *testing.T) {
		agent := NewAgent()

		ctx, cancel := context.WithCancel(context.Background())

		// 启动 goroutine 来取消 context
		go func() {
			time.Sleep(50 * time.Millisecond)
			cancel()
		}()

		// Run 应该在 context 取消后返回
		err := agent.Run(ctx)
		require.NoError(t, err)
	})

	t.Run("Run 在已取消的 context 上立即返回", func(t *testing.T) {
		agent := NewAgent()

		ctx, cancel := context.WithCancel(context.Background())
		cancel() // 立即取消

		err := agent.Run(ctx)
		require.NoError(t, err)
	})

	t.Run("Run 正确处理带超时的 context", func(t *testing.T) {
		agent := NewAgent()

		ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
		defer cancel()

		start := time.Now()
		err := agent.Run(ctx)
		elapsed := time.Since(start)

		require.NoError(t, err)
		// 验证大约等待了 50ms（允许一定误差）
		assert.GreaterOrEqual(t, elapsed.Milliseconds(), int64(40))
	})
}

func TestAgentOptions(t *testing.T) {
	t.Run("所有选项链式调用", func(t *testing.T) {
		checker := mcp.NewDefaultPermissionChecker()
		opts := NewAgentOptions().
			WithName("test").
			WithWorkingDir("/work").
			WithPermissionChecker(checker)

		require.Equal(t, "test", opts.Name)
		require.Equal(t, "/work", opts.WorkingDir)
		require.Equal(t, checker, opts.PermissionChecker)
	})
}

func TestAgent_GetSession(t *testing.T) {
	t.Run("获取已存在的会话", func(t *testing.T) {
		agent := NewAgent()
		_, err := agent.CreateSession("session-1")
		require.NoError(t, err)

		session := agent.GetSession("session-1")
		require.NotNil(t, session)
		require.Equal(t, "session-1", session.SessionID)
	})

	t.Run("获取不存在的会话返回 nil", func(t *testing.T) {
		agent := NewAgent()

		session := agent.GetSession("non-existent")
		require.Nil(t, session)
	})

	t.Run("并发获取会话", func(t *testing.T) {
		agent := NewAgent()
		_, err := agent.CreateSession("concurrent-session")
		require.NoError(t, err)

		// 并发读取
		done := make(chan bool)
		for i := 0; i < 10; i++ {
			go func() {
				session := agent.GetSession("concurrent-session")
				assert.NotNil(t, session)
				done <- true
			}()
		}

		for i := 0; i < 10; i++ {
			<-done
		}
	})
}

func TestAgent_RemoveSession(t *testing.T) {
	t.Run("移除已存在的会话", func(t *testing.T) {
		agent := NewAgent()
		_, err := agent.CreateSession("session-to-remove")
		require.NoError(t, err)

		// 确认会话存在
		session := agent.GetSession("session-to-remove")
		require.NotNil(t, session)

		// 移除会话
		agent.RemoveSession("session-to-remove")

		// 确认会话已被移除
		session = agent.GetSession("session-to-remove")
		require.Nil(t, session)
	})

	t.Run("移除不存在的会话不报错", func(t *testing.T) {
		agent := NewAgent()

		// 移除不存在的会话应该安全地什么都不做
		agent.RemoveSession("non-existent")
	})

	t.Run("移除会话时清理终端", func(t *testing.T) {
		agent := NewAgent()
		session, err := agent.CreateSession("session-with-terminal")
		require.NoError(t, err)

		// 添加终端到会话
		session.MCPContext.AddTerminal("terminal-1", "echo test")

		// 移除会话
		agent.RemoveSession("session-with-terminal")

		// 确认会话已被移除
		session = agent.GetSession("session-with-terminal")
		require.Nil(t, session)
	})

	t.Run("并发移除不同会话", func(t *testing.T) {
		// 使用多个 agent 来避免工具重复注册问题
		agents := make([]*Agent, 5)
		for i := 0; i < 5; i++ {
			agents[i] = NewAgent()
			_, err := agents[i].CreateSession(string(rune('a' + i)))
			require.NoError(t, err)
		}

		// 并发移除
		done := make(chan bool)
		for i := 0; i < 5; i++ {
			go func(agent *Agent, id string) {
				agent.RemoveSession(id)
				done <- true
			}(agents[i], string(rune('a' + i)))
		}

		for i := 0; i < 5; i++ {
			<-done
		}

		// 确认所有会话已被移除
		for i := 0; i < 5; i++ {
			session := agents[i].GetSession(string(rune('a' + i)))
			require.Nil(t, session)
		}
	})
}

// ============================================================
// ACP 协议方法测试
// ============================================================

func TestAgent_Initialize(t *testing.T) {
	t.Run("初始化代理", func(t *testing.T) {
		agent := NewAgent()
		ctx := context.Background()

		req := acpsdk.InitializeRequest{
			ProtocolVersion: 1,
			ClientCapabilities: acpsdk.ClientCapabilities{
				Terminal: true,
			},
		}

		resp, err := agent.Initialize(ctx, req)
		require.NoError(t, err)
		assert.Equal(t, acpsdk.ProtocolVersion(1), resp.ProtocolVersion)
		assert.True(t, resp.AgentCapabilities.LoadSession)
		assert.True(t, resp.AgentCapabilities.PromptCapabilities.Image)
		assert.True(t, resp.AgentCapabilities.PromptCapabilities.Audio)
		assert.NotNil(t, resp.AgentInfo)
	})

	t.Run("返回代理信息", func(t *testing.T) {
		opts := NewAgentOptions().WithName("test-agent")
		agent := NewAgentWithOptions(opts)
		ctx := context.Background()

		req := acpsdk.InitializeRequest{
			ProtocolVersion: 1,
		}

		resp, err := agent.Initialize(ctx, req)
		require.NoError(t, err)
		assert.Equal(t, "test-agent", resp.AgentInfo.Name)
		assert.NotEmpty(t, resp.AgentInfo.Version)
	})
}

func TestAgent_NewSession_ACP(t *testing.T) {
	t.Run("创建新会话", func(t *testing.T) {
		agent := NewAgent()
		ctx := context.Background()

		req := acpsdk.NewSessionRequest{
			Cwd:        "/tmp",
			McpServers: []acpsdk.McpServer{},
		}

		resp, err := agent.NewSession(ctx, req)
		require.NoError(t, err)
		assert.NotEmpty(t, resp.SessionId)
	})

	t.Run("使用默认工作目录", func(t *testing.T) {
		opts := NewAgentOptions().WithWorkingDir("/default")
		agent := NewAgentWithOptions(opts)
		ctx := context.Background()

		req := acpsdk.NewSessionRequest{
			Cwd:        "", // 空字符串使用默认值
			McpServers: []acpsdk.McpServer{},
		}

		resp, err := agent.NewSession(ctx, req)
		require.NoError(t, err)
		assert.NotEmpty(t, resp.SessionId)
	})
}

func TestAgent_LoadSession(t *testing.T) {
	t.Run("加载已存在的会话", func(t *testing.T) {
		agent := NewAgent()
		ctx := context.Background()

		// 先创建会话
		createReq := acpsdk.NewSessionRequest{
			Cwd:        "/tmp",
			McpServers: []acpsdk.McpServer{},
		}
		createResp, err := agent.NewSession(ctx, createReq)
		require.NoError(t, err)

		// 加载会话
		loadReq := acpsdk.LoadSessionRequest{
			SessionId:  createResp.SessionId,
			Cwd:        "/tmp",
			McpServers: []acpsdk.McpServer{},
		}

		_, err = agent.LoadSession(ctx, loadReq)
		require.NoError(t, err)
	})

	t.Run("加载不存在的会话失败", func(t *testing.T) {
		agent := NewAgent()
		ctx := context.Background()

		req := acpsdk.LoadSessionRequest{
			SessionId:  "non-existent-session",
			Cwd:        "/tmp",
			McpServers: []acpsdk.McpServer{},
		}

		_, err := agent.LoadSession(ctx, req)
		require.Error(t, err)
		require.Contains(t, err.Error(), "session not found")
	})
}

func TestAgent_Cancel(t *testing.T) {
	t.Run("取消已存在的会话", func(t *testing.T) {
		agent := NewAgent()
		ctx := context.Background()

		// 先创建会话
		createReq := acpsdk.NewSessionRequest{
			Cwd:        "/tmp",
			McpServers: []acpsdk.McpServer{},
		}
		createResp, err := agent.NewSession(ctx, createReq)
		require.NoError(t, err)

		// 取消会话
		cancelReq := acpsdk.CancelNotification{
			SessionId: createResp.SessionId,
		}

		err = agent.Cancel(ctx, cancelReq)
		require.NoError(t, err)

		// 确认会话已被移除
		session := agent.GetSession(string(createResp.SessionId))
		require.Nil(t, session)
	})

	t.Run("取消不存在的会话是幂等的", func(t *testing.T) {
		agent := NewAgent()
		ctx := context.Background()

		req := acpsdk.CancelNotification{
			SessionId: "non-existent-session",
		}

		err := agent.Cancel(ctx, req)
		require.NoError(t, err)
	})
}

func TestAgent_Prompt(t *testing.T) {
	t.Run("会话不存在时返回错误", func(t *testing.T) {
		agent := NewAgent()
		ctx := context.Background()

		req := acpsdk.PromptRequest{
			SessionId: "non-existent-session",
			Prompt:    []acpsdk.ContentBlock{acpsdk.TextBlock("Hello")},
		}

		_, err := agent.Prompt(ctx, req)
		require.Error(t, err)
		require.Contains(t, err.Error(), "session not found")
	})

	t.Run("空 prompt 返回错误", func(t *testing.T) {
		agent := NewAgent()
		ctx := context.Background()

		// 先创建会话
		createResp, err := agent.NewSession(ctx, acpsdk.NewSessionRequest{
			Cwd:        "/tmp",
			McpServers: []acpsdk.McpServer{},
		})
		require.NoError(t, err)

		req := acpsdk.PromptRequest{
			SessionId: createResp.SessionId,
			Prompt:    []acpsdk.ContentBlock{}, // 空 prompt
		}

		_, err = agent.Prompt(ctx, req)
		require.Error(t, err)
		require.Contains(t, err.Error(), "prompt is required")
	})

	t.Run("有效 prompt 但 AI 未实现", func(t *testing.T) {
		agent := NewAgent()
		ctx := context.Background()

		// 先创建会话
		createResp, err := agent.NewSession(ctx, acpsdk.NewSessionRequest{
			Cwd:        "/tmp",
			McpServers: []acpsdk.McpServer{},
		})
		require.NoError(t, err)

		req := acpsdk.PromptRequest{
			SessionId: createResp.SessionId,
			Prompt:    []acpsdk.ContentBlock{acpsdk.TextBlock("Hello, World!")},
		}

		_, err = agent.Prompt(ctx, req)
		require.Error(t, err)
		require.Contains(t, err.Error(), "not implemented")
	})

	t.Run("多模态 prompt 验证", func(t *testing.T) {
		agent := NewAgent()
		ctx := context.Background()

		// 先创建会话
		createResp, err := agent.NewSession(ctx, acpsdk.NewSessionRequest{
			Cwd:        "/tmp",
			McpServers: []acpsdk.McpServer{},
		})
		require.NoError(t, err)

		// 包含文本和图片的 prompt
		req := acpsdk.PromptRequest{
			SessionId: createResp.SessionId,
			Prompt: []acpsdk.ContentBlock{
				acpsdk.TextBlock("Please analyze this image:"),
				acpsdk.ImageBlock("base64imagedata", "image/png"),
			},
		}

		_, err = agent.Prompt(ctx, req)
		require.Error(t, err)
		require.Contains(t, err.Error(), "not implemented")
	})
}

func TestAgent_SetSessionMode(t *testing.T) {
	t.Run("SetSessionMode 未实现", func(t *testing.T) {
		agent := NewAgent()
		ctx := context.Background()

		req := acpsdk.SetSessionModeRequest{}

		_, err := agent.SetSessionMode(ctx, req)
		require.Error(t, err)
		require.Contains(t, err.Error(), "not implemented")
	})
}

func TestAgent_Authenticate(t *testing.T) {
	t.Run("Authenticate 未实现", func(t *testing.T) {
		agent := NewAgent()
		ctx := context.Background()

		req := acpsdk.AuthenticateRequest{}

		_, err := agent.Authenticate(ctx, req)
		require.Error(t, err)
		require.Contains(t, err.Error(), "not implemented")
	})
}

// ============================================================
// 连接感知接口测试
// ============================================================

func TestAgent_SetAgentConnection(t *testing.T) {
	t.Run("设置连接", func(t *testing.T) {
		agent := NewAgent()
		require.Nil(t, agent.conn)

		// 模拟连接设置
		agent.SetAgentConnection(nil)
		// 连接可以是 nil（用于测试）
		require.Nil(t, agent.conn)
	})
}

// ============================================================
// 通知发送方法测试
// ============================================================

func TestAgent_SendSessionUpdate_NoConnection(t *testing.T) {
	t.Run("无连接时发送失败", func(t *testing.T) {
		agent := NewAgent()
		ctx := context.Background()

		err := agent.SendSessionUpdate(ctx, "session-1", acpsdk.UpdateAgentMessageText("test"))
		require.Error(t, err)
		require.Contains(t, err.Error(), "connection not established")
	})
}

func TestAgent_SendAgentMessage_NoConnection(t *testing.T) {
	t.Run("无连接时发送消息失败", func(t *testing.T) {
		agent := NewAgent()
		ctx := context.Background()

		err := agent.SendAgentMessage(ctx, "session-1", "Hello")
		require.Error(t, err)
		require.Contains(t, err.Error(), "connection not established")
	})
}

func TestAgent_SendAgentThought_NoConnection(t *testing.T) {
	t.Run("无连接时发送思考失败", func(t *testing.T) {
		agent := NewAgent()
		ctx := context.Background()

		err := agent.SendAgentThought(ctx, "session-1", "Thinking...")
		require.Error(t, err)
		require.Contains(t, err.Error(), "connection not established")
	})
}

func TestAgent_SendPlanUpdate_NoConnection(t *testing.T) {
	t.Run("无连接时发送计划失败", func(t *testing.T) {
		agent := NewAgent()
		ctx := context.Background()

		entries := []acpsdk.PlanEntry{
			{Content: "Step 1", Status: acpsdk.PlanEntryStatusPending},
		}
		err := agent.SendPlanUpdate(ctx, "session-1", entries...)
		require.Error(t, err)
		require.Contains(t, err.Error(), "connection not established")
	})
}

func TestAgent_StartToolCall_NoConnection(t *testing.T) {
	t.Run("无连接时开始工具调用失败", func(t *testing.T) {
		agent := NewAgent()
		ctx := context.Background()

		err := agent.StartToolCall(ctx, "session-1", "call-1", "Reading file",
			acpsdk.WithStartKind(acpsdk.ToolKindRead),
			acpsdk.WithStartStatus(acpsdk.ToolCallStatusPending),
		)
		require.Error(t, err)
		require.Contains(t, err.Error(), "connection not established")
	})
}

func TestAgent_UpdateToolCall_NoConnection(t *testing.T) {
	t.Run("无连接时更新工具调用失败", func(t *testing.T) {
		agent := NewAgent()
		ctx := context.Background()

		err := agent.UpdateToolCall(ctx, "session-1", "call-1",
			acpsdk.WithUpdateStatus(acpsdk.ToolCallStatusCompleted),
		)
		require.Error(t, err)
		require.Contains(t, err.Error(), "connection not established")
	})
}

func TestAgent_RequestPermission_NoConnection(t *testing.T) {
	t.Run("无连接时请求权限失败", func(t *testing.T) {
		agent := NewAgent()
		ctx := context.Background()

		toolCall := acpsdk.ToolCallUpdate{
			ToolCallId: "call-1",
			Title:      acpsdk.Ptr("Test tool call"),
		}
		options := []acpsdk.PermissionOption{
			{Kind: acpsdk.PermissionOptionKindAllowOnce, Name: "Allow", OptionId: "allow"},
		}

		_, err := agent.RequestPermission(ctx, "session-1", toolCall, options)
		require.Error(t, err)
		require.Contains(t, err.Error(), "connection not established")
	})
}

// ============================================================
// 通知集成测试（使用管道连接）
// ============================================================

// mockClient 实现 acpsdk.Client 接口
type mockClient struct {
	sessionUpdates []acpsdk.SessionNotification
	mu             sync.Mutex
}

func (m *mockClient) WriteTextFile(ctx context.Context, p acpsdk.WriteTextFileRequest) (acpsdk.WriteTextFileResponse, error) {
	return acpsdk.WriteTextFileResponse{}, nil
}

func (m *mockClient) ReadTextFile(ctx context.Context, p acpsdk.ReadTextFileRequest) (acpsdk.ReadTextFileResponse, error) {
	return acpsdk.ReadTextFileResponse{}, nil
}

func (m *mockClient) RequestPermission(ctx context.Context, p acpsdk.RequestPermissionRequest) (acpsdk.RequestPermissionResponse, error) {
	return acpsdk.RequestPermissionResponse{
		Outcome: acpsdk.RequestPermissionOutcome{
			Selected: &acpsdk.RequestPermissionOutcomeSelected{
				OptionId: "allow",
			},
		},
	}, nil
}

func (m *mockClient) SessionUpdate(ctx context.Context, n acpsdk.SessionNotification) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.sessionUpdates = append(m.sessionUpdates, n)
	return nil
}

func (m *mockClient) CreateTerminal(ctx context.Context, p acpsdk.CreateTerminalRequest) (acpsdk.CreateTerminalResponse, error) {
	return acpsdk.CreateTerminalResponse{}, nil
}

func (m *mockClient) KillTerminalCommand(ctx context.Context, p acpsdk.KillTerminalCommandRequest) (acpsdk.KillTerminalCommandResponse, error) {
	return acpsdk.KillTerminalCommandResponse{}, nil
}

func (m *mockClient) ReleaseTerminal(ctx context.Context, p acpsdk.ReleaseTerminalRequest) (acpsdk.ReleaseTerminalResponse, error) {
	return acpsdk.ReleaseTerminalResponse{}, nil
}

func (m *mockClient) TerminalOutput(ctx context.Context, p acpsdk.TerminalOutputRequest) (acpsdk.TerminalOutputResponse, error) {
	return acpsdk.TerminalOutputResponse{}, nil
}

func (m *mockClient) WaitForTerminalExit(ctx context.Context, p acpsdk.WaitForTerminalExitRequest) (acpsdk.WaitForTerminalExitResponse, error) {
	return acpsdk.WaitForTerminalExitResponse{}, nil
}

func (m *mockClient) getUpdates() []acpsdk.SessionNotification {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.sessionUpdates
}

func TestAgent_NotificationIntegration(t *testing.T) {
	t.Run("通过连接发送通知", func(t *testing.T) {
		// 创建管道连接
		c2aR, c2aW := io.Pipe()
		a2cR, a2cW := io.Pipe()
		defer c2aR.Close()
		defer c2aW.Close()
		defer a2cR.Close()
		defer a2cW.Close()

		// 创建 mock client
		client := &mockClient{}

		// 创建客户端连接
		clientConn := acpsdk.NewClientSideConnection(client, c2aW, a2cR)
		_ = clientConn // 避免未使用警告

		// 创建 agent
		agent := NewAgent()

		// 创建 agent 连接
		agentConn := acpsdk.NewAgentSideConnection(agent, a2cW, c2aR)
		agent.SetAgentConnection(agentConn)

		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()

		// 发送消息
		err := agent.SendAgentMessage(ctx, "session-1", "Hello, World!")
		require.NoError(t, err)

		// 等待消息传输
		time.Sleep(100 * time.Millisecond)

		// 验证客户端收到消息
		updates := client.getUpdates()
		require.GreaterOrEqual(t, len(updates), 1)
		assert.Equal(t, acpsdk.SessionId("session-1"), updates[0].SessionId)
	})

	t.Run("发送思考消息", func(t *testing.T) {
		c2aR, c2aW := io.Pipe()
		a2cR, a2cW := io.Pipe()
		defer c2aR.Close()
		defer c2aW.Close()
		defer a2cR.Close()
		defer a2cW.Close()

		client := &mockClient{}
		clientConn := acpsdk.NewClientSideConnection(client, c2aW, a2cR)
		_ = clientConn

		agent := NewAgent()
		agentConn := acpsdk.NewAgentSideConnection(agent, a2cW, c2aR)
		agent.SetAgentConnection(agentConn)

		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()

		err := agent.SendAgentThought(ctx, "session-2", "Thinking...")
		require.NoError(t, err)

		time.Sleep(100 * time.Millisecond)

		updates := client.getUpdates()
		require.GreaterOrEqual(t, len(updates), 1)
	})

	t.Run("发送计划更新", func(t *testing.T) {
		c2aR, c2aW := io.Pipe()
		a2cR, a2cW := io.Pipe()
		defer c2aR.Close()
		defer c2aW.Close()
		defer a2cR.Close()
		defer a2cW.Close()

		client := &mockClient{}
		clientConn := acpsdk.NewClientSideConnection(client, c2aW, a2cR)
		_ = clientConn

		agent := NewAgent()
		agentConn := acpsdk.NewAgentSideConnection(agent, a2cW, c2aR)
		agent.SetAgentConnection(agentConn)

		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()

		entries := []acpsdk.PlanEntry{
			{Content: "Step 1", Status: acpsdk.PlanEntryStatusCompleted},
			{Content: "Step 2", Status: acpsdk.PlanEntryStatusInProgress},
		}
		err := agent.SendPlanUpdate(ctx, "session-3", entries...)
		require.NoError(t, err)

		time.Sleep(100 * time.Millisecond)

		updates := client.getUpdates()
		require.GreaterOrEqual(t, len(updates), 1)
	})

	t.Run("工具调用通知", func(t *testing.T) {
		c2aR, c2aW := io.Pipe()
		a2cR, a2cW := io.Pipe()
		defer c2aR.Close()
		defer c2aW.Close()
		defer a2cR.Close()
		defer a2cW.Close()

		client := &mockClient{}
		clientConn := acpsdk.NewClientSideConnection(client, c2aW, a2cR)
		_ = clientConn

		agent := NewAgent()
		agentConn := acpsdk.NewAgentSideConnection(agent, a2cW, c2aR)
		agent.SetAgentConnection(agentConn)

		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()

		// 开始工具调用
		err := agent.StartToolCall(ctx, "session-4", "call-1", "Reading file",
			acpsdk.WithStartKind(acpsdk.ToolKindRead),
			acpsdk.WithStartStatus(acpsdk.ToolCallStatusPending),
			acpsdk.WithStartLocations([]acpsdk.ToolCallLocation{{Path: "/tmp/test.txt"}}),
		)
		require.NoError(t, err)

		// 更新工具调用
		err = agent.UpdateToolCall(ctx, "session-4", "call-1",
			acpsdk.WithUpdateStatus(acpsdk.ToolCallStatusCompleted),
		)
		require.NoError(t, err)

		time.Sleep(100 * time.Millisecond)

		updates := client.getUpdates()
		require.GreaterOrEqual(t, len(updates), 2)
	})
}
