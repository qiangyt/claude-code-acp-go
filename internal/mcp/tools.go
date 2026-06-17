// Package mcp 实现 MCP (Model Context Protocol) 服务器
//
// 本文件实现 MCP 工具适配层，将 ACP 工具暴露为 MCP 工具
package mcp

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/qiangyt/go-comm/v2"
)

// ============================================================
// Read 工具
// ============================================================

// ReadToolInput Read 工具输入参数
type ReadToolInput struct {
	FilePath string  `json:"file_path"`
	Offset   float64 `json:"offset,omitempty"`
	Limit    float64 `json:"limit,omitempty"`
}

// NewReadTool 创建 Read 工具
func NewReadTool() Tool {
	return Tool{
		Name:        "Read",
		Description: "读取文件内容。支持读取部分行（通过 offset 和 limit 参数）。如果路径是目录，返回目录列表。",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"file_path": map[string]any{
					"type":        "string",
					"description": "要读取的文件绝对路径",
				},
				"offset": map[string]any{
					"type":        "number",
					"description": "起始行号（从 0 开始）",
				},
				"limit": map[string]any{
					"type":        "number",
					"description": "读取的最大行数",
				},
			},
			"required": []string{"file_path"},
		},
		Handler: handleReadTool,
	}
}

func handleReadTool(ctx context.Context, args map[string]any) (any, error) {
	filePath, ok := args["file_path"].(string)
	if !ok {
		return nil, fmt.Errorf("file_path is required")
	}

	// 检查路径是否存在
	info, err := os.Stat(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to access path: %w", err)
	}

	// 如果是目录，返回目录列表
	if info.IsDir() {
		return listDirectory(filePath)
	}

	// 读取文件
	content, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	// 处理 offset 和 limit
	lines := strings.Split(string(content), "\n")
	offset := 0
	limit := len(lines)

	if o, ok := args["offset"].(float64); ok {
		offset = int(o)
	}
	if l, ok := args["limit"].(float64); ok {
		limit = int(l)
	}

	// 边界检查
	if offset > len(lines) {
		offset = len(lines)
	}
	end := offset + limit
	if end > len(lines) {
		end = len(lines)
	}

	result := strings.Join(lines[offset:end], "\n")
	return result, nil
}

func listDirectory(dirPath string) (any, error) {
	entries, err := os.ReadDir(dirPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read directory: %w", err)
	}

	var buf bytes.Buffer
	buf.WriteString(fmt.Sprintf("total %d\n", len(entries)))
	for _, entry := range entries {
		info, err := entry.Info()
		if err != nil {
			continue
		}
		mode := info.Mode().String()
		size := info.Size()
		name := entry.Name()
		if entry.IsDir() {
			name += "/"
		}
		buf.WriteString(fmt.Sprintf("%s %8d %s\n", mode, size, name))
	}
	return buf.String(), nil
}

// ============================================================
// Write 工具
// ============================================================

// WriteToolInput Write 工具输入参数
type WriteToolInput struct {
	FilePath string `json:"file_path"`
	Content  string `json:"content"`
}

// NewWriteTool 创建 Write 工具
func NewWriteTool() Tool {
	return Tool{
		Name:        "Write",
		Description: "写入文件内容。如果文件存在则覆盖，如果不存在则创建（包括父目录）。",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"file_path": map[string]any{
					"type":        "string",
					"description": "要写入的文件绝对路径",
				},
				"content": map[string]any{
					"type":        "string",
					"description": "要写入的内容",
				},
			},
			"required": []string{"file_path", "content"},
		},
		Handler: handleWriteTool,
	}
}

func handleWriteTool(ctx context.Context, args map[string]any) (any, error) {
	filePath, ok := args["file_path"].(string)
	if !ok {
		return nil, fmt.Errorf("file_path is required")
	}
	content, ok := args["content"].(string)
	if !ok {
		return nil, fmt.Errorf("content is required")
	}

	// 创建父目录（如果不存在）
	dir := filepath.Dir(filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create directory: %w", err)
	}

	// 写入文件
	if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
		return nil, fmt.Errorf("failed to write file: %w", err)
	}

	return fmt.Sprintf("Successfully wrote %d bytes to %s", len(content), filePath), nil
}

// ============================================================
// Edit 工具
// ============================================================

// EditToolInput Edit 工具输入参数
type EditToolInput struct {
	FilePath    string `json:"file_path"`
	OldString   string `json:"old_string"`
	NewString   string `json:"new_string"`
	ReplaceAll  bool   `json:"replace_all,omitempty"`
}

// NewEditTool 创建 Edit 工具
func NewEditTool() Tool {
	return Tool{
		Name:        "Edit",
		Description: "编辑文件内容，执行字符串替换。可以替换所有匹配或仅替换第一个。",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"file_path": map[string]any{
					"type":        "string",
					"description": "要编辑的文件绝对路径",
				},
				"old_string": map[string]any{
					"type":        "string",
					"description": "要替换的旧字符串",
				},
				"new_string": map[string]any{
					"type":        "string",
					"description": "替换后的新字符串",
				},
				"replace_all": map[string]any{
					"type":        "boolean",
					"description": "是否替换所有匹配（默认只替换第一个）",
				},
			},
			"required": []string{"file_path", "old_string", "new_string"},
		},
		Handler: handleEditTool,
	}
}

func handleEditTool(ctx context.Context, args map[string]any) (any, error) {
	filePath, ok := args["file_path"].(string)
	if !ok {
		return nil, fmt.Errorf("file_path is required")
	}
	oldString, ok := args["old_string"].(string)
	if !ok {
		return nil, fmt.Errorf("old_string is required")
	}
	newString, ok := args["new_string"].(string)
	if !ok {
		return nil, fmt.Errorf("new_string is required")
	}
	replaceAll := false
	if r, ok := args["replace_all"].(bool); ok {
		replaceAll = r
	}

	// 读取文件
	content, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	// 检查是否存在匹配
	if !strings.Contains(string(content), oldString) {
		return nil, fmt.Errorf("old_string not found in file")
	}

	// 执行替换
	var newContent string
	var count int
	if replaceAll {
		newContent = strings.ReplaceAll(string(content), oldString, newString)
		count = strings.Count(string(content), oldString)
	} else {
		newContent = strings.Replace(string(content), oldString, newString, 1)
		count = 1
	}

	// 写入文件
	if err := os.WriteFile(filePath, []byte(newContent), 0644); err != nil {
		return nil, fmt.Errorf("failed to write file: %w", err)
	}

	return fmt.Sprintf("Successfully replaced %d occurrence(s) in %s", count, filePath), nil
}

// ============================================================
// Bash 工具
// ============================================================

// BashToolInput Bash 工具输入参数
type BashToolInput struct {
	Command           string  `json:"command"`
	Timeout           float64 `json:"timeout,omitempty"`
	RunInBackground   bool    `json:"run_in_background,omitempty"`
	TerminalID        string  `json:"terminal_id,omitempty"`
	WorkingDirectory  string  `json:"working_directory,omitempty"`
}

// DefaultBashTimeout 默认 Bash 超时时间（秒）
const DefaultBashTimeout = 120

// NewBashTool 创建 Bash 工具（无会话上下文）
func NewBashTool() Tool {
	return NewBashToolWithSession(nil)
}

// NewBashToolWithSession 创建带会话上下文的 Bash 工具
func NewBashToolWithSession(sessionCtx SessionContext) Tool {
	return Tool{
		Name:        "Bash",
		Description: "执行 Shell 命令。支持超时控制和后台运行。",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"command": map[string]any{
					"type":        "string",
					"description": "要执行的 Shell 命令",
				},
				"timeout": map[string]any{
					"type":        "number",
					"description": fmt.Sprintf("超时时间（秒），默认 %d", DefaultBashTimeout),
				},
				"run_in_background": map[string]any{
					"type":        "boolean",
					"description": "是否在后台运行",
				},
				"terminal_id": map[string]any{
					"type":        "string",
					"description": "终端 ID（后台运行时使用）",
				},
				"working_directory": map[string]any{
					"type":        "string",
					"description": "工作目录",
				},
			},
			"required": []string{"command"},
		},
		Handler: func(ctx context.Context, args map[string]any) (any, error) {
			return handleBashTool(ctx, args, sessionCtx)
		},
	}
}

func handleBashTool(ctx context.Context, args map[string]any, sessionCtx SessionContext) (any, error) {
	command, ok := args["command"].(string)
	if !ok {
		return nil, fmt.Errorf("command is required")
	}

	timeout := DefaultBashTimeout
	if t, ok := args["timeout"].(float64); ok && t > 0 {
		timeout = int(t)
	}

	runInBackground := false
	if r, ok := args["run_in_background"].(bool); ok {
		runInBackground = r
	}

	workingDir := ""
	if w, ok := args["working_directory"].(string); ok {
		workingDir = w
	}

	terminalID := ""
	if t, ok := args["terminal_id"].(string); ok {
		terminalID = t
	}

	// 后台运行
	if runInBackground {
		return runBashBackground(command, workingDir, terminalID, sessionCtx)
	}

	// 前台运行
	return runBashForeground(command, workingDir, timeout)
}

// getBlacklistRules 获取黑名单规则（前台和后台命令共用）
func getBlacklistRules() []comm.CommandRule {
	return []comm.CommandRule{
		// 阻止危险的 rm -rf 操作
		comm.NewCommandRule("rm", comm.MatchExact).
			WithArgsFilter(comm.ArgMatcher{Position: -1, Pattern: "-r*", Mode: comm.MatchGlob}),
		// 阻止格式化命令
		comm.NewCommandRule("mkfs*", comm.MatchGlob),
		// 阻止 dd 命令（可能覆盖磁盘）
		comm.NewCommandRule("dd", comm.MatchExact),
	}
}

// createSecurityChecker 创建安全检查器（用于后台命令预检查）
func createSecurityChecker() comm.SecurityChecker {
	checker := comm.NewSecurityChecker()
	for _, rule := range getBlacklistRules() {
		checker.WithBlacklist(rule)
	}
	return checker
}

func runBashForeground(command, workingDir string, timeout int) (any, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeout)*time.Second)
	defer cancel()

	// 使用 GoshExecutor 执行命令
	// 配置：设置超时时间和安全规则（与 runBashBackground 保持一致）
	config := comm.DefaultGoshConfig().
		WithKillTimeout(time.Duration(timeout) * time.Second).
		WithBlacklist(getBlacklistRules()...)

	executor := comm.NewGoshExecutor(config)

	var out bytes.Buffer
	err := executor.Run(ctx, workingDir, command, nil, &out, &out)

	output := out.String()

	// 检查超时
	if ctx.Err() == context.DeadlineExceeded {
		return fmt.Sprintf("Command timed out after %d seconds\n%s", timeout, output), nil
	}

	// 检查错误
	if err != nil {
		errMsg := err.Error()
		// 尝试从错误信息中提取退出码（格式: "exit status N"）
		if strings.Contains(errMsg, "exit status") {
			// 提取退出码
			parts := strings.Split(errMsg, "exit status ")
			if len(parts) > 1 {
				exitCode := strings.TrimSpace(parts[len(parts)-1])
				// 取第一个空格前的内容作为退出码
				if spaceIdx := strings.Index(exitCode, " "); spaceIdx > 0 {
					exitCode = exitCode[:spaceIdx]
				}
				return fmt.Sprintf("%s\nexit code: %s", output, exitCode), nil
			}
		}
		return fmt.Sprintf("%s\nerror: %v", output, err), nil
	}

	return output, nil
}

func runBashBackground(command, workingDir, terminalID string, sessionCtx SessionContext) (any, error) {
	if terminalID == "" {
		terminalID = fmt.Sprintf("terminal-%d", time.Now().UnixNano())
	}

	// 安全预检查：解析命令并检查是否允许执行
	extractor := comm.NewCommandExtractor()
	cmds, err := extractor.Extract(command)
	if err != nil {
		return nil, fmt.Errorf("failed to parse command: %w", err)
	}

	checker := createSecurityChecker()
	if err := checker.Check(cmds); err != nil {
		return nil, fmt.Errorf("command blocked: %w", err)
	}

	// 记录终端
	if sessionCtx != nil {
		sessionCtx.AddTerminal(terminalID, command)
	}

	// 检查通过，启动后台命令
	cmd := exec.Command("bash", "-c", command)
	if workingDir != "" {
		cmd.Dir = workingDir
	}

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("failed to start command: %w", err)
	}

	// 在 goroutine 中等待命令完成并更新状态
	go func() {
		err := cmd.Wait()
		if sessionCtx != nil {
			if output := stdout.String() + stderr.String(); output != "" {
				sessionCtx.UpdateTerminalOutput(terminalID, output)
			}
			if err != nil {
				sessionCtx.UpdateTerminalStatus(terminalID, TerminalStatusFailed)
			} else {
				sessionCtx.UpdateTerminalStatus(terminalID, TerminalStatusCompleted)
			}
		}
	}()

	return fmt.Sprintf(`{"terminal_id": "%s", "status": "running"}`, terminalID), nil
}

// ============================================================
// BashOutput 工具
// ============================================================

// NewBashOutputTool 创建 BashOutput 工具（无会话上下文）
func NewBashOutputTool() Tool {
	return NewBashOutputToolWithSession(nil)
}

// NewBashOutputToolWithSession 创建带会话上下文的 BashOutput 工具
func NewBashOutputToolWithSession(sessionCtx SessionContext) Tool {
	return Tool{
		Name:        "BashOutput",
		Description: "获取后台命令的输出。",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"terminal_id": map[string]any{
					"type":        "string",
					"description": "终端 ID",
				},
			},
			"required": []string{"terminal_id"},
		},
		Handler: func(ctx context.Context, args map[string]any) (any, error) {
			return handleBashOutputTool(ctx, args, sessionCtx)
		},
	}
}

func handleBashOutputTool(ctx context.Context, args map[string]any, sessionCtx SessionContext) (any, error) {
	terminalID, ok := args["terminal_id"].(string)
	if !ok {
		return nil, fmt.Errorf("terminal_id is required")
	}

	if sessionCtx == nil {
		return nil, fmt.Errorf("no session context available")
	}

	terminal := sessionCtx.GetTerminal(terminalID)
	if terminal == nil {
		return nil, fmt.Errorf("terminal not found: %s", terminalID)
	}

	return fmt.Sprintf(`{"terminal_id": "%s", "status": "%s", "output": %q}`,
		terminalID, terminal.Status, terminal.Output), nil
}

// ============================================================
// KillShell 工具
// ============================================================

// NewKillShellTool 创建 KillShell 工具（无会话上下文）
func NewKillShellTool() Tool {
	return NewKillShellToolWithSession(nil)
}

// NewKillShellToolWithSession 创建带会话上下文的 KillShell 工具
func NewKillShellToolWithSession(sessionCtx SessionContext) Tool {
	return Tool{
		Name:        "KillShell",
		Description: "终止后台命令。",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"terminal_id": map[string]any{
					"type":        "string",
					"description": "终端 ID",
				},
			},
			"required": []string{"terminal_id"},
		},
		Handler: func(ctx context.Context, args map[string]any) (any, error) {
			return handleKillShellTool(ctx, args, sessionCtx)
		},
	}
}

func handleKillShellTool(ctx context.Context, args map[string]any, sessionCtx SessionContext) (any, error) {
	terminalID, ok := args["terminal_id"].(string)
	if !ok {
		return nil, fmt.Errorf("terminal_id is required")
	}

	if sessionCtx == nil {
		return nil, fmt.Errorf("no session context available")
	}

	terminal := sessionCtx.GetTerminal(terminalID)
	if terminal == nil {
		return nil, fmt.Errorf("terminal not found: %s", terminalID)
	}

	// 更新终端状态为已终止
	sessionCtx.UpdateTerminalStatus(terminalID, TerminalStatusKilled)
	sessionCtx.RemoveTerminal(terminalID)

	return fmt.Sprintf(`{"terminal_id": "%s", "status": "terminated"}`, terminalID), nil
}

// ============================================================
// 工具注册
// ============================================================

// RegisterBuiltinTools 注册所有内置工具到服务器
func RegisterBuiltinTools(server Server, sessionCtx SessionContext) error {
	tools := []Tool{
		NewReadTool(),
		NewWriteTool(),
		NewEditTool(),
		NewBashToolWithSession(sessionCtx),
		NewBashOutputToolWithSession(sessionCtx),
		NewKillShellToolWithSession(sessionCtx),
	}

	return server.RegisterTools(tools...)
}
