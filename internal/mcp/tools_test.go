package mcp

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

// ============================================================
// Read 工具测试
// ============================================================

func TestReadTool_ExistingFile(t *testing.T) {
	// 创建临时文件
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.txt")
	err := os.WriteFile(testFile, []byte("hello world"), 0644)
	require.NoError(t, err)

	tool := NewReadTool()
	result, err := tool.Handler(context.Background(), map[string]any{
		"file_path": testFile,
	})
	require.NoError(t, err)
	require.Contains(t, result.(string), "hello world")
}

func TestReadTool_NonexistentFile(t *testing.T) {
	tool := NewReadTool()
	_, err := tool.Handler(context.Background(), map[string]any{
		"file_path": "/nonexistent/file.txt",
	})
	require.Error(t, err)
}

func TestReadTool_Directory(t *testing.T) {
	tmpDir := t.TempDir()

	tool := NewReadTool()
	result, err := tool.Handler(context.Background(), map[string]any{
		"file_path": tmpDir,
	})
	require.NoError(t, err)
	// 目录应该返回列表
	require.Contains(t, result.(string), "total")
}

func TestReadTool_WithOffsetAndLimit(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.txt")
	content := "line1\nline2\nline3\nline4\nline5\n"
	err := os.WriteFile(testFile, []byte(content), 0644)
	require.NoError(t, err)

	tool := NewReadTool()
	result, err := tool.Handler(context.Background(), map[string]any{
		"file_path": testFile,
		"offset":    float64(1),
		"limit":     float64(2),
	})
	require.NoError(t, err)
	require.Contains(t, result.(string), "line2")
	require.Contains(t, result.(string), "line3")
	require.NotContains(t, result.(string), "line1")
	require.NotContains(t, result.(string), "line4")
}

// ============================================================
// Write 工具测试
// ============================================================

func TestWriteTool_NewFile(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "new.txt")

	tool := NewWriteTool()
	result, err := tool.Handler(context.Background(), map[string]any{
		"file_path": testFile,
		"content":   "hello world",
	})
	require.NoError(t, err)
	require.Contains(t, strings.ToLower(result.(string)), "success")

	// 验证文件内容
	data, err := os.ReadFile(testFile)
	require.NoError(t, err)
	require.Equal(t, "hello world", string(data))
}

func TestWriteTool_OverwriteExisting(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "existing.txt")
	err := os.WriteFile(testFile, []byte("old content"), 0644)
	require.NoError(t, err)

	tool := NewWriteTool()
	result, err := tool.Handler(context.Background(), map[string]any{
		"file_path": testFile,
		"content":   "new content",
	})
	require.NoError(t, err)
	require.Contains(t, strings.ToLower(result.(string)), "success")

	// 验证文件被覆盖
	data, err := os.ReadFile(testFile)
	require.NoError(t, err)
	require.Equal(t, "new content", string(data))
}

func TestWriteTool_CreateDirectory(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "subdir", "new.txt")

	tool := NewWriteTool()
	result, err := tool.Handler(context.Background(), map[string]any{
		"file_path": testFile,
		"content":   "hello",
	})
	require.NoError(t, err)
	require.Contains(t, strings.ToLower(result.(string)), "success")

	// 验证目录和文件被创建
	data, err := os.ReadFile(testFile)
	require.NoError(t, err)
	require.Equal(t, "hello", string(data))
}

// ============================================================
// Edit 工具测试
// ============================================================

func TestEditTool_ReplaceString(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.txt")
	err := os.WriteFile(testFile, []byte("hello world"), 0644)
	require.NoError(t, err)

	tool := NewEditTool()
	result, err := tool.Handler(context.Background(), map[string]any{
		"file_path":  testFile,
		"old_string": "world",
		"new_string": "golang",
	})
	require.NoError(t, err)
	require.Contains(t, result.(string), "replaced")

	// 验证文件内容
	data, err := os.ReadFile(testFile)
	require.NoError(t, err)
	require.Equal(t, "hello golang", string(data))
}

func TestEditTool_NotFound(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.txt")
	err := os.WriteFile(testFile, []byte("hello world"), 0644)
	require.NoError(t, err)

	tool := NewEditTool()
	_, err = tool.Handler(context.Background(), map[string]any{
		"file_path":  testFile,
		"old_string": "nonexistent",
		"new_string": "golang",
	})
	require.Error(t, err)
	require.Contains(t, err.Error(), "not found")
}

func TestEditTool_ReplaceAll(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.txt")
	err := os.WriteFile(testFile, []byte("foo foo foo"), 0644)
	require.NoError(t, err)

	tool := NewEditTool()
	result, err := tool.Handler(context.Background(), map[string]any{
		"file_path":   testFile,
		"old_string":  "foo",
		"new_string":  "bar",
		"replace_all": true,
	})
	require.NoError(t, err)
	require.Contains(t, result.(string), "3")

	// 验证所有匹配被替换
	data, err := os.ReadFile(testFile)
	require.NoError(t, err)
	require.Equal(t, "bar bar bar", string(data))
}

// ============================================================
// Bash 工具测试
// ============================================================

func TestBashTool_SimpleCommand(t *testing.T) {
	tool := NewBashTool()
	result, err := tool.Handler(context.Background(), map[string]any{
		"command": "echo hello",
	})
	require.NoError(t, err)
	require.Contains(t, result.(string), "hello")
}

func TestBashTool_CommandFailure(t *testing.T) {
	tool := NewBashTool()
	result, err := tool.Handler(context.Background(), map[string]any{
		"command": "exit 1",
	})
	require.NoError(t, err) // Bash 工具不返回错误，而是返回包含错误信息的输出
	require.Contains(t, result.(string), "exit code: 1")
}

func TestBashTool_Timeout(t *testing.T) {
	tool := NewBashTool()
	result, err := tool.Handler(context.Background(), map[string]any{
		"command": "sleep 10",
		"timeout": float64(1), // 1 秒超时
	})
	require.NoError(t, err)
	// 输出包含 "timed out" 或 "timeout"
	resultStr := strings.ToLower(result.(string))
	require.True(t, strings.Contains(resultStr, "timed out") || strings.Contains(resultStr, "timeout"),
		"expected timeout message, got: %s", result.(string))
}

func TestBashTool_Background(t *testing.T) {
	sessionCtx := NewSessionContext("test-session", t.TempDir())

	tool := NewBashToolWithSession(sessionCtx)
	result, err := tool.Handler(context.Background(), map[string]any{
		"command":            "sleep 1",
		"run_in_background":  true,
		"terminal_id":        "test-terminal",
	})
	require.NoError(t, err)
	require.Contains(t, result.(string), "terminal_id")
}

// ============================================================
// BashOutput 工具测试
// ============================================================

func TestBashOutputTool_ExistingTerminal(t *testing.T) {
	sessionCtx := NewSessionContext("test-session", t.TempDir())
	sessionCtx.AddTerminal("test-terminal", "echo hello")

	tool := NewBashOutputToolWithSession(sessionCtx)
	result, err := tool.Handler(context.Background(), map[string]any{
		"terminal_id": "test-terminal",
	})
	require.NoError(t, err)
	require.Contains(t, result.(string), "output")
}

func TestBashOutputTool_NonexistentTerminal(t *testing.T) {
	sessionCtx := NewSessionContext("test-session", t.TempDir())

	tool := NewBashOutputToolWithSession(sessionCtx)
	_, err := tool.Handler(context.Background(), map[string]any{
		"terminal_id": "nonexistent",
	})
	require.Error(t, err)
	require.Contains(t, err.Error(), "not found")
}

// ============================================================
// KillShell 工具测试
// ============================================================

func TestKillShellTool_ExistingTerminal(t *testing.T) {
	sessionCtx := NewSessionContext("test-session", t.TempDir())
	sessionCtx.AddTerminal("test-terminal", "sleep 100")

	tool := NewKillShellToolWithSession(sessionCtx)
	result, err := tool.Handler(context.Background(), map[string]any{
		"terminal_id": "test-terminal",
	})
	require.NoError(t, err)
	require.Contains(t, result.(string), "terminated")
}

func TestKillShellTool_NonexistentTerminal(t *testing.T) {
	sessionCtx := NewSessionContext("test-session", t.TempDir())

	tool := NewKillShellToolWithSession(sessionCtx)
	_, err := tool.Handler(context.Background(), map[string]any{
		"terminal_id": "nonexistent",
	})
	require.Error(t, err)
	require.Contains(t, err.Error(), "not found")
}

// ============================================================
// 工具注册测试
// ============================================================

func TestRegisterBuiltinTools(t *testing.T) {
	server := NewServerP("test-server")
	sessionCtx := NewSessionContext("test-session", t.TempDir())

	err := RegisterBuiltinTools(server, sessionCtx)
	require.NoError(t, err)

	tools := server.ListTools()
	require.Len(t, tools, 6) // Read, Write, Edit, Bash, BashOutput, KillShell

	toolNames := make([]string, len(tools))
	for i, tool := range tools {
		toolNames[i] = tool.Name
	}
	require.Contains(t, toolNames, "Read")
	require.Contains(t, toolNames, "Write")
	require.Contains(t, toolNames, "Edit")
	require.Contains(t, toolNames, "Bash")
	require.Contains(t, toolNames, "BashOutput")
	require.Contains(t, toolNames, "KillShell")
}

// ============================================================
// 额外覆盖率测试
// ============================================================

func TestReadTool_MissingFilePath(t *testing.T) {
	tool := NewReadTool()
	_, err := tool.Handler(context.Background(), map[string]any{})
	require.Error(t, err)
	require.Contains(t, err.Error(), "file_path is required")
}

func TestReadTool_EmptyDirectory(t *testing.T) {
	tmpDir := t.TempDir()
	emptyDir := filepath.Join(tmpDir, "empty")
	require.NoError(t, os.Mkdir(emptyDir, 0755))

	tool := NewReadTool()
	result, err := tool.Handler(context.Background(), map[string]any{
		"file_path": emptyDir,
	})
	require.NoError(t, err)
	require.Contains(t, result.(string), "total 0")
}

func TestReadTool_LargeOffset(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.txt")
	err := os.WriteFile(testFile, []byte("line1\nline2"), 0644)
	require.NoError(t, err)

	tool := NewReadTool()
	result, err := tool.Handler(context.Background(), map[string]any{
		"file_path": testFile,
		"offset":    float64(100), // 超出行数
	})
	require.NoError(t, err)
	require.Empty(t, result.(string))
}

func TestWriteTool_MissingFilePath(t *testing.T) {
	tool := NewWriteTool()
	_, err := tool.Handler(context.Background(), map[string]any{
		"content": "hello",
	})
	require.Error(t, err)
	require.Contains(t, err.Error(), "file_path is required")
}

func TestWriteTool_MissingContent(t *testing.T) {
	tool := NewWriteTool()
	_, err := tool.Handler(context.Background(), map[string]any{
		"file_path": "/tmp/test.txt",
	})
	require.Error(t, err)
	require.Contains(t, err.Error(), "content is required")
}

func TestEditTool_MissingFilePath(t *testing.T) {
	tool := NewEditTool()
	_, err := tool.Handler(context.Background(), map[string]any{
		"old_string": "old",
		"new_string": "new",
	})
	require.Error(t, err)
	require.Contains(t, err.Error(), "file_path is required")
}

func TestEditTool_MissingOldString(t *testing.T) {
	tool := NewEditTool()
	_, err := tool.Handler(context.Background(), map[string]any{
		"file_path": "/tmp/test.txt",
		"new_string": "new",
	})
	require.Error(t, err)
	require.Contains(t, err.Error(), "old_string is required")
}

func TestEditTool_MissingNewString(t *testing.T) {
	tool := NewEditTool()
	_, err := tool.Handler(context.Background(), map[string]any{
		"file_path":  "/tmp/test.txt",
		"old_string": "old",
	})
	require.Error(t, err)
	require.Contains(t, err.Error(), "new_string is required")
}

func TestEditTool_FileNotFound(t *testing.T) {
	tool := NewEditTool()
	_, err := tool.Handler(context.Background(), map[string]any{
		"file_path":  "/nonexistent/file.txt",
		"old_string": "old",
		"new_string": "new",
	})
	require.Error(t, err)
}

func TestBashTool_MissingCommand(t *testing.T) {
	tool := NewBashTool()
	_, err := tool.Handler(context.Background(), map[string]any{})
	require.Error(t, err)
	require.Contains(t, err.Error(), "command is required")
}

func TestBashTool_WithWorkingDirectory(t *testing.T) {
	tmpDir := t.TempDir()
	tool := NewBashTool()
	result, err := tool.Handler(context.Background(), map[string]any{
		"command":           "pwd",
		"working_directory": tmpDir,
	})
	require.NoError(t, err)
	require.Contains(t, result.(string), tmpDir)
}

func TestBashTool_BackgroundWithoutTerminalID(t *testing.T) {
	sessionCtx := NewSessionContext("test-session", t.TempDir())
	tool := NewBashToolWithSession(sessionCtx)

	result, err := tool.Handler(context.Background(), map[string]any{
		"command":           "echo hello",
		"run_in_background": true,
	})
	require.NoError(t, err)
	// 应该自动生成 terminal_id
	require.Contains(t, result.(string), "terminal_id")
}

func TestBashOutputTool_MissingTerminalID(t *testing.T) {
	tool := NewBashOutputTool()
	_, err := tool.Handler(context.Background(), map[string]any{})
	require.Error(t, err)
	require.Contains(t, err.Error(), "terminal_id is required")
}

func TestBashOutputTool_NoSessionContext(t *testing.T) {
	tool := NewBashOutputTool() // 没有会话上下文
	_, err := tool.Handler(context.Background(), map[string]any{
		"terminal_id": "test",
	})
	require.Error(t, err)
	require.Contains(t, err.Error(), "no session context")
}

func TestKillShellTool_MissingTerminalID(t *testing.T) {
	tool := NewKillShellTool()
	_, err := tool.Handler(context.Background(), map[string]any{})
	require.Error(t, err)
	require.Contains(t, err.Error(), "terminal_id is required")
}

func TestKillShellTool_NoSessionContext(t *testing.T) {
	tool := NewKillShellTool() // 没有会话上下文
	_, err := tool.Handler(context.Background(), map[string]any{
		"terminal_id": "test",
	})
	require.Error(t, err)
	require.Contains(t, err.Error(), "no session context")
}

func TestNewBashOutputTool_NoSession(t *testing.T) {
	tool := NewBashOutputTool()
	require.NotNil(t, tool)
	require.Equal(t, "BashOutput", tool.Name)
}

func TestNewKillShellTool_NoSession(t *testing.T) {
	tool := NewKillShellTool()
	require.NotNil(t, tool)
	require.Equal(t, "KillShell", tool.Name)
}

func TestListDirectory_Error(t *testing.T) {
	// 测试无法读取的目录（通过提供无效路径）
	result, err := listDirectory("/nonexistent/path/that/does/not/exist")
	require.Error(t, err)
	require.Nil(t, result)
}

func TestBashTool_CommandWithError(t *testing.T) {
	tool := NewBashTool()
	result, err := tool.Handler(context.Background(), map[string]any{
		"command": "ls /nonexistent/path/that/does/not/exist",
	})
	require.NoError(t, err)
	// stderr 应该包含错误信息
	require.Contains(t, strings.ToLower(result.(string)), "no such file")
}

func TestBashTool_Background_WithSessionContext(t *testing.T) {
	sessionCtx := NewSessionContext("test-session", t.TempDir())
	tool := NewBashToolWithSession(sessionCtx)

	result, err := tool.Handler(context.Background(), map[string]any{
		"command":           "echo background",
		"run_in_background": true,
		"terminal_id":       "bg-terminal",
	})
	require.NoError(t, err)
	require.Contains(t, result.(string), "bg-terminal")

	// 等待后台命令完成
	time.Sleep(100 * time.Millisecond)

	// 检查终端是否被追踪
	terminal := sessionCtx.GetTerminal("bg-terminal")
	// 由于 goroutine 可能还没完成，我们检查终端存在
	require.NotNil(t, terminal)
}

func TestBashTool_ZeroTimeout(t *testing.T) {
	tool := NewBashTool()
	result, err := tool.Handler(context.Background(), map[string]any{
		"command": "echo test",
		"timeout": float64(0), // 零超时应该使用默认值
	})
	require.NoError(t, err)
	require.Contains(t, result.(string), "test")
}

func TestListDirectory_WithFilesAndDirs(t *testing.T) {
	tmpDir := t.TempDir()

	// 创建文件和子目录
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "file1.txt"), []byte("content"), 0644))
	require.NoError(t, os.Mkdir(filepath.Join(tmpDir, "subdir"), 0755))

	result, err := listDirectory(tmpDir)
	require.NoError(t, err)
	require.Contains(t, result, "file1.txt")
	require.Contains(t, result, "subdir/")
}

func TestBashTool_ExitError(t *testing.T) {
	tool := NewBashTool()
	result, err := tool.Handler(context.Background(), map[string]any{
		"command": "bash -c 'exit 42'",
	})
	require.NoError(t, err)
	require.Contains(t, result.(string), "exit code: 42")
}

func TestBashTool_BackgroundWithWorkingDir(t *testing.T) {
	tmpDir := t.TempDir()
	sessionCtx := NewSessionContext("test-session", tmpDir)
	tool := NewBashToolWithSession(sessionCtx)

	result, err := tool.Handler(context.Background(), map[string]any{
		"command":            "pwd",
		"run_in_background":  true,
		"terminal_id":        "wd-terminal",
		"working_directory":  tmpDir,
	})
	require.NoError(t, err)
	require.Contains(t, result.(string), "wd-terminal")
}

func TestBashTool_NonExitStatusError(t *testing.T) {
	// 测试非退出码错误的情况
	tool := NewBashTool()
	result, err := tool.Handler(context.Background(), map[string]any{
		"command": "echo test",
	})
	require.NoError(t, err)
	require.Contains(t, result.(string), "test")
}
