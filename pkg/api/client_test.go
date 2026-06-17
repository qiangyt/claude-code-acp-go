package api

import (
	"context"
	"io"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	acpsdk "github.com/coder/acp-go-sdk"
)

func TestNewClient(t *testing.T) {
	t.Run("创建默认客户端", func(t *testing.T) {
		client := NewClient()
		require.NotNil(t, client)
		require.NotNil(t, client.agent)
	})
}

func TestNewClientWithOptions(t *testing.T) {
	t.Run("使用选项创建客户端", func(t *testing.T) {
		opts := NewOptions().
			WithName("test-client").
			WithWorkingDir("/tmp")

		client := NewClientWithOptions(opts)
		require.NotNil(t, client)
	})

	t.Run("使用 nil 选项创建客户端", func(t *testing.T) {
		client := NewClientWithOptions(nil)
		require.NotNil(t, client)
	})
}

func TestClient_Initialize(t *testing.T) {
	t.Run("初始化客户端", func(t *testing.T) {
		client := NewClient()
		ctx := context.Background()

		req := acpsdk.InitializeRequest{
			ProtocolVersion: 1,
			ClientCapabilities: acpsdk.ClientCapabilities{
				Terminal: true,
			},
		}

		resp, err := client.Initialize(ctx, req)
		require.NoError(t, err)
		assert.Equal(t, acpsdk.ProtocolVersion(1), resp.ProtocolVersion)
		assert.True(t, resp.AgentCapabilities.LoadSession)
	})
}

func TestClient_NewSession(t *testing.T) {
	t.Run("创建新会话", func(t *testing.T) {
		client := NewClient()
		ctx := context.Background()

		req := acpsdk.NewSessionRequest{
			Cwd:        "/tmp",
			McpServers: []acpsdk.McpServer{},
		}

		resp, err := client.NewSession(ctx, req)
		require.NoError(t, err)
		assert.NotEmpty(t, resp.SessionId)
	})
}

func TestClient_LoadSession(t *testing.T) {
	t.Run("加载已存在的会话", func(t *testing.T) {
		client := NewClient()
		ctx := context.Background()

		// 先创建会话
		createResp, err := client.NewSession(ctx, acpsdk.NewSessionRequest{
			Cwd:        "/tmp",
			McpServers: []acpsdk.McpServer{},
		})
		require.NoError(t, err)

		// 加载会话
		loadReq := acpsdk.LoadSessionRequest{
			SessionId:  createResp.SessionId,
			Cwd:        "/tmp",
			McpServers: []acpsdk.McpServer{},
		}

		_, err = client.LoadSession(ctx, loadReq)
		require.NoError(t, err)
	})

	t.Run("加载不存在的会话失败", func(t *testing.T) {
		client := NewClient()
		ctx := context.Background()

		req := acpsdk.LoadSessionRequest{
			SessionId:  "non-existent-session",
			Cwd:        "/tmp",
			McpServers: []acpsdk.McpServer{},
		}

		_, err := client.LoadSession(ctx, req)
		require.Error(t, err)
		require.Contains(t, err.Error(), "session not found")
	})
}

func TestClient_Cancel(t *testing.T) {
	t.Run("取消已存在的会话", func(t *testing.T) {
		client := NewClient()
		ctx := context.Background()

		// 先创建会话
		createResp, err := client.NewSession(ctx, acpsdk.NewSessionRequest{
			Cwd:        "/tmp",
			McpServers: []acpsdk.McpServer{},
		})
		require.NoError(t, err)

		// 取消会话
		cancelReq := acpsdk.CancelNotification{
			SessionId: createResp.SessionId,
		}

		err = client.Cancel(ctx, cancelReq)
		require.NoError(t, err)

		// 确认会话已被移除
		session := client.GetSession(string(createResp.SessionId))
		require.Nil(t, session)
	})

	t.Run("取消不存在的会话是幂等的", func(t *testing.T) {
		client := NewClient()
		ctx := context.Background()

		req := acpsdk.CancelNotification{
			SessionId: "non-existent-session",
		}

		err := client.Cancel(ctx, req)
		require.NoError(t, err)
	})
}

func TestClient_Prompt(t *testing.T) {
	t.Run("会话不存在时返回错误", func(t *testing.T) {
		client := NewClient()
		ctx := context.Background()

		req := acpsdk.PromptRequest{
			SessionId: "non-existent-session",
			Prompt:    []acpsdk.ContentBlock{acpsdk.TextBlock("Hello")},
		}

		_, err := client.Prompt(ctx, req)
		require.Error(t, err)
		require.Contains(t, err.Error(), "session not found")
	})
}

func TestClient_GetSession(t *testing.T) {
	t.Run("获取已存在的会话", func(t *testing.T) {
		client := NewClient()
		ctx := context.Background()

		createResp, err := client.NewSession(ctx, acpsdk.NewSessionRequest{
			Cwd:        "/tmp",
			McpServers: []acpsdk.McpServer{},
		})
		require.NoError(t, err)

		session := client.GetSession(string(createResp.SessionId))
		require.NotNil(t, session)
		assert.Equal(t, string(createResp.SessionId), session.SessionID)
	})

	t.Run("获取不存在的会话返回 nil", func(t *testing.T) {
		client := NewClient()

		session := client.GetSession("non-existent")
		require.Nil(t, session)
	})
}

func TestClient_RemoveSession(t *testing.T) {
	t.Run("移除已存在的会话", func(t *testing.T) {
		client := NewClient()
		ctx := context.Background()

		createResp, err := client.NewSession(ctx, acpsdk.NewSessionRequest{
			Cwd:        "/tmp",
			McpServers: []acpsdk.McpServer{},
		})
		require.NoError(t, err)

		// 确认会话存在
		session := client.GetSession(string(createResp.SessionId))
		require.NotNil(t, session)

		// 移除会话
		client.RemoveSession(string(createResp.SessionId))

		// 确认会话已被移除
		session = client.GetSession(string(createResp.SessionId))
		require.Nil(t, session)
	})
}

func TestClient_SendSessionUpdate(t *testing.T) {
	t.Run("无连接时发送失败", func(t *testing.T) {
		client := NewClient()
		ctx := context.Background()

		err := client.SendSessionUpdate(ctx, "session-1", acpsdk.UpdateAgentMessageText("test"))
		require.Error(t, err)
		require.Contains(t, err.Error(), "connection not established")
	})
}

func TestClient_SendAgentMessage(t *testing.T) {
	t.Run("无连接时发送失败", func(t *testing.T) {
		client := NewClient()
		ctx := context.Background()

		err := client.SendAgentMessage(ctx, "session-1", "Hello")
		require.Error(t, err)
		require.Contains(t, err.Error(), "connection not established")
	})
}

func TestClient_StartFromStdio(t *testing.T) {
	t.Run("从 stdio 启动", func(t *testing.T) {
		client := NewClient()

		// 使用管道模拟 stdio
		r1, w1 := io.Pipe()
		r2, w2 := io.Pipe()
		defer r1.Close()
		defer w1.Close()
		defer r2.Close()
		defer w2.Close()

		ctx := context.Background()
		conn, err := client.StartFromStdio(ctx, w1, r2)
		require.NoError(t, err)
		require.NotNil(t, conn)
	})
}

func TestClient_SetSessionMode(t *testing.T) {
	t.Run("SetSessionMode 未实现", func(t *testing.T) {
		client := NewClient()
		ctx := context.Background()

		req := acpsdk.SetSessionModeRequest{}
		_, err := client.SetSessionMode(ctx, req)
		require.Error(t, err)
		require.Contains(t, err.Error(), "not implemented")
	})
}

func TestClient_Authenticate(t *testing.T) {
	t.Run("Authenticate 未实现", func(t *testing.T) {
		client := NewClient()
		ctx := context.Background()

		req := acpsdk.AuthenticateRequest{}
		_, err := client.Authenticate(ctx, req)
		require.Error(t, err)
		require.Contains(t, err.Error(), "not implemented")
	})
}

func TestClient_Run(t *testing.T) {
	t.Run("Run 在 context 取消时返回", func(t *testing.T) {
		client := NewClient()

		ctx, cancel := context.WithCancel(context.Background())

		// 启动 goroutine 来取消 context
		go func() {
			time.Sleep(50 * time.Millisecond)
			cancel()
		}()

		err := client.Run(ctx)
		require.NoError(t, err)
	})
}
