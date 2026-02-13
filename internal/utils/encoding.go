package utils

import (
	"net/url"
	"path/filepath"
	"strings"
)

// EncodePath 对路径进行 URL 编码
func EncodePath(path string) string {
	return url.PathEscape(path)
}

// DecodePath 解码 URL 编码的路径
func DecodePath(encoded string) (string, error) {
	return url.PathUnescape(encoded)
}

// RelPath 计算相对于 base 的路径
func RelPath(base, target string) (string, error) {
	return filepath.Rel(base, target)
}

// AbsPath 将相对路径转换为绝对路径
func AbsPath(base, path string) (string, error) {
	if filepath.IsAbs(path) {
		return path, nil
	}
	return filepath.Abs(filepath.Join(base, path))
}

// NormalizePath 规范化路径（移除冗余的 . 和 ..）
func NormalizePath(path string) string {
	return filepath.Clean(path)
}

// IsAbsPath 检查路径是否为绝对路径
func IsAbsPath(path string) bool {
	return filepath.IsAbs(path)
}

// PathInDir 检查路径是否在指定目录内
func PathInDir(dir, path string) (bool, error) {
	absDir, err := filepath.Abs(dir)
	if err != nil {
		return false, err
	}

	absPath, err := filepath.Abs(path)
	if err != nil {
		return false, err
	}

	// 确保目录以分隔符结尾，避免部分匹配
	if !strings.HasSuffix(absDir, string(filepath.Separator)) {
		absDir += string(filepath.Separator)
	}

	return strings.HasPrefix(absPath+string(filepath.Separator), absDir), nil
}

// FileExtension 获取文件扩展名
func FileExtension(path string) string {
	return filepath.Ext(path)
}

// JoinPath 连接路径元素
func JoinPath(elem ...string) string {
	// 过滤空元素
	var parts []string
	for _, e := range elem {
		if e != "" {
			parts = append(parts, e)
		}
	}
	return filepath.Join(parts...)
}

// BaseName 获取路径的最后一部分
func BaseName(path string) string {
	return filepath.Base(path)
}

// DirName 获取路径的目录部分
func DirName(path string) string {
	return filepath.Dir(path)
}
