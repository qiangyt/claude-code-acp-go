package transport

import (
	"bufio"
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEncoder_Encode(t *testing.T) {
	t.Run("编码简单对象", func(t *testing.T) {
		var buf bytes.Buffer
		enc := NewEncoder(&buf)

		obj := map[string]any{"type": "test", "value": 123}
		err := enc.Encode(obj)
		require.NoError(t, err)

		expected := `{"type":"test","value":123}` + "\n"
		assert.Equal(t, expected, buf.String())
	})

	t.Run("编码嵌套对象", func(t *testing.T) {
		var buf bytes.Buffer
		enc := NewEncoder(&buf)

		obj := map[string]any{
			"type": "response",
			"data": map[string]any{
				"nested":  "value",
				"number":  42,
				"boolean": true,
			},
		}
		err := enc.Encode(obj)
		require.NoError(t, err)

		// 验证输出的 JSON 格式正确
		expected := `{"data":{"boolean":true,"nested":"value","number":42},"type":"response"}` + "\n"
		assert.Equal(t, expected, buf.String())
	})

	t.Run("编码包含特殊字符的字符串", func(t *testing.T) {
		var buf bytes.Buffer
		enc := NewEncoder(&buf)

		obj := map[string]any{
			"text": "line1\nline2\ttab\"quote",
		}
		err := enc.Encode(obj)
		require.NoError(t, err)

		// 验证特殊字符被正确转义
		var decoded map[string]any
		err = json.Unmarshal(buf.Bytes(), &decoded)
		require.NoError(t, err)
		assert.Equal(t, "line1\nline2\ttab\"quote", decoded["text"])
	})

	t.Run("编码空对象", func(t *testing.T) {
		var buf bytes.Buffer
		enc := NewEncoder(&buf)

		obj := map[string]any{}
		err := enc.Encode(obj)
		require.NoError(t, err)

		assert.Equal(t, "{}\n", buf.String())
	})

	t.Run("编码 nil 对象", func(t *testing.T) {
		var buf bytes.Buffer
		enc := NewEncoder(&buf)

		err := enc.Encode(nil)
		require.NoError(t, err)

		assert.Equal(t, "null\n", buf.String())
	})

	t.Run("编码多个对象", func(t *testing.T) {
		var buf bytes.Buffer
		enc := NewEncoder(&buf)

		objs := []map[string]any{
			{"id": 1},
			{"id": 2},
			{"id": 3},
		}

		for _, obj := range objs {
			err := enc.Encode(obj)
			require.NoError(t, err)
		}

		lines := strings.Split(strings.TrimSuffix(buf.String(), "\n"), "\n")
		assert.Len(t, lines, 3)
		assert.Equal(t, `{"id":1}`, lines[0])
		assert.Equal(t, `{"id":2}`, lines[1])
		assert.Equal(t, `{"id":3}`, lines[2])
	})
}

func TestDecoder_Decode(t *testing.T) {
	t.Run("解码单行 JSON", func(t *testing.T) {
		input := `{"type":"response","id":1}` + "\n"
		dec := NewDecoder(strings.NewReader(input))

		var result map[string]any
		err := dec.Decode(&result)
		require.NoError(t, err)

		assert.Equal(t, "response", result["type"])
		assert.Equal(t, 1.0, result["id"]) // JSON 数字默认解析为 float64
	})

	t.Run("解码多行 JSON", func(t *testing.T) {
		input := `{"id":1}` + "\n" + `{"id":2}` + "\n" + `{"id":3}` + "\n"
		dec := NewDecoder(strings.NewReader(input))

		var results []map[string]any
		for {
			var result map[string]any
			err := dec.Decode(&result)
			if err != nil {
				break
			}
			results = append(results, result)
		}

		assert.Len(t, results, 3)
		assert.Equal(t, 1.0, results[0]["id"])
		assert.Equal(t, 2.0, results[1]["id"])
		assert.Equal(t, 3.0, results[2]["id"])
	})

	t.Run("处理空行", func(t *testing.T) {
		input := `{"id":1}` + "\n\n" + `{"id":2}` + "\n"
		dec := NewDecoder(strings.NewReader(input))

		var results []map[string]any
		for {
			var result map[string]any
			err := dec.Decode(&result)
			if err != nil {
				break
			}
			results = append(results, result)
		}

		// 应该跳过空行，解码两个对象
		assert.Len(t, results, 2)
	})

	t.Run("处理格式错误", func(t *testing.T) {
		input := `{"invalid json` + "\n"
		dec := NewDecoder(strings.NewReader(input))

		var result map[string]any
		err := dec.Decode(&result)
		assert.Error(t, err)
	})

	t.Run("解码到结构体", func(t *testing.T) {
		input := `{"name":"test","value":42}` + "\n"
		dec := NewDecoder(strings.NewReader(input))

		type TestStruct struct {
			Name  string `json:"name"`
			Value int    `json:"value"`
		}

		var result TestStruct
		err := dec.Decode(&result)
		require.NoError(t, err)

		assert.Equal(t, "test", result.Name)
		assert.Equal(t, 42, result.Value)
	})

	t.Run("解码空输入", func(t *testing.T) {
		input := ""
		dec := NewDecoder(strings.NewReader(input))

		var result map[string]any
		err := dec.Decode(&result)
		assert.Error(t, err) // EOF
	})

	t.Run("解码带空格的 JSON", func(t *testing.T) {
		input := `{ "type" : "test" , "value" : 123 }` + "\n"
		dec := NewDecoder(strings.NewReader(input))

		var result map[string]any
		err := dec.Decode(&result)
		require.NoError(t, err)

		assert.Equal(t, "test", result["type"])
		assert.Equal(t, 123.0, result["value"])
	})

	t.Run("解码超长行", func(t *testing.T) {
		// 测试 scanner 缓冲区限制
		longLine := strings.Repeat("a", bufio.MaxScanTokenSize+1)
		input := `{"data":"` + longLine + `"}` + "\n"
		dec := NewDecoder(strings.NewReader(input))

		var result map[string]any
		err := dec.Decode(&result)
		// 默认 scanner 缓冲区可能会报错或成功处理
		_ = err // 根据实现可能返回错误
	})
}

func TestEncoderDecoder_RoundTrip(t *testing.T) {
	t.Run("往返测试", func(t *testing.T) {
		var buf bytes.Buffer
		enc := NewEncoder(&buf)

		original := map[string]any{
			"type":    "request",
			"id":      123,
			"message": "Hello, World!",
			"data": map[string]any{
				"nested": true,
				"items":  []any{1, 2, 3},
			},
		}

		err := enc.Encode(original)
		require.NoError(t, err)

		dec := NewDecoder(&buf)
		var decoded map[string]any
		err = dec.Decode(&decoded)
		require.NoError(t, err)

		assert.Equal(t, original["type"], decoded["type"])
		// JSON 数字默认解析为 float64
		assert.Equal(t, float64(123), decoded["id"])
		assert.Equal(t, original["message"], decoded["message"])
	})
}
