// Package api 提供 claude-code-acp-go 的公开 API
package api

import (
	"claude-code-acp-go/internal/acp"
	"claude-code-acp-go/internal/mcp"
)

// OptionsT 客户端配置选项
type OptionsT struct {
	// Name Agent 名称
	Name string
	// WorkingDir 工作目录
	WorkingDir string
	// PermissionChecker 自定义权限检查器
	PermissionChecker mcp.PermissionChecker
}

// Options 选项类型别名
type Options = *OptionsT

// NewOptions 创建默认选项
func NewOptions() Options {
	return &OptionsT{
		Name:       "claude-code-acp",
		WorkingDir: ".",
	}
}

// WithName 设置 Agent 名称
func (o *OptionsT) WithName(name string) Options {
	o.Name = name
	return o
}

// WithWorkingDir 设置工作目录
func (o *OptionsT) WithWorkingDir(dir string) Options {
	o.WorkingDir = dir
	return o
}

// WithPermissionChecker 设置权限检查器
func (o *OptionsT) WithPermissionChecker(checker mcp.PermissionChecker) Options {
	o.PermissionChecker = checker
	return o
}

// toAgentOptions 转换为内部 Agent 选项
func (o Options) toAgentOptions() acp.AgentOptions {
	opts := acp.NewAgentOptions().
		WithName(o.Name).
		WithWorkingDir(o.WorkingDir)

	if o.PermissionChecker != nil {
		opts = opts.WithPermissionChecker(o.PermissionChecker)
	}

	return opts
}
