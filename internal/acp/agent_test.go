package acp

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewAgent(t *testing.T) {
	t.Run("创建新的 Agent", func(t *testing.T) {
		agent := NewAgent()
		assert.NotNil(t, agent)
	})
}

func TestAgent_Run(t *testing.T) {
	t.Run("Run 在 context 取消时返回", func(t *testing.T) {
		agent := NewAgent()

		ctx, cancel := context.WithCancel(context.Background())

		// 启动 goroutine 来取消 context
		go func() {
			time.Sleep(50 * time.Millisecond)
			cancel()
		}()

		// Run 应该在 context 取消后返回
		err := agent.Run(ctx)
		require.NoError(t, err)
	})

	t.Run("Run 在已取消的 context 上立即返回", func(t *testing.T) {
		agent := NewAgent()

		ctx, cancel := context.WithCancel(context.Background())
		cancel() // 立即取消

		err := agent.Run(ctx)
		require.NoError(t, err)
	})

	t.Run("Run 正确处理带超时的 context", func(t *testing.T) {
		agent := NewAgent()

		ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
		defer cancel()

		start := time.Now()
		err := agent.Run(ctx)
		elapsed := time.Since(start)

		require.NoError(t, err)
		// 验证大约等待了 50ms（允许一定误差）
		assert.GreaterOrEqual(t, elapsed.Milliseconds(), int64(40))
	})
}
