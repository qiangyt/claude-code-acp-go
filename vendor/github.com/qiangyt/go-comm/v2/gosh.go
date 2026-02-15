package comm

import (
	"context"
	"io"
	"strings"
)

// RunGoshCommandP 执行 shell 命令（失败时 panic）
func RunGoshCommandP(vars map[string]string, dir string, cmd string, passwordInput FnInput) CommandOutput {
	r, err := RunGoshCommand(vars, dir, cmd, passwordInput)
	if err != nil {
		panic(err)
	}
	return r
}

// RunGoshCommand 执行 shell 命令
func RunGoshCommand(vars map[string]string, dir string, cmd string, passwordInput FnInput) (CommandOutput, error) {
	var stdin io.Reader
	if IsSudoCommand(cmd) {
		password := InputSudoPassword(passwordInput)
		if len(password) > 0 {
			stdin = strings.NewReader(password + "\n")
			cmd = InstrumentSudoCommand(cmd)
		}
	}

	// 创建默认配置，注册 zenity 处理器
	config := DefaultGoshConfig().
		WithGoHandler("zenity", ExecZenityHandler)

	executor := NewGoshExecutor(config)

	out := strings.Builder{}
	err := executor.RunWithVars(context.TODO(), vars, dir, cmd, stdin, &out, &out)
	if err != nil {
		return nil, err
	}

	return ParseCommandOutput(out.String())
}
