本项目使用 `go-comm` 作为通用工具库，所有开发活动必须遵循 `go-comm` 的使用规范。

## 核心原则

### 1. go-comm 优先原则

在编写任何代码前，**必须**先检查 `go-comm` 是否已提供相同或相似功能：

1. **搜索 go-comm**: 在 `/data1/baton/go-comm` 中搜索相关功能
2. **阅读源码**: 查看相关源文件，了解功能和使用方法
3. **优先复用**: 如果 `go-comm` 已有该功能，直接使用，不要重新实现

### 2. 依赖一致性原则

本项目的所有依赖必须与 `go-comm` 保持完全一致：

- 查看 `go-comm/go.mod` 中的依赖版本
- 确保 `batoning/go.mod` 使用相同版本
- 使用本地依赖：`replace github.com/qiangyt/go-comm/v2 => ../go-comm`

### 3. 代码风格一致原则

遵循 `go-comm` 的代码风格和编码约定：

- 命名约定（类型用 `T` 后缀，函数用 `P` 后缀表示 panic 版本）
- 错误处理: 只使用 panic 模式
- 类型转换（Required/Optional 模式）
- 国际化（所有用户可见消息必须使用 i18n）

### 4. 功能提取原则

当编写新功能时，如果该功能可能被其他项目使用：

- 评估功能的通用性
- 将可重用部分提取到 `go-comm`
- 遵循 `go-comm` 的代码风格
- 添加完整的测试和文档
- 在项目中引用 `go-comm` 的功能

1. **GO_COMM_USAGE.md** - go-comm 使用规范（位于项目根目录）
   - 详细说明 `go-comm` 的所有功能模块
   - 代码复用原则和最佳实践
   - 依赖管理规范
   - 代码风格和约定
   - 开发工作流程

2. **go-comm 源码** - 通用库源码（位于 `/data1/batoning-workspace/go-comm`）
   - 核心功能实现
   - API 使用示例
   - 代码风格参考

## 开发流程

### 功能开发流程

```
1. 需求分析
   ↓
2. 搜索 go-comm（查找现有功能）
   ↓
3. 评估复用性
   ├─ 直接复用 → 使用 go-comm 功能
   ├─ 增强后复用 → 更新 go-comm → 使用更新后的功能
   └─ 新建功能 → 在项目中实现 → 评估通用性 → 可能提取到 go-comm
   ↓
4. 编写代码（遵循 go-comm 代码风格），执行 `mise build` 执行构建
   ↓
5. 测试，执行 `mise test` 执行构建
   ↓
6. 文档更新
```
##
1. 任何场合（包括代码注释）、除了 UI 界面要支持国际化，其它都使用中文
2. 当我报告一个 Bug 时，不要急着修复。先写一个能复现这个 Bug 的测试，然后让子代理去修复，并用通过的测试来证明修复有效
3. TDD：先写测试，再写生产代码，然后迭代生产代码直至所有测试都通过，最后根据 coverage，补充测试 case，要求行 coverage===100%和 branch coverage===100%
4. 始终阅读 /data1/baton/batoning/doc/GO_COMM_USAGE.md 并遵守 /data1/baton/batoning/doc/GO_COMM_USAGE.md 的要求。
5. 每次任务完成做总结时，都应该回顾整个session，如果有用户指出的犯过的错误，记录到 /data1/batoning-workspace/openspec/project.md，如果有适合做成服用的skill，请创建 project scope 的skill。
6. 目标项目：/data1/batoning-workspace/batoning

### 检查清单

在编写代码时，使用以下清单确保符合规范：

#### 复用性检查
- [ ] 已检查 `go-comm` 是否提供相同功能
- [ ] 如果功能略有偏差，已考虑更新 `go-comm`
- [ ] 如果新功能具有通用性，已提取到 `go-comm`

#### 依赖一致性检查
- [ ] 所有依赖版本与 `go-comm/go.mod` 一致
- [ ] 新依赖已评估是否应该添加到 `go-comm`

#### 代码风格检查
- [ ] 使用 `go-comm` 的命名约定
- [ ] 遵循双函数模式（普通 + Panic 版本）
- [ ] 使用 Required/Optional 模式进行类型转换
- [ ] 用户可见消息使用 i18n

#### 日志和错误处理检查
- [ ] 使用 `go-comm` 的日志系统（`comm.Logger`）
- [ ] 错误消息已本地化（使用 `comm.T()` 或 `comm.LocalizeError()`）
- [ ] 使用适当的日志级别（Debug/Info/Error/Warn）

## go-comm 核心功能速查

### 日志系统
```go
import "github.com/qiangyt/go-comm/v2"

logger := comm.NewLoggerP(console, &comm.LoggerConfigT{
    MaxSize:    100,
    MaxAge:     30,
    MaxBackups: 10,
    Compress:   true,
}, "/var/log/app.log")

logger.Info().Str("key", "value").Msg("message")
logger.Error(err).Msg("error occurred")
```

### 配置管理
```go
cfg := &MyConfig{}
result, _, err := comm.DecodeWithYaml(yamlText, comm.StrictConfigConfig(), cfg, defaults)
```

### 国际化
```go
msg := comm.T("error.required", map[string]any{
    "Hint": "config",
    "Key":  "name",
})
```

### 类型转换
```go
// 必需字段
name := comm.RequiredStringP("config", "name", m)

// 可选字段
port, has := comm.OptionalStringP("config", "port", m, "8080")

// Map 转换
opts := comm.RequiredMapP("config", "options", m)
```

### 文件操作
```go
ops := comm.NewFileOps(fs)
ops.CopyFileWithMode(src, dest, "0644", "user:group", false, "")
```

### Shell 命令
```go
output, err := comm.RunGoshCommand(vars, "/tmp", "ls -la", nil)
```

## 常见错误示例

### ❌ 错误：重复实现 go-comm 已有功能
```go
// 不要这样做
import "log"

log.Println("Application started")
```

### ✅ 正确：使用 go-comm 的日志系统
```go
import "github.com/qiangyt/go-comm/v2"

logger.Info().Msg("Application started")
```

### ❌ 错误：手动类型断言
```go
// 不要这样做
name, ok := m["name"].(string)
if !ok {
    return fmt.Errorf("name is required")
}
```

### ✅ 正确：使用 go-comm 的类型转换
```go
import "github.com/qiangyt/go-comm/v2"

name := comm.RequiredStringP("config", "name", m)
```

### ❌ 错误：硬编码错误消息
```go
// 不要这样做
return fmt.Errorf("field %s is required", fieldName)
```

### ✅ 正确：使用 i18n
```go
import "github.com/qiangyt/go-comm/v2"

return comm.LocalizeError("error.required", map[string]any{
    "Hint": "config",
    "Key":  fieldName,
})
```

## 项目结构

```
batoning/
├── CLAUDE.md              # 本文件
├── GO_COMM_USAGE.md       # go-comm 使用规范
├── go.mod                 # Go 模块定义
├── README.md              # 项目说明
└── scripts/               # 脚本目录
```

## 依赖关系

```go
// go.mod
module github.com/qiangyt/batoning

require (
    github.com/qiangyt/go-comm/v2 v2.6.11
    // ... 其他依赖，必须与 go-comm 保持版本一致
)

// 使用本地依赖（开发时）
replace github.com/qiangyt/go-comm/v2 => ../go-comm
```

## 关键提醒

1. **永不重复造轮子**: 在编写任何工具函数前，先检查 `go-comm`
2. **保持依赖一致**: 所有依赖版本必须与 `go-comm` 一致
3. **遵循代码风格**: 命名、错误处理、类型转换都要遵循 `go-comm` 的约定
4. **考虑通用性**: 编写代码时，思考是否应该提取到 `go-comm`
5. **使用 i18n**: 所有用户可见的消息都必须国际化
6. **使用 go-comm 日志**: 不使用标准库的 log 包

## 获取帮助

- 查看 `GO_COMM_USAGE.md` 了解 go-comm 的详细功能和使用方法
- 阅读 `/data1/batoning-workspace/go-comm` 中的源码和测试
- 参考 `go-comm` 中的示例代码（playground 目录）
