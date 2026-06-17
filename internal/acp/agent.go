package acp

import (
	"context"
	"fmt"
	"sync"

	"claude-code-acp-go/internal/mcp"

	acpsdk "github.com/coder/acp-go-sdk"
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
	// conn ACP 连接（用于发送通知）
	conn *acpsdk.AgentSideConnection
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

// ============================================================
// 连接感知接口 (实现 acpsdk.AgentConnAware 接口)
// ============================================================

// SetAgentConnection 设置 ACP 连接
// 实现 acpsdk.AgentConnAware 接口，在连接建立后由框架调用
func (a *Agent) SetAgentConnection(conn *acpsdk.AgentSideConnection) {
	a.conn = conn
}

// ============================================================
// 通知发送方法
// ============================================================

// SendSessionUpdate 发送会话更新通知
func (a *Agent) SendSessionUpdate(ctx context.Context, sessionID string, update acpsdk.SessionUpdate) error {
	if a.conn == nil {
		return fmt.Errorf("agent connection not established")
	}
	return a.conn.SessionUpdate(ctx, acpsdk.SessionNotification{
		SessionId: acpsdk.SessionId(sessionID),
		Update:    update,
	})
}

// SendAgentMessage 发送代理消息
func (a *Agent) SendAgentMessage(ctx context.Context, sessionID, text string) error {
	return a.SendSessionUpdate(ctx, sessionID, acpsdk.UpdateAgentMessageText(text))
}

// SendAgentThought 发送代理思考
func (a *Agent) SendAgentThought(ctx context.Context, sessionID, text string) error {
	return a.SendSessionUpdate(ctx, sessionID, acpsdk.UpdateAgentThoughtText(text))
}

// SendPlanUpdate 发送计划更新
func (a *Agent) SendPlanUpdate(ctx context.Context, sessionID string, entries ...acpsdk.PlanEntry) error {
	return a.SendSessionUpdate(ctx, sessionID, acpsdk.UpdatePlan(entries...))
}

// StartToolCall 开始工具调用通知
func (a *Agent) StartToolCall(ctx context.Context, sessionID string, id acpsdk.ToolCallId, title string, opts ...acpsdk.ToolCallStartOpt) error {
	return a.SendSessionUpdate(ctx, sessionID, acpsdk.StartToolCall(id, title, opts...))
}

// UpdateToolCall 更新工具调用通知
func (a *Agent) UpdateToolCall(ctx context.Context, sessionID string, id acpsdk.ToolCallId, opts ...acpsdk.ToolCallUpdateOpt) error {
	return a.SendSessionUpdate(ctx, sessionID, acpsdk.UpdateToolCall(id, opts...))
}

// RequestPermission 请求权限
func (a *Agent) RequestPermission(ctx context.Context, sessionID string, toolCall acpsdk.ToolCallUpdate, options []acpsdk.PermissionOption) (acpsdk.RequestPermissionResponse, error) {
	if a.conn == nil {
		return acpsdk.RequestPermissionResponse{}, fmt.Errorf("agent connection not established")
	}
	return a.conn.RequestPermission(ctx, acpsdk.RequestPermissionRequest{
		SessionId: acpsdk.SessionId(sessionID),
		ToolCall:  toolCall,
		Options:   options,
	})
}

// ============================================================
// ACP 协议方法 (实现 acpsdk.Agent 接口)
// ============================================================

// Initialize 初始化代理
func (a *Agent) Initialize(ctx context.Context, req acpsdk.InitializeRequest) (acpsdk.InitializeResponse, error) {
	// 构建响应
	resp := acpsdk.InitializeResponse{
		ProtocolVersion: req.ProtocolVersion,
		AgentCapabilities: acpsdk.AgentCapabilities{
			LoadSession: true,
			PromptCapabilities: acpsdk.PromptCapabilities{
				Image:           true,
				Audio:           true,
				EmbeddedContext: true,
			},
			McpCapabilities: acpsdk.McpCapabilities{
				Http: true,
				Sse:  true,
			},
			SessionCapabilities: acpsdk.SessionCapabilities{},
		},
		AgentInfo: &acpsdk.Implementation{
			Name:    a.name,
			Version: "1.0.0",
		},
	}

	return resp, nil
}

// NewSession 创建新会话 (ACP 协议方法)
func (a *Agent) NewSession(ctx context.Context, req acpsdk.NewSessionRequest) (acpsdk.NewSessionResponse, error) {
	// 生成会话 ID
	sessionID := generateSessionID()

	// 确定工作目录
	cwd := a.workingDir
	if req.Cwd != "" {
		cwd = req.Cwd
	}

	// 创建会话
	session, err := a.CreateSession(string(sessionID))
	if err != nil {
		return acpsdk.NewSessionResponse{}, fmt.Errorf("failed to create session: %w", err)
	}

	// 更新工作目录
	session.WorkingDir = cwd

	// 构建响应
	resp := acpsdk.NewSessionResponse{
		SessionId: sessionID,
	}

	return resp, nil
}

// LoadSession 加载已存在的会话 (实现 acpsdk.AgentLoader 接口)
func (a *Agent) LoadSession(ctx context.Context, req acpsdk.LoadSessionRequest) (acpsdk.LoadSessionResponse, error) {
	sessionID := string(req.SessionId)

	// 检查会话是否存在
	session := a.GetSession(sessionID)
	if session == nil {
		return acpsdk.LoadSessionResponse{}, fmt.Errorf("session not found: %s", sessionID)
	}

	// 构建响应
	resp := acpsdk.LoadSessionResponse{}

	return resp, nil
}

// Cancel 取消会话操作
func (a *Agent) Cancel(ctx context.Context, params acpsdk.CancelNotification) error {
	sessionID := string(params.SessionId)

	// 检查会话是否存在
	session := a.GetSession(sessionID)
	if session == nil {
		// 会话不存在，幂等操作直接返回
		return nil
	}

	// 清理并移除会话
	a.RemoveSession(sessionID)

	return nil
}

// Prompt 处理用户提示
func (a *Agent) Prompt(ctx context.Context, params acpsdk.PromptRequest) (acpsdk.PromptResponse, error) {
	// 验证会话存在
	sessionID := string(params.SessionId)
	session := a.GetSession(sessionID)
	if session == nil {
		return acpsdk.PromptResponse{}, fmt.Errorf("session not found: %s", sessionID)
	}

	// 验证 prompt 内容
	if len(params.Prompt) == 0 {
		return acpsdk.PromptResponse{}, fmt.Errorf("prompt is required")
	}

	// TODO: 实现实际的 prompt 处理
	// 这里应该:
	// 1. 解析 ContentBlock (text, image, audio, resource_link, resource)
	// 2. 调用 AI 模型处理
	// 3. 通过 SendSessionUpdate 发送流式响应
	// 4. 返回最终结果

	// 当前返回未实现错误
	return acpsdk.PromptResponse{}, fmt.Errorf("not implemented")
}

// SetSessionMode 设置会话模式
func (a *Agent) SetSessionMode(ctx context.Context, params acpsdk.SetSessionModeRequest) (acpsdk.SetSessionModeResponse, error) {
	// TODO: 实现会话模式设置
	return acpsdk.SetSessionModeResponse{}, fmt.Errorf("not implemented")
}

// Authenticate 认证
func (a *Agent) Authenticate(ctx context.Context, params acpsdk.AuthenticateRequest) (acpsdk.AuthenticateResponse, error) {
	// TODO: 实现认证
	return acpsdk.AuthenticateResponse{}, fmt.Errorf("not implemented")
}

// ============================================================
// 辅助函数
// ============================================================

// sessionCounter 会话计数器
var sessionCounter uint64

// sessionMutex 会话计数器锁
var sessionMutex sync.Mutex

// generateSessionID 生成唯一的会话 ID
func generateSessionID() acpsdk.SessionId {
	sessionMutex.Lock()
	defer sessionMutex.Unlock()
	sessionCounter++
	return acpsdk.SessionId(fmt.Sprintf("session-%d", sessionCounter))
}
