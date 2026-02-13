# go-comm 使用规范

## 概述

`go-comm` 是一个项目无关的 Go 通用工具库，位于 `/data1/vibe-deploying-workspace/go-comm`。本文档规范了在 `vibe-deploying` 项目中使用 `go-comm` 的准则和最佳实践。

## 1. go-comm 核心功能

### 1.1 日志系统 (Logger)
- **文件**: `logger.go`
- **功能**:
  - 结构化日志（基于 `phuslu/log`）
  - 支持文件轮转（基于 `lumberjack`）
  - 日志上下文和追踪 ID
  - 子日志器创建
  - 事件日志集成

```go
// 创建日志器
logger := comm.NewLoggerP(console, &comm.LoggerConfigT{
    MaxSize:    100,    // MB
    MaxAge:     30,     // days
    MaxBackups: 10,
    Compress:   true,
}, "/var/log/app.log")

// 创建带追踪ID的上下文
logCtx := comm.NewLogContext(true)
subLogger := logger.NewSubLogger(logCtx)

// 记录日志
logger.Info().Str("key", "value").Msg("message")
logger.Error(err).Msg("error occurred")
```

### 1.2 配置管理 (Config)
- **文件**: `config.go`
- **功能**:
  - YAML/TOML 配置文件解析
  - 环境变量加载（支持 `.env`、shell 脚本）
  - 配置验证和类型转换
  - 严格模式和动态模式
  - 默认值合并

```go
// 解析 YAML 配置
cfg := &MyConfig{}
result, metadata, err := comm.DecodeWithYaml(yamlText, comm.StrictConfigConfig(), cfg, defaults)

// 加载环境变量
vars, err := comm.LoadEnvScripts(fs, map[string]string{}, ".env", "/etc/profile")
```

### 1.3 国际化 (i18n)
- **文件**: `i18n.go`
- **功能**:
  - 多语言支持（zh, en, ru, fr, es, it, de, hu, ko, ja, vi, th, id）
  - 自动语言检测
  - 模板化消息翻译
  - 内置本地化文件

```go
// 翻译消息
msg := comm.T("error.required", map[string]any{
    "Hint": "config",
    "Key":  "name",
})

// 带格式的翻译
msg := comm.Tf("Hello %s", "world")

// 设置语言
comm.SetLanguage("zh")
```

### 1.4 文件操作 (FileOps)
- **文件**: `fileops.go`, `file.go`, `file_cache.go`
- **功能**:
  - 文件复制（支持权限和所有者设置）
  - 文件缓存（支持 HTTP/HTTPS 远程文件）
  - 文件回退机制
  - 本地化错误消息

```go
// 创建文件操作器
ops := comm.NewFileOps(fs)

// 复制文件（支持 chmod 和 chown）
ops.CopyFileWithMode(src, dest, "0644", "user:group", false, "")

// 创建远程文件缓存
cache := comm.NewFileCache(fs, "/tmp/cache", true)
cachedFile, err := cache.AcquireFile("https://example.com/file.zip")
```

### 1.5 Shell 命令执行 (Gosh)
- **文件**: `gosh.go`, `command.go`
- **功能**:
  - Shell 脚本解析和执行（基于 `mvdan.cc/sh`）
  - Sudo 密码输入处理
  - 环境变量管理
  - 命令输出解析

```go
// 执行 shell 命令
output, err := comm.RunGoshCommand(vars, "/tmp", "source /etc/profile && echo $PATH", nil)

// 带 sudo 的命令
output, err := comm.RunGoshCommand(vars, "", "sudo ls /root", passwordInput)
```

### 1.6 插件系统 (Plugin)
- **文件**: `plugin.go`, `base_plugin.go`, `fs_plugin_loader.go`, `external_go_plugin.go`
- **功能**:
  - 插件接口定义
  - 文件系统插件加载器
  - 外部 Go 插件支持
  - 插件清单和注册表
  - 生命周期管理（启动/停止）

```go
// 创建插件加载器
loader := comm.NewFsPluginLoader(fs, "/plugins", "app")

// 启动插件
err := loader.Start(logger)

// 停止插件
err := loader.Stop(logger)
```

### 1.7 类型转换和验证
- **文件**: `string.go`, `int.go`, `bool.go`, `float.go`, `map.go`, `collection.go`
- **功能**:
  - 安全的类型转换（Required/Optional 模式）
  - 本地化错误消息
  - 集合工具函数

```go
// 必需字段
name := comm.RequiredStringP("config", "name", m)

// 可选字段
port, has := comm.OptionalStringP("config", "port", m, "8080")

// Map 转换
opts := comm.RequiredMapP("config", "options", m)
```

### 1.8 反射和工具函数
- **文件**: `reflect.go`, `map.go`, `text.go`, `hash.go`
- **功能**:
  - 泛型类型转换
  - Map 合并和深拷贝
  - 哈希计算
  - 文本处理

```go
// Map 合并
merged := comm.MergeMap(baseMap, overrideMap)

// 深拷贝
copy := comm.DeepCopyMap(original)

// 切片转 Map
m := comm.Slice2Map(items, func(v Item) string {
    return v.ID
})
```

### 1.9 网络和系统信息
- **文件**: `net.go`, `sysinfo.go`, `ssh_config.go`
- **功能**:
  - 网络接口检测
  - 系统信息检测
  - SSH 配置解析

```go
// 获取系统信息
info := comm.NewSystemDetector().DetectSystem()

// 解析 SSH 配置
sshConfig, err := comm.ParseSSHConfig(fs, "/etc/ssh/ssh_config")
```

### 1.10 进度和 IO
- **文件**: `progress.go`, `io.go`
- **功能**:
  - 进度读取器和写入器
  - IO 工具函数

```go
// 创建进度读取器
reader := comm.NewProgressReader(fileReader, size, func(read int, total int) {
    fmt.Printf("Progress: %d/%d\n", read, total)
})
```

## 2. 依赖管理规范

### 2.1 依赖版本同步
`vibe-deploying` 必须与 `go-comm` 保持依赖版本完全一致：

```go
// vibe-deploying/go.mod
require (
    github.com/qiangyt/go-comm/v2 v2.6.11
    // 其他依赖必须与 go-comm/go.mod 中的版本一致
    github.com/BurntSushi/toml v1.6.0
    github.com/phuslu/log v1.0.121
    github.com/pkg/errors v0.9.1
    // ... 其他依赖
)
```

### 2.2 使用本地依赖
开发时使用本地 `go-comm`：

```go
// vibe-deploying/go.mod
replace github.com/qiangyt/go-comm/v2 => ../go-comm
```

### 2.3 依赖检查清单
在添加新依赖时，必须检查 `go-comm/go.mod`：
- 如果 `go-comm` 已有该依赖，使用相同版本
- 如果 `go-comm` 没有该依赖，评估是否应该添加到 `go-comm`

## 3. 代码风格和约定

### 3.1 命名约定
- **类型**: 使用 `T` 后缀（如 `LoggerT`、`ConfigT`）
- **类型别名**: 使用 `type Type = *TypeT`（如 `type Logger = *LoggerT`）
- **函数**: 驼峰命名， Panic 版本使用 `P` 后缀（如 `NewLoggerP`、`RequiredStringP`）
- **接口**: 无 `T` 后缀（如 `Plugin`、`PluginLoader`）

### 3.2 错误处理
- **双函数模式**: 每个可能出错的函数提供两个版本
  - 普通版本：返回 `(result, error)`
  - Panic 版本：以 `P` 结尾，出错时 panic

```go
// 推荐：使用 Panic 版本（简化错误处理）
// vibe-deploying 应当尽可能使用 panic 版本，仅在最外层处理这些 panic
name := comm.RequiredStringP("config", "name", m)

// 不推荐：普通版本（需要处理错误）
// 仅在特殊场景下使用（如需要细粒度错误控制）
name, err := comm.RequiredString("config", "name", m)
if err != nil {
    return err
}
```

### 3.3 类型转换
- **Required 模式**: 字段必需，缺失时返回错误
- **Optional 模式**: 字段可选，返回默认值

```go
// 必需字段
name := comm.RequiredStringP("config", "name", m)

// 可选字段（带默认值）
port, has := comm.OptionalStringP("config", "port", m, "8080")
```

### 3.4 本地化
- 所有用户可见的错误消息必须使用 `i18n`
- 使用 `comm.T()` 或 `comm.LocalizeError()`

```go
// 使用本地化错误
return comm.LocalizeError("error.required", map[string]any{
    "Hint": "config",
    "Key":  "name",
})

// 翻译消息
msg := comm.T("error.file.not_found", map[string]any{
    "Path": filePath,
})
```

### 3.5 日志记录
- 使用结构化日志
- 记录关键操作和错误
- 使用适当的日志级别

```go
// Info: 正常操作
logger.Info().Str("action", "deploy").Msg("Deployment started")

// Error: 错误
logger.Error(err).Str("file", path).Msg("Failed to read file")

// Debug: 调试信息
logger.Debug().Int("count", len(items)).Msg("Processing items")
```

## 4. 代码复用原则

### 4.1 复用前检查
在编写新代码前，必须检查 `go-comm` 是否已提供相同功能：

1. **搜索 go-comm**: 在 `go-comm` 中搜索相关功能
2. **查看文档**: 阅读相关源文件和注释
3. **评估适配性**: 判断是否可直接复用

### 4.2 功能适配策略
当 `go-comm` 中的功能与需求略有偏差时：

1. **直接复用**: 如果偏差很小，直接使用 `go-comm` 功能
2. **增强 go-comm**: 如果改进具有通用性，更新 `go-comm` 后复用
   - 在 `go-comm` 中添加新功能或参数
   - 确保 `go-comm` 的向后兼容性
   - 更新 `go-comm` 的测试和文档
3. **项目特定实现**: 仅当功能特定于当前项目时，在项目内实现

### 4.3 可重用代码提取
当编写新代码时，如果发现该功能可能被其他项目使用：

1. **评估通用性**: 判断功能是否具有通用价值
2. **提取到 go-comm**:
   - 在 `go-comm` 中添加新功能
   - 遵循 `go-comm` 的代码风格和约定
   - 添加完整的测试
   - 更新文档
3. **在项目中引用**: 从 `go-comm` 导入并使用

### 4.4 复用示例

#### 示例 1: 日志系统
```go
// ✅ 正确: 使用 go-comm 的日志系统
import "github.com/qiangyt/go-comm/v2"

logger := comm.NewLoggerP(nil, nil, "/var/log/app.log")
logger.Info().Msg("Application started")

// ❌ 错误: 重新实现日志系统
log.Println("Application started")  // 不要使用标准 log 包
```

#### 示例 2: 配置管理
```go
// ✅ 正确: 使用 go-comm 的配置解析
import "github.com/qiangyt/go-comm/v2"

cfg := &Config{}
result, _, err := comm.DecodeWithYaml(yamlText, comm.StrictConfigConfig(), cfg, defaults)

// ❌ 错误: 使用其他 YAML 库
import "gopkg.in/yaml.v3"
err := yaml.Unmarshal(data, &cfg)  // 不要直接使用
```

#### 示例 3: 类型转换
```go
// ✅ 正确: 使用 go-comm 的类型转换
import "github.com/qiangyt/go-comm/v2"

name := comm.RequiredStringP("config", "name", m)
port, has := comm.OptionalStringP("config", "port", m, "8080")

// ❌ 错误: 手动类型断言
name, ok := m["name"].(string)
if !ok {
    return fmt.Errorf("name is required")
}
```

## 5. 开发工作流

### 5.1 功能开发流程
1. **检查 go-comm**: 在 `go-comm` 中查找现有功能
2. **评估复用**: 决定是复用、增强还是新建
3. **实现功能**:
   - 如果复用：直接从 `go-comm` 导入
   - 如果增强：先更新 `go-comm`，再导入
   - 如果新建：在项目中实现，考虑未来提取
4. **测试**: 确保功能正常工作
5. **文档**: 更新相关文档

### 5.2 依赖更新流程
当需要更新依赖时：

1. **检查 go-comm**: 查看 `go-comm/go.mod` 中的依赖版本
2. **同步版本**: 确保 `vibe-deploying/go.mod` 使用相同版本
3. **测试**: 运行测试确保兼容性
4. **更新 go-comm**: 如果需要新版本，先更新 `go-comm`

### 5.3 go-comm 更新流程
当需要向 `go-comm` 添加功能时：

1. **设计 API**: 遵循 `go-comm` 的代码风格和约定
2. **TDD 开发流程**:
   - 先编写测试用例（覆盖所有场景）
   - 编写生产代码使测试通过
   - 迭代生产代码直至所有测试都通过
   - 根据 coverage 报告补充测试用例
   - **要求**: 100% 行覆盖率和 100% 分支覆盖率
3. **更新版本**: 按照语义化版本更新版本号
4. **同步到项目**: 在 `vibe-deploying` 中使用新版本

## 6. 检查清单

在代码审查时，使用此清单确保符合规范：

### 6.1 功能复用检查
- [ ] 检查 `go-comm` 是否已提供相同功能
- [ ] 如果功能略有偏差，考虑更新 `go-comm`
- [ ] 如果新功能具有通用性，提取到 `go-comm`

### 6.2 依赖一致性检查
- [ ] 所有依赖版本与 `go-comm/go.mod` 一致
- [ ] 新依赖已评估是否应该添加到 `go-comm`

### 6.3 代码风格检查
- [ ] 使用 `go-comm` 的命名约定
- [ ] 遵循双函数模式（普通 + Panic 版本）
- [ ] 使用 Required/Optional 模式进行类型转换
- [ ] 用户可见消息使用 i18n

### 6.4 日志和错误处理检查
- [ ] 使用 `go-comm` 的日志系统
- [ ] 错误消息本地化
- [ ] 使用适当的日志级别

## 7. 常见问题和最佳实践

### Q1: 如何判断功能是否应该提取到 go-comm？
**A**: 如果满足以下任一条件，应该提取：
- 功能在多个项目中使用
- 功能是通用工具或工具函数
- 功能不依赖特定业务逻辑

### Q2: go-comm 功能不够用怎么办？
**A**:
1. 检查是否可以通过参数或配置支持
2. 考虑增强 `go-comm` 的功能
3. 在项目中扩展，但保持可提取性

### Q3: 如何处理 go-comm 的 bug？
**A**:
1. 在 `go-comm` 中修复 bug
2. 添加回归测试
3. 更新版本号
4. 在项目中使用修复后的版本

### Q4: 可以使用其他库的相同功能吗？
**A**:
- 如果 `go-comm` 已提供该功能，必须使用 `go-comm`
- 如果 `go-comm` 未提供，评估是否应该添加到 `go-comm`
- 仅在功能特定于项目时使用其他库

## 8. 参考资料

### 8.1 go-comm 项目信息
- **路径**: `/data1/vibe-deploying-workspace/go-comm`
- **模块**: `github.com/qiangyt/go-comm/v2`
- **当前版本**: v2.6.11

### 8.2 核心文件清单
- `logger.go` - 日志系统
- `config.go` - 配置管理
- `i18n.go` - 国际化
- `fileops.go` - 文件操作
- `gosh.go` - Shell 命令执行
- `plugin.go` - 插件系统
- `string.go`, `int.go`, `bool.go`, `map.go` - 类型转换
- `collection.go` - 集合工具
- `reflect.go` - 反射工具
- `net.go` - 网络工具
- `sysinfo.go` - 系统信息

### 8.3 相关文档
- `go-comm/README.md` - 项目概述
- `GO_COMM_USAGE.md` - 本文档
- `CLAUDE.md` - Claude AI 开发指南

---

## 版本历史

- v1.0.0 (2025-02-03) - 初始版本
