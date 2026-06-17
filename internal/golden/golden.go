// Package golden 提供黄金文件测试功能
//
// 黄金文件测试是一种将实际输出与预先录制的"黄金"输出进行比较的测试方法。
// 这对于测试协议消息序列化特别有用。
package golden

import (
	"bufio"
	"encoding/json"
	"flag"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

// updateFlag 控制是否更新黄金文件
var updateFlag = flag.Bool("update-golden", false, "Update golden files instead of comparing")

// Dir 黄金文件目录
var Dir = "golden"

// SetDir 设置黄金文件目录
func SetDir(dir string) {
	Dir = dir
}

// Read 读取黄金文件内容
func Read(t *testing.T, name string) []byte {
	t.Helper()
	path := filepath.Join(Dir, name)
	data, err := os.ReadFile(path)
	require.NoError(t, err, "Failed to read golden file: %s", path)
	return data
}

// Write 写入黄金文件
func Write(t *testing.T, name string, data []byte) {
	t.Helper()
	path := filepath.Join(Dir, name)
	err := os.MkdirAll(filepath.Dir(path), 0755)
	require.NoError(t, err, "Failed to create golden directory")
	err = os.WriteFile(path, data, 0644)
	require.NoError(t, err, "Failed to write golden file: %s", path)
}

// Compare 比较实际输出与黄金文件
// 如果使用 -update-golden 标志，则更新黄金文件而不是比较
func Compare(t *testing.T, name string, actual []byte) {
	t.Helper()

	// 规范化 JSON 输出
	normalizedActual := normalizeJSON(t, actual)

	if *updateFlag {
		Write(t, name, normalizedActual)
		return
	}

	expected := Read(t, name)
	normalizedExpected := normalizeJSON(t, expected)
	CompareBytes(t, normalizedExpected, normalizedActual)
}

// normalizeJSON 规范化 JSON 输出
func normalizeJSON(t *testing.T, data []byte) []byte {
	t.Helper()
	if json.Valid(data) {
		var v any
		if err := json.Unmarshal(data, &v); err == nil {
			normalized, err := json.MarshalIndent(v, "", "  ")
			if err == nil {
				return normalized
			}
		}
	}
	return data
}

// CompareBytes 比较字节数组
func CompareBytes(t *testing.T, expected, actual []byte) {
	t.Helper()

	// 规范化比较（处理空白差异）
	expectedStr := strings.TrimSpace(string(expected))
	actualStr := strings.TrimSpace(string(actual))

	require.Equal(t, expectedStr, actualStr, "Output does not match golden file")
}

// CompareJSONL 比较 JSONL 格式的输出
// JSONL 是每行一个 JSON 对象的格式
func CompareJSONL(t *testing.T, name string, actual []byte) {
	t.Helper()

	// 解析并规范化每一行
	actualLines := parseJSONL(t, actual)

	if *updateFlag {
		// 重新格式化为标准 JSONL
		var formatted []byte
		for _, line := range actualLines {
			formatted = append(formatted, line...)
			formatted = append(formatted, '\n')
		}
		Write(t, name, formatted)
		return
	}

	expectedData := Read(t, name)
	expectedLines := parseJSONL(t, expectedData)

	require.Equal(t, len(expectedLines), len(actualLines),
		"Number of lines differs: expected %d, got %d", len(expectedLines), len(actualLines))

	for i, expected := range expectedLines {
		actual := actualLines[i]
		compareJSONObjects(t, i+1, expected, actual)
	}
}

// parseJSONL 解析 JSONL 格式的数据
func parseJSONL(t *testing.T, data []byte) [][]byte {
	t.Helper()

	var lines [][]byte
	scanner := bufio.NewScanner(strings.NewReader(string(data)))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		// 规范化 JSON
		var v any
		require.NoError(t, json.Unmarshal([]byte(line), &v), "Invalid JSON on line: %s", line)
		normalized, err := json.Marshal(v)
		require.NoError(t, err, "Failed to normalize JSON")
		lines = append(lines, normalized)
	}
	require.NoError(t, scanner.Err(), "Failed to scan JSONL")
	return lines
}

// compareJSONObjects 比较两个 JSON 对象
func compareJSONObjects(t *testing.T, lineNum int, expected, actual []byte) {
	t.Helper()

	var expectedObj, actualObj map[string]any
	require.NoError(t, json.Unmarshal(expected, &expectedObj), "Invalid expected JSON on line %d", lineNum)
	require.NoError(t, json.Unmarshal(actual, &actualObj), "Invalid actual JSON on line %d", lineNum)

	require.Equal(t, expectedObj, actualObj, "JSON objects differ on line %d", lineNum)
}

// Equals 检查实际输出是否等于黄金文件内容
func Equals(t *testing.T, name string, actual []byte) bool {
	t.Helper()

	if *updateFlag {
		Write(t, name, actual)
		return true
	}

	expected := Read(t, name)
	return string(expected) == string(actual)
}
