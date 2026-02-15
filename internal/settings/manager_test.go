package settings

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewManager(t *testing.T) {
	t.Run("创建默认管理器", func(t *testing.T) {
		manager := NewManager()
		require.NotNil(t, manager)
	})
}

func TestManager_Load(t *testing.T) {
	t.Run("加载空目录", func(t *testing.T) {
		tmpDir := t.TempDir()
		manager := NewManager()
		settings := manager.Load(tmpDir)
		require.NotNil(t, settings)
		// 没有设置文件时使用默认值
		assert.Equal(t, "", settings.Model)
	})

	t.Run("加载用户设置", func(t *testing.T) {
		// 创建临时用户目录
		tmpHome := t.TempDir()
		claudeDir := filepath.Join(tmpHome, ".claude")
		require.NoError(t, os.MkdirAll(claudeDir, 0755))

		settingsFile := filepath.Join(claudeDir, "settings.json")
		content := `{"model": "claude-3-opus"}`
		require.NoError(t, os.WriteFile(settingsFile, []byte(content), 0644))

		manager := NewManager().WithHomeDir(tmpHome)
		settings := manager.Load(t.TempDir())

		require.NotNil(t, settings)
		assert.Equal(t, "claude-3-opus", settings.Model)
	})

	t.Run("加载项目设置", func(t *testing.T) {
		tmpDir := t.TempDir()
		claudeDir := filepath.Join(tmpDir, ".claude")
		require.NoError(t, os.MkdirAll(claudeDir, 0755))

		settingsFile := filepath.Join(claudeDir, "settings.json")
		content := `{"model": "claude-3-sonnet", "env": {"API_KEY": "test"}}`
		require.NoError(t, os.WriteFile(settingsFile, []byte(content), 0644))

		manager := NewManager()
		settings := manager.Load(tmpDir)

		require.NotNil(t, settings)
		assert.Equal(t, "claude-3-sonnet", settings.Model)
		assert.Equal(t, "test", settings.Env["API_KEY"])
	})

	t.Run("加载本地设置", func(t *testing.T) {
		tmpDir := t.TempDir()
		claudeDir := filepath.Join(tmpDir, ".claude")
		require.NoError(t, os.MkdirAll(claudeDir, 0755))

		localFile := filepath.Join(claudeDir, "settings.local.json")
		content := `{"model": "claude-3-haiku"}`
		require.NoError(t, os.WriteFile(localFile, []byte(content), 0644))

		manager := NewManager()
		settings := manager.Load(tmpDir)

		require.NotNil(t, settings)
		assert.Equal(t, "claude-3-haiku", settings.Model)
	})

	t.Run("设置合并 - 本地覆盖项目", func(t *testing.T) {
		tmpDir := t.TempDir()
		tmpHome := t.TempDir()

		// 用户设置
		userDir := filepath.Join(tmpHome, ".claude")
		require.NoError(t, os.MkdirAll(userDir, 0755))
		require.NoError(t, os.WriteFile(filepath.Join(userDir, "settings.json"),
			[]byte(`{"model": "claude-3-opus", "env": {"A": "1"}}`), 0644))

		// 项目设置
		projectDir := filepath.Join(tmpDir, ".claude")
		require.NoError(t, os.MkdirAll(projectDir, 0755))
		require.NoError(t, os.WriteFile(filepath.Join(projectDir, "settings.json"),
			[]byte(`{"model": "claude-3-sonnet", "env": {"B": "2"}}`), 0644))

		// 本地设置
		require.NoError(t, os.WriteFile(filepath.Join(projectDir, "settings.local.json"),
			[]byte(`{"env": {"C": "3"}}`), 0644))

		manager := NewManager().WithHomeDir(tmpHome)
		settings := manager.Load(tmpDir)

		require.NotNil(t, settings)
		// 本地设置没有指定 model，所以使用项目设置
		assert.Equal(t, "claude-3-sonnet", settings.Model)
		// 环境变量合并
		assert.Equal(t, "1", settings.Env["A"])
		assert.Equal(t, "2", settings.Env["B"])
		assert.Equal(t, "3", settings.Env["C"])
	})

	t.Run("无效 JSON 跳过", func(t *testing.T) {
		tmpDir := t.TempDir()
		claudeDir := filepath.Join(tmpDir, ".claude")
		require.NoError(t, os.MkdirAll(claudeDir, 0755))

		settingsFile := filepath.Join(claudeDir, "settings.json")
		require.NoError(t, os.WriteFile(settingsFile, []byte(`{invalid json}`), 0644))

		manager := NewManager()
		settings := manager.Load(tmpDir)

		// 无效 JSON 应该跳过，返回默认设置
		require.NotNil(t, settings)
	})
}

func TestSettings_Permissions(t *testing.T) {
	t.Run("获取权限设置", func(t *testing.T) {
		tmpDir := t.TempDir()
		claudeDir := filepath.Join(tmpDir, ".claude")
		require.NoError(t, os.MkdirAll(claudeDir, 0755))

		settingsFile := filepath.Join(claudeDir, "settings.json")
		content := `{
			"permissions": {
				"allow": ["Read", "Bash(ls:*)"],
				"deny": ["Bash(rm:*)"]
			}
		}`
		require.NoError(t, os.WriteFile(settingsFile, []byte(content), 0644))

		manager := NewManager()
		settings := manager.Load(tmpDir)

		require.NotNil(t, settings.Permissions)
		assert.Contains(t, settings.Permissions.Allow, "Read")
		assert.Contains(t, settings.Permissions.Allow, "Bash(ls:*)")
		assert.Contains(t, settings.Permissions.Deny, "Bash(rm:*)")
	})
}

func TestSettingsT_GetEnv(t *testing.T) {
	t.Run("获取存在的环境变量", func(t *testing.T) {
		settings := &SettingsT{
			Env: map[string]string{
				"API_KEY": "secret",
				"DEBUG":   "true",
			},
		}

		val, ok := settings.GetEnv("API_KEY")
		assert.True(t, ok)
		assert.Equal(t, "secret", val)

		val, ok = settings.GetEnv("DEBUG")
		assert.True(t, ok)
		assert.Equal(t, "true", val)
	})

	t.Run("获取不存在的环境变量", func(t *testing.T) {
		settings := &SettingsT{
			Env: map[string]string{},
		}

		val, ok := settings.GetEnv("NONEXISTENT")
		assert.False(t, ok)
		assert.Equal(t, "", val)
	})

	t.Run("获取环境变量带默认值", func(t *testing.T) {
		settings := &SettingsT{
			Env: map[string]string{
				"PORT": "8080",
			},
		}

		val := settings.GetEnvWithDefault("PORT", "3000")
		assert.Equal(t, "8080", val)

		val = settings.GetEnvWithDefault("TIMEOUT", "30")
		assert.Equal(t, "30", val)
	})
}

func TestSettingsT_Clone(t *testing.T) {
	t.Run("克隆设置", func(t *testing.T) {
		settings := &SettingsT{
			Model: "claude-3-opus",
			Env: map[string]string{
				"KEY": "value",
			},
			Permissions: &PermissionsT{
				Allow: []string{"Read"},
				Deny:  []string{"Bash(rm:*)"},
			},
		}

		cloned := settings.Clone()

		// 修改克隆不影响原对象
		cloned.Model = "claude-3-sonnet"
		cloned.Env["KEY"] = "modified"
		cloned.Permissions.Allow[0] = "Write"

		assert.Equal(t, "claude-3-opus", settings.Model)
		assert.Equal(t, "value", settings.Env["KEY"])
		assert.Equal(t, "Read", settings.Permissions.Allow[0])
	})
}

func TestSettingsT_Merge(t *testing.T) {
	t.Run("合并设置", func(t *testing.T) {
		base := &SettingsT{
			Model: "claude-3-opus",
			Env: map[string]string{
				"A": "1",
				"B": "2",
			},
		}

		override := &SettingsT{
			Model: "claude-3-sonnet",
			Env: map[string]string{
				"B": "3",
				"C": "4",
			},
		}

		result := base.Merge(override)

		assert.Equal(t, "claude-3-sonnet", result.Model)
		assert.Equal(t, "1", result.Env["A"])
		assert.Equal(t, "3", result.Env["B"])
		assert.Equal(t, "4", result.Env["C"])
	})

	t.Run("合并空设置", func(t *testing.T) {
		base := &SettingsT{
			Model: "claude-3-opus",
		}

		result := base.Merge(nil)
		assert.Equal(t, "claude-3-opus", result.Model)
	})

	t.Run("合并带元数据", func(t *testing.T) {
		base := &SettingsT{
			Model: "claude-3-opus",
			Meta: map[string]any{
				"key1": "value1",
			},
		}

		override := &SettingsT{
			Model: "claude-3-sonnet",
			Meta: map[string]any{
				"key2": "value2",
			},
		}

		result := base.Merge(override)

		assert.Equal(t, "claude-3-sonnet", result.Model)
		assert.Equal(t, "value1", result.Meta["key1"])
		assert.Equal(t, "value2", result.Meta["key2"])
	})

	t.Run("合并带权限", func(t *testing.T) {
		base := &SettingsT{
			Permissions: &PermissionsT{
				Allow: []string{"Read"},
			},
		}

		override := &SettingsT{
			Permissions: &PermissionsT{
				Allow: []string{"Write"},
				Deny:  []string{"Bash(rm:*)"},
			},
		}

		result := base.Merge(override)

		assert.Contains(t, result.Permissions.Allow, "Read")
		assert.Contains(t, result.Permissions.Allow, "Write")
		assert.Contains(t, result.Permissions.Deny, "Bash(rm:*)")
	})
}

func TestPermissionsT_Merge(t *testing.T) {
	t.Run("合并权限", func(t *testing.T) {
		base := &PermissionsT{
			Allow:                 []string{"Read"},
			Deny:                  []string{"Bash(rm:*)"},
			Ask:                   []string{"Write"},
			AdditionalDirectories: []string{"/tmp"},
		}

		other := &PermissionsT{
			Allow:                 []string{"Edit"},
			Deny:                  []string{"Bash(sudo:*)"},
			Ask:                   []string{"Delete"},
			AdditionalDirectories: []string{"/home"},
		}

		result := base.Merge(other)

		assert.Contains(t, result.Allow, "Read")
		assert.Contains(t, result.Allow, "Edit")
		assert.Contains(t, result.Deny, "Bash(rm:*)")
		assert.Contains(t, result.Deny, "Bash(sudo:*)")
		assert.Contains(t, result.Ask, "Write")
		assert.Contains(t, result.Ask, "Delete")
		assert.Contains(t, result.AdditionalDirectories, "/tmp")
		assert.Contains(t, result.AdditionalDirectories, "/home")
	})

	t.Run("合并空权限", func(t *testing.T) {
		base := &PermissionsT{
			Allow: []string{"Read"},
		}

		result := base.Merge(nil)
		assert.Equal(t, []string{"Read"}, result.Allow)
	})
}

func TestManager_ClearCache(t *testing.T) {
	t.Run("清除缓存", func(t *testing.T) {
		tmpDir := t.TempDir()
		claudeDir := filepath.Join(tmpDir, ".claude")
		require.NoError(t, os.MkdirAll(claudeDir, 0755))

		settingsFile := filepath.Join(claudeDir, "settings.json")
		require.NoError(t, os.WriteFile(settingsFile, []byte(`{"model": "test"}`), 0644))

		manager := NewManager()

		// 第一次加载
		settings1 := manager.Load(tmpDir)
		assert.Equal(t, "test", settings1.Model)

		// 修改文件
		require.NoError(t, os.WriteFile(settingsFile, []byte(`{"model": "test2"}`), 0644))

		// 未清除缓存，仍然返回旧值
		settings2 := manager.Load(tmpDir)
		assert.Equal(t, "test", settings2.Model)

		// 清除缓存
		manager.ClearCache()

		// 重新加载获取新值
		settings3 := manager.Load(tmpDir)
		assert.Equal(t, "test2", settings3.Model)
	})
}

func TestManager_Reload(t *testing.T) {
	t.Run("重新加载", func(t *testing.T) {
		tmpDir := t.TempDir()
		claudeDir := filepath.Join(tmpDir, ".claude")
		require.NoError(t, os.MkdirAll(claudeDir, 0755))

		settingsFile := filepath.Join(claudeDir, "settings.json")
		require.NoError(t, os.WriteFile(settingsFile, []byte(`{"model": "v1"}`), 0644))

		manager := NewManager()
		settings := manager.Load(tmpDir)
		assert.Equal(t, "v1", settings.Model)

		// 修改文件
		require.NoError(t, os.WriteFile(settingsFile, []byte(`{"model": "v2"}`), 0644))

		// 重新加载
		settings = manager.Reload(tmpDir)
		assert.Equal(t, "v2", settings.Model)
	})
}

func TestManager_GetSourcePaths(t *testing.T) {
	t.Run("获取来源路径", func(t *testing.T) {
		manager := NewManager().WithHomeDir("/home/user")
		paths := manager.GetSourcePaths("/project")

		require.Len(t, paths, 3)
		assert.Equal(t, "/home/user/.claude/settings.json", paths[0])
		assert.Equal(t, "/project/.claude/settings.json", paths[1])
		assert.Equal(t, "/project/.claude/settings.local.json", paths[2])
	})

	t.Run("空工作目录", func(t *testing.T) {
		manager := NewManager().WithHomeDir("/home/user")
		paths := manager.GetSourcePaths("")

		require.Len(t, paths, 1)
		assert.Equal(t, "/home/user/.claude/settings.json", paths[0])
	})
}

func TestSettingsT_Clone_WithAllFields(t *testing.T) {
	t.Run("克隆完整设置", func(t *testing.T) {
		settings := &SettingsT{
			Model: "claude-3-opus",
			Env: map[string]string{
				"KEY": "value",
			},
			Permissions: &PermissionsT{
				Allow: []string{"Read"},
			},
			Meta: map[string]any{
				"custom": "data",
			},
		}

		cloned := settings.Clone()

		// 修改克隆不影响原对象
		cloned.Model = "modified"
		cloned.Env["KEY"] = "modified"
		cloned.Permissions.Allow[0] = "modified"
		cloned.Meta["custom"] = "modified"

		assert.Equal(t, "claude-3-opus", settings.Model)
		assert.Equal(t, "value", settings.Env["KEY"])
		assert.Equal(t, "Read", settings.Permissions.Allow[0])
		assert.Equal(t, "data", settings.Meta["custom"])
	})
}

func TestSettingsT_GetEnv_NilEnv(t *testing.T) {
	t.Run("Env 为 nil", func(t *testing.T) {
		settings := &SettingsT{
			Env: nil,
		}

		val, ok := settings.GetEnv("KEY")
		assert.False(t, ok)
		assert.Equal(t, "", val)
	})
}
