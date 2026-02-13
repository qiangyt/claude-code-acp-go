package main

import (
	"context"
	"errors"
	"os"
	"testing"
	"time"

	"claude-code-acp-go/internal/acp"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewCLI(t *testing.T) {
	t.Run("创建新的 CLI 实例", func(t *testing.T) {
		cli := NewCLI()
		assert.NotNil(t, cli)
		assert.NotNil(t, cli.Agent)
		assert.NotNil(t, cli.Stderr)
		assert.NotNil(t, cli.Context)
		assert.NotNil(t, cli.Cancel)
	})
}

// mockAgent 是用于测试的 mock Agent
type mockAgent struct {
	runErr error
}

func (m *mockAgent) Run(ctx context.Context) error {
	if m.runErr != nil {
		return m.runErr
	}
	<-ctx.Done()
	return nil
}

func TestCLI_Run(t *testing.T) {
	t.Run("Run 成功返回 0", func(t *testing.T) {
		cli := &CLI{
			Agent:  acp.NewAgent(),
			Stderr: os.Stderr,
			Context: func() context.Context {
				ctx, cancel := context.WithCancel(context.Background())
				cancel()
				return ctx
			}(),
			Cancel: func() {},
		}

		exitCode := cli.Run()
		assert.Equal(t, 0, exitCode)
	})

	t.Run("Run 失败返回 1", func(t *testing.T) {
		expectedErr := errors.New("test error")
		cli := &CLI{
			Agent:  &mockAgent{runErr: expectedErr},
			Stderr: os.Stderr,
			Context: func() context.Context {
				ctx, cancel := context.WithCancel(context.Background())
				cancel()
				return ctx
			}(),
			Cancel: func() {},
		}

		exitCode := cli.Run()
		assert.Equal(t, 1, exitCode)
	})

	t.Run("Run 处理信号取消", func(t *testing.T) {
		cli := NewCLI()

		// 启动 goroutine 来取消 context
		go func() {
			time.Sleep(50 * time.Millisecond)
			cli.Cancel()
		}()

		exitCode := cli.Run()
		assert.Equal(t, 0, exitCode)
	})
}

// TestCLI_Integration 是一个集成测试，验证完整的 CLI 流程
func TestCLI_Integration(t *testing.T) {
	t.Run("完整 CLI 流程", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
		defer cancel()

		cli := &CLI{
			Agent:   acp.NewAgent(),
			Stderr:  os.Stderr,
			Context: ctx,
			Cancel:  cancel,
		}

		exitCode := cli.Run()
		require.Equal(t, 0, exitCode)
	})
}
