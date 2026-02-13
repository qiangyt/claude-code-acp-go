// claude-code-acp Go 实现 - CLI 入口点
//go:build !test

package main

import "os"

func main() {
	cli := NewCLI()
	os.Exit(cli.Run())
}
