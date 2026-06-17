package acp

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	"claude-code-acp-go/internal/golden"

	acpsdk "github.com/coder/acp-go-sdk"
)

func TestMain(m *testing.M) {
	// 设置黄金文件目录为项目根目录下的 golden 目录
	golden.SetDir(filepath.Join("..", "..", "golden"))
	os.Exit(m.Run())
}

// ============================================================
// 核心协议: Initialize
// ============================================================

func TestGolden_Initialize(t *testing.T) {
	t.Run("initialize 基本流程", func(t *testing.T) {
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

		// 序列化响应
		respJSON, err := json.MarshalIndent(resp, "", "  ")
		require.NoError(t, err)

		// 与黄金文件比较
		golden.Compare(t, "initialize-response.json", respJSON)
	})
}

// ============================================================
// 核心协议: NewSession
// ============================================================

func TestGolden_NewSession(t *testing.T) {
	t.Run("session/new 基本流程", func(t *testing.T) {
		agent := NewAgent()
		ctx := context.Background()

		req := acpsdk.NewSessionRequest{
			Cwd: "/tmp/test-project",
		}

		resp, err := agent.NewSession(ctx, req)
		require.NoError(t, err)
		require.NotEmpty(t, resp.SessionId)

		// 验证 SessionId 格式
		require.Contains(t, string(resp.SessionId), "session-")
	})
}

// ============================================================
// SessionUpdate: AgentMessage
// ============================================================

func TestGolden_SessionUpdate(t *testing.T) {
	t.Run("SessionUpdate 消息格式", func(t *testing.T) {
		update := acpsdk.UpdateAgentMessageText("Hello, World!")

		updateJSON, err := json.MarshalIndent(update, "", "  ")
		require.NoError(t, err)

		golden.Compare(t, "session-update-agent-message.json", updateJSON)
	})
}

// ============================================================
// SessionUpdate: ToolCall
// ============================================================

func TestGolden_ToolCall(t *testing.T) {
	t.Run("ToolCall 开始消息格式", func(t *testing.T) {
		update := acpsdk.StartToolCall(
			"call-1",
			"Reading file",
			acpsdk.WithStartKind(acpsdk.ToolKindRead),
			acpsdk.WithStartStatus(acpsdk.ToolCallStatusPending),
			acpsdk.WithStartLocations([]acpsdk.ToolCallLocation{{Path: "/tmp/test.txt"}}),
		)

		updateJSON, err := json.MarshalIndent(update, "", "  ")
		require.NoError(t, err)

		golden.Compare(t, "tool-call-start.json", updateJSON)
	})
}

func TestGolden_ToolCallUpdate(t *testing.T) {
	t.Run("ToolCall 更新消息格式", func(t *testing.T) {
		update := acpsdk.UpdateToolCall(
			"call-1",
			acpsdk.WithUpdateStatus(acpsdk.ToolCallStatusCompleted),
			acpsdk.WithUpdateContent([]acpsdk.ToolCallContent{
				acpsdk.ToolContent(acpsdk.TextBlock("file content here")),
			}),
		)

		updateJSON, err := json.MarshalIndent(update, "", "  ")
		require.NoError(t, err)

		golden.Compare(t, "tool-call-update.json", updateJSON)
	})
}

// ============================================================
// SessionUpdate: AgentThought
// ============================================================

func TestGolden_AgentThought(t *testing.T) {
	t.Run("AgentThought 消息格式", func(t *testing.T) {
		update := acpsdk.UpdateAgentThoughtText("Let me analyze this code...")

		updateJSON, err := json.MarshalIndent(update, "", "  ")
		require.NoError(t, err)

		golden.Compare(t, "session-update-agent-thought.json", updateJSON)
	})
}

// ============================================================
// SessionUpdate: Plan
// ============================================================

func TestGolden_PlanUpdate(t *testing.T) {
	t.Run("Plan 更新消息格式", func(t *testing.T) {
		entries := []acpsdk.PlanEntry{
			{Content: "Read configuration file", Status: acpsdk.PlanEntryStatusCompleted},
			{Content: "Analyze code structure", Status: acpsdk.PlanEntryStatusInProgress},
			{Content: "Implement changes", Status: acpsdk.PlanEntryStatusPending},
		}
		update := acpsdk.UpdatePlan(entries...)

		updateJSON, err := json.MarshalIndent(update, "", "  ")
		require.NoError(t, err)

		golden.Compare(t, "session-update-plan.json", updateJSON)
	})
}

// ============================================================
// SessionUpdate: UserMessage
// ============================================================

func TestGolden_UserMessage(t *testing.T) {
	t.Run("UserMessage 消息格式", func(t *testing.T) {
		update := acpsdk.UpdateUserMessageText("Please help me with this code")

		updateJSON, err := json.MarshalIndent(update, "", "  ")
		require.NoError(t, err)

		golden.Compare(t, "session-update-user-message.json", updateJSON)
	})
}

// ============================================================
// ContentBlock 完整覆盖 (5种)
// ============================================================

func TestGolden_ContentBlock_Text(t *testing.T) {
	t.Run("Text 内容块格式", func(t *testing.T) {
		textBlock := acpsdk.TextBlock("Hello, this is a text message")
		textJSON, err := json.MarshalIndent(textBlock, "", "  ")
		require.NoError(t, err)

		golden.Compare(t, "content-block-text.json", textJSON)
	})
}

func TestGolden_ContentBlock_Image(t *testing.T) {
	t.Run("Image 内容块格式", func(t *testing.T) {
		imageBlock := acpsdk.ImageBlock("base64imagedata", "image/png")
		imageJSON, err := json.MarshalIndent(imageBlock, "", "  ")
		require.NoError(t, err)

		golden.Compare(t, "content-block-image.json", imageJSON)
	})
}

func TestGolden_ContentBlock_Audio(t *testing.T) {
	t.Run("Audio 内容块格式", func(t *testing.T) {
		audioBlock := acpsdk.AudioBlock("base64audiodata", "audio/wav")
		audioJSON, err := json.MarshalIndent(audioBlock, "", "  ")
		require.NoError(t, err)

		golden.Compare(t, "content-block-audio.json", audioJSON)
	})
}

func TestGolden_ContentBlock_ResourceLink(t *testing.T) {
	t.Run("ResourceLink 内容块格式", func(t *testing.T) {
		resourceLink := acpsdk.ResourceLinkBlock("config.json", "file:///project/config.json")
		linkJSON, err := json.MarshalIndent(resourceLink, "", "  ")
		require.NoError(t, err)

		golden.Compare(t, "content-block-resource-link.json", linkJSON)
	})
}

func TestGolden_ContentBlock_Resource(t *testing.T) {
	t.Run("Resource 内容块格式", func(t *testing.T) {
		// 使用 SDK helper 函数创建 Resource 内容块
		textContent := acpsdk.TextResourceContents{
			Uri:  "file:///project/main.go",
			Text: "package main\n\nfunc main() {}",
		}
		resource := acpsdk.EmbeddedResourceResource{
			TextResourceContents: &textContent,
		}
		resourceBlock := acpsdk.ResourceBlock(resource)
		resourceJSON, err := json.MarshalIndent(resourceBlock, "", "  ")
		require.NoError(t, err)

		golden.Compare(t, "content-block-resource.json", resourceJSON)
	})
}

// ============================================================
// ToolKind 完整覆盖 (10种)
// ============================================================

func TestGolden_ToolKind_Delete(t *testing.T) {
	update := acpsdk.StartToolCall(
		acpsdk.ToolCallId("call-delete"),
		"Deleting file",
		acpsdk.WithStartKind(acpsdk.ToolKindDelete),
		acpsdk.WithStartStatus(acpsdk.ToolCallStatusPending),
		acpsdk.WithStartLocations([]acpsdk.ToolCallLocation{{Path: "/tmp/old.txt"}}),
	)
	updateJSON, _ := json.MarshalIndent(update, "", "  ")
	golden.Compare(t, "tool-kind-delete.json", updateJSON)
}

func TestGolden_ToolKind_Move(t *testing.T) {
	update := acpsdk.StartToolCall(
		acpsdk.ToolCallId("call-move"),
		"Moving file",
		acpsdk.WithStartKind(acpsdk.ToolKindMove),
		acpsdk.WithStartStatus(acpsdk.ToolCallStatusPending),
		acpsdk.WithStartLocations([]acpsdk.ToolCallLocation{
			{Path: "/tmp/old.txt"},
			{Path: "/tmp/new.txt"},
		}),
	)
	updateJSON, _ := json.MarshalIndent(update, "", "  ")
	golden.Compare(t, "tool-kind-move.json", updateJSON)
}

func TestGolden_ToolKind_Search(t *testing.T) {
	update := acpsdk.StartToolCall(
		acpsdk.ToolCallId("call-search"),
		"Searching in files",
		acpsdk.WithStartKind(acpsdk.ToolKindSearch),
		acpsdk.WithStartStatus(acpsdk.ToolCallStatusPending),
	)
	updateJSON, _ := json.MarshalIndent(update, "", "  ")
	golden.Compare(t, "tool-kind-search.json", updateJSON)
}

func TestGolden_ToolKind_Execute(t *testing.T) {
	update := acpsdk.StartToolCall(
		acpsdk.ToolCallId("call-execute"),
		"Executing command",
		acpsdk.WithStartKind(acpsdk.ToolKindExecute),
		acpsdk.WithStartStatus(acpsdk.ToolCallStatusPending),
	)
	updateJSON, _ := json.MarshalIndent(update, "", "  ")
	golden.Compare(t, "tool-kind-execute.json", updateJSON)
}

func TestGolden_ToolKind_Think(t *testing.T) {
	update := acpsdk.StartToolCall(
		acpsdk.ToolCallId("call-think"),
		"Thinking",
		acpsdk.WithStartKind(acpsdk.ToolKindThink),
		acpsdk.WithStartStatus(acpsdk.ToolCallStatusPending),
	)
	updateJSON, _ := json.MarshalIndent(update, "", "  ")
	golden.Compare(t, "tool-kind-think.json", updateJSON)
}

func TestGolden_ToolKind_Fetch(t *testing.T) {
	update := acpsdk.StartToolCall(
		acpsdk.ToolCallId("call-fetch"),
		"Fetching URL",
		acpsdk.WithStartKind(acpsdk.ToolKindFetch),
		acpsdk.WithStartStatus(acpsdk.ToolCallStatusPending),
	)
	updateJSON, _ := json.MarshalIndent(update, "", "  ")
	golden.Compare(t, "tool-kind-fetch.json", updateJSON)
}

func TestGolden_ToolKind_SwitchMode(t *testing.T) {
	update := acpsdk.StartToolCall(
		acpsdk.ToolCallId("call-switch-mode"),
		"Switching mode",
		acpsdk.WithStartKind(acpsdk.ToolKindSwitchMode),
		acpsdk.WithStartStatus(acpsdk.ToolCallStatusPending),
	)
	updateJSON, _ := json.MarshalIndent(update, "", "  ")
	golden.Compare(t, "tool-kind-switch-mode.json", updateJSON)
}

func TestGolden_ToolKind_Other(t *testing.T) {
	update := acpsdk.StartToolCall(
		acpsdk.ToolCallId("call-other"),
		"Other operation",
		acpsdk.WithStartKind(acpsdk.ToolKindOther),
		acpsdk.WithStartStatus(acpsdk.ToolCallStatusPending),
	)
	updateJSON, _ := json.MarshalIndent(update, "", "  ")
	golden.Compare(t, "tool-kind-other.json", updateJSON)
}

func TestGolden_ToolKind_Edit(t *testing.T) {
	update := acpsdk.StartEditToolCall(
		acpsdk.ToolCallId("call-edit-1"),
		"Editing configuration file",
		"/project/config.json",
		map[string]any{"new_value": "updated"},
	)
	updateJSON, _ := json.MarshalIndent(update, "", "  ")
	golden.Compare(t, "tool-call-edit.json", updateJSON)
}

// ============================================================
// ToolCallStatus 完整覆盖 (4种)
// ============================================================

func TestGolden_ToolCallStatus_InProgress(t *testing.T) {
	update := acpsdk.UpdateToolCall(
		acpsdk.ToolCallId("call-in-progress"),
		acpsdk.WithUpdateStatus(acpsdk.ToolCallStatusInProgress),
		acpsdk.WithUpdateTitle("In progress operation"),
	)
	updateJSON, _ := json.MarshalIndent(update, "", "  ")
	golden.Compare(t, "tool-status-in-progress.json", updateJSON)
}

func TestGolden_ToolCallStatus_Failed(t *testing.T) {
	update := acpsdk.UpdateToolCall(
		acpsdk.ToolCallId("call-failed"),
		acpsdk.WithUpdateStatus(acpsdk.ToolCallStatusFailed),
		acpsdk.WithUpdateTitle("Failed operation"),
	)
	updateJSON, _ := json.MarshalIndent(update, "", "  ")
	golden.Compare(t, "tool-status-failed.json", updateJSON)
}

// ============================================================
// 请求类型
// ============================================================

func TestGolden_PromptRequest(t *testing.T) {
	t.Run("Prompt 请求格式", func(t *testing.T) {
		req := acpsdk.PromptRequest{
			SessionId: "session-test-123",
			Prompt: []acpsdk.ContentBlock{
				acpsdk.TextBlock("Please analyze this code"),
				acpsdk.ResourceLinkBlock("main.go", "file:///project/main.go"),
			},
		}

		reqJSON, err := json.MarshalIndent(req, "", "  ")
		require.NoError(t, err)

		golden.Compare(t, "prompt-request.json", reqJSON)
	})
}

func TestGolden_LoadSessionRequest(t *testing.T) {
	t.Run("LoadSession 请求格式", func(t *testing.T) {
		req := acpsdk.LoadSessionRequest{
			SessionId: "session-12345",
		}

		reqJSON, err := json.MarshalIndent(req, "", "  ")
		require.NoError(t, err)

		golden.Compare(t, "load-session-request.json", reqJSON)
	})
}

func TestGolden_AuthenticateRequest(t *testing.T) {
	t.Run("Authenticate 请求格式", func(t *testing.T) {
		req := acpsdk.AuthenticateRequest{
			MethodId: "auth-123",
		}

		reqJSON, err := json.MarshalIndent(req, "", "  ")
		require.NoError(t, err)

		golden.Compare(t, "authenticate-request.json", reqJSON)
	})
}

func TestGolden_SetSessionModeRequest(t *testing.T) {
	t.Run("SetSessionMode 请求格式", func(t *testing.T) {
		req := acpsdk.SetSessionModeRequest{
			SessionId: "session-123",
			ModeId:    acpsdk.SessionModeId("code"),
		}

		reqJSON, err := json.MarshalIndent(req, "", "  ")
		require.NoError(t, err)

		golden.Compare(t, "set-session-mode-request.json", reqJSON)
	})
}

// ============================================================
// 通知类型
// ============================================================

func TestGolden_CancelNotification(t *testing.T) {
	t.Run("Cancel 通知格式", func(t *testing.T) {
		notification := acpsdk.CancelNotification{
			SessionId: "session-test-123",
		}

		notifJSON, err := json.MarshalIndent(notification, "", "  ")
		require.NoError(t, err)

		golden.Compare(t, "cancel-notification.json", notifJSON)
	})
}

func TestGolden_SessionAvailableCommandsUpdate(t *testing.T) {
	t.Run("AvailableCommands 更新格式", func(t *testing.T) {
		update := acpsdk.SessionAvailableCommandsUpdate{
			SessionUpdate: "available_commands",
			AvailableCommands: []acpsdk.AvailableCommand{
				{
					Name:        "test",
					Description: "Test command",
				},
			},
		}

		updateJSON, err := json.MarshalIndent(update, "", "  ")
		require.NoError(t, err)

		golden.Compare(t, "session-available-commands-update.json", updateJSON)
	})
}

func TestGolden_SessionCurrentModeUpdate(t *testing.T) {
	t.Run("CurrentMode 更新格式", func(t *testing.T) {
		update := acpsdk.SessionCurrentModeUpdate{
			SessionUpdate: "current_mode",
			CurrentModeId: acpsdk.SessionModeId("code"),
		}

		updateJSON, err := json.MarshalIndent(update, "", "  ")
		require.NoError(t, err)

		golden.Compare(t, "session-current-mode-update.json", updateJSON)
	})
}

// ============================================================
// 权限请求
// ============================================================

func TestGolden_RequestPermissionResponse(t *testing.T) {
	t.Run("RequestPermission 响应格式", func(t *testing.T) {
		resp := acpsdk.RequestPermissionResponse{
			Outcome: acpsdk.RequestPermissionOutcome{
				Selected: &acpsdk.RequestPermissionOutcomeSelected{
					OptionId: "allow",
				},
			},
		}

		respJSON, err := json.MarshalIndent(resp, "", "  ")
		require.NoError(t, err)

		golden.Compare(t, "request-permission-response.json", respJSON)
	})
}

func TestGolden_RequestPermissionResponse_Cancelled(t *testing.T) {
	t.Run("RequestPermission 取消响应格式", func(t *testing.T) {
		resp := acpsdk.RequestPermissionResponse{
			Outcome: acpsdk.RequestPermissionOutcome{
				Cancelled: &acpsdk.RequestPermissionOutcomeCancelled{},
			},
		}

		respJSON, err := json.MarshalIndent(resp, "", "  ")
		require.NoError(t, err)

		golden.Compare(t, "request-permission-response-cancelled.json", respJSON)
	})
}

// ============================================================
// 终端相关
// ============================================================

func TestGolden_CreateTerminalRequest(t *testing.T) {
	t.Run("CreateTerminal 请求格式", func(t *testing.T) {
		req := acpsdk.CreateTerminalRequest{
			SessionId: "session-123",
			Command:   "/bin/bash",
			Args:      []string{"-l"},
			Env: []acpsdk.EnvVariable{
				{Name: "TERM", Value: "xterm-256color"},
			},
		}

		reqJSON, err := json.MarshalIndent(req, "", "  ")
		require.NoError(t, err)

		golden.Compare(t, "create-terminal-request.json", reqJSON)
	})
}

func TestGolden_TerminalOutputRequest(t *testing.T) {
	t.Run("TerminalOutput 请求格式", func(t *testing.T) {
		req := acpsdk.TerminalOutputRequest{
			SessionId:  "session-123",
			TerminalId: "terminal-1",
		}

		reqJSON, err := json.MarshalIndent(req, "", "  ")
		require.NoError(t, err)

		golden.Compare(t, "terminal-output-request.json", reqJSON)
	})
}

// ============================================================
// 文件操作
// ============================================================

func TestGolden_ReadTextFileRequest(t *testing.T) {
	t.Run("ReadTextFile 请求格式", func(t *testing.T) {
		req := acpsdk.ReadTextFileRequest{
			SessionId: "session-123",
			Path:      "/project/main.go",
		}

		reqJSON, err := json.MarshalIndent(req, "", "  ")
		require.NoError(t, err)

		golden.Compare(t, "read-text-file-request.json", reqJSON)
	})
}

func TestGolden_WriteTextFileRequest(t *testing.T) {
	t.Run("WriteTextFile 请求格式", func(t *testing.T) {
		req := acpsdk.WriteTextFileRequest{
			SessionId: "session-123",
			Path:      "/project/main.go",
			Content:   "package main\n\nfunc main() {}\n",
		}

		reqJSON, err := json.MarshalIndent(req, "", "  ")
		require.NoError(t, err)

		golden.Compare(t, "write-text-file-request.json", reqJSON)
	})
}

// ============================================================
// MCP 服务器配置
// ============================================================

func TestGolden_McpServer_Stdio(t *testing.T) {
	t.Run("MCP Server Stdio 配置格式", func(t *testing.T) {
		server := acpsdk.McpServer{
			Stdio: &acpsdk.McpServerStdio{
				Command: "mcp-filesystem",
				Args:    []string{"/home/user"},
				Env: []acpsdk.EnvVariable{
					{Name: "DEBUG", Value: "1"},
				},
			},
		}

		serverJSON, err := json.MarshalIndent(server, "", "  ")
		require.NoError(t, err)

		golden.Compare(t, "mcp-server-stdio.json", serverJSON)
	})
}

func TestGolden_McpServer_Http(t *testing.T) {
	t.Run("MCP Server HTTP 配置格式", func(t *testing.T) {
		server := acpsdk.McpServer{
			Http: &acpsdk.McpServerHttpInline{
				Url: "http://localhost:8080/mcp",
				Headers: []acpsdk.HttpHeader{
					{Name: "Authorization", Value: "Bearer token"},
				},
			},
		}

		serverJSON, err := json.MarshalIndent(server, "", "  ")
		require.NoError(t, err)

		golden.Compare(t, "mcp-server-http.json", serverJSON)
	})
}

func TestGolden_McpServer_Sse(t *testing.T) {
	t.Run("MCP Server SSE 配置格式", func(t *testing.T) {
		server := acpsdk.McpServer{
			Sse: &acpsdk.McpServerSseInline{
				Url: "http://localhost:8080/sse",
			},
		}

		serverJSON, err := json.MarshalIndent(server, "", "  ")
		require.NoError(t, err)

		golden.Compare(t, "mcp-server-sse.json", serverJSON)
	})
}

// ============================================================
// 错误类型
// ============================================================

func TestGolden_Error(t *testing.T) {
	t.Run("Error 格式", func(t *testing.T) {
		err := acpsdk.Error{
			Code:    acpsdk.ErrorCode{Other: &acpsdk.ErrorCodeOther{}},
			Message: "Something went wrong",
			Data:    map[string]any{"detail": "Additional info"},
		}

		errJSON, marshalErr := json.MarshalIndent(err, "", "  ")
		require.NoError(t, marshalErr)

		golden.Compare(t, "error.json", errJSON)
	})
}

// ============================================================
// 响应类型 - 完整覆盖
// ============================================================

func TestGolden_PromptResponse(t *testing.T) {
	t.Run("Prompt 响应格式", func(t *testing.T) {
		resp := acpsdk.PromptResponse{
			StopReason: acpsdk.StopReasonEndTurn,
		}

		respJSON, err := json.MarshalIndent(resp, "", "  ")
		require.NoError(t, err)

		golden.Compare(t, "prompt-response.json", respJSON)
	})
}

func TestGolden_PromptResponse_MaxTokens(t *testing.T) {
	t.Run("Prompt 响应 MaxTokens", func(t *testing.T) {
		resp := acpsdk.PromptResponse{
			StopReason: acpsdk.StopReasonMaxTokens,
		}

		respJSON, err := json.MarshalIndent(resp, "", "  ")
		require.NoError(t, err)

		golden.Compare(t, "prompt-response-max-tokens.json", respJSON)
	})
}

func TestGolden_PromptResponse_Cancelled(t *testing.T) {
	t.Run("Prompt 响应 Cancelled", func(t *testing.T) {
		resp := acpsdk.PromptResponse{
			StopReason: acpsdk.StopReasonCancelled,
		}

		respJSON, err := json.MarshalIndent(resp, "", "  ")
		require.NoError(t, err)

		golden.Compare(t, "prompt-response-cancelled.json", respJSON)
	})
}

func TestGolden_CreateTerminalResponse(t *testing.T) {
	t.Run("CreateTerminal 响应格式", func(t *testing.T) {
		resp := acpsdk.CreateTerminalResponse{
			TerminalId: "terminal-123",
		}

		respJSON, err := json.MarshalIndent(resp, "", "  ")
		require.NoError(t, err)

		golden.Compare(t, "create-terminal-response.json", respJSON)
	})
}

func TestGolden_TerminalOutputResponse(t *testing.T) {
	t.Run("TerminalOutput 响应格式", func(t *testing.T) {
		resp := acpsdk.TerminalOutputResponse{
			Output: "terminal output here",
		}

		respJSON, err := json.MarshalIndent(resp, "", "  ")
		require.NoError(t, err)

		golden.Compare(t, "terminal-output-response.json", respJSON)
	})
}

func TestGolden_ReadTextFileResponse(t *testing.T) {
	t.Run("ReadTextFile 响应格式", func(t *testing.T) {
		resp := acpsdk.ReadTextFileResponse{
			Content: "package main\n\nfunc main() {}\n",
		}

		respJSON, err := json.MarshalIndent(resp, "", "  ")
		require.NoError(t, err)

		golden.Compare(t, "read-text-file-response.json", respJSON)
	})
}

func TestGolden_WriteTextFileResponse(t *testing.T) {
	t.Run("WriteTextFile 响应格式", func(t *testing.T) {
		resp := acpsdk.WriteTextFileResponse{}

		respJSON, err := json.MarshalIndent(resp, "", "  ")
		require.NoError(t, err)

		golden.Compare(t, "write-text-file-response.json", respJSON)
	})
}

func TestGolden_ReleaseTerminalRequest(t *testing.T) {
	t.Run("ReleaseTerminal 请求格式", func(t *testing.T) {
		req := acpsdk.ReleaseTerminalRequest{
			SessionId:  "session-123",
			TerminalId: "terminal-1",
		}

		reqJSON, err := json.MarshalIndent(req, "", "  ")
		require.NoError(t, err)

		golden.Compare(t, "release-terminal-request.json", reqJSON)
	})
}

func TestGolden_ReleaseTerminalResponse(t *testing.T) {
	t.Run("ReleaseTerminal 响应格式", func(t *testing.T) {
		resp := acpsdk.ReleaseTerminalResponse{}

		respJSON, err := json.MarshalIndent(resp, "", "  ")
		require.NoError(t, err)

		golden.Compare(t, "release-terminal-response.json", respJSON)
	})
}

func TestGolden_WaitForTerminalExitRequest(t *testing.T) {
	t.Run("WaitForTerminalExit 请求格式", func(t *testing.T) {
		req := acpsdk.WaitForTerminalExitRequest{
			SessionId:  "session-123",
			TerminalId: "terminal-1",
		}

		reqJSON, err := json.MarshalIndent(req, "", "  ")
		require.NoError(t, err)

		golden.Compare(t, "wait-for-terminal-exit-request.json", reqJSON)
	})
}

func TestGolden_WaitForTerminalExitResponse(t *testing.T) {
	t.Run("WaitForTerminalExit 响应格式", func(t *testing.T) {
		exitCode := 0
		resp := acpsdk.WaitForTerminalExitResponse{
			ExitCode: &exitCode,
		}

		respJSON, err := json.MarshalIndent(resp, "", "  ")
		require.NoError(t, err)

		golden.Compare(t, "wait-for-terminal-exit-response.json", respJSON)
	})
}

func TestGolden_KillTerminalCommandRequest(t *testing.T) {
	t.Run("KillTerminalCommand 请求格式", func(t *testing.T) {
		req := acpsdk.KillTerminalCommandRequest{
			SessionId:  "session-123",
			TerminalId: "terminal-1",
		}

		reqJSON, err := json.MarshalIndent(req, "", "  ")
		require.NoError(t, err)

		golden.Compare(t, "kill-terminal-command-request.json", reqJSON)
	})
}

func TestGolden_KillTerminalCommandResponse(t *testing.T) {
	t.Run("KillTerminalCommand 响应格式", func(t *testing.T) {
		resp := acpsdk.KillTerminalCommandResponse{}

		respJSON, err := json.MarshalIndent(resp, "", "  ")
		require.NoError(t, err)

		golden.Compare(t, "kill-terminal-command-response.json", respJSON)
	})
}

func TestGolden_SetSessionModeResponse(t *testing.T) {
	t.Run("SetSessionMode 响应格式", func(t *testing.T) {
		resp := acpsdk.SetSessionModeResponse{}

		respJSON, err := json.MarshalIndent(resp, "", "  ")
		require.NoError(t, err)

		golden.Compare(t, "set-session-mode-response.json", respJSON)
	})
}

// ============================================================
// 会话模式相关
// ============================================================

func TestGolden_SessionMode(t *testing.T) {
	t.Run("SessionMode 格式", func(t *testing.T) {
		desc := "Default coding mode"
		mode := acpsdk.SessionMode{
			Id:          acpsdk.SessionModeId("code"),
			Name:        "Code Mode",
			Description: &desc,
		}

		modeJSON, err := json.MarshalIndent(mode, "", "  ")
		require.NoError(t, err)

		golden.Compare(t, "session-mode.json", modeJSON)
	})
}

func TestGolden_SessionModeState(t *testing.T) {
	t.Run("SessionModeState 格式", func(t *testing.T) {
		desc := "Coding mode"
		state := acpsdk.SessionModeState{
			CurrentModeId: acpsdk.SessionModeId("code"),
			AvailableModes: []acpsdk.SessionMode{
				{
					Id:          acpsdk.SessionModeId("code"),
					Name:        "Code",
					Description: &desc,
				},
			},
		}

		stateJSON, err := json.MarshalIndent(state, "", "  ")
		require.NoError(t, err)

		golden.Compare(t, "session-mode-state.json", stateJSON)
	})
}

// ============================================================
// 终端对象
// ============================================================

func TestGolden_Terminal(t *testing.T) {
	t.Run("Terminal 格式", func(t *testing.T) {
		terminal := acpsdk.Terminal{
			TerminalId: "terminal-123",
		}

		terminalJSON, err := json.MarshalIndent(terminal, "", "  ")
		require.NoError(t, err)

		golden.Compare(t, "terminal.json", terminalJSON)
	})
}

// ============================================================
// Diff 类型
// ============================================================

func TestGolden_Diff(t *testing.T) {
	t.Run("Diff 格式", func(t *testing.T) {
		oldText := "old content"
		diff := acpsdk.Diff{
			OldText: &oldText,
			NewText: "new content",
		}

		diffJSON, err := json.MarshalIndent(diff, "", "  ")
		require.NoError(t, err)

		golden.Compare(t, "diff.json", diffJSON)
	})
}

// ============================================================
// ToolCallContent 变体
// ============================================================

func TestGolden_ToolCallContent_Diff(t *testing.T) {
	t.Run("ToolCallContent Diff 格式", func(t *testing.T) {
		oldText := "old text"
		content := acpsdk.ToolCallContent{
			Diff: &acpsdk.ToolCallContentDiff{
				OldText: &oldText,
				NewText: "new text",
			},
		}

		contentJSON, err := json.MarshalIndent(content, "", "  ")
		require.NoError(t, err)

		golden.Compare(t, "tool-call-content-diff.json", contentJSON)
	})
}

func TestGolden_ToolCallContent_Terminal(t *testing.T) {
	t.Run("ToolCallContent Terminal 格式", func(t *testing.T) {
		content := acpsdk.ToolCallContent{
			Terminal: &acpsdk.ToolCallContentTerminal{
				TerminalId: "terminal-1",
			},
		}

		contentJSON, err := json.MarshalIndent(content, "", "  ")
		require.NoError(t, err)

		golden.Compare(t, "tool-call-content-terminal.json", contentJSON)
	})
}

// ============================================================
// 权限请求完整格式
// ============================================================

func TestGolden_RequestPermissionRequest(t *testing.T) {
	t.Run("RequestPermission 请求格式", func(t *testing.T) {
		req := acpsdk.RequestPermissionRequest{
			SessionId: "session-123",
			Options: []acpsdk.PermissionOption{
				{Kind: acpsdk.PermissionOptionKindAllowOnce, Name: "Allow Once", OptionId: "allow-once"},
				{Kind: acpsdk.PermissionOptionKindRejectOnce, Name: "Reject Once", OptionId: "reject-once"},
			},
			ToolCall: acpsdk.ToolCallUpdate{
				ToolCallId: acpsdk.ToolCallId("call-1"),
			},
		}

		reqJSON, err := json.MarshalIndent(req, "", "  ")
		require.NoError(t, err)

		golden.Compare(t, "request-permission-request.json", reqJSON)
	})
}

// ============================================================
// 错误码类型 - 完整覆盖
// ============================================================

func TestGolden_ErrorCode_ParseError(t *testing.T) {
	t.Run("ParseError 错误码", func(t *testing.T) {
		err := acpsdk.Error{
			Code:    acpsdk.ErrorCode{ParseError: &acpsdk.ErrorCodeParseError{}},
			Message: "Parse error",
		}

		errJSON, marshalErr := json.MarshalIndent(err, "", "  ")
		require.NoError(t, marshalErr)

		golden.Compare(t, "error-code-parse-error.json", errJSON)
	})
}

func TestGolden_ErrorCode_InvalidRequest(t *testing.T) {
	t.Run("InvalidRequest 错误码", func(t *testing.T) {
		err := acpsdk.Error{
			Code:    acpsdk.ErrorCode{InvalidRequest: &acpsdk.ErrorCodeInvalidRequest{}},
			Message: "Invalid request",
		}

		errJSON, marshalErr := json.MarshalIndent(err, "", "  ")
		require.NoError(t, marshalErr)

		golden.Compare(t, "error-code-invalid-request.json", errJSON)
	})
}

func TestGolden_ErrorCode_MethodNotFound(t *testing.T) {
	t.Run("MethodNotFound 错误码", func(t *testing.T) {
		err := acpsdk.Error{
			Code:    acpsdk.ErrorCode{MethodNotFound: &acpsdk.ErrorCodeMethodNotFound{}},
			Message: "Method not found",
		}

		errJSON, marshalErr := json.MarshalIndent(err, "", "  ")
		require.NoError(t, marshalErr)

		golden.Compare(t, "error-code-method-not-found.json", errJSON)
	})
}

func TestGolden_ErrorCode_InvalidParams(t *testing.T) {
	t.Run("InvalidParams 错误码", func(t *testing.T) {
		err := acpsdk.Error{
			Code:    acpsdk.ErrorCode{InvalidParams: &acpsdk.ErrorCodeInvalidParams{}},
			Message: "Invalid params",
		}

		errJSON, marshalErr := json.MarshalIndent(err, "", "  ")
		require.NoError(t, marshalErr)

		golden.Compare(t, "error-code-invalid-params.json", errJSON)
	})
}

func TestGolden_ErrorCode_InternalError(t *testing.T) {
	t.Run("InternalError 错误码", func(t *testing.T) {
		err := acpsdk.Error{
			Code:    acpsdk.ErrorCode{InternalError: &acpsdk.ErrorCodeInternalError{}},
			Message: "Internal error",
		}

		errJSON, marshalErr := json.MarshalIndent(err, "", "  ")
		require.NoError(t, marshalErr)

		golden.Compare(t, "error-code-internal-error.json", errJSON)
	})
}

func TestGolden_ErrorCode_AuthenticationRequired(t *testing.T) {
	t.Run("AuthenticationRequired 错误码", func(t *testing.T) {
		err := acpsdk.Error{
			Code:    acpsdk.ErrorCode{AuthenticationRequired: &acpsdk.ErrorCodeAuthenticationRequired{}},
			Message: "Authentication required",
		}

		errJSON, marshalErr := json.MarshalIndent(err, "", "  ")
		require.NoError(t, marshalErr)

		golden.Compare(t, "error-code-authentication-required.json", errJSON)
	})
}

func TestGolden_ErrorCode_ResourceNotFound(t *testing.T) {
	t.Run("ResourceNotFound 错误码", func(t *testing.T) {
		err := acpsdk.Error{
			Code:    acpsdk.ErrorCode{ResourceNotFound: &acpsdk.ErrorCodeResourceNotFound{}},
			Message: "Resource not found",
		}

		errJSON, marshalErr := json.MarshalIndent(err, "", "  ")
		require.NoError(t, marshalErr)

		golden.Compare(t, "error-code-resource-not-found.json", errJSON)
	})
}

// ============================================================
// ContentChunk 类型
// ============================================================

func TestGolden_ContentChunk(t *testing.T) {
	t.Run("ContentChunk 格式", func(t *testing.T) {
		chunk := acpsdk.ContentChunk{
			Content: acpsdk.TextBlock("chunk content"),
		}

		chunkJSON, err := json.MarshalIndent(chunk, "", "  ")
		require.NoError(t, err)

		golden.Compare(t, "content-chunk.json", chunkJSON)
	})
}

// ============================================================
// AvailableCommand 完整格式
// ============================================================

func TestGolden_AvailableCommand(t *testing.T) {
	t.Run("AvailableCommand 完整格式", func(t *testing.T) {
		cmd := acpsdk.AvailableCommand{
			Name:        "commit",
			Description: "Create a git commit",
			Input: &acpsdk.AvailableCommandInput{
				Unstructured: &acpsdk.UnstructuredCommandInput{
					Hint: "Create a commit with the staged changes",
				},
			},
		}

		cmdJSON, err := json.MarshalIndent(cmd, "", "  ")
		require.NoError(t, err)

		golden.Compare(t, "available-command.json", cmdJSON)
	})
}

// ============================================================
// ResourceLink 独立类型
// ============================================================

func TestGolden_ResourceLink(t *testing.T) {
	t.Run("ResourceLink 格式", func(t *testing.T) {
		link := acpsdk.ResourceLink{
			Name: "config.json",
			Uri:  "file:///project/config.json",
		}

		linkJSON, err := json.MarshalIndent(link, "", "  ")
		require.NoError(t, err)

		golden.Compare(t, "resource-link.json", linkJSON)
	})
}

// ============================================================
// EmbeddedResource 类型
// ============================================================

func TestGolden_EmbeddedResource(t *testing.T) {
	t.Run("EmbeddedResource 格式", func(t *testing.T) {
		res := acpsdk.EmbeddedResource{
			Resource: acpsdk.EmbeddedResourceResource{
				TextResourceContents: &acpsdk.TextResourceContents{
					Uri:  "file:///project/main.go",
					Text: "package main",
				},
			},
		}

		resJSON, err := json.MarshalIndent(res, "", "  ")
		require.NoError(t, err)

		golden.Compare(t, "embedded-resource.json", resJSON)
	})
}

// ============================================================
// BlobResourceContents 类型
// ============================================================

func TestGolden_BlobResourceContents(t *testing.T) {
	t.Run("BlobResourceContents 格式", func(t *testing.T) {
		mimeType := "image/png"
		blob := acpsdk.BlobResourceContents{
			Uri:      "file:///project/image.png",
			Blob:     "base64encodeddata",
			MimeType: &mimeType,
		}

		blobJSON, err := json.MarshalIndent(blob, "", "  ")
		require.NoError(t, err)

		golden.Compare(t, "blob-resource-contents.json", blobJSON)
	})
}

// ============================================================
// 能力类型 - 完整覆盖
// ============================================================

func TestGolden_AgentCapabilities(t *testing.T) {
	cap := acpsdk.AgentCapabilities{
		LoadSession: true,
		McpCapabilities: acpsdk.McpCapabilities{
			Http: true,
			Sse:  true,
		},
		PromptCapabilities: acpsdk.PromptCapabilities{
			Audio:           true,
			EmbeddedContext: true,
			Image:           true,
		},
	}
	capJSON, _ := json.MarshalIndent(cap, "", "  ")
	golden.Compare(t, "agent-capabilities.json", capJSON)
}

func TestGolden_ClientCapabilities(t *testing.T) {
	cap := acpsdk.ClientCapabilities{
		Terminal: true,
	}
	capJSON, _ := json.MarshalIndent(cap, "", "  ")
	golden.Compare(t, "client-capabilities.json", capJSON)
}

func TestGolden_McpCapabilities(t *testing.T) {
	cap := acpsdk.McpCapabilities{
		Http: true,
		Sse:  true,
	}
	capJSON, _ := json.MarshalIndent(cap, "", "  ")
	golden.Compare(t, "mcp-capabilities.json", capJSON)
}

func TestGolden_PromptCapabilities(t *testing.T) {
	cap := acpsdk.PromptCapabilities{
		Audio:           true,
		EmbeddedContext: true,
		Image:           true,
	}
	capJSON, _ := json.MarshalIndent(cap, "", "  ")
	golden.Compare(t, "prompt-capabilities.json", capJSON)
}

// ============================================================
// 基础类型
// ============================================================

func TestGolden_AuthMethod(t *testing.T) {
	desc := "OAuth authentication"
	method := acpsdk.AuthMethod{
		Id:          "oauth",
		Name:        "OAuth",
		Description: &desc,
	}
	methodJSON, _ := json.MarshalIndent(method, "", "  ")
	golden.Compare(t, "auth-method.json", methodJSON)
}

func TestGolden_Implementation(t *testing.T) {
	impl := acpsdk.Implementation{
		Name:    "claude-code-acp",
		Version: "1.0.0",
	}
	implJSON, _ := json.MarshalIndent(impl, "", "  ")
	golden.Compare(t, "implementation.json", implJSON)
}

func TestGolden_EnvVariable(t *testing.T) {
	env := acpsdk.EnvVariable{
		Name:  "PATH",
		Value: "/usr/bin:/bin",
	}
	envJSON, _ := json.MarshalIndent(env, "", "  ")
	golden.Compare(t, "env-variable.json", envJSON)
}

func TestGolden_HttpHeader(t *testing.T) {
	header := acpsdk.HttpHeader{
		Name:  "Authorization",
		Value: "Bearer token123",
	}
	headerJSON, _ := json.MarshalIndent(header, "", "  ")
	golden.Compare(t, "http-header.json", headerJSON)
}

// ============================================================
// 内容类型
// ============================================================

func TestGolden_TextContent(t *testing.T) {
	content := acpsdk.TextContent{
		Text: "Hello, World!",
	}
	contentJSON, _ := json.MarshalIndent(content, "", "  ")
	golden.Compare(t, "text-content.json", contentJSON)
}

func TestGolden_AudioContent(t *testing.T) {
	content := acpsdk.AudioContent{
		Data:     "base64audiodata",
		MimeType: "audio/wav",
	}
	contentJSON, _ := json.MarshalIndent(content, "", "  ")
	golden.Compare(t, "audio-content.json", contentJSON)
}

func TestGolden_ImageContent(t *testing.T) {
	content := acpsdk.ImageContent{
		Data:     "base64imagedata",
		MimeType: "image/png",
	}
	contentJSON, _ := json.MarshalIndent(content, "", "  ")
	golden.Compare(t, "image-content.json", contentJSON)
}

func TestGolden_TextResourceContents(t *testing.T) {
	mimeType := "text/x-go"
	content := acpsdk.TextResourceContents{
		Uri:      "file:///project/main.go",
		MimeType: &mimeType,
		Text:     "package main",
	}
	contentJSON, _ := json.MarshalIndent(content, "", "  ")
	golden.Compare(t, "text-resource-contents.json", contentJSON)
}

// ============================================================
// 计划/权限
// ============================================================

func TestGolden_Plan(t *testing.T) {
	plan := acpsdk.Plan{
		Entries: []acpsdk.PlanEntry{
			{Content: "Step 1", Status: acpsdk.PlanEntryStatusCompleted},
			{Content: "Step 2", Status: acpsdk.PlanEntryStatusInProgress},
		},
	}
	planJSON, _ := json.MarshalIndent(plan, "", "  ")
	golden.Compare(t, "plan.json", planJSON)
}

func TestGolden_PlanEntry(t *testing.T) {
	entry := acpsdk.PlanEntry{
		Content: "Read configuration file",
		Status:  acpsdk.PlanEntryStatusCompleted,
	}
	entryJSON, _ := json.MarshalIndent(entry, "", "  ")
	golden.Compare(t, "plan-entry.json", entryJSON)
}

func TestGolden_PermissionOption(t *testing.T) {
	opt := acpsdk.PermissionOption{
		Kind:    acpsdk.PermissionOptionKindAllowOnce,
		Name:    "Allow Once",
		OptionId: "allow-once",
	}
	optJSON, _ := json.MarshalIndent(opt, "", "  ")
	golden.Compare(t, "permission-option.json", optJSON)
}

func TestGolden_PermissionOption_AllowAlways(t *testing.T) {
	opt := acpsdk.PermissionOption{
		Kind:    acpsdk.PermissionOptionKindAllowAlways,
		Name:    "Allow Always",
		OptionId: "allow-always",
	}
	optJSON, _ := json.MarshalIndent(opt, "", "  ")
	golden.Compare(t, "permission-option-allow-always.json", optJSON)
}

func TestGolden_PermissionOption_RejectOnce(t *testing.T) {
	opt := acpsdk.PermissionOption{
		Kind:    acpsdk.PermissionOptionKindRejectOnce,
		Name:    "Reject Once",
		OptionId: "reject-once",
	}
	optJSON, _ := json.MarshalIndent(opt, "", "  ")
	golden.Compare(t, "permission-option-reject-once.json", optJSON)
}

func TestGolden_PermissionOption_RejectAlways(t *testing.T) {
	opt := acpsdk.PermissionOption{
		Kind:    acpsdk.PermissionOptionKindRejectAlways,
		Name:    "Reject Always",
		OptionId: "reject-always",
	}
	optJSON, _ := json.MarshalIndent(opt, "", "  ")
	golden.Compare(t, "permission-option-reject-always.json", optJSON)
}

// ============================================================
// 工具调用
// ============================================================

func TestGolden_ToolCallLocation(t *testing.T) {
	loc := acpsdk.ToolCallLocation{
		Path: "/project/main.go",
	}
	locJSON, _ := json.MarshalIndent(loc, "", "  ")
	golden.Compare(t, "tool-call-location.json", locJSON)
}

func TestGolden_ToolCallContent_Content(t *testing.T) {
	content := acpsdk.ToolCallContent{
		Content: &acpsdk.ToolCallContentContent{
			Content: acpsdk.TextBlock("file contents"),
		},
	}
	contentJSON, _ := json.MarshalIndent(content, "", "  ")
	golden.Compare(t, "tool-call-content-content-type.json", contentJSON)
}

// ============================================================
// 响应类型补充
// ============================================================

func TestGolden_LoadSessionResponse(t *testing.T) {
	resp := acpsdk.LoadSessionResponse{}
	respJSON, _ := json.MarshalIndent(resp, "", "  ")
	golden.Compare(t, "load-session-response.json", respJSON)
}

func TestGolden_AuthenticateResponse(t *testing.T) {
	resp := acpsdk.AuthenticateResponse{}
	respJSON, _ := json.MarshalIndent(resp, "", "  ")
	golden.Compare(t, "authenticate-response.json", respJSON)
}

func TestGolden_NewSessionRequest_Full(t *testing.T) {
	req := acpsdk.NewSessionRequest{
		Cwd: "/home/user/project",
	}
	reqJSON, _ := json.MarshalIndent(req, "", "  ")
	golden.Compare(t, "new-session-request.json", reqJSON)
}

func TestGolden_NewSessionResponse(t *testing.T) {
	resp := acpsdk.NewSessionResponse{
		SessionId: "session-12345",
	}
	respJSON, _ := json.MarshalIndent(resp, "", "  ")
	golden.Compare(t, "new-session-response.json", respJSON)
}

// ============================================================
// StopReason 完整覆盖 (5种)
// ============================================================

func TestGolden_StopReason_Refusal(t *testing.T) {
	resp := acpsdk.PromptResponse{
		StopReason: acpsdk.StopReasonRefusal,
	}
	respJSON, _ := json.MarshalIndent(resp, "", "  ")
	golden.Compare(t, "prompt-response-refusal.json", respJSON)
}

func TestGolden_StopReason_MaxTurnRequests(t *testing.T) {
	resp := acpsdk.PromptResponse{
		StopReason: acpsdk.StopReasonMaxTurnRequests,
	}
	respJSON, _ := json.MarshalIndent(resp, "", "  ")
	golden.Compare(t, "prompt-response-max-turn-requests.json", respJSON)
}

// ============================================================
// 缺失的类型补充测试
// ============================================================

func TestGolden_Annotations(t *testing.T) {
	t.Run("Annotations 格式", func(t *testing.T) {
		lastMod := "2024-01-01T00:00:00Z"
		priority := 0.8
		annotations := acpsdk.Annotations{
			Audience:     []acpsdk.Role{acpsdk.RoleUser, acpsdk.RoleAssistant},
			LastModified: &lastMod,
			Priority:     &priority,
		}

		annotationsJSON, err := json.MarshalIndent(annotations, "", "  ")
		require.NoError(t, err)

		golden.Compare(t, "annotations.json", annotationsJSON)
	})
}

func TestGolden_Content(t *testing.T) {
	t.Run("Content 格式", func(t *testing.T) {
		content := acpsdk.Content{
			Content: acpsdk.TextBlock("This is content"),
		}

		contentJSON, err := json.MarshalIndent(content, "", "  ")
		require.NoError(t, err)

		golden.Compare(t, "content.json", contentJSON)
	})
}

func TestGolden_FileSystemCapability(t *testing.T) {
	t.Run("FileSystemCapability 格式", func(t *testing.T) {
		cap := acpsdk.FileSystemCapability{
			ReadTextFile:  true,
			WriteTextFile: true,
		}

		capJSON, err := json.MarshalIndent(cap, "", "  ")
		require.NoError(t, err)

		golden.Compare(t, "file-system-capability.json", capJSON)
	})
}

func TestGolden_SessionCapabilities(t *testing.T) {
	t.Run("SessionCapabilities 格式", func(t *testing.T) {
		cap := acpsdk.SessionCapabilities{}

		capJSON, err := json.MarshalIndent(cap, "", "  ")
		require.NoError(t, err)

		golden.Compare(t, "session-capabilities.json", capJSON)
	})
}

func TestGolden_SessionNotification(t *testing.T) {
	t.Run("SessionNotification 格式", func(t *testing.T) {
		notification := acpsdk.SessionNotification{
			SessionId: "session-123",
			Update: acpsdk.SessionUpdate{
				AgentMessageChunk: &acpsdk.SessionUpdateAgentMessageChunk{
					Content:       acpsdk.TextBlock("Notification message"),
					SessionUpdate: "agent_message_chunk",
				},
			},
		}

		notifJSON, err := json.MarshalIndent(notification, "", "  ")
		require.NoError(t, err)

		golden.Compare(t, "session-notification.json", notifJSON)
	})
}

func TestGolden_TerminalExitStatus(t *testing.T) {
	t.Run("TerminalExitStatus 格式", func(t *testing.T) {
		exitCode := 0
		status := acpsdk.TerminalExitStatus{
			ExitCode: &exitCode,
		}

		statusJSON, err := json.MarshalIndent(status, "", "  ")
		require.NoError(t, err)

		golden.Compare(t, "terminal-exit-status.json", statusJSON)
	})
}

func TestGolden_ToolKind_Read(t *testing.T) {
	t.Run("ToolKind Read", func(t *testing.T) {
		update := acpsdk.StartToolCall(
			acpsdk.ToolCallId("call-read"),
			"Reading file",
			acpsdk.WithStartKind(acpsdk.ToolKindRead),
			acpsdk.WithStartStatus(acpsdk.ToolCallStatusPending),
			acpsdk.WithStartLocations([]acpsdk.ToolCallLocation{{Path: "/tmp/test.txt"}}),
		)
		updateJSON, _ := json.MarshalIndent(update, "", "  ")
		golden.Compare(t, "tool-kind-read.json", updateJSON)
	})
}

func TestGolden_AvailableCommandInput(t *testing.T) {
	t.Run("AvailableCommandInput 格式", func(t *testing.T) {
		input := acpsdk.AvailableCommandInput{
			Unstructured: &acpsdk.UnstructuredCommandInput{
				Hint: "Enter commit message",
			},
		}

		inputJSON, err := json.MarshalIndent(input, "", "  ")
		require.NoError(t, err)

		golden.Compare(t, "available-command-input.json", inputJSON)
	})
}

func TestGolden_EmbeddedResourceResource_Text(t *testing.T) {
	t.Run("EmbeddedResourceResource Text 格式", func(t *testing.T) {
		res := acpsdk.EmbeddedResourceResource{
			TextResourceContents: &acpsdk.TextResourceContents{
				Uri:  "file:///project/main.go",
				Text: "package main",
			},
		}

		resJSON, err := json.MarshalIndent(res, "", "  ")
		require.NoError(t, err)

		golden.Compare(t, "embedded-resource-resource-text.json", resJSON)
	})
}

func TestGolden_EmbeddedResourceResource_Blob(t *testing.T) {
	t.Run("EmbeddedResourceResource Blob 格式", func(t *testing.T) {
		mimeType := "image/png"
		res := acpsdk.EmbeddedResourceResource{
			BlobResourceContents: &acpsdk.BlobResourceContents{
				Uri:      "file:///project/image.png",
				Blob:     "base64encodeddata",
				MimeType: &mimeType,
			},
		}

		resJSON, err := json.MarshalIndent(res, "", "  ")
		require.NoError(t, err)

		golden.Compare(t, "embedded-resource-resource-blob.json", resJSON)
	})
}

func TestGolden_TerminalExitStatus_WithSignal(t *testing.T) {
	t.Run("TerminalExitStatus with Signal", func(t *testing.T) {
		signal := "SIGTERM"
		status := acpsdk.TerminalExitStatus{
			Signal: &signal,
		}

		statusJSON, err := json.MarshalIndent(status, "", "  ")
		require.NoError(t, err)

		golden.Compare(t, "terminal-exit-status-signal.json", statusJSON)
	})
}

func TestGolden_RequestPermissionOutcome_Selected(t *testing.T) {
	t.Run("RequestPermissionOutcome Selected", func(t *testing.T) {
		outcome := acpsdk.RequestPermissionOutcome{
			Selected: &acpsdk.RequestPermissionOutcomeSelected{
				OptionId: "allow-once",
			},
		}

		outcomeJSON, err := json.MarshalIndent(outcome, "", "  ")
		require.NoError(t, err)

		golden.Compare(t, "request-permission-outcome-selected.json", outcomeJSON)
	})
}

func TestGolden_RequestPermissionOutcome_Cancelled(t *testing.T) {
	t.Run("RequestPermissionOutcome Cancelled", func(t *testing.T) {
		outcome := acpsdk.RequestPermissionOutcome{
			Cancelled: &acpsdk.RequestPermissionOutcomeCancelled{},
		}

		outcomeJSON, err := json.MarshalIndent(outcome, "", "  ")
		require.NoError(t, err)

		golden.Compare(t, "request-permission-outcome-cancelled.json", outcomeJSON)
	})
}

func TestGolden_SelectedPermissionOutcome(t *testing.T) {
	t.Run("SelectedPermissionOutcome 格式", func(t *testing.T) {
		outcome := acpsdk.SelectedPermissionOutcome{
			OptionId: "allow-once",
		}

		outcomeJSON, err := json.MarshalIndent(outcome, "", "  ")
		require.NoError(t, err)

		golden.Compare(t, "selected-permission-outcome.json", outcomeJSON)
	})
}

func TestGolden_UnstructuredCommandInput(t *testing.T) {
	t.Run("UnstructuredCommandInput 格式", func(t *testing.T) {
		input := acpsdk.UnstructuredCommandInput{
			Hint: "Enter your command arguments",
		}

		inputJSON, err := json.MarshalIndent(input, "", "  ")
		require.NoError(t, err)

		golden.Compare(t, "unstructured-command-input.json", inputJSON)
	})
}

// ============================================================
// 边缘情况: ToolCall 独立结构体
// ============================================================

func TestGolden_ToolCall_Standalone(t *testing.T) {
	t.Run("ToolCall standalone", func(t *testing.T) {
		toolCall := acpsdk.ToolCall{
			ToolCallId: acpsdk.ToolCallId("call-123"),
			Title:      "Reading file",
			Kind:       acpsdk.ToolKindRead,
			Status:     acpsdk.ToolCallStatusPending,
			Locations: []acpsdk.ToolCallLocation{
				{Path: "/tmp/test.txt"},
			},
		}

		toolCallJSON, err := json.MarshalIndent(toolCall, "", "  ")
		require.NoError(t, err)

		golden.Compare(t, "tool-call.json", toolCallJSON)
	})
}

// ============================================================
// 边缘情况: ErrorCodeOther 独立测试
// ============================================================

func TestGolden_ErrorCode_Other(t *testing.T) {
	t.Run("ErrorCodeOther", func(t *testing.T) {
		sdkErr := acpsdk.Error{
			Code:    acpsdk.ErrorCode{Other: &acpsdk.ErrorCodeOther{}},
			Message: "Custom error",
		}

		errJSON, err := json.MarshalIndent(sdkErr, "", "  ")
		require.NoError(t, err)

		golden.Compare(t, "error-code-other.json", errJSON)
	})
}

// ============================================================
// 边缘情况: McpServer Inline 类型
// ============================================================

func TestGolden_McpServer_HttpInline(t *testing.T) {
	t.Run("McpServerHttpInline", func(t *testing.T) {
		server := acpsdk.McpServerHttpInline{
			Url: "http://localhost:8080/mcp",
			Headers: []acpsdk.HttpHeader{
				{Name: "Authorization", Value: "Bearer token"},
			},
		}

		serverJSON, err := json.MarshalIndent(server, "", "  ")
		require.NoError(t, err)

		golden.Compare(t, "mcp-server-http-inline.json", serverJSON)
	})
}

func TestGolden_McpServer_SseInline(t *testing.T) {
	t.Run("McpServerSseInline", func(t *testing.T) {
		server := acpsdk.McpServerSseInline{
			Url: "http://localhost:8080/sse",
		}

		serverJSON, err := json.MarshalIndent(server, "", "  ")
		require.NoError(t, err)

		golden.Compare(t, "mcp-server-sse-inline.json", serverJSON)
	})
}

// ============================================================
// 边缘情况: Agent JSON-RPC 类型
// ============================================================

func TestGolden_AgentRequest(t *testing.T) {
	t.Run("AgentRequest", func(t *testing.T) {
		req := acpsdk.AgentRequest{
			Id:     acpsdk.RequestId{Str: &acpsdk.RequestIdStr{}},
			Method: "session/prompt",
			Params: map[string]any{"key": "value"},
		}

		reqJSON, err := json.MarshalIndent(req, "", "  ")
		require.NoError(t, err)

		golden.Compare(t, "agent-request.json", reqJSON)
	})
}

func TestGolden_AgentNotification(t *testing.T) {
	t.Run("AgentNotification", func(t *testing.T) {
		notif := acpsdk.AgentNotification{
			Method: "session/update",
			Params: map[string]any{"key": "value"},
		}

		notifJSON, err := json.MarshalIndent(notif, "", "  ")
		require.NoError(t, err)

		golden.Compare(t, "agent-notification.json", notifJSON)
	})
}

func TestGolden_AgentResult(t *testing.T) {
	t.Run("AgentResult", func(t *testing.T) {
		result := acpsdk.AgentResult{
			Id:     acpsdk.RequestId{Str: &acpsdk.RequestIdStr{}},
			Result: map[string]any{"status": "success"},
		}

		resultJSON, err := json.MarshalIndent(result, "", "  ")
		require.NoError(t, err)

		golden.Compare(t, "agent-result.json", resultJSON)
	})
}

func TestGolden_AgentError(t *testing.T) {
	t.Run("AgentError", func(t *testing.T) {
		agentErr := acpsdk.AgentError{
			Id: acpsdk.RequestId{Str: &acpsdk.RequestIdStr{}},
			Error: acpsdk.Error{
				Code:    acpsdk.ErrorCode{InternalError: &acpsdk.ErrorCodeInternalError{}},
				Message: "Internal error occurred",
			},
		}

		errJSON, err := json.MarshalIndent(agentErr, "", "  ")
		require.NoError(t, err)

		golden.Compare(t, "agent-error.json", errJSON)
	})
}

func TestGolden_AgentResponse_Result(t *testing.T) {
	t.Run("AgentResponse with Result", func(t *testing.T) {
		resp := acpsdk.AgentResponse{
			Result: &acpsdk.AgentResult{
				Id:     acpsdk.RequestId{Str: &acpsdk.RequestIdStr{}},
				Result: map[string]any{"status": "ok"},
			},
		}

		respJSON, err := json.MarshalIndent(resp, "", "  ")
		require.NoError(t, err)

		golden.Compare(t, "agent-response-result.json", respJSON)
	})
}

func TestGolden_AgentResponse_Error(t *testing.T) {
	t.Run("AgentResponse with Error", func(t *testing.T) {
		resp := acpsdk.AgentResponse{
			Error: &acpsdk.AgentError{
				Id: acpsdk.RequestId{Str: &acpsdk.RequestIdStr{}},
				Error: acpsdk.Error{
					Code:    acpsdk.ErrorCode{InvalidParams: &acpsdk.ErrorCodeInvalidParams{}},
					Message: "Invalid parameters",
				},
			},
		}

		respJSON, err := json.MarshalIndent(resp, "", "  ")
		require.NoError(t, err)

		golden.Compare(t, "agent-response-error.json", respJSON)
	})
}

// ============================================================
// 边缘情况: Client JSON-RPC 类型
// ============================================================

func TestGolden_ClientRequest(t *testing.T) {
	t.Run("ClientRequest", func(t *testing.T) {
		req := acpsdk.ClientRequest{
			Id:     acpsdk.RequestId{Str: &acpsdk.RequestIdStr{}},
			Method: "initialize",
			Params: map[string]any{"protocolVersion": 1},
		}

		reqJSON, err := json.MarshalIndent(req, "", "  ")
		require.NoError(t, err)

		golden.Compare(t, "client-request.json", reqJSON)
	})
}

func TestGolden_ClientNotification(t *testing.T) {
	t.Run("ClientNotification", func(t *testing.T) {
		notif := acpsdk.ClientNotification{
			Method: "session/cancel",
			Params: map[string]any{"sessionId": "session-123"},
		}

		notifJSON, err := json.MarshalIndent(notif, "", "  ")
		require.NoError(t, err)

		golden.Compare(t, "client-notification.json", notifJSON)
	})
}

func TestGolden_ClientResult(t *testing.T) {
	t.Run("ClientResult", func(t *testing.T) {
		result := acpsdk.ClientResult{
			Id: acpsdk.RequestId{Str: &acpsdk.RequestIdStr{}},
			Result: acpsdk.RequestPermissionResponse{
				Outcome: acpsdk.RequestPermissionOutcome{
					Selected: &acpsdk.RequestPermissionOutcomeSelected{
						OptionId: "allow",
					},
				},
			},
		}

		resultJSON, err := json.MarshalIndent(result, "", "  ")
		require.NoError(t, err)

		golden.Compare(t, "client-result.json", resultJSON)
	})
}

func TestGolden_ClientError(t *testing.T) {
	t.Run("ClientError", func(t *testing.T) {
		clientErr := acpsdk.ClientError{
			Id: acpsdk.RequestId{Str: &acpsdk.RequestIdStr{}},
			Error: acpsdk.Error{
				Code:    acpsdk.ErrorCode{ResourceNotFound: &acpsdk.ErrorCodeResourceNotFound{}},
				Message: "Resource not found",
			},
		}

		errJSON, err := json.MarshalIndent(clientErr, "", "  ")
		require.NoError(t, err)

		golden.Compare(t, "client-error.json", errJSON)
	})
}

func TestGolden_ClientResponse_Result(t *testing.T) {
	t.Run("ClientResponse with Result", func(t *testing.T) {
		resp := acpsdk.ClientResponse{
			Result: &acpsdk.ClientResult{
				Id: acpsdk.RequestId{Str: &acpsdk.RequestIdStr{}},
				Result: acpsdk.RequestPermissionResponse{
					Outcome: acpsdk.RequestPermissionOutcome{
						Selected: &acpsdk.RequestPermissionOutcomeSelected{
							OptionId: "allow",
						},
					},
				},
			},
		}

		respJSON, err := json.MarshalIndent(resp, "", "  ")
		require.NoError(t, err)

		golden.Compare(t, "client-response-result.json", respJSON)
	})
}

func TestGolden_ClientResponse_Error(t *testing.T) {
	t.Run("ClientResponse with Error", func(t *testing.T) {
		resp := acpsdk.ClientResponse{
			Error: &acpsdk.ClientError{
				Id: acpsdk.RequestId{Str: &acpsdk.RequestIdStr{}},
				Error: acpsdk.Error{
					Code:    acpsdk.ErrorCode{MethodNotFound: &acpsdk.ErrorCodeMethodNotFound{}},
					Message: "Method not supported",
				},
			},
		}

		respJSON, err := json.MarshalIndent(resp, "", "  ")
		require.NoError(t, err)

		golden.Compare(t, "client-response-error.json", respJSON)
	})
}

// ============================================================
// 边缘情况: RequestId 变体
// ============================================================

func TestGolden_RequestId_Null(t *testing.T) {
	t.Run("RequestId Null", func(t *testing.T) {
		reqId := acpsdk.RequestId{
			Null: &acpsdk.RequestIdNull{},
		}

		reqIdJSON, err := json.MarshalIndent(reqId, "", "  ")
		require.NoError(t, err)

		golden.Compare(t, "request-id-null.json", reqIdJSON)
	})
}

func TestGolden_RequestId_Number(t *testing.T) {
	t.Run("RequestId Number", func(t *testing.T) {
		reqId := acpsdk.RequestId{
			Number: &acpsdk.RequestIdNumber{},
		}

		reqIdJSON, err := json.MarshalIndent(reqId, "", "  ")
		require.NoError(t, err)

		golden.Compare(t, "request-id-number.json", reqIdJSON)
	})
}

func TestGolden_RequestId_String(t *testing.T) {
	t.Run("RequestId String", func(t *testing.T) {
		reqId := acpsdk.RequestId{
			Str: &acpsdk.RequestIdStr{},
		}

		reqIdJSON, err := json.MarshalIndent(reqId, "", "  ")
		require.NoError(t, err)

		golden.Compare(t, "request-id-string.json", reqIdJSON)
	})
}

// ============================================================
// 边缘情况: SessionUpdate 所有变体
// ============================================================

func TestGolden_SessionUpdate_ToolCall(t *testing.T) {
	t.Run("SessionUpdate ToolCall", func(t *testing.T) {
		update := acpsdk.SessionUpdate{
			ToolCall: &acpsdk.SessionUpdateToolCall{
				SessionUpdate: "tool_call",
				ToolCallId:    acpsdk.ToolCallId("call-1"),
				Title:         "Executing tool",
				Kind:          acpsdk.ToolKindExecute,
				Status:        acpsdk.ToolCallStatusInProgress,
			},
		}

		updateJSON, err := json.MarshalIndent(update, "", "  ")
		require.NoError(t, err)

		golden.Compare(t, "session-update-tool-call.json", updateJSON)
	})
}

func TestGolden_SessionUpdate_ToolCallUpdate(t *testing.T) {
	t.Run("SessionUpdate ToolCallUpdate", func(t *testing.T) {
		status := acpsdk.ToolCallStatusCompleted
		update := acpsdk.SessionUpdate{
			ToolCallUpdate: &acpsdk.SessionToolCallUpdate{
				SessionUpdate: "tool_call_update",
				ToolCallId:    acpsdk.ToolCallId("call-1"),
				Status:        &status,
			},
		}

		updateJSON, err := json.MarshalIndent(update, "", "  ")
		require.NoError(t, err)

		golden.Compare(t, "session-update-tool-call-update.json", updateJSON)
	})
}

func TestGolden_SessionUpdate_AvailableCommandsUpdate(t *testing.T) {
	t.Run("SessionUpdate AvailableCommandsUpdate", func(t *testing.T) {
		update := acpsdk.SessionUpdate{
			AvailableCommandsUpdate: &acpsdk.SessionAvailableCommandsUpdate{
				SessionUpdate: "available_commands",
				AvailableCommands: []acpsdk.AvailableCommand{
					{Name: "test", Description: "Test command"},
				},
			},
		}

		updateJSON, err := json.MarshalIndent(update, "", "  ")
		require.NoError(t, err)

		golden.Compare(t, "session-update-available-commands.json", updateJSON)
	})
}

func TestGolden_SessionUpdate_CurrentModeUpdate(t *testing.T) {
	t.Run("SessionUpdate CurrentModeUpdate", func(t *testing.T) {
		update := acpsdk.SessionUpdate{
			CurrentModeUpdate: &acpsdk.SessionCurrentModeUpdate{
				SessionUpdate: "current_mode",
				CurrentModeId: acpsdk.SessionModeId("code"),
			},
		}

		updateJSON, err := json.MarshalIndent(update, "", "  ")
		require.NoError(t, err)

		golden.Compare(t, "session-update-current-mode.json", updateJSON)
	})
}

func TestGolden_SessionUpdate_UserMessageChunk(t *testing.T) {
	t.Run("SessionUpdate UserMessageChunk", func(t *testing.T) {
		update := acpsdk.SessionUpdate{
			UserMessageChunk: &acpsdk.SessionUpdateUserMessageChunk{
				SessionUpdate: "user_message_chunk",
				Content:       acpsdk.TextBlock("User message"),
			},
		}

		updateJSON, err := json.MarshalIndent(update, "", "  ")
		require.NoError(t, err)

		golden.Compare(t, "session-update-user-message-chunk.json", updateJSON)
	})
}

func TestGolden_SessionUpdate_AgentMessageChunk(t *testing.T) {
	t.Run("SessionUpdate AgentMessageChunk", func(t *testing.T) {
		update := acpsdk.SessionUpdate{
			AgentMessageChunk: &acpsdk.SessionUpdateAgentMessageChunk{
				SessionUpdate: "agent_message_chunk",
				Content:       acpsdk.TextBlock("Agent response"),
			},
		}

		updateJSON, err := json.MarshalIndent(update, "", "  ")
		require.NoError(t, err)

		golden.Compare(t, "session-update-agent-message-chunk.json", updateJSON)
	})
}

func TestGolden_SessionUpdate_AgentThoughtChunk(t *testing.T) {
	t.Run("SessionUpdate AgentThoughtChunk", func(t *testing.T) {
		update := acpsdk.SessionUpdate{
			AgentThoughtChunk: &acpsdk.SessionUpdateAgentThoughtChunk{
				SessionUpdate: "agent_thought_chunk",
				Content:       acpsdk.TextBlock("Thinking..."),
			},
		}

		updateJSON, err := json.MarshalIndent(update, "", "  ")
		require.NoError(t, err)

		golden.Compare(t, "session-update-agent-thought-chunk.json", updateJSON)
	})
}

func TestGolden_SessionUpdate_Plan(t *testing.T) {
	t.Run("SessionUpdate Plan", func(t *testing.T) {
		update := acpsdk.SessionUpdate{
			Plan: &acpsdk.SessionUpdatePlan{
				SessionUpdate: "plan",
				Entries: []acpsdk.PlanEntry{
					{Content: "Step 1", Status: acpsdk.PlanEntryStatusCompleted},
				},
			},
		}

		updateJSON, err := json.MarshalIndent(update, "", "  ")
		require.NoError(t, err)

		golden.Compare(t, "session-update-plan-chunk.json", updateJSON)
	})
}

// ============================================================
// 边缘情况: ContentBlock 所有变体 (Union类型)
// ============================================================

func TestGolden_ContentBlock_Union_Text(t *testing.T) {
	t.Run("ContentBlock Union Text", func(t *testing.T) {
		block := acpsdk.ContentBlock{
			Text: &acpsdk.ContentBlockText{
				Type: "text",
				Text: "Hello, World!",
			},
		}

		blockJSON, err := json.MarshalIndent(block, "", "  ")
		require.NoError(t, err)

		golden.Compare(t, "content-block-text-full.json", blockJSON)
	})
}

func TestGolden_ContentBlock_Union_Image(t *testing.T) {
	t.Run("ContentBlock Union Image", func(t *testing.T) {
		block := acpsdk.ContentBlock{
			Image: &acpsdk.ContentBlockImage{
				Type:     "image",
				Data:     "base64imagedata",
				MimeType: "image/png",
			},
		}

		blockJSON, err := json.MarshalIndent(block, "", "  ")
		require.NoError(t, err)

		golden.Compare(t, "content-block-image-full.json", blockJSON)
	})
}

func TestGolden_ContentBlock_Union_Audio(t *testing.T) {
	t.Run("ContentBlock Union Audio", func(t *testing.T) {
		block := acpsdk.ContentBlock{
			Audio: &acpsdk.ContentBlockAudio{
				Type:     "audio",
				Data:     "base64audiodata",
				MimeType: "audio/wav",
			},
		}

		blockJSON, err := json.MarshalIndent(block, "", "  ")
		require.NoError(t, err)

		golden.Compare(t, "content-block-audio-full.json", blockJSON)
	})
}

func TestGolden_ContentBlock_Union_ResourceLink(t *testing.T) {
	t.Run("ContentBlock Union ResourceLink", func(t *testing.T) {
		block := acpsdk.ContentBlock{
			ResourceLink: &acpsdk.ContentBlockResourceLink{
				Type: "resource_link",
				Name: "config.json",
				Uri:  "file:///project/config.json",
			},
		}

		blockJSON, err := json.MarshalIndent(block, "", "  ")
		require.NoError(t, err)

		golden.Compare(t, "content-block-resource-link-full.json", blockJSON)
	})
}

func TestGolden_ContentBlock_Union_Resource(t *testing.T) {
	t.Run("ContentBlock Union Resource", func(t *testing.T) {
		block := acpsdk.ContentBlock{
			Resource: &acpsdk.ContentBlockResource{
				Type: "resource",
				Resource: acpsdk.EmbeddedResourceResource{
					TextResourceContents: &acpsdk.TextResourceContents{
						Uri:  "file:///project/main.go",
						Text: "package main",
					},
				},
			},
		}

		blockJSON, err := json.MarshalIndent(block, "", "  ")
		require.NoError(t, err)

		golden.Compare(t, "content-block-resource-full.json", blockJSON)
	})
}

// ============================================================
// 边缘情况: ErrorCode 所有变体
// ============================================================

func TestGolden_ErrorCode_ParseError_Full(t *testing.T) {
	t.Run("ErrorCode ParseError full", func(t *testing.T) {
		code := acpsdk.ErrorCode{
			ParseError: &acpsdk.ErrorCodeParseError{},
		}

		codeJSON, err := json.MarshalIndent(code, "", "  ")
		require.NoError(t, err)

		golden.Compare(t, "error-code-parse-error-full.json", codeJSON)
	})
}

func TestGolden_ErrorCode_InvalidRequest_Full(t *testing.T) {
	t.Run("ErrorCode InvalidRequest full", func(t *testing.T) {
		code := acpsdk.ErrorCode{
			InvalidRequest: &acpsdk.ErrorCodeInvalidRequest{},
		}

		codeJSON, err := json.MarshalIndent(code, "", "  ")
		require.NoError(t, err)

		golden.Compare(t, "error-code-invalid-request-full.json", codeJSON)
	})
}

func TestGolden_ErrorCode_MethodNotFound_Full(t *testing.T) {
	t.Run("ErrorCode MethodNotFound full", func(t *testing.T) {
		code := acpsdk.ErrorCode{
			MethodNotFound: &acpsdk.ErrorCodeMethodNotFound{},
		}

		codeJSON, err := json.MarshalIndent(code, "", "  ")
		require.NoError(t, err)

		golden.Compare(t, "error-code-method-not-found-full.json", codeJSON)
	})
}

func TestGolden_ErrorCode_InvalidParams_Full(t *testing.T) {
	t.Run("ErrorCode InvalidParams full", func(t *testing.T) {
		code := acpsdk.ErrorCode{
			InvalidParams: &acpsdk.ErrorCodeInvalidParams{},
		}

		codeJSON, err := json.MarshalIndent(code, "", "  ")
		require.NoError(t, err)

		golden.Compare(t, "error-code-invalid-params-full.json", codeJSON)
	})
}

func TestGolden_ErrorCode_InternalError_Full(t *testing.T) {
	t.Run("ErrorCode InternalError full", func(t *testing.T) {
		code := acpsdk.ErrorCode{
			InternalError: &acpsdk.ErrorCodeInternalError{},
		}

		codeJSON, err := json.MarshalIndent(code, "", "  ")
		require.NoError(t, err)

		golden.Compare(t, "error-code-internal-error-full.json", codeJSON)
	})
}

func TestGolden_ErrorCode_AuthenticationRequired_Full(t *testing.T) {
	t.Run("ErrorCode AuthenticationRequired full", func(t *testing.T) {
		code := acpsdk.ErrorCode{
			AuthenticationRequired: &acpsdk.ErrorCodeAuthenticationRequired{},
		}

		codeJSON, err := json.MarshalIndent(code, "", "  ")
		require.NoError(t, err)

		golden.Compare(t, "error-code-authentication-required-full.json", codeJSON)
	})
}

func TestGolden_ErrorCode_ResourceNotFound_Full(t *testing.T) {
	t.Run("ErrorCode ResourceNotFound full", func(t *testing.T) {
		code := acpsdk.ErrorCode{
			ResourceNotFound: &acpsdk.ErrorCodeResourceNotFound{},
		}

		codeJSON, err := json.MarshalIndent(code, "", "  ")
		require.NoError(t, err)

		golden.Compare(t, "error-code-resource-not-found-full.json", codeJSON)
	})
}

func TestGolden_ErrorCode_Other_Full(t *testing.T) {
	t.Run("ErrorCode Other full", func(t *testing.T) {
		code := acpsdk.ErrorCode{
			Other: &acpsdk.ErrorCodeOther{},
		}

		codeJSON, err := json.MarshalIndent(code, "", "  ")
		require.NoError(t, err)

		golden.Compare(t, "error-code-other-full.json", codeJSON)
	})
}
