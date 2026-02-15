// Package mcp 实现 MCP (Model Context Protocol) 服务器
//
// 本包为 ACP 提供 MCP 工具支持，将 ACP 工具暴露为 MCP 工具供 SDK 使用。
//
// # 核心组件
//
//   - Server: MCP 服务器核心，处理 JSON-RPC 消息
//   - Tool: 工具定义，包含名称、描述、输入 Schema 和处理函数
//   - SessionContext: 会话上下文，追踪终端进程
//   - PermissionChecker: 权限检查器，控制工具调用权限
//
// # 内置工具
//
//   - Read: 读取文件内容
//   - Write: 写入文件内容
//   - Edit: 编辑文件（字符串替换）
//   - Bash: 执行 Shell 命令
//   - BashOutput: 获取后台命令输出
//   - KillShell: 终止后台命令
//
// # 使用示例
//
//	server := mcp.NewServerP("my-server")
//	sessionCtx := mcp.NewSessionContext("session-1", "/work")
//
//	// 注册内置工具
//	err := mcp.RegisterBuiltinTools(server, sessionCtx)
//	if err != nil {
//	    panic(err)
//	}
//
//	// 处理 MCP 消息
//	response, err := server.HandleMessage(message)
package mcp
