// claude-code-acp Go 实现 - CLI 核心逻辑
// 此文件包含 CLI 的核心逻辑，可被测试

package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"claude-code-acp-go/internal/acp"
)

// Runner 定义 Agent 的运行接口
type Runner interface {
	Run(ctx context.Context) error
}

// CLI 提供 CLI 应用程序的核心逻辑
type CLI struct {
	Agent   Runner
	Stderr  *os.File
	Context context.Context
	Cancel  context.CancelFunc
}

// NewCLI 创建新的 CLI 实例
func NewCLI() *CLI {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	return &CLI{
		Agent:   acp.NewAgent(),
		Stderr:  os.Stderr,
		Context: ctx,
		Cancel:  cancel,
	}
}

// Run 执行 CLI 应用程序
func (c *CLI) Run() int {
	defer c.Cancel()

	if err := c.Agent.Run(c.Context); err != nil {
		fmt.Fprintf(c.Stderr, "错误: %v\n", err)
		return 1
	}
	return 0
}
