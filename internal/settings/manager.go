// Package settings 实现 Claude Code 设置管理
//
// 本包管理多个来源的设置（用户、项目、本地、企业托管），
// 并提供权限规则解析和检查功能。
package settings

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
)

// ============================================================
// 权限设置
// ============================================================

// PermissionsT 权限设置结构
type PermissionsT struct {
	// Allow 允许的规则列表
	Allow []string `json:"allow,omitempty"`
	// Deny 拒绝的规则列表
	Deny []string `json:"deny,omitempty"`
	// Ask 需要询问的规则列表
	Ask []string `json:"ask,omitempty"`
	// AdditionalDirectories 额外允许的目录
	AdditionalDirectories []string `json:"additionalDirectories,omitempty"`
}

// Permissions 权限设置类型别名
type Permissions = *PermissionsT

// Clone 克隆权限设置
func (p *PermissionsT) Clone() Permissions {
	return &PermissionsT{
		Allow:                 append([]string{}, p.Allow...),
		Deny:                  append([]string{}, p.Deny...),
		Ask:                   append([]string{}, p.Ask...),
		AdditionalDirectories: append([]string{}, p.AdditionalDirectories...),
	}
}

// Merge 合并权限设置
func (p *PermissionsT) Merge(other Permissions) Permissions {
	if other == nil {
		return p.Clone()
	}

	result := p.Clone()
	result.Allow = append(result.Allow, other.Allow...)
	result.Deny = append(result.Deny, other.Deny...)
	result.Ask = append(result.Ask, other.Ask...)
	result.AdditionalDirectories = append(result.AdditionalDirectories, other.AdditionalDirectories...)
	return result
}

// ============================================================
// 设置结构
// ============================================================

// SettingsT 设置结构
type SettingsT struct {
	// Model 模型选择
	Model string `json:"model,omitempty"`
	// Env 环境变量
	Env map[string]string `json:"env,omitempty"`
	// Permissions 权限配置
	Permissions Permissions `json:"permissions,omitempty"`
	// Meta 元数据
	Meta map[string]any `json:"_meta,omitempty"`
}

// Settings 设置类型别名
type Settings = *SettingsT

// NewSettings 创建默认设置
func NewSettings() Settings {
	return &SettingsT{
		Env: make(map[string]string),
	}
}

// Clone 克隆设置
func (s *SettingsT) Clone() Settings {
	env := make(map[string]string)
	for k, v := range s.Env {
		env[k] = v
	}

	var permissions Permissions
	if s.Permissions != nil {
		permissions = s.Permissions.Clone()
	}

	meta := make(map[string]any)
	for k, v := range s.Meta {
		meta[k] = v
	}

	return &SettingsT{
		Model:       s.Model,
		Env:         env,
		Permissions: permissions,
		Meta:        meta,
	}
}

// Merge 合并设置（other 覆盖 s）
func (s *SettingsT) Merge(other Settings) Settings {
	if other == nil {
		return s.Clone()
	}

	result := s.Clone()

	// 覆盖非空值
	if other.Model != "" {
		result.Model = other.Model
	}

	// 合并环境变量
	for k, v := range other.Env {
		result.Env[k] = v
	}

	// 合并权限
	if other.Permissions != nil {
		if result.Permissions == nil {
			result.Permissions = &PermissionsT{}
		}
		result.Permissions = result.Permissions.Merge(other.Permissions)
	}

	// 合并元数据
	for k, v := range other.Meta {
		result.Meta[k] = v
	}

	return result
}

// GetEnv 获取环境变量
func (s *SettingsT) GetEnv(key string) (string, bool) {
	if s.Env == nil {
		return "", false
	}
	val, ok := s.Env[key]
	return val, ok
}

// GetEnvWithDefault 获取环境变量，带默认值
func (s *SettingsT) GetEnvWithDefault(key, defaultValue string) string {
	if val, ok := s.GetEnv(key); ok {
		return val
	}
	return defaultValue
}

// ============================================================
// 设置管理器
// ============================================================

// ManagerT 设置管理器结构
type ManagerT struct {
	// homeDir 用户主目录
	homeDir string
	// cache 缓存的设置
	cache map[string]Settings
	// mu 读写锁
	mu sync.RWMutex
}

// Manager 设置管理器类型别名
type Manager = *ManagerT

// NewManager 创建设置管理器
func NewManager() Manager {
	homeDir, _ := os.UserHomeDir()
	return &ManagerT{
		homeDir: homeDir,
		cache:   make(map[string]Settings),
	}
}

// WithHomeDir 设置用户主目录
func (m *ManagerT) WithHomeDir(dir string) Manager {
	return &ManagerT{
		homeDir: dir,
		cache:   make(map[string]Settings),
	}
}

// Load 加载设置
func (m *ManagerT) Load(cwd string) Settings {
	m.mu.RLock()
	if cached, ok := m.cache[cwd]; ok {
		m.mu.RUnlock()
		return cached
	}
	m.mu.RUnlock()

	// 加载各来源设置并合并
	result := NewSettings()

	// 1. 用户设置（最低优先级）
	if m.homeDir != "" {
		userSettings := m.loadFromFile(filepath.Join(m.homeDir, ".claude", "settings.json"))
		if userSettings != nil {
			result = result.Merge(userSettings)
		}
	}

	// 2. 项目设置
	if cwd != "" {
		projectSettings := m.loadFromFile(filepath.Join(cwd, ".claude", "settings.json"))
		if projectSettings != nil {
			result = result.Merge(projectSettings)
		}

		// 3. 本地项目设置（最高优先级）
		localSettings := m.loadFromFile(filepath.Join(cwd, ".claude", "settings.local.json"))
		if localSettings != nil {
			result = result.Merge(localSettings)
		}
	}

	// 缓存结果
	m.mu.Lock()
	m.cache[cwd] = result
	m.mu.Unlock()

	return result
}

// loadFromFile 从文件加载设置
func (m *ManagerT) loadFromFile(path string) Settings {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil // 文件不存在，跳过
	}

	var settings SettingsT
	if err := json.Unmarshal(data, &settings); err != nil {
		// JSON 解析失败，跳过
		return nil
	}

	// 确保 Env 不为 nil
	if settings.Env == nil {
		settings.Env = make(map[string]string)
	}

	return &settings
}

// ClearCache 清除缓存
func (m *ManagerT) ClearCache() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.cache = make(map[string]Settings)
}

// Reload 重新加载设置
func (m *ManagerT) Reload(cwd string) Settings {
	m.ClearCache()
	return m.Load(cwd)
}

// GetSourcePaths 获取所有设置来源路径
func (m *ManagerT) GetSourcePaths(cwd string) []string {
	paths := make([]string, 0)

	if m.homeDir != "" {
		paths = append(paths, filepath.Join(m.homeDir, ".claude", "settings.json"))
	}

	if cwd != "" {
		paths = append(paths,
			filepath.Join(cwd, ".claude", "settings.json"),
			filepath.Join(cwd, ".claude", "settings.local.json"),
		)
	}

	return paths
}
