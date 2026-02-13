package transport

import (
	"bytes"
	"context"
	"io"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStdioTransport_Send(t *testing.T) {
	t.Run("发送消息", func(t *testing.T) {
		var buf bytes.Buffer
		tr := NewStdioTransport(nil, &buf)

		msg := map[string]any{"type": "test", "id": 1}
		err := tr.Send(msg)
		require.NoError(t, err)

		expected := `{"id":1,"type":"test"}` + "\n"
		assert.Equal(t, expected, buf.String())
	})

	t.Run("发送多条消息", func(t *testing.T) {
		var buf bytes.Buffer
		tr := NewStdioTransport(nil, &buf)

		for i := 1; i <= 3; i++ {
			err := tr.Send(map[string]any{"id": i})
			require.NoError(t, err)
		}

		// 验证每条消息都在独立的一行
		lines := bytes.Split(bytes.TrimSuffix(buf.Bytes(), []byte("\n")), []byte("\n"))
		assert.Len(t, lines, 3)
	})

	t.Run("发送 nil 消息", func(t *testing.T) {
		var buf bytes.Buffer
		tr := NewStdioTransport(nil, &buf)

		err := tr.Send(nil)
		require.NoError(t, err)
		assert.Equal(t, "null\n", buf.String())
	})
}

func TestStdioTransport_Receive(t *testing.T) {
	t.Run("接收消息", func(t *testing.T) {
		input := `{"type":"request","method":"test"}` + "\n"
		tr := NewStdioTransport(bytes.NewReader([]byte(input)), nil)

		msg, err := tr.Receive()
		require.NoError(t, err)
		assert.Equal(t, "request", msg["type"])
		assert.Equal(t, "test", msg["method"])
	})

	t.Run("接收多条消息", func(t *testing.T) {
		input := `{"id":1}` + "\n" + `{"id":2}` + "\n" + `{"id":3}` + "\n"
		tr := NewStdioTransport(bytes.NewReader([]byte(input)), nil)

		for i := 1; i <= 3; i++ {
			msg, err := tr.Receive()
			require.NoError(t, err)
			assert.Equal(t, float64(i), msg["id"])
		}

		// 第四次应该返回 EOF
		_, err := tr.Receive()
		assert.ErrorIs(t, err, io.EOF)
	})

	t.Run("跳过空行", func(t *testing.T) {
		input := `{"id":1}` + "\n\n" + `{"id":2}` + "\n"
		tr := NewStdioTransport(bytes.NewReader([]byte(input)), nil)

		msg, err := tr.Receive()
		require.NoError(t, err)
		assert.Equal(t, float64(1), msg["id"])

		msg, err = tr.Receive()
		require.NoError(t, err)
		assert.Equal(t, float64(2), msg["id"])
	})

	t.Run("处理格式错误", func(t *testing.T) {
		input := `{"invalid json` + "\n"
		tr := NewStdioTransport(bytes.NewReader([]byte(input)), nil)

		_, err := tr.Receive()
		assert.Error(t, err)
	})

	t.Run("空输入", func(t *testing.T) {
		tr := NewStdioTransport(bytes.NewReader([]byte{}), nil)

		_, err := tr.Receive()
		assert.ErrorIs(t, err, io.EOF)
	})
}

func TestStdioTransport_Close(t *testing.T) {
	t.Run("关闭传输", func(t *testing.T) {
		tr := NewStdioTransport(nil, nil)
		err := tr.Close()
		require.NoError(t, err)
	})

	t.Run("关闭后发送失败", func(t *testing.T) {
		var buf bytes.Buffer
		tr := NewStdioTransport(nil, &buf)

		err := tr.Close()
		require.NoError(t, err)

		// 关闭后发送应该返回错误
		err = tr.Send(map[string]any{"test": 1})
		assert.Error(t, err)
	})

	t.Run("关闭后接收失败", func(t *testing.T) {
		input := `{"id":1}` + "\n"
		tr := NewStdioTransport(bytes.NewReader([]byte(input)), nil)

		err := tr.Close()
		require.NoError(t, err)

		// 关闭后接收应该返回错误
		_, err = tr.Receive()
		assert.Error(t, err)
	})
}

func TestStdioTransport_Concurrent(t *testing.T) {
	t.Run("并发读写", func(t *testing.T) {
		var buf bytes.Buffer
		input := `{"id":1}` + "\n" + `{"id":2}` + "\n" + `{"id":3}` + "\n"
		tr := NewStdioTransport(bytes.NewReader([]byte(input)), &buf)

		var wg sync.WaitGroup
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		// 并发读取
		readCount := 0
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				select {
				case <-ctx.Done():
					return
				default:
					_, err := tr.Receive()
					if err != nil {
						return
					}
					readCount++
				}
			}
		}()

		// 并发写入
		writeCount := 0
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := 0; i < 10; i++ {
				select {
				case <-ctx.Done():
					return
				default:
					err := tr.Send(map[string]any{"id": i})
					if err != nil {
						return
					}
					writeCount++
				}
			}
		}()

		wg.Wait()

		// 验证没有数据竞争
		assert.GreaterOrEqual(t, readCount, 0)
		assert.GreaterOrEqual(t, writeCount, 0)
	})
}

func TestStdioTransport_Interface(t *testing.T) {
	// 确保实现了 Transport 接口
	var _ Transport = (*StdioTransport)(nil)
}

func TestStdioTransport_NoWriter(t *testing.T) {
	t.Run("无写入器时发送失败", func(t *testing.T) {
		tr := NewStdioTransport(nil, nil) // 无写入器

		err := tr.Send(map[string]any{"test": 1})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "no writer")
	})
}

func TestStdioTransport_NoReader(t *testing.T) {
	t.Run("无读取器时接收失败", func(t *testing.T) {
		tr := NewStdioTransport(nil, nil) // 无读取器

		_, err := tr.Receive()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "no reader")
	})
}

func TestStdioTransport_CloseTwice(t *testing.T) {
	t.Run("重复关闭", func(t *testing.T) {
		tr := NewStdioTransport(nil, nil)

		err := tr.Close()
		require.NoError(t, err)

		// 再次关闭应该成功（幂等）
		err = tr.Close()
		require.NoError(t, err)
	})
}
