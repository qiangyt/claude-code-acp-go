package api

import (
	"context"
	"io"

	"claude-code-acp-go/internal/acp"

	acpsdk "github.com/coder/acp-go-sdk"
)

// ClientT 是 claude-code-acp 的公开 API 客户端
type ClientT struct {
	agent *acp.Agent
}

// Client 客户端类型别名
type Client = *ClientT

// NewClient 创建新的客户端
func NewClient() Client {
	return NewClientWithOptions(nil)
}

// NewClientWithOptions 使用选项创建客户端
func NewClientWithOptions(opts Options) Client {
	if opts == nil {
		opts = NewOptions()
	}

	agent := acp.NewAgentWithOptions(opts.toAgentOptions())
	return &ClientT{agent: agent}
}

// Initialize 初始化代理
func (c *ClientT) Initialize(ctx context.Context, req acpsdk.InitializeRequest) (acpsdk.InitializeResponse, error) {
	return c.agent.Initialize(ctx, req)
}

// NewSession 创建新会话
func (c *ClientT) NewSession(ctx context.Context, req acpsdk.NewSessionRequest) (acpsdk.NewSessionResponse, error) {
	return c.agent.NewSession(ctx, req)
}

// LoadSession 加载已存在的会话
func (c *ClientT) LoadSession(ctx context.Context, req acpsdk.LoadSessionRequest) (acpsdk.LoadSessionResponse, error) {
	return c.agent.LoadSession(ctx, req)
}

// Cancel 取消会话
func (c *ClientT) Cancel(ctx context.Context, notification acpsdk.CancelNotification) error {
	return c.agent.Cancel(ctx, notification)
}

// Prompt 发送用户提示
func (c *ClientT) Prompt(ctx context.Context, req acpsdk.PromptRequest) (acpsdk.PromptResponse, error) {
	return c.agent.Prompt(ctx, req)
}

// SetSessionMode 设置会话模式
func (c *ClientT) SetSessionMode(ctx context.Context, req acpsdk.SetSessionModeRequest) (acpsdk.SetSessionModeResponse, error) {
	return c.agent.SetSessionMode(ctx, req)
}

// Authenticate 认证
func (c *ClientT) Authenticate(ctx context.Context, req acpsdk.AuthenticateRequest) (acpsdk.AuthenticateResponse, error) {
	return c.agent.Authenticate(ctx, req)
}

// Run 启动代理
func (c *ClientT) Run(ctx context.Context) error {
	return c.agent.Run(ctx)
}

// SetAgentConnection 设置 ACP 连接
func (c *ClientT) SetAgentConnection(conn *acpsdk.AgentSideConnection) {
	c.agent.SetAgentConnection(conn)
}

// SendSessionUpdate 发送会话更新通知
func (c *ClientT) SendSessionUpdate(ctx context.Context, sessionID string, update acpsdk.SessionUpdate) error {
	return c.agent.SendSessionUpdate(ctx, sessionID, update)
}

// SendAgentMessage 发送代理消息
func (c *ClientT) SendAgentMessage(ctx context.Context, sessionID, text string) error {
	return c.agent.SendAgentMessage(ctx, sessionID, text)
}

// GetSession 获取会话
func (c *ClientT) GetSession(sessionID string) acp.Session {
	return c.agent.GetSession(sessionID)
}

// RemoveSession 移除会话
func (c *ClientT) RemoveSession(sessionID string) {
	c.agent.RemoveSession(sessionID)
}

// StartFromStdio 从 stdio 启动代理连接
func (c *ClientT) StartFromStdio(ctx context.Context, out io.Writer, in io.Reader) (*acpsdk.AgentSideConnection, error) {
	conn := acpsdk.NewAgentSideConnection(c.agent, out, in)
	c.SetAgentConnection(conn)
	return conn, nil
}
