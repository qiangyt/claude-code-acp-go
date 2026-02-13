package acp

import (
	"context"
)

// Agent 实现 ACP Agent 接口
type Agent struct {
	// TODO: 添加字段
}

// NewAgent 创建新的 ACP 代理
func NewAgent() *Agent {
	return &Agent{}
}

// Run 启动代理，监听 stdin/stdout
func (a *Agent) Run(ctx context.Context) error {
	// TODO: 实现主事件循环
	<-ctx.Done()
	return nil
}
