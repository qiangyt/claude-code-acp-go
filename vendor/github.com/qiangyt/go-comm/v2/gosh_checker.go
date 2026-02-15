package comm

import (
	"fmt"
	"path/filepath"
	"regexp"

	"github.com/pkg/errors"
)

// ============================================================
// MatchMode 匹配模式
// ============================================================

// MatchMode 表示匹配模式
type MatchMode int

const (
	// MatchGlob 通配符匹配 (使用 filepath.Match)
	MatchGlob MatchMode = iota
	// MatchExact 精确匹配
	MatchExact
	// MatchRegex 正则表达式匹配
	MatchRegex
)

// ============================================================
// ArgMatcher 参数匹配器
// ============================================================

// ArgMatcher 参数匹配器
type ArgMatcher struct {
	// Position 参数位置 (-1 表示任意位置)
	Position int
	// Pattern 匹配模式
	Pattern string
	// Mode 匹配类型
	Mode MatchMode
	// Required 是否必需 (用于组合匹配)
	Required bool
}

// ============================================================
// CommandRule 命令规则
// ============================================================

// CommandRuleT 命令规则
type CommandRuleT struct {
	// Pattern 命令模式
	Pattern string
	// ArgsMatchers 参数匹配器列表
	ArgsMatchers []ArgMatcher
	// MatchMode 匹配模式
	MatchMode MatchMode
	// SourceFilter 命令来源过滤 (空表示所有来源)
	SourceFilter []CommandSource
}

// CommandRule 是 CommandRuleT 的指针别名
type CommandRule = *CommandRuleT

// NewCommandRule 创建命令规则
func NewCommandRule(pattern string, matchMode MatchMode) CommandRule {
	return &CommandRuleT{
		Pattern:     pattern,
		MatchMode:   matchMode,
		ArgsMatchers: []ArgMatcher{},
		SourceFilter: []CommandSource{},
	}
}

// WithArgsFilter 添加参数过滤器
func (me CommandRule) WithArgsFilter(matchers ...ArgMatcher) CommandRule {
	me.ArgsMatchers = append(me.ArgsMatchers, matchers...)
	return me
}

// WithSourceFilter 设置命令来源过滤
func (me CommandRule) WithSourceFilter(sources ...CommandSource) CommandRule {
	me.SourceFilter = append(me.SourceFilter, sources...)
	return me
}

// ============================================================
// SecurityCheckError 安全检查错误
// ============================================================

// SecurityCheckErrorT 安全检查错误
type SecurityCheckErrorT struct {
	// Command 触发错误的命令
	Command ExtractedCommand
	// Rule 触发错误的规则 (可能为 nil)
	Rule CommandRule
	// Message 错误消息
	Message string
	// ViolationType 违规类型
	ViolationType string
}

// SecurityCheckError 是 SecurityCheckErrorT 的指针别名
type SecurityCheckError = *SecurityCheckErrorT

// Error 实现 error 接口
func (e SecurityCheckError) Error() string {
	if e.Rule != nil {
		return fmt.Sprintf("command '%s' %s (rule: %s)", e.Command.Name, e.Message, e.Rule.Pattern)
	}
	return fmt.Sprintf("command '%s' %s", e.Command.Name, e.Message)
}

// ============================================================
// SecurityChecker 安全检查器
// ============================================================

// SecurityCheckerT 安全检查器
type SecurityCheckerT struct {
	// blacklist 黑名单规则
	blacklist []CommandRule
	// whitelist 白名单规则
	whitelist []CommandRule
	// whitelistMode 是否启用白名单模式
	whitelistMode bool
}

// SecurityChecker 是 SecurityCheckerT 的指针别名
type SecurityChecker = *SecurityCheckerT

// NewSecurityChecker 创建安全检查器
func NewSecurityChecker() SecurityChecker {
	return &SecurityCheckerT{
		blacklist:     []CommandRule{},
		whitelist:     []CommandRule{},
		whitelistMode: false,
	}
}

// WithBlacklist 添加黑名单规则
func (me SecurityChecker) WithBlacklist(rules ...CommandRule) SecurityChecker {
	me.blacklist = append(me.blacklist, rules...)
	return me
}

// WithWhitelist 添加白名单规则
func (me SecurityChecker) WithWhitelist(rules ...CommandRule) SecurityChecker {
	me.whitelist = append(me.whitelist, rules...)
	return me
}

// WithWhitelistMode 设置白名单模式
func (me SecurityChecker) WithWhitelistMode(enabled bool) SecurityChecker {
	me.whitelistMode = enabled
	return me
}

// Check 检查命令列表是否通过安全检查
func (me SecurityChecker) Check(cmds []ExtractedCommand) error {
	for _, cmd := range cmds {
		// 1. 检查黑名单
		if rule, matched := me.matchesBlacklist(cmd); matched {
			return &SecurityCheckErrorT{
				Command:       cmd,
				Rule:          rule,
				Message:       "is blocked by blacklist",
				ViolationType: "blacklist",
			}
		}

		// 2. 检查白名单模式
		if me.whitelistMode {
			if _, matched := me.matchesWhitelist(cmd); !matched {
				return &SecurityCheckErrorT{
					Command:       cmd,
					Rule:          nil,
					Message:       "is not in whitelist",
					ViolationType: "whitelist",
				}
			}
		}
	}
	return nil
}

// matchesBlacklist 检查命令是否匹配黑名单
func (me SecurityChecker) matchesBlacklist(cmd ExtractedCommand) (CommandRule, bool) {
	for _, rule := range me.blacklist {
		if me.matchesRule(cmd, rule) {
			return rule, true
		}
	}
	return nil, false
}

// matchesWhitelist 检查命令是否匹配白名单
func (me SecurityChecker) matchesWhitelist(cmd ExtractedCommand) (CommandRule, bool) {
	for _, rule := range me.whitelist {
		if me.matchesRule(cmd, rule) {
			return rule, true
		}
	}
	return nil, false
}

// matchesRule 检查命令是否匹配规则
func (me SecurityChecker) matchesRule(cmd ExtractedCommand, rule CommandRule) bool {
	// 1. 命令名匹配
	if !me.matchPattern(rule.MatchMode, rule.Pattern, cmd.Name) {
		return false
	}

	// 2. 来源过滤
	if len(rule.SourceFilter) > 0 {
		found := false
		for _, s := range rule.SourceFilter {
			if s == cmd.Source {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}

	// 3. 参数匹配
	if len(rule.ArgsMatchers) > 0 {
		if !me.matchArgs(cmd.Args, rule.ArgsMatchers) {
			return false
		}
	}

	return true
}

// matchPattern 模式匹配
func (me SecurityChecker) matchPattern(mode MatchMode, pattern, s string) bool {
	switch mode {
	case MatchExact:
		return pattern == s
	case MatchGlob:
		matched, _ := filepath.Match(pattern, s)
		return matched
	case MatchRegex:
		re, err := regexp.Compile(pattern)
		if err != nil {
			return false
		}
		return re.MatchString(s)
	}
	return false
}

// matchArgs 参数匹配
func (me SecurityChecker) matchArgs(args []string, matchers []ArgMatcher) bool {
	// 如果没有任何 Required 匹配器，只要任意一个匹配即可
	hasRequired := false
	for _, m := range matchers {
		if m.Required {
			hasRequired = true
			break
		}
	}

	if hasRequired {
		// 所有 Required 匹配器都必须匹配
		for _, m := range matchers {
			if m.Required {
				if !me.matchArg(args, m) {
					return false
				}
			}
		}
		return true
	}

	// 没有必需匹配器，任意一个匹配即可（或没有匹配器时返回 true）
	if len(matchers) == 0 {
		return true
	}

	for _, m := range matchers {
		if me.matchArg(args, m) {
			return true
		}
	}
	return false
}

// matchArg 检查单个参数匹配器
func (me SecurityChecker) matchArg(args []string, matcher ArgMatcher) bool {
	if matcher.Position == -1 {
		// 任意位置匹配
		for _, arg := range args {
			if me.matchPattern(matcher.Mode, matcher.Pattern, arg) {
				return true
			}
		}
		return false
	}

	// 指定位置匹配
	if matcher.Position >= 0 && matcher.Position < len(args) {
		return me.matchPattern(matcher.Mode, matcher.Pattern, args[matcher.Position])
	}
	return false
}

// ============================================================
// 便捷函数
// ============================================================

// CheckCommands 检查命令是否通过安全检查的便捷函数
func CheckCommands(cmds []ExtractedCommand, checker SecurityChecker) error {
	if checker == nil {
		return nil
	}
	return checker.Check(cmds)
}

// IsSecurityCheckError 检查错误是否为安全检查错误
func IsSecurityCheckError(err error) bool {
	_, ok := err.(SecurityCheckError)
	return ok
}

// GetSecurityCheckError 获取安全检查错误详情
func GetSecurityCheckError(err error) SecurityCheckError {
	if checkErr, ok := err.(SecurityCheckError); ok {
		return checkErr
	}
	if wrapped := errors.Cause(err); wrapped != nil {
		return GetSecurityCheckError(wrapped)
	}
	return nil
}
