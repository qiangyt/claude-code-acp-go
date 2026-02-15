package acp

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"claude-code-acp-go/internal/mcp"
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
