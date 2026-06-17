package api

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"claude-code-acp-go/internal/mcp"
)

func TestNewOptions(t *testing.T) {
	t.Run("创建默认选项", func(t *testing.T) {
		opts := NewOptions()
		require.NotNil(t, opts)
		assert.Equal(t, "claude-code-acp", opts.Name)
		assert.Equal(t, ".", opts.WorkingDir)
	})
}

func TestOptions_WithName(t *testing.T) {
	t.Run("设置名称", func(t *testing.T) {
		opts := NewOptions().WithName("test-agent")
		assert.Equal(t, "test-agent", opts.Name)
	})
}

func TestOptions_WithWorkingDir(t *testing.T) {
	t.Run("设置工作目录", func(t *testing.T) {
		opts := NewOptions().WithWorkingDir("/work")
		assert.Equal(t, "/work", opts.WorkingDir)
	})
}

func TestOptions_WithPermissionChecker(t *testing.T) {
	t.Run("设置权限检查器", func(t *testing.T) {
		checker := mcp.NewDefaultPermissionChecker()
		opts := NewOptions().WithPermissionChecker(checker)
		assert.Equal(t, checker, opts.PermissionChecker)
	})
}

func TestOptions_Chaining(t *testing.T) {
	t.Run("链式调用所有选项", func(t *testing.T) {
		checker := mcp.NewDefaultPermissionChecker()
		opts := NewOptions().
			WithName("test").
			WithWorkingDir("/work").
			WithPermissionChecker(checker)

		assert.Equal(t, "test", opts.Name)
		assert.Equal(t, "/work", opts.WorkingDir)
		assert.Equal(t, checker, opts.PermissionChecker)
	})
}
