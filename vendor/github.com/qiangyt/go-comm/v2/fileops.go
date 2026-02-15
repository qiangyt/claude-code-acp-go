package comm

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"

	"github.com/spf13/afero"
)

// FileOps 通用文件操作
type FileOpsT struct {
	fs       afero.Fs
	localize LocalizeFunc
}

type FileOps = *FileOpsT

// NewFileOps 创建文件操作器
func NewFileOps(fs afero.Fs) FileOps {
	if fs == nil {
		fs = afero.NewOsFs()
	}
	return &FileOpsT{
		fs:       fs,
		localize: DefaultLocalizeFunc,
	}
}

// SetLocalizeFunc 设置本地化函数
func (f FileOps) SetLocalizeFunc(localize LocalizeFunc) {
	if localize != nil {
		f.localize = localize
	}
}

// SetPermissions 设置文件权限
// chmod: 权限字符串，如 "755", "0644"
func (f FileOps) SetPermissions(filePath, chmod string) {
	if chmod == "" {
		return
	}

	if runtime.GOOS == "windows" {
		// Windows不支持chmod
		return
	}

	// 解析权限字符串（支持0644、644等格式）
	chmod = strings.TrimPrefix(chmod, "0")
	perm, err := strconv.ParseUint(chmod, 8, 32)
	if err != nil {
		panic(fmt.Sprintf("%s: %s", f.localize("InvalidChmodValue", map[string]any{"value": chmod}), err.Error()))
	}

	if err := f.fs.Chmod(filePath, os.FileMode(perm)); err != nil {
		panic(fmt.Sprintf("%s: %v", f.localize("FailedToChmod"), err))
	}
}

// SetOwner 设置文件所有者
// chown: 所有者字符串，如 "user:group"
func (f FileOps) SetOwner(filePath, chown string, useSudo bool, sudoPassword string) {
	if chown == "" {
		return
	}

	if runtime.GOOS == "windows" {
		// Windows不支持chown
		return
	}

	// 构建chown命令
	var cmd *exec.Cmd
	if useSudo {
		if sudoPassword != "" {
			cmd = exec.Command("sudo", "-S", "chown", chown, filePath)
			cmd.Stdin = strings.NewReader(sudoPassword + "\n")
		} else {
			cmd = exec.Command("sudo", "chown", chown, filePath)
		}
	} else {
		cmd = exec.Command("chown", chown, filePath)
	}

	output, err := cmd.CombinedOutput()
	if err != nil {
		panic(f.localize("FailedToChown", map[string]any{"output": string(output)}))
	}
}

// CopyFile 复制文件
func (f FileOps) CopyFile(srcPath, destPath string) {
	// 确保目标目录存在
	destDir := filepath.Dir(destPath)
	if err := f.fs.MkdirAll(destDir, 0o755); err != nil {
		panic(fmt.Sprintf("%s: %v", f.localize("FailedToCreateDestDir"), err))
	}

	// 打开源文件
	srcFile, err := f.fs.Open(srcPath)
	if err != nil {
		panic(fmt.Sprintf("%s: %v", f.localize("FailedToOpenSource"), err))
	}
	defer srcFile.Close()

	// 创建目标文件
	destFile, err := f.fs.Create(destPath)
	if err != nil {
		panic(fmt.Sprintf("%s: %v", f.localize("FailedToCreateDest"), err))
	}
	defer destFile.Close()

	// 复制内容
	if _, err := io.Copy(destFile, srcFile); err != nil {
		panic(fmt.Sprintf("%s: %v", f.localize("FailedToCopyFile"), err))
	}

	// 复制文件权限
	srcInfo, err := f.fs.Stat(srcPath)
	if err == nil {
		f.fs.Chmod(destPath, srcInfo.Mode())
	}
}

// MoveFile 移动文件
func (f FileOps) MoveFile(srcPath, destPath string) {
	// 首先尝试直接重命名（同一文件系统）
	if err := f.fs.Rename(srcPath, destPath); err == nil {
		return
	}

	// 如果失败，则复制后删除
	f.CopyFile(srcPath, destPath)

	if err := f.fs.Remove(srcPath); err != nil {
		panic(fmt.Sprintf("%s: %v", f.localize("FailedToRemoveSource"), err))
	}
}

// CreateSymlink 创建符号链接
func (f FileOps) CreateSymlink(targetPath, linkPath string, useSudo bool, sudoPassword string) {
	if linkPath == "" {
		return
	}

	if runtime.GOOS == "windows" {
		// Windows符号链接需要特殊处理
		return
	}

	// 确保链接目录存在
	linkDir := filepath.Dir(linkPath)
	if err := f.fs.MkdirAll(linkDir, 0o755); err != nil {
		panic(fmt.Sprintf("%s: %v", f.localize("FailedToCreateLinkDir"), err))
	}

	// 删除已存在的链接
	if _, err := f.fs.Stat(linkPath); err == nil {
		if useSudo {
			var cmd *exec.Cmd
			if sudoPassword != "" {
				cmd = exec.Command("sudo", "-S", "rm", "-f", linkPath)
				cmd.Stdin = strings.NewReader(sudoPassword + "\n")
			} else {
				cmd = exec.Command("sudo", "rm", "-f", linkPath)
			}
			cmd.Run()
		} else {
			f.fs.Remove(linkPath)
		}
	}

	// 创建符号链接
	var cmd *exec.Cmd
	if useSudo {
		if sudoPassword != "" {
			cmd = exec.Command("sudo", "-S", "ln", "-s", targetPath, linkPath)
			cmd.Stdin = strings.NewReader(sudoPassword + "\n")
		} else {
			cmd = exec.Command("sudo", "ln", "-s", targetPath, linkPath)
		}
	} else {
		cmd = exec.Command("ln", "-s", targetPath, linkPath)
	}

	output, err := cmd.CombinedOutput()
	if err != nil {
		panic(f.localize("FailedToCreateSymlink", map[string]any{"output": string(output)}))
	}
}

// EnsureDir 确保目录存在
func (f FileOps) EnsureDir(dirPath string) {
	if err := f.fs.MkdirAll(dirPath, 0o755); err != nil {
		panic(fmt.Sprintf("%s: %v", f.localize("FailedToCreateDir"), err))
	}
}

// FileExists 检查文件是否存在
func (f FileOps) FileExists(filePath string) bool {
	_, err := f.fs.Stat(filePath)
	return err == nil
}

// DirExists 检查目录是否存在
func (f FileOps) DirExists(dirPath string) bool {
	info, err := f.fs.Stat(dirPath)
	return err == nil && info.IsDir()
}

// GetFileSize 获取文件大小
func (f FileOps) GetFileSize(filePath string) int64 {
	info, err := f.fs.Stat(filePath)
	if err != nil {
		panic(fmt.Sprintf("%s: %v", f.localize("FailedToGetFileSize"), err))
	}
	return info.Size()
}

// RenameFile 重命名文件
func (f FileOps) RenameFile(oldPath, newName string) string {
	if newName == "" {
		return oldPath
	}

	dir := filepath.Dir(oldPath)
	newPath := filepath.Join(dir, newName)

	if err := f.fs.Rename(oldPath, newPath); err != nil {
		panic(fmt.Sprintf("%s: %v", f.localize("FailedToRenameFile"), err))
	}

	return newPath
}

// CreateDirectory 创建目录
func (f FileOps) CreateDirectory(dirPath string, chmod, chown string, useSudo bool, sudoPassword string) {
	// 创建目录
	if useSudo {
		var cmd *exec.Cmd
		if sudoPassword != "" {
			cmd = exec.Command("sudo", "-S", "mkdir", "-p", dirPath)
			cmd.Stdin = strings.NewReader(sudoPassword + "\n")
		} else {
			cmd = exec.Command("sudo", "mkdir", "-p", dirPath)
		}
		output, err := cmd.CombinedOutput()
		if err != nil {
			panic(f.localize("FailedToCreateDirSudo", map[string]any{"output": string(output)}))
		}
	} else {
		if err := f.fs.MkdirAll(dirPath, 0o755); err != nil {
			panic(fmt.Sprintf("%s: %v", f.localize("FailedToCreateDir"), err))
		}
	}

	// 设置权限
	if chmod != "" {
		f.SetPermissions(dirPath, chmod)
	}

	// 设置所有者
	if chown != "" {
		f.SetOwner(dirPath, chown, useSudo, sudoPassword)
	}
}

// RemoveFile 删除文件
func (f FileOps) RemoveFile(filePath string, useSudo bool, sudoPassword string) {
	if filePath == "" {
		return
	}

	var cmd *exec.Cmd
	if useSudo {
		if sudoPassword != "" {
			cmd = exec.Command("sudo", "-S", "rm", "-f", filePath)
			cmd.Stdin = strings.NewReader(sudoPassword + "\n")
		} else {
			cmd = exec.Command("sudo", "rm", "-f", filePath)
		}
		cmd.CombinedOutput()
	} else {
		f.fs.Remove(filePath)
	}
}

// GetFs 获取文件系统
func (f FileOps) GetFs() afero.Fs {
	return f.fs
}
