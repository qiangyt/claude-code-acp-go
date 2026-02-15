package comm

import (
	"fmt"
	"io"
	"sync"
	"time"
)

// ProgressReader 进度读取器，包装io.Reader并报告读取进度
type ProgressReaderT struct {
	reader      io.Reader
	total       int64
	transferred int64
	onProgress  func(transferred, total int64, speed float64)
	lastUpdate  time.Time
	lastBytes   int64
	startTime   time.Time
	mu          sync.Mutex
}

type ProgressReader = *ProgressReaderT

// NewProgressReader 创建进度读取器
// reader: 要包装的io.Reader
// total: 总大小（如果未知则为0）
// onProgress: 进度回调函数，参数为(已传输字节数, 总字节数, 传输速度bytes/s)
func NewProgressReader(reader io.Reader, total int64, onProgress func(transferred, total int64, speed float64)) ProgressReader {
	now := time.Now()
	return &ProgressReaderT{
		reader:      reader,
		total:       total,
		transferred: 0,
		onProgress:  onProgress,
		lastUpdate:  now,
		lastBytes:   0,
		startTime:   now,
	}
}

func (p ProgressReader) Read(buf []byte) (int, error) {
	n, err := p.reader.Read(buf)

	p.mu.Lock()
	p.transferred += int64(n)
	p.mu.Unlock()

	// 每秒更新一次进度
	if time.Since(p.lastUpdate) >= time.Second {
		p.reportProgress()
	}

	return n, err
}

func (p ProgressReader) reportProgress() {
	p.mu.Lock()
	defer p.mu.Unlock()

	now := time.Now()
	elapsed := now.Sub(p.lastUpdate).Seconds()
	if elapsed == 0 {
		return
	}

	// 计算速度 (bytes/秒)
	bytesSinceLastUpdate := p.transferred - p.lastBytes
	speed := float64(bytesSinceLastUpdate) / elapsed

	if p.onProgress != nil {
		p.onProgress(p.transferred, p.total, speed)
	}

	p.lastUpdate = now
	p.lastBytes = p.transferred
}

// GetStatistics 获取传输统计信息
func (p ProgressReader) GetStatistics() (transferred int64, total int64, avgSpeed float64, duration time.Duration) {
	p.mu.Lock()
	defer p.mu.Unlock()

	duration = time.Since(p.startTime)
	avgSpeed = float64(p.transferred) / duration.Seconds()
	return p.transferred, p.total, avgSpeed, duration
}

// Finish 完成传输，触发最后一次进度报告
func (p ProgressReader) Finish() {
	p.reportProgress()
}

// ProgressWriter 进度写入器，包装io.Writer并报告写入进度
type ProgressWriterT struct {
	writer      io.Writer
	total       int64
	transferred int64
	onProgress  func(transferred, total int64, speed float64)
	lastUpdate  time.Time
	lastBytes   int64
	startTime   time.Time
	mu          sync.Mutex
}

type ProgressWriter = *ProgressWriterT

// NewProgressWriter 创建进度写入器
func NewProgressWriter(writer io.Writer, total int64, onProgress func(transferred, total int64, speed float64)) ProgressWriter {
	now := time.Now()
	return &ProgressWriterT{
		writer:      writer,
		total:       total,
		transferred: 0,
		onProgress:  onProgress,
		lastUpdate:  now,
		lastBytes:   0,
		startTime:   now,
	}
}

func (p ProgressWriter) Write(buf []byte) (int, error) {
	n, err := p.writer.Write(buf)

	p.mu.Lock()
	p.transferred += int64(n)
	p.mu.Unlock()

	// 每秒更新一次进度
	if time.Since(p.lastUpdate) >= time.Second {
		p.reportProgress()
	}

	return n, err
}

func (p ProgressWriter) reportProgress() {
	p.mu.Lock()
	defer p.mu.Unlock()

	now := time.Now()
	elapsed := now.Sub(p.lastUpdate).Seconds()
	if elapsed == 0 {
		return
	}

	// 计算速度 (bytes/秒)
	bytesSinceLastUpdate := p.transferred - p.lastBytes
	speed := float64(bytesSinceLastUpdate) / elapsed

	if p.onProgress != nil {
		p.onProgress(p.transferred, p.total, speed)
	}

	p.lastUpdate = now
	p.lastBytes = p.transferred
}

// GetStatistics 获取传输统计信息
func (p ProgressWriter) GetStatistics() (transferred int64, total int64, avgSpeed float64, duration time.Duration) {
	p.mu.Lock()
	defer p.mu.Unlock()

	duration = time.Since(p.startTime)
	avgSpeed = float64(p.transferred) / duration.Seconds()
	return p.transferred, p.total, avgSpeed, duration
}

// Finish 完成传输，触发最后一次进度报告
func (p ProgressWriter) Finish() {
	p.reportProgress()
}

// FormatBytes 格式化字节数为人类可读格式
func FormatBytes(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

// FormatBits 格式化比特数为人类可读格式
func FormatBits(bits float64) string {
	const unit = 1000.0
	if bits < unit {
		return fmt.Sprintf("%.0f bit", bits)
	}
	div, exp := unit, 0
	for n := bits / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cbit", bits/div, "KMGTPE"[exp])
}

// FormatDuration 格式化时间为人类可读格式
func FormatDuration(seconds float64) string {
	if seconds < 60 {
		return fmt.Sprintf("%.0fs", seconds)
	}
	minutes := int(seconds / 60)
	secs := int(seconds) % 60
	if minutes < 60 {
		return fmt.Sprintf("%dm%ds", minutes, secs)
	}
	hours := minutes / 60
	minutes = minutes % 60
	return fmt.Sprintf("%dh%dm%ds", hours, minutes, secs)
}

// FormatSpeed 格式化传输速度
func FormatSpeed(bytesPerSec float64, format string) string {
	switch format {
	case "bits":
		bitsPerSec := bytesPerSec * 8
		return FormatBits(bitsPerSec) + "/s"
	case "bytes":
		return FormatBytes(int64(bytesPerSec)) + "/s"
	case "human", "":
		return FormatBytes(int64(bytesPerSec)) + "/s"
	default:
		return FormatBytes(int64(bytesPerSec)) + "/s"
	}
}
