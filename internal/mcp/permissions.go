// Package mcp 实现 MCP (Model Context Protocol) 服务器
//
// 本文件实现权限检查功能
package mcp

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"sync"
)

// ============================================================
// 权限规则
// ============================================================

// PermissionRuleT 权限规则结构
type PermissionRuleT struct {
	// ToolName 工具名称
	ToolName string
	// Pattern 匹配模式（支持通配符和正则表达式）
	Pattern string
	// compiledPattern 编译后的正则表达式
	compiledPattern *regexp.Regexp
}

// PermissionRule 权限规则类型别名
type PermissionRule = *PermissionRuleT

// Match 检查参数是否匹配规则
func (r *PermissionRuleT) Match(params map[string]any) bool {
	if r.Pattern == "" {
		return true // 空规则匹配所有
	}

	// 尝试匹配常见的参数字段
	for _, key := range []string{"command", "file_path", "content"} {
		if value, ok := params[key].(string); ok {
			if r.matchPattern(value) {
				return true
			}
		}
	}

	return false
}

// matchPattern 检查字符串是否匹配模式
func (r *PermissionRuleT) matchPattern(value string) bool {
	// 简单通配符匹配
	if strings.Contains(r.Pattern, "*") {
		// 将通配符转换为正则表达式
		pattern := regexp.QuoteMeta(r.Pattern)
		pattern = strings.ReplaceAll(pattern, `\*`, ".*")
		pattern = "^" + pattern + "$"

		matched, err := regexp.MatchString(pattern, value)
		if err == nil && matched {
			return true
		}
	}

	// 尝试作为正则表达式匹配
	if r.compiledPattern == nil {
		compiled, err := regexp.Compile(r.Pattern)
		if err == nil {
			r.compiledPattern = compiled
		}
	}

	if r.compiledPattern != nil {
		return r.compiledPattern.MatchString(value)
	}

	// 精确匹配
	return value == r.Pattern
}

// ============================================================
// 默认权限检查器
// ============================================================

// DefaultPermissionCheckerT 默认权限检查器
type DefaultPermissionCheckerT struct {
	allowRules []PermissionRule
	denyRules  []PermissionRule
	mu         sync.RWMutex
}

// DefaultPermissionChecker 默认权限检查器类型别名
type DefaultPermissionChecker = *DefaultPermissionCheckerT

// NewDefaultPermissionChecker 创建默认权限检查器
func NewDefaultPermissionChecker() DefaultPermissionChecker {
	return &DefaultPermissionCheckerT{
		allowRules: make([]PermissionRule, 0),
		denyRules:  make([]PermissionRule, 0),
	}
}

// AddAllowRule 添加允许规则
func (c *DefaultPermissionCheckerT) AddAllowRule(toolName, pattern string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.allowRules = append(c.allowRules, &PermissionRuleT{
		ToolName: toolName,
		Pattern:  pattern,
	})
}

// AddDenyRule 添加拒绝规则
func (c *DefaultPermissionCheckerT) AddDenyRule(toolName, pattern string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.denyRules = append(c.denyRules, &PermissionRuleT{
		ToolName: toolName,
		Pattern:  pattern,
	})
}

// ClearRules 清除所有规则
func (c *DefaultPermissionCheckerT) ClearRules() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.allowRules = make([]PermissionRule, 0)
	c.denyRules = make([]PermissionRule, 0)
}

// Check 检查工具调用权限
func (c *DefaultPermissionCheckerT) Check(ctx context.Context, toolName string, params map[string]any) PermissionDecision {
	c.mu.RLock()
	defer c.mu.RUnlock()

	// 首先检查拒绝规则（优先级更高）
	for _, rule := range c.denyRules {
		if rule.ToolName == toolName || rule.ToolName == "*" {
			if rule.Match(params) {
				return PermissionDecisionDeny
			}
		}
	}

	// 然后检查允许规则
	for _, rule := range c.allowRules {
		if rule.ToolName == toolName || rule.ToolName == "*" {
			if rule.Match(params) {
				return PermissionDecisionAllow
			}
		}
	}

	// 没有匹配的规则，默认需要询问
	return PermissionDecisionAsk
}

// ============================================================
// 权限错误
// ============================================================

// PermissionError 权限错误
type PermissionError struct {
	ToolName string
	Message  string
}

// Error 实现 error 接口
func (e *PermissionError) Error() string {
	return fmt.Sprintf("permission denied for tool %s: %s", e.ToolName, e.Message)
}

// NewPermissionError 创建权限错误
func NewPermissionError(toolName, message string) *PermissionError {
	return &PermissionError{
		ToolName: toolName,
		Message:  message,
	}
}
