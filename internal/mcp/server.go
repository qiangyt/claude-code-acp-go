// Package mcp 实现 MCP (Model Context Protocol) 服务器
//
// 本包为 ACP 提供 MCP 工具支持，将 ACP 工具暴露为 MCP 工具供 SDK 使用。
package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"

	sdktypes "github.com/schlunsen/claude-agent-sdk-go/types"
)

// ============================================================
// 权限决策类型
// ============================================================

// PermissionDecision 权限决策结果
type PermissionDecision string

const (
	// PermissionDecisionAllow 允许执行
	PermissionDecisionAllow PermissionDecision = "allow"
	// PermissionDecisionDeny 拒绝执行
	PermissionDecisionDeny PermissionDecision = "deny"
	// PermissionDecisionAsk 需要询问用户
	PermissionDecisionAsk PermissionDecision = "ask"
)

// ============================================================
// 工具类型
// ============================================================

// ToolHandler 工具处理函数类型
type ToolHandler func(ctx context.Context, args map[string]any) (any, error)

// Tool MCP 工具定义
type Tool struct {
	// Name 工具名称（唯一标识）
	Name string
	// Description 工具描述
	Description string
	// InputSchema 输入参数 JSON Schema
	InputSchema map[string]any
	// Handler 工具处理函数
	Handler ToolHandler
}

// Validate 验证工具配置
func (t *Tool) Validate() error {
	if t.Name == "" {
		return fmt.Errorf("tool name is required")
	}
	if t.Handler == nil {
		return fmt.Errorf("tool handler is required")
	}
	return nil
}

// ============================================================
// 权限检查器接口
// ============================================================

// PermissionChecker 权限检查器接口
type PermissionChecker interface {
	// Check 检查工具调用权限
	Check(ctx context.Context, toolName string, params map[string]any) PermissionDecision
}

// ============================================================
// 会话上下文
// ============================================================

// SessionContextT 会话上下文结构
type SessionContextT struct {
	// SessionID 会话 ID
	SessionID string
	// WorkingDir 工作目录
	WorkingDir string
	// Terminals 终端进程映射
	Terminals map[string]*TerminalInfo
	// mu 读写锁
	mu sync.RWMutex
}

// Session 会话上下文类型别名
type SessionContext = *SessionContextT

// TerminalInfo 终端进程信息
type TerminalInfo struct {
	// ID 终端 ID
	ID string
	// Cmd 执行的命令
	Cmd string
	// Status 终端状态
	Status TerminalStatus
	// Output 累积输出
	Output string
}

// TerminalStatus 终端状态
type TerminalStatus string

const (
	// TerminalStatusRunning 运行中
	TerminalStatusRunning TerminalStatus = "running"
	// TerminalStatusCompleted 已完成
	TerminalStatusCompleted TerminalStatus = "completed"
	// TerminalStatusFailed 已失败
	TerminalStatusFailed TerminalStatus = "failed"
	// TerminalStatusKilled 已终止
	TerminalStatusKilled TerminalStatus = "killed"
)

// NewSessionContext 创建新的会话上下文
func NewSessionContext(sessionID, workingDir string) SessionContext {
	return &SessionContextT{
		SessionID:  sessionID,
		WorkingDir: workingDir,
		Terminals:  make(map[string]*TerminalInfo),
	}
}

// AddTerminal 添加终端
func (s *SessionContextT) AddTerminal(id, cmd string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Terminals[id] = &TerminalInfo{
		ID:     id,
		Cmd:    cmd,
		Status: TerminalStatusRunning,
	}
}

// GetTerminal 获取终端信息
func (s *SessionContextT) GetTerminal(id string) *TerminalInfo {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.Terminals[id]
}

// UpdateTerminalOutput 更新终端输出
func (s *SessionContextT) UpdateTerminalOutput(id, output string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if terminal, ok := s.Terminals[id]; ok {
		terminal.Output += output
	}
}

// UpdateTerminalStatus 更新终端状态
func (s *SessionContextT) UpdateTerminalStatus(id string, status TerminalStatus) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if terminal, ok := s.Terminals[id]; ok {
		terminal.Status = status
	}
}

// RemoveTerminal 移除终端
func (s *SessionContextT) RemoveTerminal(id string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.Terminals, id)
}

// ClearTerminals 清除所有终端
func (s *SessionContextT) ClearTerminals() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Terminals = make(map[string]*TerminalInfo)
}

// ============================================================
// 服务器选项
// ============================================================

// ServerOptionsT 服务器选项结构
type ServerOptionsT struct {
	// PermissionChecker 权限检查器
	PermissionChecker PermissionChecker
	// SessionContext 会话上下文
	SessionContext SessionContext
}

// ServerOptions 服务器选项类型别名
type ServerOptions = *ServerOptionsT

// NewServerOptions 创建默认服务器选项
func NewServerOptions() ServerOptions {
	return &ServerOptionsT{}
}

// WithPermissionChecker 设置权限检查器
func (o *ServerOptionsT) WithPermissionChecker(checker PermissionChecker) ServerOptions {
	o.PermissionChecker = checker
	return o
}

// WithSessionContext 设置会话上下文
func (o *ServerOptionsT) WithSessionContext(ctx SessionContext) ServerOptions {
	o.SessionContext = ctx
	return o
}

// ============================================================
// MCP 服务器
// ============================================================

// ServerT MCP 服务器结构
type ServerT struct {
	// name 服务器名称
	name string
	// version 服务器版本
	version string
	// tools 工具映射
	tools map[string]*Tool
	// mu 读写锁
	mu sync.RWMutex
	// permissionChecker 权限检查器
	permissionChecker PermissionChecker
	// sessionContext 会话上下文
	sessionContext SessionContext
}

// Server MCP 服务器类型别名
type Server = *ServerT

// NewServer 创建新的 MCP 服务器
func NewServer(name string, opts ...ServerOptions) (Server, error) {
	if name == "" {
		return nil, fmt.Errorf("server name is required")
	}

	server := &ServerT{
		name:    name,
		version: "1.0.0",
		tools:   make(map[string]*Tool),
	}

	if len(opts) > 0 && opts[0] != nil {
		server.permissionChecker = opts[0].PermissionChecker
		server.sessionContext = opts[0].SessionContext
	}

	return server, nil
}

// NewServerP 创建新的 MCP 服务器（Panic 版本）
func NewServerP(name string) Server {
	server, err := NewServer(name, nil)
	if err != nil {
		panic(err)
	}
	return server
}

// Name 返回服务器名称
func (s *ServerT) Name() string {
	return s.name
}

// Version 返回服务器版本
func (s *ServerT) Version() string {
	return s.version
}

// ============================================================
// 工具注册
// ============================================================

// RegisterTool 注册单个工具
func (s *ServerT) RegisterTool(tool Tool) error {
	if err := tool.Validate(); err != nil {
		return err
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.tools[tool.Name]; exists {
		return fmt.Errorf("duplicate tool name: %s", tool.Name)
	}

	s.tools[tool.Name] = &tool
	return nil
}

// RegisterTools 批量注册工具
func (s *ServerT) RegisterTools(tools ...Tool) error {
	for _, tool := range tools {
		if err := s.RegisterTool(tool); err != nil {
			return err
		}
	}
	return nil
}

// ListTools 列出所有工具
func (s *ServerT) ListTools() []Tool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	tools := make([]Tool, 0, len(s.tools))
	for _, tool := range s.tools {
		tools = append(tools, *tool)
	}
	return tools
}

// ============================================================
// MCP 消息处理
// ============================================================

// HandleMessage 处理 MCP 消息
func (s *ServerT) HandleMessage(message map[string]any) (map[string]any, error) {
	// 提取 method 字段
	method, ok := message["method"].(string)
	if !ok {
		return s.errorResponse(message, -32600, "Invalid Request: missing or non-string method"), nil
	}

	// 路由到对应的处理器
	switch method {
	case "tools/list":
		return s.handleListTools(message)
	case "tools/call":
		return s.handleCallTool(message)
	default:
		return s.errorResponse(message, -32601, "Method not found: "+method), nil
	}
}

// handleListTools 处理 tools/list 请求
func (s *ServerT) handleListTools(message map[string]any) (map[string]any, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	tools := make([]map[string]any, 0, len(s.tools))
	for _, tool := range s.tools {
		toolMap := map[string]any{
			"name":        tool.Name,
			"description": tool.Description,
		}
		if tool.InputSchema != nil {
			toolMap["inputSchema"] = tool.InputSchema
		}
		tools = append(tools, toolMap)
	}

	id := message["id"]
	return map[string]any{
		"jsonrpc": "2.0",
		"id":      id,
		"result": map[string]any{
			"tools": tools,
		},
	}, nil
}

// handleCallTool 处理 tools/call 请求
func (s *ServerT) handleCallTool(message map[string]any) (map[string]any, error) {
	// 提取参数
	params, ok := message["params"].(map[string]any)
	if !ok {
		return s.errorResponse(message, -32602, "Invalid params: expected object"), nil
	}

	toolName, ok := params["name"].(string)
	if !ok {
		return s.errorResponse(message, -32602, "Invalid params: missing tool name"), nil
	}

	args, _ := params["arguments"].(map[string]any)
	if args == nil {
		args = make(map[string]any)
	}

	// 查找工具
	s.mu.RLock()
	tool, exists := s.tools[toolName]
	s.mu.RUnlock()

	if !exists {
		return s.errorResponse(message, -32603, "Tool not found: "+toolName), nil
	}

	// 执行工具处理器
	ctx := context.Background()
	result, err := tool.Handler(ctx, args)
	if err != nil {
		return s.errorResponse(message, -32603, "Tool execution failed: "+err.Error()), nil
	}

	// 格式化结果
	id := message["id"]
	content := s.formatResult(result)

	return map[string]any{
		"jsonrpc": "2.0",
		"id":      id,
		"result": map[string]any{
			"content": content,
		},
	}, nil
}

// formatResult 将工具结果格式化为 MCP 内容块
func (s *ServerT) formatResult(result any) []map[string]any {
	contentBlocks := make([]map[string]any, 0)

	switch v := result.(type) {
	case string:
		contentBlocks = append(contentBlocks, map[string]any{
			"type": "text",
			"text": v,
		})
	case map[string]any:
		if text, ok := v["text"].(string); ok {
			contentBlocks = append(contentBlocks, map[string]any{
				"type": "text",
				"text": text,
			})
		} else {
			// 返回 JSON 格式
			text := formatMapAsJSON(v)
			contentBlocks = append(contentBlocks, map[string]any{
				"type": "text",
				"text": text,
			})
		}
	case []any:
		for _, item := range v {
			if m, ok := item.(map[string]any); ok {
				contentBlocks = append(contentBlocks, m)
			}
		}
	default:
		text := formatAsJSON(v)
		contentBlocks = append(contentBlocks, map[string]any{
			"type": "text",
			"text": text,
		})
	}

	return contentBlocks
}

// errorResponse 创建错误响应
func (s *ServerT) errorResponse(message map[string]any, code int, errMsg string) map[string]any {
	id := message["id"]
	return map[string]any{
		"jsonrpc": "2.0",
		"id":      id,
		"error": map[string]any{
			"code":    code,
			"message": errMsg,
		},
	}
}

// formatAsJSON 格式化为 JSON 字符串
func formatAsJSON(v any) string {
	data, _ := json.Marshal(v)
	return string(data)
}

// formatMapAsJSON 格式化 map 为 JSON 字符串
func formatMapAsJSON(m map[string]any) string {
	data, _ := json.Marshal(m)
	return string(data)
}

// ============================================================
// SDK 兼容性
// ============================================================

// ToSDKMCPServer 转换为 SDK 的 MCPServer 接口
func (s *ServerT) ToSDKMCPServer() (sdktypes.MCPServer, error) {
	tools := make([]sdktypes.Tool, 0)
	for _, tool := range s.tools {
		// 需要转换函数类型
		handler := tool.Handler
		sdkHandler := func(ctx context.Context, args map[string]any) (any, error) {
			return handler(ctx, args)
		}
		tools = append(tools, sdktypes.Tool{
			Name:        tool.Name,
			Description: tool.Description,
			InputSchema: tool.InputSchema,
			Handler:     sdkHandler,
		})
	}

	return sdktypes.NewSDKMCPServer(s.name, tools...)
}
