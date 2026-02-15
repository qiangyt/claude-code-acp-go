package mcp

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

// ============================================================
// 1.1 Server 接口和结构测试
// ============================================================

func TestServerOptions_WithPermissionChecker(t *testing.T) {
	checker := &mockPermissionChecker{}
	opts := NewServerOptions().WithPermissionChecker(checker)
	require.NotNil(t, opts)
	require.Equal(t, checker, opts.PermissionChecker)
}

func TestServerOptions_WithSessionContext(t *testing.T) {
	ctx := NewSessionContext("test-session", "/tmp")
	opts := NewServerOptions().WithSessionContext(ctx)
	require.NotNil(t, opts)
	require.Equal(t, ctx, opts.SessionContext)
}

func TestNewServer_WithDefaultOptions(t *testing.T) {
	server, err := NewServer("test-server")
	require.NoError(t, err)
	require.NotNil(t, server)
	require.Equal(t, "test-server", server.Name())
}

func TestNewServer_WithNilOptions(t *testing.T) {
	server, err := NewServer("test-server", nil)
	require.NoError(t, err)
	require.NotNil(t, server)
}

func TestNewServer_WithEmptyName(t *testing.T) {
	_, err := NewServer("")
	require.Error(t, err)
	require.Contains(t, err.Error(), "server name is required")
}

func TestNewServerP_Success(t *testing.T) {
	server := NewServerP("test-server")
	require.NotNil(t, server)
	require.Equal(t, "test-server", server.Name())
}

func TestServer_Name(t *testing.T) {
	server := NewServerP("my-server")
	require.Equal(t, "my-server", server.Name())
}

func TestServer_Version(t *testing.T) {
	server := NewServerP("my-server")
	require.Equal(t, "1.0.0", server.Version())
}

// ============================================================
// 1.2 工具注册测试
// ============================================================

func TestServer_RegisterTool(t *testing.T) {
	server := NewServerP("test-server")
	tool := Tool{
		Name:        "read",
		Description: "读取文件",
		Handler:     func(ctx context.Context, args map[string]any) (any, error) { return "ok", nil },
	}

	err := server.RegisterTool(tool)
	require.NoError(t, err)

	// 验证工具已注册
	tools := server.ListTools()
	require.Len(t, tools, 1)
	require.Equal(t, "read", tools[0].Name)
}

func TestServer_RegisterTool_EmptyName(t *testing.T) {
	server := NewServerP("test-server")
	tool := Tool{
		Name:        "",
		Description: "空名称工具",
		Handler:     func(ctx context.Context, args map[string]any) (any, error) { return "ok", nil },
	}

	err := server.RegisterTool(tool)
	require.Error(t, err)
	require.Contains(t, err.Error(), "tool name is required")
}

func TestServer_RegisterTool_NilHandler(t *testing.T) {
	server := NewServerP("test-server")
	tool := Tool{
		Name:        "test",
		Description: "空处理器工具",
		Handler:     nil,
	}

	err := server.RegisterTool(tool)
	require.Error(t, err)
	require.Contains(t, err.Error(), "tool handler is required")
}

func TestServer_RegisterTool_Duplicate(t *testing.T) {
	server := NewServerP("test-server")
	tool := Tool{
		Name:        "read",
		Description: "读取文件",
		Handler:     func(ctx context.Context, args map[string]any) (any, error) { return "ok", nil },
	}

	err := server.RegisterTool(tool)
	require.NoError(t, err)

	// 注册同名工具
	err = server.RegisterTool(tool)
	require.Error(t, err)
	require.Contains(t, err.Error(), "duplicate tool name")
}

func TestServer_RegisterTools(t *testing.T) {
	server := NewServerP("test-server")
	tools := []Tool{
		{
			Name:        "read",
			Description: "读取文件",
			Handler:     func(ctx context.Context, args map[string]any) (any, error) { return "read", nil },
		},
		{
			Name:        "write",
			Description: "写入文件",
			Handler:     func(ctx context.Context, args map[string]any) (any, error) { return "write", nil },
		},
	}

	err := server.RegisterTools(tools...)
	require.NoError(t, err)

	list := server.ListTools()
	require.Len(t, list, 2)
}

func TestServer_ListTools_Empty(t *testing.T) {
	server := NewServerP("test-server")
	tools := server.ListTools()
	require.Empty(t, tools)
}

// ============================================================
// 1.3 HandleMessage 测试 (MCP 协议处理)
// ============================================================

func TestServer_HandleMessage_ListTools(t *testing.T) {
	server := NewServerP("test-server")
	_ = server.RegisterTool(Tool{
		Name:        "read",
		Description: "读取文件",
		Handler:     func(ctx context.Context, args map[string]any) (any, error) { return "ok", nil },
	})

	msg := map[string]any{
		"jsonrpc": "2.0",
		"id":      float64(1),
		"method":  "tools/list",
	}

	resp, err := server.HandleMessage(msg)
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.Equal(t, "2.0", resp["jsonrpc"])
	require.Equal(t, float64(1), resp["id"])

	result := resp["result"].(map[string]any)
	tools := result["tools"].([]map[string]any)
	require.Len(t, tools, 1)
	require.Equal(t, "read", tools[0]["name"])
}

func TestServer_HandleMessage_CallTool(t *testing.T) {
	server := NewServerP("test-server")
	_ = server.RegisterTool(Tool{
		Name:        "add",
		Description: "加法运算",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"a": map[string]any{"type": "number"},
				"b": map[string]any{"type": "number"},
			},
		},
		Handler: func(ctx context.Context, args map[string]any) (any, error) {
			a := args["a"].(float64)
			b := args["b"].(float64)
			return map[string]any{"result": a + b}, nil
		},
	})

	msg := map[string]any{
		"jsonrpc": "2.0",
		"id":      float64(1),
		"method":  "tools/call",
		"params": map[string]any{
			"name": "add",
			"arguments": map[string]any{
				"a": float64(3),
				"b": float64(5),
			},
		},
	}

	resp, err := server.HandleMessage(msg)
	require.NoError(t, err)
	require.NotNil(t, resp)

	result := resp["result"].(map[string]any)
	content := result["content"].([]map[string]any)
	require.Len(t, content, 1)
	require.Equal(t, "text", content[0]["type"])
}

func TestServer_HandleMessage_InvalidMethod(t *testing.T) {
	server := NewServerP("test-server")

	msg := map[string]any{
		"jsonrpc": "2.0",
		"id":      float64(1),
		"method":  "unknown/method",
	}

	resp, err := server.HandleMessage(msg)
	require.NoError(t, err)
	require.NotNil(t, resp)

	errResp := resp["error"].(map[string]any)
	require.Equal(t, -32601, errResp["code"])
}

func TestServer_HandleMessage_MissingMethod(t *testing.T) {
	server := NewServerP("test-server")

	msg := map[string]any{
		"jsonrpc": "2.0",
		"id":      float64(1),
	}

	resp, err := server.HandleMessage(msg)
	require.NoError(t, err)
	require.NotNil(t, resp)

	errResp := resp["error"].(map[string]any)
	require.Equal(t, -32600, errResp["code"])
}

func TestServer_HandleMessage_ToolNotFound(t *testing.T) {
	server := NewServerP("test-server")

	msg := map[string]any{
		"jsonrpc": "2.0",
		"id":      float64(1),
		"method":  "tools/call",
		"params": map[string]any{
			"name":      "nonexistent",
			"arguments": map[string]any{},
		},
	}

	resp, err := server.HandleMessage(msg)
	require.NoError(t, err)
	require.NotNil(t, resp)

	errResp := resp["error"].(map[string]any)
	require.Equal(t, -32603, errResp["code"])
}

func TestServer_HandleMessage_MissingToolName(t *testing.T) {
	server := NewServerP("test-server")

	msg := map[string]any{
		"jsonrpc": "2.0",
		"id":      float64(1),
		"method":  "tools/call",
		"params": map[string]any{
			"arguments": map[string]any{},
		},
	}

	resp, err := server.HandleMessage(msg)
	require.NoError(t, err)
	require.NotNil(t, resp)

	errResp := resp["error"].(map[string]any)
	require.Equal(t, -32602, errResp["code"])
}

func TestServer_HandleMessage_MissingParams(t *testing.T) {
	server := NewServerP("test-server")

	msg := map[string]any{
		"jsonrpc": "2.0",
		"id":      float64(1),
		"method":  "tools/call",
	}

	resp, err := server.HandleMessage(msg)
	require.NoError(t, err)
	require.NotNil(t, resp)

	errResp := resp["error"].(map[string]any)
	require.Equal(t, -32602, errResp["code"])
}

// ============================================================
// Mock 类型
// ============================================================

type mockPermissionChecker struct {
	allow bool
	err   error
}

func (m *mockPermissionChecker) Check(ctx context.Context, toolName string, params map[string]any) PermissionDecision {
	if m.err != nil {
		return PermissionDecisionDeny
	}
	if m.allow {
		return PermissionDecisionAllow
	}
	return PermissionDecisionAsk
}

// ============================================================
// SessionContext 测试
// ============================================================

func TestSessionContext_AddTerminal(t *testing.T) {
	ctx := NewSessionContext("test-session", "/tmp")
	ctx.AddTerminal("term-1", "ls -la")

	terminal := ctx.GetTerminal("term-1")
	require.NotNil(t, terminal)
	require.Equal(t, "term-1", terminal.ID)
	require.Equal(t, "ls -la", terminal.Cmd)
	require.Equal(t, TerminalStatusRunning, terminal.Status)
}

func TestSessionContext_GetTerminal_NotFound(t *testing.T) {
	ctx := NewSessionContext("test-session", "/tmp")
	terminal := ctx.GetTerminal("nonexistent")
	require.Nil(t, terminal)
}

func TestSessionContext_UpdateTerminalOutput(t *testing.T) {
	ctx := NewSessionContext("test-session", "/tmp")
	ctx.AddTerminal("term-1", "ls")

	ctx.UpdateTerminalOutput("term-1", "line1\n")
	ctx.UpdateTerminalOutput("term-1", "line2\n")

	terminal := ctx.GetTerminal("term-1")
	require.Equal(t, "line1\nline2\n", terminal.Output)
}

func TestSessionContext_UpdateTerminalOutput_Nonexistent(t *testing.T) {
	ctx := NewSessionContext("test-session", "/tmp")
	// 不应该 panic
	ctx.UpdateTerminalOutput("nonexistent", "output")
}

func TestSessionContext_UpdateTerminalStatus(t *testing.T) {
	ctx := NewSessionContext("test-session", "/tmp")
	ctx.AddTerminal("term-1", "ls")

	ctx.UpdateTerminalStatus("term-1", TerminalStatusCompleted)

	terminal := ctx.GetTerminal("term-1")
	require.Equal(t, TerminalStatusCompleted, terminal.Status)
}

func TestSessionContext_UpdateTerminalStatus_Nonexistent(t *testing.T) {
	ctx := NewSessionContext("test-session", "/tmp")
	// 不应该 panic
	ctx.UpdateTerminalStatus("nonexistent", TerminalStatusCompleted)
}

func TestSessionContext_RemoveTerminal(t *testing.T) {
	ctx := NewSessionContext("test-session", "/tmp")
	ctx.AddTerminal("term-1", "ls")

	ctx.RemoveTerminal("term-1")

	terminal := ctx.GetTerminal("term-1")
	require.Nil(t, terminal)
}

func TestSessionContext_ClearTerminals(t *testing.T) {
	ctx := NewSessionContext("test-session", "/tmp")
	ctx.AddTerminal("term-1", "ls")
	ctx.AddTerminal("term-2", "pwd")

	ctx.ClearTerminals()

	require.Nil(t, ctx.GetTerminal("term-1"))
	require.Nil(t, ctx.GetTerminal("term-2"))
}

// ============================================================
// formatResult 测试
// ============================================================

func TestServer_FormatResult_String(t *testing.T) {
	server := NewServerP("test-server")
	_ = server.RegisterTool(Tool{
		Name:        "test",
		Description: "test",
		Handler:     func(ctx context.Context, args map[string]any) (any, error) { return "hello", nil },
	})

	msg := map[string]any{
		"jsonrpc": "2.0",
		"id":      float64(1),
		"method":  "tools/call",
		"params": map[string]any{
			"name":      "test",
			"arguments": map[string]any{},
		},
	}

	resp, err := server.HandleMessage(msg)
	require.NoError(t, err)

	result := resp["result"].(map[string]any)
	content := result["content"].([]map[string]any)
	require.Equal(t, "text", content[0]["type"])
	require.Equal(t, "hello", content[0]["text"])
}

func TestServer_FormatResult_MapWithText(t *testing.T) {
	server := NewServerP("test-server")
	_ = server.RegisterTool(Tool{
		Name:        "test",
		Description: "test",
		Handler:     func(ctx context.Context, args map[string]any) (any, error) { return map[string]any{"text": "my text"}, nil },
	})

	msg := map[string]any{
		"jsonrpc": "2.0",
		"id":      float64(1),
		"method":  "tools/call",
		"params": map[string]any{
			"name":      "test",
			"arguments": map[string]any{},
		},
	}

	resp, err := server.HandleMessage(msg)
	require.NoError(t, err)

	result := resp["result"].(map[string]any)
	content := result["content"].([]map[string]any)
	require.Equal(t, "my text", content[0]["text"])
}

func TestServer_FormatResult_Array(t *testing.T) {
	server := NewServerP("test-server")
	_ = server.RegisterTool(Tool{
		Name:        "test",
		Description: "test",
		Handler: func(ctx context.Context, args map[string]any) (any, error) {
			return []any{
				map[string]any{"type": "text", "text": "item1"},
				map[string]any{"type": "text", "text": "item2"},
			}, nil
		},
	})

	msg := map[string]any{
		"jsonrpc": "2.0",
		"id":      float64(1),
		"method":  "tools/call",
		"params": map[string]any{
			"name":      "test",
			"arguments": map[string]any{},
		},
	}

	resp, err := server.HandleMessage(msg)
	require.NoError(t, err)

	result := resp["result"].(map[string]any)
	content := result["content"].([]map[string]any)
	require.Len(t, content, 2)
}

func TestServer_FormatResult_Default(t *testing.T) {
	server := NewServerP("test-server")
	_ = server.RegisterTool(Tool{
		Name:        "test",
		Description: "test",
		Handler:     func(ctx context.Context, args map[string]any) (any, error) { return 123, nil }, // int 类型
	})

	msg := map[string]any{
		"jsonrpc": "2.0",
		"id":      float64(1),
		"method":  "tools/call",
		"params": map[string]any{
			"name":      "test",
			"arguments": map[string]any{},
		},
	}

	resp, err := server.HandleMessage(msg)
	require.NoError(t, err)

	result := resp["result"].(map[string]any)
	content := result["content"].([]map[string]any)
	require.Equal(t, "123", content[0]["text"])
}

// ============================================================
// 其他测试
// ============================================================

func TestNewServer_WithOptions(t *testing.T) {
	checker := &mockPermissionChecker{allow: true}
	sessionCtx := NewSessionContext("session-1", "/work")
	opts := NewServerOptions().WithPermissionChecker(checker).WithSessionContext(sessionCtx)

	server, err := NewServer("test-server", opts)
	require.NoError(t, err)
	require.NotNil(t, server)
	require.Equal(t, checker, server.permissionChecker)
	require.Equal(t, sessionCtx, server.sessionContext)
}

func TestServer_RegisterTools_Error(t *testing.T) {
	server := NewServerP("test-server")
	tools := []Tool{
		{
			Name:        "read",
			Description: "读取文件",
			Handler:     func(ctx context.Context, args map[string]any) (any, error) { return "read", nil },
		},
		{
			Name:        "", // 空名称，应该失败
			Description: "空名称工具",
			Handler:     func(ctx context.Context, args map[string]any) (any, error) { return "write", nil },
		},
	}

	err := server.RegisterTools(tools...)
	require.Error(t, err)
}

func TestServer_HandleMessage_ToolError(t *testing.T) {
	server := NewServerP("test-server")
	_ = server.RegisterTool(Tool{
		Name:        "fail",
		Description: "失败工具",
		Handler:     func(ctx context.Context, args map[string]any) (any, error) { return nil, fmt.Errorf("tool failed") },
	})

	msg := map[string]any{
		"jsonrpc": "2.0",
		"id":      float64(1),
		"method":  "tools/call",
		"params": map[string]any{
			"name":      "fail",
			"arguments": map[string]any{},
		},
	}

	resp, err := server.HandleMessage(msg)
	require.NoError(t, err)

	errResp := resp["error"].(map[string]any)
	require.Equal(t, -32603, errResp["code"])
	require.Contains(t, errResp["message"].(string), "tool failed")
}

func TestServer_HandleMessage_CallTool_NoArgs(t *testing.T) {
	server := NewServerP("test-server")
	_ = server.RegisterTool(Tool{
		Name:        "test",
		Description: "test",
		Handler:     func(ctx context.Context, args map[string]any) (any, error) { return "ok", nil },
	})

	// params 中没有 arguments 字段
	msg := map[string]any{
		"jsonrpc": "2.0",
		"id":      float64(1),
		"method":  "tools/call",
		"params": map[string]any{
			"name": "test",
		},
	}

	resp, err := server.HandleMessage(msg)
	require.NoError(t, err)

	result := resp["result"].(map[string]any)
	require.NotNil(t, result)
}

func TestServer_HandleMessage_ListTools_WithInputSchema(t *testing.T) {
	server := NewServerP("test-server")
	_ = server.RegisterTool(Tool{
		Name:        "read",
		Description: "读取文件",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"path": map[string]any{"type": "string"},
			},
		},
		Handler: func(ctx context.Context, args map[string]any) (any, error) { return "ok", nil },
	})

	msg := map[string]any{
		"jsonrpc": "2.0",
		"id":      float64(1),
		"method":  "tools/list",
	}

	resp, err := server.HandleMessage(msg)
	require.NoError(t, err)

	result := resp["result"].(map[string]any)
	tools := result["tools"].([]map[string]any)
	require.NotNil(t, tools[0]["inputSchema"])
}

func TestServer_ToSDKMCPServer(t *testing.T) {
	server := NewServerP("test-server")
	_ = server.RegisterTool(Tool{
		Name:        "add",
		Description: "加法运算",
		InputSchema: map[string]any{
			"type": "object",
		},
		Handler: func(ctx context.Context, args map[string]any) (any, error) {
			return map[string]any{"result": 1}, nil
		},
	})

	sdkServer, err := server.ToSDKMCPServer()
	require.NoError(t, err)
	require.NotNil(t, sdkServer)
	require.Equal(t, "test-server", sdkServer.Name())
}

func TestNewServerP_Panic(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("Expected panic for empty server name")
		}
	}()
	NewServerP("")
}

func TestServer_ToSDKMCPServer_EmptyTools(t *testing.T) {
	server := NewServerP("test-server")
	// 不注册任何工具，应该返回错误

	_, err := server.ToSDKMCPServer()
	require.Error(t, err)
}

func TestServer_ToSDKMCPServer_HandleMessage(t *testing.T) {
	server := NewServerP("test-server")
	_ = server.RegisterTool(Tool{
		Name:        "add",
		Description: "加法运算",
		InputSchema: map[string]any{
			"type": "object",
		},
		Handler: func(ctx context.Context, args map[string]any) (any, error) {
			a := args["a"].(float64)
			b := args["b"].(float64)
			return map[string]any{"result": a + b}, nil
		},
	})

	sdkServer, err := server.ToSDKMCPServer()
	require.NoError(t, err)

	// 通过 SDK 服务器调用工具
	msg := map[string]any{
		"jsonrpc": "2.0",
		"id":      float64(1),
		"method":  "tools/call",
		"params": map[string]any{
			"name": "add",
			"arguments": map[string]any{
				"a": float64(3),
				"b": float64(5),
			},
		},
	}

	resp, err := sdkServer.HandleMessage(msg)
	require.NoError(t, err)
	require.NotNil(t, resp)

	result := resp["result"].(map[string]any)
	content := result["content"].([]map[string]any)
	require.Len(t, content, 1)
}
