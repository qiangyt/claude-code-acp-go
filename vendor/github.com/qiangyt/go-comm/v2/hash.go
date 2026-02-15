package comm

import (
	"crypto/md5"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
)

// HashCalculator 哈希计算器
type HashCalculatorT struct{}

type HashCalculator = *HashCalculatorT

// NewHashCalculator 创建哈希计算器
func NewHashCalculator() HashCalculator {
	return &HashCalculatorT{}
}

// CalculateMD5 计算文件的MD5哈希值
func (h HashCalculator) CalculateMD5(filePath string) string {
	file, err := os.Open(filePath)
	if err != nil {
		panic(fmt.Sprintf("failed to open file for MD5 calculation: %v", err))
	}
	defer file.Close()

	hash := md5.New()
	if _, err := io.Copy(hash, file); err != nil {
		panic(fmt.Sprintf("failed to calculate MD5: %v", err))
	}

	return hex.EncodeToString(hash.Sum(nil))
}

// CalculateSHA256 计算文件的SHA256哈希值
func (h HashCalculator) CalculateSHA256(filePath string) string {
	file, err := os.Open(filePath)
	if err != nil {
		panic(fmt.Sprintf("failed to open file for SHA256 calculation: %v", err))
	}
	defer file.Close()

	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		panic(fmt.Sprintf("failed to calculate SHA256: %v", err))
	}

	return hex.EncodeToString(hash.Sum(nil))
}

// CalculateMD5FromReader 从io.Reader计算MD5
func (h HashCalculator) CalculateMD5FromReader(r io.Reader) string {
	hash := md5.New()
	if _, err := io.Copy(hash, r); err != nil {
		panic(fmt.Sprintf("failed to calculate MD5 from reader: %v", err))
	}
	return hex.EncodeToString(hash.Sum(nil))
}

// CalculateSHA256FromReader 从io.Reader计算SHA256
func (h HashCalculator) CalculateSHA256FromReader(r io.Reader) string {
	hash := sha256.New()
	if _, err := io.Copy(hash, r); err != nil {
		panic(fmt.Sprintf("failed to calculate SHA256 from reader: %v", err))
	}
	return hex.EncodeToString(hash.Sum(nil))
}
