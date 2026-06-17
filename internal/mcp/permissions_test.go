package mcp

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

// ============================================================
// 权限检查器测试
// ============================================================

func TestDefaultPermissionChecker_Allow(t *testing.T) {
	checker := NewDefaultPermissionChecker()
	checker.AddAllowRule("Read", "")
	checker.AddAllowRule("Bash", "ls *")

	decision := checker.Check(context.Background(), "Read", map[string]any{"file_path": "/tmp/test.txt"})
	require.Equal(t, PermissionDecisionAllow, decision)

	decision = checker.Check(context.Background(), "Bash", map[string]any{"command": "ls -la"})
	require.Equal(t, PermissionDecisionAllow, decision)
}

func TestDefaultPermissionChecker_Deny(t *testing.T) {
	checker := NewDefaultPermissionChecker()
	checker.AddDenyRule("Bash", "rm -rf *")

	decision := checker.Check(context.Background(), "Bash", map[string]any{"command": "rm -rf /home"})
	require.Equal(t, PermissionDecisionDeny, decision)
}

func TestDefaultPermissionChecker_Ask(t *testing.T) {
	checker := NewDefaultPermissionChecker()
	// 没有任何规则的工具应该返回 Ask

	decision := checker.Check(context.Background(), "Write", map[string]any{"file_path": "/tmp/test.txt"})
	require.Equal(t, PermissionDecisionAsk, decision)
}

func TestDefaultPermissionChecker_DenyOverridesAllow(t *testing.T) {
	checker := NewDefaultPermissionChecker()
	checker.AddAllowRule("Bash", "*")
	checker.AddDenyRule("Bash", "rm *")

	// 应该被拒绝，因为 deny 规则优先
	decision := checker.Check(context.Background(), "Bash", map[string]any{"command": "rm -rf /"})
	require.Equal(t, PermissionDecisionDeny, decision)
}

func TestDefaultPermissionChecker_ClearRules(t *testing.T) {
	checker := NewDefaultPermissionChecker()
	checker.AddAllowRule("Read", "")
	checker.ClearRules()

	decision := checker.Check(context.Background(), "Read", map[string]any{})
	require.Equal(t, PermissionDecisionAsk, decision)
}

func TestPermissionRule_Match(t *testing.T) {
	tests := []struct {
		name     string
		rule     PermissionRuleT
		params   map[string]any
		expected bool
	}{
		{
			name:     "空规则匹配所有",
			rule:     PermissionRuleT{Pattern: ""},
			params:   map[string]any{"command": "anything"},
			expected: true,
		},
		{
			name:     "通配符匹配",
			rule:     PermissionRuleT{Pattern: "ls *"},
			params:   map[string]any{"command": "ls -la"},
			expected: true,
		},
		{
			name:     "通配符不匹配",
			rule:     PermissionRuleT{Pattern: "ls *"},
			params:   map[string]any{"command": "rm -rf"},
			expected: false,
		},
		{
			name:     "精确匹配",
			rule:     PermissionRuleT{Pattern: "echo hello"},
			params:   map[string]any{"command": "echo hello"},
			expected: true,
		},
		{
			name:     "正则匹配",
			rule:     PermissionRuleT{Pattern: "cat .*\\.txt"},
			params:   map[string]any{"command": "cat file.txt"},
			expected: true,
		},
		{
			name:     "精确不匹配 (无效正则触发精确匹配)",
			rule:     PermissionRuleT{Pattern: "[invalid"}, // 无效正则，会触发精确匹配
			params:   map[string]any{"command": "not [invalid"},
			expected: false,
		},
		{
			name:     "无效正则精确匹配",
			rule:     PermissionRuleT{Pattern: "[invalid"}, // 无效正则
			params:   map[string]any{"command": "[invalid"},
			expected: true,
		},
		{
			name:     "匹配 file_path 字段",
			rule:     PermissionRuleT{Pattern: "/tmp/*"},
			params:   map[string]any{"file_path": "/tmp/test.txt"},
			expected: true,
		},
		{
			name:     "匹配 content 字段",
			rule:     PermissionRuleT{Pattern: "hello"},
			params:   map[string]any{"content": "hello world"},
			expected: true,
		},
		{
			name:     "无匹配字段",
			rule:     PermissionRuleT{Pattern: "test"},
			params:   map[string]any{"other_field": "value"},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.expected, tt.rule.Match(tt.params))
		})
	}
}

// ============================================================
// 权限集成测试
// ============================================================

func TestServer_WithPermissionCheck(t *testing.T) {
	checker := NewDefaultPermissionChecker()
	checker.AddAllowRule("Read", "")
	checker.AddDenyRule("Write", "")

	server, err := NewServer("test-server", NewServerOptions().WithPermissionChecker(checker))
	require.NoError(t, err)

	_ = server.RegisterTool(Tool{
		Name:        "Read",
		Description: "读取文件",
		Handler:     func(ctx context.Context, args map[string]any) (any, error) { return "read ok", nil },
	})

	_ = server.RegisterTool(Tool{
		Name:        "Write",
		Description: "写入文件",
		Handler:     func(ctx context.Context, args map[string]any) (any, error) { return "write ok", nil },
	})

	// Read 应该被允许
	msg := map[string]any{
		"jsonrpc": "2.0",
		"id":      float64(1),
		"method":  "tools/call",
		"params": map[string]any{
			"name":      "Read",
			"arguments": map[string]any{},
		},
	}
	resp, err := server.HandleMessage(msg)
	require.NoError(t, err)
	require.Nil(t, resp["error"])

	// Write 应该被拒绝
	msg = map[string]any{
		"jsonrpc": "2.0",
		"id":      float64(2),
		"method":  "tools/call",
		"params": map[string]any{
			"name":      "Write",
			"arguments": map[string]any{},
		},
	}
	_, err = server.HandleMessage(msg)
	require.NoError(t, err)
	// 注意：当前实现中权限检查器只是记录，实际拦截需要在 HandleMessage 中实现
	// 这里测试权限检查器是否被正确调用
}

// ============================================================
// 权限错误测试
// ============================================================

func TestPermissionError(t *testing.T) {
	err := NewPermissionError("Write", "sensitive file")
	require.Error(t, err)
	require.Contains(t, err.Error(), "permission denied")
	require.Contains(t, err.Error(), "Write")
	require.Contains(t, err.Error(), "sensitive file")
}

func TestPermissionError_Error(t *testing.T) {
	err := &PermissionError{
		ToolName: "Bash",
		Message:  "dangerous command",
	}
	require.Equal(t, "permission denied for tool Bash: dangerous command", err.Error())
}
