package comm

import (
	"os"
	"runtime"
	"strings"
)

// SystemInfo 系统信息
type SystemInfo struct {
	OS      string // 操作系统: linux, windows, macos
	Dist    string // Linux发行版: ubuntu, debian, centos, fedora等
	Version string // 发行版版本
	Arch    string // 架构: x86, amd64, arm64, armhf
}

// SystemDetector 系统检测器
type SystemDetectorT struct {
	info SystemInfo
}

type SystemDetector = *SystemDetectorT

// NewSystemDetector 创建系统检测器
func NewSystemDetector() SystemDetector {
	detector := &SystemDetectorT{}
	detector.detect()
	return detector
}

// detect 检测系统信息
func (d *SystemDetectorT) detect() {
	// 检测操作系统
	d.info.OS = runtime.GOOS
	switch d.info.OS {
	case "darwin":
		d.info.OS = "macos"
	case "windows":
		d.info.OS = "windows"
	case "linux":
		d.info.OS = "linux"
	}

	// 检测架构
	d.info.Arch = runtime.GOARCH
	switch d.info.Arch {
	case "386":
		d.info.Arch = "x86"
	case "amd64":
		d.info.Arch = "amd64"
	case "arm64":
		d.info.Arch = "arm64"
	case "arm":
		d.info.Arch = "armhf"
	}

	// 如果是Linux，检测发行版
	if d.info.OS == "linux" {
		d.detectLinuxDist()
	}
}

// detectLinuxDist 检测Linux发行版
func (d *SystemDetectorT) detectLinuxDist() {
	// 尝试读取 /etc/os-release
	data, err := os.ReadFile("/etc/os-release")
	if err != nil {
		// 尝试读取 /etc/lsb-release
		data, err = os.ReadFile("/etc/lsb-release")
		if err != nil {
			return
		}
	}

	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "ID=") {
			d.info.Dist = strings.Trim(strings.TrimPrefix(line, "ID="), "\"")
		} else if strings.HasPrefix(line, "VERSION_ID=") {
			d.info.Version = strings.Trim(strings.TrimPrefix(line, "VERSION_ID="), "\"")
		}
	}
}

// GetSystemInfo 获取系统信息
func (d SystemDetector) GetSystemInfo() SystemInfo {
	return d.info
}

// IsLinux 是否为Linux系统
func (d SystemDetector) IsLinux() bool {
	return d.info.OS == "linux"
}

// IsWindows 是否为Windows系统
func (d SystemDetector) IsWindows() bool {
	return d.info.OS == "windows"
}

// IsMacOS 是否为macOS系统
func (d SystemDetector) IsMacOS() bool {
	return d.info.OS == "macos"
}

// MatchSystem 检查系统是否匹配指定条件
// 支持逗号或空格分隔的多个值
func (d SystemDetector) MatchSystem(os, dist, version, arch string) bool {
	if os != "" && !matchValues(d.info.OS, os) {
		return false
	}
	if dist != "" && !matchValues(d.info.Dist, dist) {
		return false
	}
	if version != "" && !matchValues(d.info.Version, version) {
		return false
	}
	if arch != "" && !matchValues(d.info.Arch, arch) {
		return false
	}
	return true
}

// matchValues 检查值是否匹配（支持逗号或空格分隔的多个值）
func matchValues(actual, expected string) bool {
	if expected == "" {
		return true
	}

	// 分割多个值（逗号或空格）
	values := strings.FieldsFunc(expected, func(r rune) bool {
		return r == ',' || r == ' '
	})

	// 检查是否匹配任何一个值
	for _, v := range values {
		v = strings.TrimSpace(v)
		if v == actual {
			return true
		}
	}

	return false
}
