package acp

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInitializeRequest(t *testing.T) {
	t.Run("序列化基本请求", func(t *testing.T) {
		req := InitializeRequest{
			ProtocolVersion: 1,
			ClientCapabilities: ClientCapabilities{
				Terminal: true,
			},
		}

		data, err := json.Marshal(req)
		require.NoError(t, err)

		// 验证 JSON 字段名正确
		var result map[string]any
		err = json.Unmarshal(data, &result)
		require.NoError(t, err)

		assert.Equal(t, float64(1), result["protocolVersion"])
		assert.NotNil(t, result["clientCapabilities"])
	})

	t.Run("反序列化请求", func(t *testing.T) {
		jsonStr := `{"protocolVersion":1,"clientCapabilities":{"terminal":true}}`

		var req InitializeRequest
		err := json.Unmarshal([]byte(jsonStr), &req)
		require.NoError(t, err)

		assert.Equal(t, ProtocolVersion(1), req.ProtocolVersion)
		assert.True(t, req.ClientCapabilities.Terminal)
	})

	t.Run("带可选字段", func(t *testing.T) {
		jsonStr := `{"protocolVersion":1,"clientCapabilities":{},"clientInfo":{"name":"test","version":"1.0"},"_meta":{"custom":"value"}}`

		var req InitializeRequest
		err := json.Unmarshal([]byte(jsonStr), &req)
		require.NoError(t, err)

		require.NotNil(t, req.ClientInfo)
		assert.Equal(t, "test", req.ClientInfo.Name)
		assert.Equal(t, "1.0", req.ClientInfo.Version)
		assert.Equal(t, "value", req.Meta["custom"])
	})

	t.Run("文件系统能力", func(t *testing.T) {
		jsonStr := `{"protocolVersion":1,"clientCapabilities":{"fs":{"readTextFile":true,"writeTextFile":true},"terminal":false}}`

		var req InitializeRequest
		err := json.Unmarshal([]byte(jsonStr), &req)
		require.NoError(t, err)

		assert.True(t, req.ClientCapabilities.Fs.ReadTextFile)
		assert.True(t, req.ClientCapabilities.Fs.WriteTextFile)
		assert.False(t, req.ClientCapabilities.Terminal)
	})
}

func TestInitializeResponse(t *testing.T) {
	t.Run("序列化响应", func(t *testing.T) {
		agentInfo := &Implementation{
			Name:    "claude-code-acp-go",
			Version: "1.0.0",
		}
		resp := InitializeResponse{
			ProtocolVersion: 1,
			AgentCapabilities: AgentCapabilities{
				LoadSession: true,
				PromptCapabilities: PromptCapabilities{
					Image: true,
					Audio: true,
				},
			},
			AgentInfo: agentInfo,
		}

		data, err := json.Marshal(resp)
		require.NoError(t, err)

		var result map[string]any
		err = json.Unmarshal(data, &result)
		require.NoError(t, err)

		assert.Equal(t, float64(1), result["protocolVersion"])
		assert.NotNil(t, result["agentCapabilities"])
		assert.NotNil(t, result["agentInfo"])
	})

	t.Run("带认证方法", func(t *testing.T) {
		resp := InitializeResponse{
			ProtocolVersion: 1,
			AgentCapabilities: AgentCapabilities{
				LoadSession: true,
			},
			AuthMethods: []AuthMethod{
				{Id: "anthropic", Name: "Anthropic API"},
			},
		}

		data, err := json.Marshal(resp)
		require.NoError(t, err)

		var result map[string]any
		err = json.Unmarshal(data, &result)
		require.NoError(t, err)

		authMethods := result["authMethods"].([]any)
		assert.Len(t, authMethods, 1)
	})
}

func TestContentBlock(t *testing.T) {
	t.Run("文本内容块", func(t *testing.T) {
		jsonStr := `{"type":"text","text":"Hello, World!"}`

		var cb ContentBlock
		err := json.Unmarshal([]byte(jsonStr), &cb)
		require.NoError(t, err)

		require.NotNil(t, cb.Text)
		assert.Equal(t, "text", cb.Text.Type)
		assert.Equal(t, "Hello, World!", cb.Text.Text)
	})

	t.Run("图片内容块", func(t *testing.T) {
		jsonStr := `{"type":"image","data":"abc123","mimeType":"image/png"}`

		var cb ContentBlock
		err := json.Unmarshal([]byte(jsonStr), &cb)
		require.NoError(t, err)

		require.NotNil(t, cb.Image)
		assert.Equal(t, "image", cb.Image.Type)
		assert.Equal(t, "abc123", cb.Image.Data)
		assert.Equal(t, "image/png", cb.Image.MimeType)
	})

	t.Run("音频内容块", func(t *testing.T) {
		jsonStr := `{"type":"audio","data":"base64audio","mimeType":"audio/wav"}`

		var cb ContentBlock
		err := json.Unmarshal([]byte(jsonStr), &cb)
		require.NoError(t, err)

		require.NotNil(t, cb.Audio)
		assert.Equal(t, "audio", cb.Audio.Type)
		assert.Equal(t, "base64audio", cb.Audio.Data)
		assert.Equal(t, "audio/wav", cb.Audio.MimeType)
	})

	t.Run("资源链接", func(t *testing.T) {
		jsonStr := `{"type":"resource_link","uri":"file:///test.txt","name":"test.txt"}`

		var cb ContentBlock
		err := json.Unmarshal([]byte(jsonStr), &cb)
		require.NoError(t, err)

		require.NotNil(t, cb.ResourceLink)
		assert.Equal(t, "resource_link", cb.ResourceLink.Type)
		assert.Equal(t, "file:///test.txt", cb.ResourceLink.Uri)
		assert.Equal(t, "test.txt", cb.ResourceLink.Name)
	})
}

func TestPromptCapabilities(t *testing.T) {
	t.Run("序列化", func(t *testing.T) {
		caps := PromptCapabilities{
			Image:           true,
			Audio:           true,
			EmbeddedContext: true,
		}

		data, err := json.Marshal(caps)
		require.NoError(t, err)

		var result map[string]any
		err = json.Unmarshal(data, &result)
		require.NoError(t, err)

		assert.Equal(t, true, result["image"])
		assert.Equal(t, true, result["audio"])
		assert.Equal(t, true, result["embeddedContext"])
	})
}

func TestAgentCapabilities(t *testing.T) {
	t.Run("完整能力", func(t *testing.T) {
		caps := AgentCapabilities{
			LoadSession: true,
			PromptCapabilities: PromptCapabilities{
				Image: true,
				Audio: true,
			},
			McpCapabilities: McpCapabilities{
				Http: true,
				Sse:  true,
			},
		}

		data, err := json.Marshal(caps)
		require.NoError(t, err)

		var decoded AgentCapabilities
		err = json.Unmarshal(data, &decoded)
		require.NoError(t, err)

		assert.True(t, decoded.LoadSession)
		assert.True(t, decoded.PromptCapabilities.Image)
		assert.True(t, decoded.McpCapabilities.Http)
	})
}

func TestSessionUpdate(t *testing.T) {
	t.Run("工具调用更新", func(t *testing.T) {
		toolCall := SessionUpdateToolCall{
			SessionUpdate: "tool_call",
			ToolCallId:    "call-123",
			Title:         "Read file",
			Kind:          ToolKindRead,
			Status:        ToolCallStatusInProgress,
		}

		data, err := json.Marshal(toolCall)
		require.NoError(t, err)

		var result map[string]any
		err = json.Unmarshal(data, &result)
		require.NoError(t, err)

		assert.Equal(t, "tool_call", result["sessionUpdate"])
		assert.Equal(t, "call-123", result["toolCallId"])
		assert.Equal(t, "Read file", result["title"])
		assert.Equal(t, "read", result["kind"])
		assert.Equal(t, "in_progress", result["status"])
	})

	t.Run("代理消息块更新", func(t *testing.T) {
		chunk := SessionUpdateAgentMessageChunk{
			SessionUpdate: "agent_message_chunk",
			Content: ContentBlock{
				Text: &ContentBlockText{
					Type: "text",
					Text: "Hello!",
				},
			},
		}

		data, err := json.Marshal(chunk)
		require.NoError(t, err)

		var result map[string]any
		err = json.Unmarshal(data, &result)
		require.NoError(t, err)

		assert.Equal(t, "agent_message_chunk", result["sessionUpdate"])
		content := result["content"].(map[string]any)
		assert.Equal(t, "text", content["type"])
		assert.Equal(t, "Hello!", content["text"])
	})
}
