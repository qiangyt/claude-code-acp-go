package golden

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestMain(m *testing.M) {
	// 设置测试目录
	SetDir(filepath.Join("..", "..", "golden", "testdata"))
	os.Exit(m.Run())
}

func TestGolden_ReadWrite(t *testing.T) {
	t.Run("读取和写入黄金文件", func(t *testing.T) {
		name := "test-read-write.txt"
		content := []byte("test content\n")

		// 写入
		Write(t, name, content)

		// 读取
		data := Read(t, name)
		require.Equal(t, content, data)

		// 清理
		path := filepath.Join(Dir, name)
		os.Remove(path)
	})
}

func TestGolden_Compare(t *testing.T) {
	t.Run("比较相同内容", func(t *testing.T) {
		name := "test-compare-same.txt"
		content := []byte("same content")

		// 先写入黄金文件
		Write(t, name, content)

		// 比较 - 应该成功
		Compare(t, name, content)

		// 清理
		path := filepath.Join(Dir, name)
		os.Remove(path)
	})

	t.Run("比较 JSON 内容", func(t *testing.T) {
		name := "test-compare-json.json"
		content := []byte(`{"name":"test","value":123}`)

		// 先写入黄金文件
		Write(t, name, content)

		// 比较相同内容（格式可能不同）
		Compare(t, name, []byte(`{"value":123,"name":"test"}`))

		// 清理
		path := filepath.Join(Dir, name)
		os.Remove(path)
	})
}

func TestGolden_CompareJSONL(t *testing.T) {
	t.Run("比较 JSONL 内容", func(t *testing.T) {
		name := "test-compare-jsonl.jsonl"
		content := []byte(`{"type":"request","id":1}
{"type":"response","id":1}
`)

		// 先写入黄金文件
		Write(t, name, content)

		// 比较相同内容（字段顺序可以不同）
		CompareJSONL(t, name, []byte(`{"id":1,"type":"request"}
{"id":1,"type":"response"}`))

		// 清理
		path := filepath.Join(Dir, name)
		os.Remove(path)
	})
}

func TestGolden_Equals(t *testing.T) {
	t.Run("检查相等", func(t *testing.T) {
		name := "test-equals.txt"
		content := []byte("equal content")

		// 先写入黄金文件
		Write(t, name, content)

		// 检查相等
		require.True(t, Equals(t, name, content))

		// 清理
		path := filepath.Join(Dir, name)
		os.Remove(path)
	})

	t.Run("检查不相等", func(t *testing.T) {
		name := "test-equals-diff.txt"
		content := []byte("original content")

		// 先写入黄金文件
		Write(t, name, content)

		// 检查不相等
		require.False(t, Equals(t, name, []byte("different content")))

		// 清理
		path := filepath.Join(Dir, name)
		os.Remove(path)
	})
}

func TestGolden_CompareBytes(t *testing.T) {
	t.Run("比较相同字节", func(t *testing.T) {
		CompareBytes(t, []byte("same"), []byte("same"))
	})

	t.Run("比较带空白的字节", func(t *testing.T) {
		CompareBytes(t, []byte("  same  "), []byte("same"))
	})
}

func TestGolden_SetDir(t *testing.T) {
	originalDir := Dir
	defer func() { Dir = originalDir }()

	SetDir("/tmp/test-golden")
	require.Equal(t, "/tmp/test-golden", Dir)
}
