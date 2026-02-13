package tools

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestToolKind(t *testing.T) {
	t.Run("工具类型常量", func(t *testing.T) {
		assert.Equal(t, ToolKind("read"), ToolKindRead)
		assert.Equal(t, ToolKind("edit"), ToolKindEdit)
		assert.Equal(t, ToolKind("move"), ToolKindMove)
		assert.Equal(t, ToolKind("search"), ToolKindSearch)
		assert.Equal(t, ToolKind("execute"), ToolKindExecute)
		assert.Equal(t, ToolKind("think"), ToolKindThink)
		assert.Equal(t, ToolKind("fetch"), ToolKindFetch)
		assert.Equal(t, ToolKind("switch_mode"), ToolKindSwitchMode)
		assert.Equal(t, ToolKind("other"), ToolKindOther)
	})
}

func TestToolInfo(t *testing.T) {
	t.Run("序列化基本信息", func(t *testing.T) {
		info := ToolInfo{
			Title: "Read file.txt",
			Kind:  ToolKindRead,
		}

		data, err := json.Marshal(info)
		require.NoError(t, err)

		var result map[string]any
		err = json.Unmarshal(data, &result)
		require.NoError(t, err)

		assert.Equal(t, "Read file.txt", result["title"])
		assert.Equal(t, "read", result["kind"])
	})

	t.Run("带位置信息", func(t *testing.T) {
		info := ToolInfo{
			Title: "Read file.go (1-100)",
			Kind:  ToolKindRead,
			Locations: []ToolCallLocation{
				{Path: "/path/to/file.go", Line: 1},
			},
		}

		data, err := json.Marshal(info)
		require.NoError(t, err)

		var decoded ToolInfo
		err = json.Unmarshal(data, &decoded)
		require.NoError(t, err)

		assert.Len(t, decoded.Locations, 1)
		assert.Equal(t, "/path/to/file.go", decoded.Locations[0].Path)
		assert.Equal(t, 1, decoded.Locations[0].Line)
	})

	t.Run("带 diff 内容", func(t *testing.T) {
		oldText := "old content"
		info := ToolInfo{
			Title: "Write file.go",
			Kind:  ToolKindEdit,
			Content: []ToolCallContent{
				{
					Type:    "diff",
					Path:    "/path/to/file.go",
					OldText: &oldText,
					NewText: "new content",
				},
			},
		}

		data, err := json.Marshal(info)
		require.NoError(t, err)

		var decoded ToolInfo
		err = json.Unmarshal(data, &decoded)
		require.NoError(t, err)

		assert.Len(t, decoded.Content, 1)
		assert.Equal(t, "diff", decoded.Content[0].Type)
		assert.Equal(t, "old content", *decoded.Content[0].OldText)
		assert.Equal(t, "new content", decoded.Content[0].NewText)
	})
}

func TestToolCallLocation(t *testing.T) {
	t.Run("完整位置", func(t *testing.T) {
		loc := ToolCallLocation{
			Path: "/home/user/project/main.go",
			Line: 42,
		}

		data, err := json.Marshal(loc)
		require.NoError(t, err)

		var decoded ToolCallLocation
		err = json.Unmarshal(data, &decoded)
		require.NoError(t, err)

		assert.Equal(t, loc.Path, decoded.Path)
		assert.Equal(t, loc.Line, decoded.Line)
	})

	t.Run("仅路径", func(t *testing.T) {
		loc := ToolCallLocation{
			Path: "/home/user/project/main.go",
		}

		data, err := json.Marshal(loc)
		require.NoError(t, err)

		var result map[string]any
		err = json.Unmarshal(data, &result)
		require.NoError(t, err)

		assert.Equal(t, "/home/user/project/main.go", result["path"])
		assert.Nil(t, result["line"]) // omitempty
	})
}

func TestToolCallContent(t *testing.T) {
	t.Run("diff 类型", func(t *testing.T) {
		jsonStr := `{"type":"diff","path":"/file.go","oldText":"old","newText":"new"}`

		var content ToolCallContent
		err := json.Unmarshal([]byte(jsonStr), &content)
		require.NoError(t, err)

		assert.Equal(t, "diff", content.Type)
		assert.Equal(t, "/file.go", content.Path)
		require.NotNil(t, content.OldText)
		assert.Equal(t, "old", *content.OldText)
		assert.Equal(t, "new", content.NewText)
	})

	t.Run("无 oldText (新文件)", func(t *testing.T) {
		jsonStr := `{"type":"diff","path":"/file.go","newText":"content"}`

		var content ToolCallContent
		err := json.Unmarshal([]byte(jsonStr), &content)
		require.NoError(t, err)

		assert.Equal(t, "diff", content.Type)
		assert.Nil(t, content.OldText)
		assert.Equal(t, "content", content.NewText)
	})
}
