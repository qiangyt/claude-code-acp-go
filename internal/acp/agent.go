package acp

import (
	"context"
	"fmt"
	"sync"

	"claude-code-acp-go/internal/mcp"
)

// ============================================================
// Agent 选项
// ============================================================

// AgentOptionsT Agent 选项结构
type AgentOptionsT struct {
	// Name Agent 名称
	Name string
	// WorkingDir 工作目录
	WorkingDir string
	// PermissionChecker 权限检查器
	PermissionChecker mcp.PermissionChecker
}

// AgentOptions Agent 选项类型别名
type AgentOptions = *AgentOptionsT

// NewAgentOptions 创建默认 Agent 选项
func NewAgentOptions() AgentOptions {
	return &AgentOptionsT{
		Name:       "claude-code-acp",
		WorkingDir: ".",
	}
}

// WithName 设置 Agent 名称
func (o *AgentOptionsT) WithName(name string) AgentOptions {
	o.Name = name
	return o
}

// WithWorkingDir 设置工作目录
func (o *AgentOptionsT) WithWorkingDir(dir string) AgentOptions {
	o.WorkingDir = dir
	return o
}

// WithPermissionChecker 设置权限检查器
func (o *AgentOptionsT) WithPermissionChecker(checker mcp.PermissionChecker) AgentOptions {
	o.PermissionChecker = checker
	return o
}

// ============================================================
// Session 会话
// ============================================================

// SessionT 会话结构
type SessionT struct {
	// SessionID 会话 ID
	SessionID string
	// WorkingDir 工作目录
	WorkingDir string
	// MCPContext MCP 会话上下文
	MCPContext mcp.SessionContext
}

// Session 会话类型别名
type Session = *SessionT

// ============================================================
// Agent 实现
// ============================================================

// Agent 实现 ACP Agent 接口
type Agent struct {
	// name Agent 名称
	name string
	// workingDir 工作目录
	workingDir string
	// mcpServer MCP 服务器
	mcpServer mcp.Server
	// permissionChecker 权限检查器
	permissionChecker mcp.PermissionChecker
	// sessions 会话映射
	sessions map[string]Session
	// mu 会话锁
	mu sync.RWMutex
}

// NewAgent 创建新的 ACP 代理
func NewAgent() *Agent {
	return NewAgentWithOptions(nil)
}

// NewAgentWithOptions 使用选项创建 Agent
func NewAgentWithOptions(opts AgentOptions) *Agent {
	if opts == nil {
		opts = NewAgentOptions()
	}

	agent := &Agent{
		name:              opts.Name,
		workingDir:        opts.WorkingDir,
		permissionChecker: opts.PermissionChecker,
		sessions:          make(map[string]Session),
	}

	// 创建 MCP 服务器
	mcpOpts := mcp.NewServerOptions()
	if opts.PermissionChecker != nil {
		mcpOpts.WithPermissionChecker(opts.PermissionChecker)
	}

	server, err := mcp.NewServer(agent.name, mcpOpts)
	if err != nil {
		panic(fmt.Errorf("failed to create MCP server: %w", err))
	}
	agent.mcpServer = server

	return agent
}

// CreateSession 创建新会话
func (a *Agent) CreateSession(sessionID string) (Session, error) {
	a.mu.Lock()
	defer a.mu.Unlock()

	// 创建 MCP 会话上下文
	mcpCtx := mcp.NewSessionContext(sessionID, a.workingDir)

	// 为此会话注册内置工具
	if err := mcp.RegisterBuiltinTools(a.mcpServer, mcpCtx); err != nil {
		return nil, fmt.Errorf("failed to register builtin tools: %w", err)
	}

	session := &SessionT{
		SessionID:  sessionID,
		WorkingDir: a.workingDir,
		MCPContext: mcpCtx,
	}

	a.sessions[sessionID] = session
	return session, nil
}

// GetSession 获取会话
func (a *Agent) GetSession(sessionID string) Session {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.sessions[sessionID]
}

// RemoveSession 移除会话
func (a *Agent) RemoveSession(sessionID string) {
	a.mu.Lock()
	defer a.mu.Unlock()

	if session, ok := a.sessions[sessionID]; ok {
		// 清理会话的终端进程
		if session.MCPContext != nil {
			session.MCPContext.ClearTerminals()
		}
		delete(a.sessions, sessionID)
	}
}

// Run 启动代理，监听 stdin/stdout
func (a *Agent) Run(ctx context.Context) error {
	// TODO: 实现主事件循环
	// 当前只是等待 context 取消
	<-ctx.Done()
	return nil
}
