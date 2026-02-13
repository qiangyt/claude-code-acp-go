# Claude Code ACP Go 实现文档

> 本文档集记录了将 TypeScript 版本的 `claude-code-acp` 移植到 Go 语言的完整分析、设计和测试策略。

## 文档目录

| 文档 | 描述 |
|------|------|
| [analysis.md](./analysis.md) | 原始 TypeScript 项目分析 |
| [sdk-comparison.md](./sdk-comparison.md) | Claude Agent SDK 和 MCP Go SDK 对比 |
| [test-suite.md](./test-suite.md) | 完整测试验证集设计 |
| [implementation-guide.md](./implementation-guide.md) | 如何实现 100% 正确 |
| [architecture.md](./architecture.md) | Go 实现架构设计 |

## 项目概述

### 什么是 ACP?

**Agent Client Protocol (ACP)** 是一个标准化协议，用于代码编辑器/IDE 与 AI 编码代理之间的通信。类似于 LSP (Language Server Protocol) 标准化了语言服务器集成，ACP 标准化了代理-编辑器通信。

### 项目目标

将 [zed-industries/claude-code-acp](https://github.com/zed-industries/claude-code-acp) (TypeScript) 移植到 Go 语言，实现：

1. **协议兼容** - 与 ACP 规范 100% 兼容
2. **客户端兼容** - 与 Zed 编辑器等 ACP 客户端无缝集成
3. **功能对等** - 与 TypeScript 版本功能完全一致
4. **部署优化** - 单一静态二进制，无运行时依赖

### 技术栈

| 组件 | 选择 | 说明 |
|------|------|------|
| Claude Agent SDK | `schlunsen/claude-agent-sdk-go` | 非官方社区移植，生产就绪 |
| MCP SDK | `modelcontextprotocol/go-sdk` | 官方 Go SDK |
| Go 版本 | 1.24+ | 最新稳定版 |

## 快速链接

- [ACP 官方文档](https://agentclientprotocol.com)
- [原始 TypeScript 实现](https://github.com/zed-industries/claude-code-acp)
- [Claude Agent SDK Go](https://github.com/schlunsen/claude-agent-sdk-go)
- [MCP Go SDK 官方](https://github.com/modelcontextprotocol/go-sdk)
- [MCP Go SDK 社区](https://github.com/mark3labs/mcp-go)

## 项目状态

- [x] 原始项目分析
- [x] SDK 调研与对比
- [x] 测试验证集设计
- [x] 实现策略规划
- [ ] Go 代码实现
- [ ] 协议兼容性测试
- [ ] Zed 编辑器集成测试

---

*文档生成日期: 2026-02-13*