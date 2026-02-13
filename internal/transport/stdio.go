package transport

import (
	"errors"
	"io"
	"sync"
)

// Transport 定义 ACP 传输层接口
type Transport interface {
	// Send 发送消息
	Send(msg any) error
	// Receive 接收消息
	Receive() (map[string]any, error)
	// Close 关闭传输
	Close() error
}

// ErrTransportClosed 表示传输已关闭
var ErrTransportClosed = errors.New("transport closed")

// StdioTransport 实现 stdin/stdout 传输
type StdioTransport struct {
	mu     sync.RWMutex
	closed bool

	encoder *Encoder
	decoder *Decoder
	reader  io.Reader
	writer  io.Writer
}

// NewStdioTransport 创建新的 stdio 传输
func NewStdioTransport(r io.Reader, w io.Writer) *StdioTransport {
	tr := &StdioTransport{
		reader: r,
		writer: w,
	}

	if w != nil {
		tr.encoder = NewEncoder(w)
	}
	if r != nil {
		tr.decoder = NewDecoder(r)
	}

	return tr
}

// Send 发送消息到 stdout
func (t *StdioTransport) Send(msg any) error {
	t.mu.RLock()
	defer t.mu.RUnlock()

	if t.closed {
		return ErrTransportClosed
	}

	if t.encoder == nil {
		return errors.New("no writer configured")
	}

	return t.encoder.Encode(msg)
}

// Receive 从 stdin 接收消息
func (t *StdioTransport) Receive() (map[string]any, error) {
	t.mu.RLock()
	defer t.mu.RUnlock()

	if t.closed {
		return nil, ErrTransportClosed
	}

	if t.decoder == nil {
		return nil, errors.New("no reader configured")
	}

	var msg map[string]any
	err := t.decoder.Decode(&msg)
	if err != nil {
		return nil, err
	}

	return msg, nil
}

// Close 关闭传输
func (t *StdioTransport) Close() error {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.closed {
		return nil
	}

	t.closed = true
	return nil
}
