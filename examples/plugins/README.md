# AUR更新检查器插件系统

## 概述

AUR更新检查器支持插件化架构，允许您开发自定义的检查器插件，而无需重新编译整个应用程序。插件系统使用Go的plugin包实现动态加载功能。

## 插件开发指南

### 1. 创建插件

要创建一个新的检查器插件，您需要实现`PluginChecker`接口，该接口扩展了`UpstreamChecker`接口并添加了`PluginInfo`方法。

#### 必须实现的接口方法

```go
type PluginChecker interface {
    // UpstreamChecker 嵌入上游检查器接口
    UpstreamChecker

    // PluginInfo 返回插件信息
    PluginInfo() PluginInfo
}
```

#### 必须导出的函数

每个插件必须导出一个名为`NewPluginChecker`的函数，该函数返回`PluginChecker`接口的实例：

```go
func NewPluginChecker() PluginChecker {
    return NewYourChecker()
}
```

### 2. 插件示例

请参考`gitlab_checker_example`目录下的示例插件，它展示了一个完整的插件实现：

- 实现了所有必要的方法
- 包含版本检查逻辑
- 提供了插件信息

### 3. 构建插件

插件需要构建为Go插件（.so文件）。使用以下命令构建插件：

```bash
go build -buildmode=plugin -o your_plugin.so your_plugin.go
```

### 4. 加载插件

使用插件管理器加载插件：

```go
// 获取插件管理器
pluginManager := checkers.GetPluginManager()

// 加载插件
checker, err := pluginManager.LoadPlugin("default", "path/to/your_plugin.so")
if err != nil {
    // 处理错误
}
```

## 插件API参考

### PluginInfo 结构体

```go
type PluginInfo struct {
    Name        string // 插件名称
    Version     string // 插件版本
    Author      string // 插件作者
    Description string // 插件描述
}
```

### UpstreamChecker 接口

```go
type UpstreamChecker interface {
    // Check 检查上游版本
    Check(ctx context.Context, url, versionExtractKey string) (string, error)

    // CheckWithOption 根据选项检查上游版本
    CheckWithOption(ctx context.Context, url, versionExtractKey string, checkTestVersion int) (string, error)

    // Name 返回检查器名称
    Name() string

    // Supports 检查此检查器是否支持给定的URL
    Supports(url string) bool

    // Priority 返回检查器的优先级
    Priority() int
}
```

### BaseChecker 基础检查器

您可以通过嵌入`BaseChecker`来简化插件开发，它提供了一些默认实现：

```go
type BaseChecker struct {
    name          string
    versionParser *UpstreamVersionParser
}
```

## 注意事项

1. **插件兼容性**：插件必须使用与主应用程序相同版本的Go编译器编译，否则可能无法加载。

2. **依赖管理**：插件不能依赖主应用程序未使用的包。如果插件需要额外的依赖，这些依赖必须包含在插件本身中。

3. **错误处理**：插件应该妥善处理错误，并通过返回错误值而不是panic来报告问题。

4. **资源管理**：插件应该负责管理自己的资源，包括文件句柄、网络连接等。

5. **安全性**：由于插件是动态加载的，请确保插件来源可信，避免加载恶意插件。

## 故障排除

### 常见错误

1. **插件加载失败**
   - 检查插件是否使用正确的构建模式编译
   - 确保插件和主应用程序使用相同版本的Go编译器
   - 检查插件路径是否正确

2. **符号未找到**
   - 确保插件导出了`NewPluginChecker`函数
   - 检查函数签名是否正确

3. **类型断言失败**
   - 确保插件返回的类型实现了`PluginChecker`接口
   - 检查所有必需的方法是否都已实现

## 贡献指南

如果您开发了一个有用的检查器插件，并希望将其贡献给项目，请：

1. 确保您的代码遵循项目的编码规范
2. 为您的代码添加适当的测试
3. 提交一个Pull Request，并附上详细的说明

我们欢迎社区贡献，并感谢您的参与！
