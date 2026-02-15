package comm

import (
	"path/filepath"
	"regexp"
	"sort"
	"sync"
)

// ============================================================
// handlerEntry 内部处理器条目
// ============================================================

// handlerEntry 处理器注册条目
type handlerEntry struct {
	pattern     string
	handler     GoCommandHandler
	matchMode   MatchMode
	priority    int
	description string
	regex       *regexp.Regexp // 编译后的正则（如果使用正则模式）
}

// ============================================================
// HandlerOption 注册选项
// ============================================================

// HandlerOption 处理器注册选项函数
type HandlerOption func(*handlerEntry)

// WithPriority 设置处理器优先级（数值越大优先级越高）
func WithPriority(p int) HandlerOption {
	return func(e *handlerEntry) {
		e.priority = p
	}
}

// WithMatchMode 设置匹配模式
func WithMatchMode(m MatchMode) HandlerOption {
	return func(e *handlerEntry) {
		e.matchMode = m
	}
}

// WithDescription 设置处理器描述
func WithDescription(desc string) HandlerOption {
	return func(e *handlerEntry) {
		e.description = desc
	}
}

// ============================================================
// HandlerRegistry 处理器注册器
// ============================================================

// HandlerRegistryT 处理器注册器
type HandlerRegistryT struct {
	entries []handlerEntry
	mu      sync.RWMutex
}

// HandlerRegistry 是 HandlerRegistryT 的指针别名
type HandlerRegistry = *HandlerRegistryT

// NewHandlerRegistry 创建处理器注册器
func NewHandlerRegistry() HandlerRegistry {
	return &HandlerRegistryT{
		entries: []handlerEntry{},
	}
}

// Register 注册处理器
func (me HandlerRegistry) Register(pattern string, handler GoCommandHandler, opts ...HandlerOption) HandlerRegistry {
	me.mu.Lock()
	defer me.mu.Unlock()

	entry := handlerEntry{
		pattern:   pattern,
		handler:   handler,
		matchMode: MatchExact, // 默认精确匹配
		priority:  0,          // 默认优先级
	}

	// 应用选项
	for _, opt := range opts {
		opt(&entry)
	}

	// 如果是正则模式，预编译正则表达式
	if entry.matchMode == MatchRegex {
		entry.regex = regexp.MustCompile(pattern)
	}

	me.entries = append(me.entries, entry)
	return me
}

// Match 匹配命令并返回处理器
func (me HandlerRegistry) Match(cmd string) (GoCommandHandler, bool) {
	me.mu.RLock()
	defer me.mu.RUnlock()

	// 按优先级排序（高优先级在前）
	sorted := make([]handlerEntry, len(me.entries))
	copy(sorted, me.entries)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].priority > sorted[j].priority
	})

	// 查找匹配的处理器
	for _, entry := range sorted {
		if me.matchCommand(&entry, cmd) {
			return entry.handler, true
		}
	}

	return nil, false
}

// matchCommand 检查命令是否匹配条目
func (me HandlerRegistry) matchCommand(entry *handlerEntry, cmd string) bool {
	switch entry.matchMode {
	case MatchExact:
		return entry.pattern == cmd
	case MatchGlob:
		matched, _ := filepath.Match(entry.pattern, cmd)
		return matched
	case MatchRegex:
		if entry.regex != nil {
			return entry.regex.MatchString(cmd)
		}
		return false
	}
	return false
}

// List 列出所有已注册的处理器
func (me HandlerRegistry) List() []string {
	me.mu.RLock()
	defer me.mu.RUnlock()

	result := make([]string, len(me.entries))
	for i, entry := range me.entries {
		result[i] = entry.pattern
	}
	return result
}

// Count 返回已注册处理器的数量
func (me HandlerRegistry) Count() int {
	me.mu.RLock()
	defer me.mu.RUnlock()
	return len(me.entries)
}

// Clear 清除所有已注册的处理器
func (me HandlerRegistry) Clear() {
	me.mu.Lock()
	defer me.mu.Unlock()
	me.entries = []handlerEntry{}
}

// ============================================================
// 便捷函数
// ============================================================

// MatchHandler 从注册器匹配命令处理器的便捷函数
func MatchHandler(registry HandlerRegistry, cmd string) (GoCommandHandler, bool) {
	if registry == nil {
		return nil, false
	}
	return registry.Match(cmd)
}
