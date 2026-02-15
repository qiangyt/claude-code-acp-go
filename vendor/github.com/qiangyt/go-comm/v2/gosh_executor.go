package comm

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/pkg/errors"
	"mvdan.cc/sh/v3/expand"
	"mvdan.cc/sh/v3/interp"
	"mvdan.cc/sh/v3/syntax"
)

// ============================================================
// 架构设计
// ============================================================
//
// GoshExecutor 是可配置的 Shell 命令执行器，采用混合策略：
//
//	┌─────────────────────────────────────────────────────────────┐
//	│                    GoshExecutor                             │
//	├─────────────────────────────────────────────────────────────┤
//	│  GoshConfig                                                 │
//	│  ├── KillTimeout      time.Duration   // 超时时间           │
//	│  ├── Blacklist        []CommandRule   // 黑名单规则         │
//	│  ├── WhitelistMode    bool            // 白名单模式开关     │
//	│  ├── Whitelist        []CommandRule   // 白名单规则         │
//	│  └── GoHandlers       map[string]GoCommandHandler           │
//	├─────────────────────────────────────────────────────────────┤
//	│  执行流程:                                                   │
//	│  1. 解析命令 (mvdan/sh syntax)                              │
//	│  2. 检查黑名单 → 匹配则拒绝                                  │
//	│  3. 检查白名单模式 → 仅允许白名单命令                         │
//	│  4. 检查 GoHandlers → 用 Go 实现执行                         │
//	│  5. 其他命令 → exec 真实执行                                 │
//	└─────────────────────────────────────────────────────────────┘
//
// CommandRule 支持通配符和参数匹配：
//   - Pattern: "rm*" 匹配 rm, rmdir, rmmod 等
//   - ArgsFilter: ["-rf", "-r*"] 匹配 rm -rf, rm -r /path 等
//
// GoHandlers 的 key 也支持通配符：
//   - "curl" 仅匹配 curl
//   - "curl*" 匹配 curl, curlconfig 等
//
// 使用示例
//
// 简单黑名单
// config := comm.DefaultGoshConfig().
//     WithBlacklistSimple("rm", "dd")
//
// 参数级别拦截（只阻止 rm -rf）
// config := comm.DefaultGoshConfig().
//     WithBlacklist(comm.NewCommandRule("rm", "-rf", "-r"))
//
// Go 处理器替代 curl
// config := comm.DefaultGoshConfig().
//     WithGoHandler("curl", myCurlHandler)
//
// /白名单模式
// config := comm.DefaultGoshConfig().
//     WithWhitelistMode(true).
//     WithWhitelistSimple("git", "npm", "go", "echo")

// ============================================================
// GoCommandHandler
// ============================================================

// GoCommandHandler Go 实现的命令处理器
type GoCommandHandler func(ctx context.Context, hc interp.HandlerContext, args []string) error

// ============================================================
// GoshConfig
// ============================================================

// GoshConfigT 命令执行器配置
type GoshConfigT struct {
	// KillTimeout 命令超时时间
	KillTimeout time.Duration

	// Blacklist 黑名单规则（这些命令会被拒绝）
	// 支持通配符和参数匹配
	Blacklist []CommandRule

	// WhitelistMode 是否启用白名单模式
	// 启用后，只有 Whitelist 中的命令才能执行
	WhitelistMode bool

	// Whitelist 白名单规则（仅白名单模式生效）
	Whitelist []CommandRule

	// GoHandlers Go 实现的命令处理器
	// key 支持通配符，如 "curl*" 匹配 curl 和 curlconfig
	GoHandlers map[string]GoCommandHandler
}

type GoshConfig = *GoshConfigT

// DefaultGoshConfig 返回默认配置
func DefaultGoshConfig() GoshConfig {
	return &GoshConfigT{
		KillTimeout:   6 * time.Second,
		Blacklist:     []CommandRule{},
		WhitelistMode: false,
		Whitelist:     []CommandRule{},
		GoHandlers:    make(map[string]GoCommandHandler),
	}
}

// WithKillTimeout 设置超时
func (me GoshConfig) WithKillTimeout(d time.Duration) GoshConfig {
	me.KillTimeout = d
	return me
}

// WithBlacklist 设置黑名单规则
func (me GoshConfig) WithBlacklist(rules ...CommandRule) GoshConfig {
	me.Blacklist = append(me.Blacklist, rules...)
	return me
}

// WithBlacklistSimple 设置简单黑名单（仅命令名，精确匹配）
func (me GoshConfig) WithBlacklistSimple(cmds ...string) GoshConfig {
	for _, cmd := range cmds {
		me.Blacklist = append(me.Blacklist, NewCommandRule(cmd, MatchExact))
	}
	return me
}

// WithWhitelistMode 启用/禁用白名单模式
func (me GoshConfig) WithWhitelistMode(enabled bool) GoshConfig {
	me.WhitelistMode = enabled
	return me
}

// WithWhitelist 设置白名单规则
func (me GoshConfig) WithWhitelist(rules ...CommandRule) GoshConfig {
	me.Whitelist = append(me.Whitelist, rules...)
	return me
}

// WithWhitelistSimple 设置简单白名单（仅命令名，精确匹配）
func (me GoshConfig) WithWhitelistSimple(cmds ...string) GoshConfig {
	for _, cmd := range cmds {
		me.Whitelist = append(me.Whitelist, NewCommandRule(cmd, MatchExact))
	}
	return me
}

// WithGoHandler 注册 Go 命令处理器
// pattern 支持通配符，如 "curl*" 匹配 curl 和 curlconfig
func (me GoshConfig) WithGoHandler(pattern string, handler GoCommandHandler) GoshConfig {
	me.GoHandlers[pattern] = handler
	return me
}

// ============================================================
// GoshExecutor
// ============================================================

// GoshExecutorT 可配置的 Shell 命令执行器
type GoshExecutorT struct {
	config GoshConfig
}

type GoshExecutor = *GoshExecutorT

// NewGoshExecutor 创建执行器
func NewGoshExecutor(config GoshConfig) GoshExecutor {
	if config == nil {
		config = DefaultGoshConfig()
	}
	return &GoshExecutorT{config: config}
}

// Run 执行命令
func (me GoshExecutor) Run(ctx context.Context, dir string, cmd string,
	stdin io.Reader, stdout, stderr io.Writer) error {
	return me.RunWithVars(ctx, nil, dir, cmd, stdin, stdout, stderr)
}

// RunWithVars 执行命令（带变量）
func (me GoshExecutor) RunWithVars(ctx context.Context, vars map[string]string,
	dir string, cmd string, stdin io.Reader, stdout, stderr io.Writer) error {

	// 解析命令
	sf, err := syntax.NewParser().Parse(strings.NewReader(cmd), "")
	if err != nil {
		return errors.Wrapf(err, "parse command: %s", cmd)
	}

	// 构建环境变量
	environ := os.Environ()
	if vars != nil {
		envList, err := EnvironList(vars)
		if err != nil {
			return err
		}
		environ = append(environ, envList...)
	}

	// 创建执行器选项
	opts := []interp.RunnerOption{
		interp.Params("-e"),
		interp.Env(expand.ListEnviron(environ...)),
		interp.ExecHandler(me.createExecHandler()),
		interp.StdIO(stdin, stdout, stderr),
	}
	if dir != "" {
		opts = append(opts, interp.Dir(dir))
	}

	// 创建并运行
	runner, err := interp.New(opts...)
	if err != nil {
		return errors.Wrapf(err, "create runner for command: %s", cmd)
	}

	if err = runner.Run(ctx, sf); err != nil {
		return errors.Wrapf(err, "run command: %s", cmd)
	}

	return nil
}

// createExecHandler 创建执行处理器
func (me GoshExecutor) createExecHandler() interp.ExecHandlerFunc {
	return func(ctx context.Context, args []string) error {
		cmd := args[0]
		cmdArgs := args[1:]

		// 1. 检查黑名单
		if matched, rule := me.matchesRules(cmd, cmdArgs, me.config.Blacklist); matched {
			return fmt.Errorf("command '%s' is blocked by blacklist rule: %s", cmd, rule.Pattern)
		}

		// 2. 检查白名单模式
		if me.config.WhitelistMode {
			if matched, _ := me.matchesRules(cmd, cmdArgs, me.config.Whitelist); !matched {
				return fmt.Errorf("command '%s' is not in whitelist", cmd)
			}
		}

		// 3. 检查 Go 处理器
		for pattern, handler := range me.config.GoHandlers {
			if me.matchPattern(pattern, cmd) {
				return handler(ctx, interp.HandlerCtx(ctx), cmdArgs)
			}
		}

		// 4. 默认 exec 执行
		return interp.DefaultExecHandler(me.config.KillTimeout)(ctx, args)
	}
}

// matchesRules 检查命令是否匹配规则列表
func (me GoshExecutor) matchesRules(cmd string, args []string, rules []CommandRule) (bool, CommandRule) {
	checker := NewSecurityChecker()

	for _, rule := range rules {
		// 构造 ExtractedCommand 用于检查
		extractedCmd := &ExtractedCommandT{
			Name:   cmd,
			Args:   args,
			Source: SourceDirect,
		}

		if checker.matchesRule(extractedCmd, rule) {
			return true, rule
		}
	}
	return false, nil
}

// matchPattern 通配符匹配（支持 * 和 ?）
func (me GoshExecutor) matchPattern(pattern, s string) bool {
	// 简单实现：使用 filepath.Match
	matched, _ := filepath.Match(pattern, s)
	return matched
}

