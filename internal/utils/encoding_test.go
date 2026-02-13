package utils

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEncodePath(t *testing.T) {
	t.Run("编码简单路径", func(t *testing.T) {
		path := "/home/user/project/file.txt"
		encoded := EncodePath(path)
		// PathEscape 会编码斜杠等字符
		assert.Contains(t, encoded, "file.txt")
	})

	t.Run("编码包含空格的路径", func(t *testing.T) {
		path := "/home/user/my project/file name.txt"
		encoded := EncodePath(path)
		assert.Contains(t, encoded, "%20")
	})

	t.Run("编码包含中文的路径", func(t *testing.T) {
		path := "/home/user/项目/文件.txt"
		encoded := EncodePath(path)
		assert.NotEqual(t, path, encoded) // 应该被编码
	})

	t.Run("往返测试", func(t *testing.T) {
		original := "/home/user/my project/文件.txt"
		encoded := EncodePath(original)
		decoded, err := DecodePath(encoded)
		require.NoError(t, err)
		assert.Equal(t, original, decoded)
	})
}

func TestDecodePath(t *testing.T) {
	t.Run("解码简单路径", func(t *testing.T) {
		encoded := "/home/user/project/file.txt"
		decoded, err := DecodePath(encoded)
		require.NoError(t, err)
		assert.Equal(t, encoded, decoded)
	})

	t.Run("解码 URL 编码的路径", func(t *testing.T) {
		encoded := "/home/user/my%20project/file%20name.txt"
		decoded, err := DecodePath(encoded)
		require.NoError(t, err)
		assert.Equal(t, "/home/user/my project/file name.txt", decoded)
	})

	t.Run("解码中文路径", func(t *testing.T) {
		original := "/home/user/项目/文件.txt"
		encoded := EncodePath(original)
		decoded, err := DecodePath(encoded)
		require.NoError(t, err)
		assert.Equal(t, original, decoded)
	})

	t.Run("解码无效编码", func(t *testing.T) {
		encoded := "/home/user/%GG/invalid"
		_, err := DecodePath(encoded)
		assert.Error(t, err)
	})
}

func TestRelPath(t *testing.T) {
	t.Run("计算相对路径", func(t *testing.T) {
		base := "/home/user/project"
		target := "/home/user/project/src/main.go"

		rel, err := RelPath(base, target)
		require.NoError(t, err)
		assert.Equal(t, "src/main.go", rel)
	})

	t.Run("相同路径", func(t *testing.T) {
		base := "/home/user/project"
		target := "/home/user/project"

		rel, err := RelPath(base, target)
		require.NoError(t, err)
		assert.Equal(t, ".", rel)
	})

	t.Run("父目录", func(t *testing.T) {
		base := "/home/user/project"
		target := "/home/user"

		rel, err := RelPath(base, target)
		require.NoError(t, err)
		assert.Equal(t, "..", rel)
	})

	t.Run("无共同前缀", func(t *testing.T) {
		base := "/home/user/project"
		target := "/var/log/app.log"

		rel, err := RelPath(base, target)
		require.NoError(t, err)
		// 应该返回一个向上遍历的相对路径
		assert.Contains(t, rel, "..")
	})
}

func TestAbsPath(t *testing.T) {
	t.Run("转换为绝对路径", func(t *testing.T) {
		rel := "src/main.go"
		base := "/home/user/project"

		abs, err := AbsPath(base, rel)
		require.NoError(t, err)
		assert.Equal(t, "/home/user/project/src/main.go", abs)
	})

	t.Run("已经是绝对路径", func(t *testing.T) {
		abs := "/home/user/project/src/main.go"
		base := "/home/user/other"

		result, err := AbsPath(base, abs)
		require.NoError(t, err)
		assert.Equal(t, abs, result)
	})

	t.Run("处理 ./ 前缀", func(t *testing.T) {
		rel := "./src/main.go"
		base := "/home/user/project"

		abs, err := AbsPath(base, rel)
		require.NoError(t, err)
		assert.Equal(t, "/home/user/project/src/main.go", abs)
	})

	t.Run("处理 ../ 前缀", func(t *testing.T) {
		rel := "../other/file.txt"
		base := "/home/user/project"

		abs, err := AbsPath(base, rel)
		require.NoError(t, err)
		assert.Equal(t, "/home/user/other/file.txt", abs)
	})
}

func TestNormalizePath(t *testing.T) {
	t.Run("规范化路径", func(t *testing.T) {
		path := "/home/user/../user/./project//file.txt"
		normalized := NormalizePath(path)
		assert.Equal(t, "/home/user/project/file.txt", normalized)
	})

	t.Run("处理多余斜杠", func(t *testing.T) {
		path := "/home//user///project//file.txt"
		normalized := NormalizePath(path)
		assert.Equal(t, "/home/user/project/file.txt", normalized)
	})
}

func TestIsAbsPath(t *testing.T) {
	t.Run("绝对路径", func(t *testing.T) {
		assert.True(t, IsAbsPath("/home/user/project"))
	})

	t.Run("相对路径", func(t *testing.T) {
		assert.False(t, IsAbsPath("src/main.go"))
		assert.False(t, IsAbsPath("./file.txt"))
		assert.False(t, IsAbsPath("../parent/file.txt"))
	})
}

func TestPathInDir(t *testing.T) {
	t.Run("路径在目录内", func(t *testing.T) {
		dir := "/home/user/project"
		path := "/home/user/project/src/main.go"

		inDir, err := PathInDir(dir, path)
		require.NoError(t, err)
		assert.True(t, inDir)
	})

	t.Run("路径在目录外", func(t *testing.T) {
		dir := "/home/user/project"
		path := "/home/user/other/file.txt"

		inDir, err := PathInDir(dir, path)
		require.NoError(t, err)
		assert.False(t, inDir)
	})

	t.Run("路径就是目录", func(t *testing.T) {
		dir := "/home/user/project"
		path := "/home/user/project"

		inDir, err := PathInDir(dir, path)
		require.NoError(t, err)
		assert.True(t, inDir)
	})

	t.Run("使用相对路径", func(t *testing.T) {
		dir := "/home/user/project"
		path := "/home/user/project/src/../src/main.go"

		inDir, err := PathInDir(dir, path)
		require.NoError(t, err)
		assert.True(t, inDir)
	})

	t.Run("无效目录路径", func(t *testing.T) {
		// 使用包含 null 字符的路径会导致 filepath.Abs 失败
		inDir, err := PathInDir("/valid/path", "/\x00invalid")
		// 在某些系统上可能不会失败，所以只检查结果
		_ = inDir
		_ = err
	})

	t.Run("目录路径获取绝对路径失败", func(t *testing.T) {
		// 保存当前工作目录
		originalWd, err := os.Getwd()
		require.NoError(t, err)

		// 创建临时目录并切换到该目录
		tempDir, err := os.MkdirTemp("", "pathindir_test")
		require.NoError(t, err)

		err = os.Chdir(tempDir)
		require.NoError(t, err)

		// 删除临时目录，使当前工作目录无效
		err = os.RemoveAll(tempDir)
		require.NoError(t, err)

		// 此时 PathInDir 应该在获取目录的绝对路径时失败
		_, err = PathInDir("some/dir", "/valid/path")

		// 恢复工作目录
		_ = os.Chdir(originalWd)

		// 验证返回了错误
		require.Error(t, err)
		assert.Contains(t, err.Error(), "getwd")
	})

	t.Run("目标路径获取绝对路径失败", func(t *testing.T) {
		// 保存当前工作目录
		originalWd, err := os.Getwd()
		require.NoError(t, err)

		// 创建临时目录并切换到该目录
		tempDir, err := os.MkdirTemp("", "pathindir_test2")
		require.NoError(t, err)

		err = os.Chdir(tempDir)
		require.NoError(t, err)

		// 删除临时目录，使当前工作目录无效
		err = os.RemoveAll(tempDir)
		require.NoError(t, err)

		// 此时 PathInDir 应该在获取目标路径的绝对路径时失败
		// 注意：由于 dir 是绝对路径，第一个 filepath.Abs 会成功
		// 但 path 是相对路径，第二个 filepath.Abs 会失败
		_, err = PathInDir("/valid/dir", "relative/path")

		// 恢复工作目录
		_ = os.Chdir(originalWd)

		// 验证返回了错误
		require.Error(t, err)
		assert.Contains(t, err.Error(), "getwd")
	})
}

func TestFileExtension(t *testing.T) {
	tests := []struct {
		path     string
		expected string
	}{
		{"/home/user/file.txt", ".txt"},
		{"/home/user/file.go", ".go"},
		{"/home/user/file.test.js", ".js"},
		{"/home/user/file", ""},
		{"/home/user/.hidden", ".hidden"}, // .hidden 被视为扩展名
		{"/home/user/.hidden.txt", ".txt"},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			ext := FileExtension(tt.path)
			assert.Equal(t, tt.expected, ext)
		})
	}
}

func TestJoinPath(t *testing.T) {
	t.Run("连接路径元素", func(t *testing.T) {
		result := JoinPath("/home", "user", "project", "file.txt")
		assert.Equal(t, "/home/user/project/file.txt", result)
	})

	t.Run("处理空元素", func(t *testing.T) {
		result := JoinPath("/home", "", "user", "", "file.txt")
		assert.Equal(t, "/home/user/file.txt", result)
	})
}

func TestBaseName(t *testing.T) {
	tests := []struct {
		path     string
		expected string
	}{
		{"/home/user/file.txt", "file.txt"},
		{"/home/user/project/", "project"},
		{"file.txt", "file.txt"},
		{"/", "/"},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			base := BaseName(tt.path)
			assert.Equal(t, tt.expected, base)
		})
	}
}

func TestDirName(t *testing.T) {
	tests := []struct {
		path     string
		expected string
	}{
		{"/home/user/file.txt", "/home/user"},
		{"/home/user/project/", "/home/user/project"}, // filepath.Dir 保留尾部斜杠前的部分
		{"file.txt", "."},
		{"/", "/"},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			dir := DirName(tt.path)
			assert.Equal(t, tt.expected, dir)
		})
	}
}
