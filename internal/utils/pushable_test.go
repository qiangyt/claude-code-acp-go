package utils

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPushable_Push(t *testing.T) {
	t.Run("推送单个元素", func(t *testing.T) {
		p := NewPushable[int](10)

		p.Push(1)
		p.End()

		var results []int
		for v := range p.Channel() {
			results = append(results, v)
		}

		assert.Equal(t, []int{1}, results)
	})

	t.Run("推送多个元素", func(t *testing.T) {
		p := NewPushable[int](10)

		for i := 1; i <= 5; i++ {
			p.Push(i)
		}
		p.End()

		var results []int
		for v := range p.Channel() {
			results = append(results, v)
		}

		assert.Equal(t, []int{1, 2, 3, 4, 5}, results)
	})

	t.Run("推送到已关闭的流", func(t *testing.T) {
		p := NewPushable[int](10)

		p.End()
		// 推送到已关闭的流应该被忽略
		p.Push(1)

		var results []int
		for v := range p.Channel() {
			results = append(results, v)
		}

		assert.Empty(t, results)
	})
}

func TestPushable_End(t *testing.T) {
	t.Run("结束流", func(t *testing.T) {
		p := NewPushable[int](10)

		p.Push(1)
		p.End()

		// 多次 End 应该是安全的
		p.End()
		p.End()

		var results []int
		for v := range p.Channel() {
			results = append(results, v)
		}

		assert.Equal(t, []int{1}, results)
	})
}

func TestPushable_Channel(t *testing.T) {
	t.Run("获取 channel", func(t *testing.T) {
		p := NewPushable[string](10)

		ch := p.Channel()
		assert.NotNil(t, ch)

		p.Push("test")
		p.End()

		val := <-ch
		assert.Equal(t, "test", val)
	})
}

func TestPushable_Concurrent(t *testing.T) {
	t.Run("并发推送", func(t *testing.T) {
		p := NewPushable[int](100)

		var wg sync.WaitGroup
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		// 10 个 goroutine 并发推送
		for i := 0; i < 10; i++ {
			wg.Add(1)
			go func(start int) {
				defer wg.Done()
				for j := 0; j < 10; j++ {
					select {
					case <-ctx.Done():
						return
					default:
						p.Push(start*10 + j)
					}
				}
			}(i)
		}

		wg.Wait()
		p.End()

		// 收集所有结果
		var results []int
		for v := range p.Channel() {
			results = append(results, v)
		}

		// 应该有 100 个元素
		assert.Len(t, results, 100)
	})

	t.Run("并发读写", func(t *testing.T) {
		p := NewPushable[int](100)

		var readWg, writeWg sync.WaitGroup
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		readCount := 0
		readWg.Add(1)
		go func() {
			defer readWg.Done()
			for {
				select {
				case <-ctx.Done():
					return
				case _, ok := <-p.Channel():
					if !ok {
						return
					}
					readCount++
				}
			}
		}()

		// 并发写入
		for i := 0; i < 50; i++ {
			writeWg.Add(1)
			go func(val int) {
				defer writeWg.Done()
				p.Push(val)
			}(i)
		}

		writeWg.Wait()
		p.End()
		readWg.Wait()

		assert.Equal(t, 50, readCount)
	})
}

func TestPushable_BufferSize(t *testing.T) {
	t.Run("带缓冲的流", func(t *testing.T) {
		p := NewPushable[int](3)

		// 缓冲区大小为 3，应该能推送 3 个元素而不阻塞
		p.Push(1)
		p.Push(2)
		p.Push(3)
		p.End()

		var results []int
		for v := range p.Channel() {
			results = append(results, v)
		}

		assert.Equal(t, []int{1, 2, 3}, results)
	})

	t.Run("无缓冲的流", func(t *testing.T) {
		p := NewPushable[int](0)

		// 无缓冲，需要消费者
		done := make(chan struct{})
		var results []int
		go func() {
			defer close(done)
			for v := range p.Channel() {
				results = append(results, v)
			}
		}()

		p.Push(1)
		p.Push(2)
		p.End()

		<-done
		assert.Equal(t, []int{1, 2}, results)
	})
}

func TestPushable_Generic(t *testing.T) {
	t.Run("字符串类型", func(t *testing.T) {
		p := NewPushable[string](10)

		p.Push("hello")
		p.Push("world")
		p.End()

		var results []string
		for v := range p.Channel() {
			results = append(results, v)
		}

		assert.Equal(t, []string{"hello", "world"}, results)
	})

	t.Run("结构体类型", func(t *testing.T) {
		type TestStruct struct {
			Name  string
			Value int
		}

		p := NewPushable[TestStruct](10)

		p.Push(TestStruct{Name: "test1", Value: 1})
		p.Push(TestStruct{Name: "test2", Value: 2})
		p.End()

		var results []TestStruct
		for v := range p.Channel() {
			results = append(results, v)
		}

		require.Len(t, results, 2)
		assert.Equal(t, "test1", results[0].Name)
		assert.Equal(t, "test2", results[1].Name)
	})
}

func TestPushable_IsClosed(t *testing.T) {
	t.Run("未关闭时", func(t *testing.T) {
		p := NewPushable[int](10)
		assert.False(t, p.IsClosed())
	})

	t.Run("关闭后", func(t *testing.T) {
		p := NewPushable[int](10)
		p.End()
		assert.True(t, p.IsClosed())
	})

	t.Run("并发检查关闭状态", func(t *testing.T) {
		p := NewPushable[int](10)

		var wg sync.WaitGroup
		for i := 0; i < 10; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				_ = p.IsClosed()
			}()
		}

		p.End()
		wg.Wait()
	})
}
