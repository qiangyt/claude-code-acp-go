# Claude AI 开发指南 - claude-code-acp-go 项目

**Claude 必须首先阅读并理解以下文档**：

## 项目概述

**项目名称**: claude-code-acp-go
**项目路径**: `/data1/baton/claude-code-acp-go`
**通用库路径**: `/data1/baton/go-comm`

本项目使用 `go-comm` 作为通用工具库，所有开发活动必须阅读 doc/go-comm.md 来获取并严格遵循 `go-comm` 的使用规范

## TDD 强制规范

### 核心原则

1. **测试先行**：任何生产代码的编写，必须先有对应的失败测试
2. **最小实现**：只编写足以让测试通过的最小代码量
3. **持续重构**：测试通过后，立即重构代码，保持代码整洁
4. **使用不同的 sub agent 来编写测试和实现代码**

### 严格流程（Red-Green-Refactor）

#### 阶段一：Red（红灯）
- 编写一个失败的测试用例
- 运行测试，确认测试失败（红灯）
- **禁止**：在此阶段编写任何生产代码

#### 阶段二：Green（绿灯）
- 编写最小化的生产代码，仅使测试通过
- 运行测试，确认测试通过（绿灯）
- **禁止**：在此阶段进行代码重构或添加额外功能

#### 阶段三：Refactor（重构）
- 在测试保护下，优化代码结构
- 每次重构后，运行测试确保仍然通过
- **禁止**：在测试失败时进行重构

### 覆盖率要求

- **行覆盖率 (Line Coverage)**: 必须 = 100%
- **分支覆盖率 (Branch Coverage)**: 必须 = 100%
- 每个函数、每个分支、每个错误路径都必须有测试覆盖

### 禁止行为

- ❌ 跳过测试直接编写生产代码
- ❌ 先写生产代码后补测试
- ❌ 为了通过测试而跳过、注释或删除测试用例
- ❌ 提交覆盖率不达标的代码
- ❌ 在测试失败时进行代码重构
- ❌ 在一个测试用例中测试多个不相关的功能

### 验证检查点

每个功能开发完成后，必须确认：
- [ ] 所有测试通过
- [ ] 行覆盖率 = 100%
- [ ] 分支覆盖率 = 100%
- [ ] 代码已通过重构优化
- [ ] 没有被注释或跳过的测试用例

## 基本规则

1. 任何场合（包括代码注释）、除了 UI 界面要支持国际化，其它都使用中文
2. 当我报告一个 Bug 时，不要急着修复。先写一个能复现这个 Bug 的测试，然后让子代理去修复，并用通过的测试来证明修复有效
3. **TDD**：严格遵循「TDD 强制规范」章节的所有要求，禁止跳过任何步骤
4. 始终阅读 /data1/baton/claude-code-acp-go/doc/GO_COMM_USAGE.md 并遵守 /data1/baton/claude-code-acp-go/doc/GO_COMM_USAGE.md 的要求。
5. 每次任务完成做总结时，都应该回顾整个session，如果有用户指出的犯过的错误，记录到 /data1/batoning-workspace/openspec/project.md，如果有适合做成服用的skill，请创建 project scope 的skill。


<!-- OPENSPEC:START -->
# OpenSpec Instructions

These instructions are for AI assistants working in this project.

Always open `@/openspec/AGENTS.md` when the request:
- Mentions planning or proposals (words like proposal, spec, change, plan)
- Introduces new capabilities, breaking changes, architecture shifts, or big performance/security work
- Sounds ambiguous and you need the authoritative spec before coding

Use `@/openspec/AGENTS.md` to learn:
- How to create and apply change proposals
- Spec format and conventions
- Project structure and guidelines

Keep this managed block so 'openspec update' can refresh the instructions.

<!-- OPENSPEC:END -->