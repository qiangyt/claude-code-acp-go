package transport

import (
	"bufio"
	"encoding/json"
	"io"
)

// Encoder 将 JSON 对象编码为 NDJSON 格式
type Encoder struct {
	encoder *json.Encoder
	writer  io.Writer
}

// NewEncoder 创建新的 NDJSON 编码器
func NewEncoder(w io.Writer) *Encoder {
	enc := json.NewEncoder(w)
	enc.SetEscapeHTML(false) // 不转义 HTML 字符，与 TypeScript 版本一致
	return &Encoder{
		encoder: enc,
		writer:  w,
	}
}

// Encode 将单个对象编码为一行 JSON
func (e *Encoder) Encode(v any) error {
	return e.encoder.Encode(v)
}

// Decoder 从 NDJSON 格式解码 JSON 对象
type Decoder struct {
	scanner *bufio.Scanner
	reader  io.Reader
}

// NewDecoder 创建新的 NDJSON 解码器
func NewDecoder(r io.Reader) *Decoder {
	return &Decoder{
		scanner: bufio.NewScanner(r),
		reader:  r,
	}
}

// Decode 从输入流解码单个 JSON 对象
// 跳过空行，返回遇到的第一个错误
func (d *Decoder) Decode(v any) error {
	for d.scanner.Scan() {
		line := d.scanner.Text()
		if len(line) == 0 {
			continue // 跳过空行
		}
		return json.Unmarshal([]byte(line), v)
	}
	if err := d.scanner.Err(); err != nil {
		return err
	}
	return io.EOF
}
