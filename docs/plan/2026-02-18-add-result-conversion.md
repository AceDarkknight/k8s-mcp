# 实现计划：MCP 结果转换功能添加

## 1. 任务背景
目前 `pkg/mcpclient` 中的 `CallTool` 方法返回原生的 `*mcp.CallToolResult`。为了方便上层逻辑直接使用结构化数据，需要提供一种将 `CallToolResult` 转换为 `internal/k8s/types.go` 中定义的结构体的方法。

## 2. 设计目标
- 支持泛型转换，能够将结果解码为任意指定的结构体。
- 自动处理 MCP 结果中的 `Content` 字段（通常为 JSON）。
- 遵循“代码改动最小”原则，复用现有库。

## 3. 实现步骤

### 步骤 1：分析 `mcp.CallToolResult`
`mcp.CallToolResult` 包含：
- `Content`: `[]mcp.Content`（通常是 `TextContent` 或 `ImageContent`）。
- `IsError`: `bool`，标识调用是否失败。

我们需要提取 `TextContent` 中的字符串，并尝试将其反序列化为目标类型。

### 步骤 2：添加转换辅助函数
在 `pkg/mcpclient/tools.go` 中添加辅助函数：

```go
// DecodeResult 将 MCP 工具调用结果解码为指定的结构体
func DecodeResult[T any](result *mcp.CallToolResult) (*T, error) {
    if result.IsError {
        return nil, fmt.Errorf("tool call returned error")
    }
    
    // 逻辑：遍历 Content，寻找 JSON 内容并解码
    // ...
}
```

### 步骤 3：实现解码逻辑
1. 检查 `result.IsError`。
2. 遍历 `result.Content`。
3. 如果内容类型是 `mcp.TextContent`，尝试将其 `Text` 字段作为 JSON 解码。
4. 返回解码后的对象或错误。

### 步骤 4：单元测试验证
- 在 `pkg/mcpclient/tools_test.go` (如果存在) 或新创建测试文件中，编写测试用例。
- 模拟包含 JSON 字符串的 `mcp.CallToolResult`。
- 验证是否能正确转换为 `k8s.Namespace` 等结构体。

## 4. 预期效果
- 开发者可以通过 `mcpclient.DecodeResult[k8s.Pod](result)` 轻松获取结构化对象。
- 减少上层业务代码中手写 JSON 解码的冗余。

## 5. 最小化改动说明
- 仅在 `pkg/mcpclient/tools.go` 中新增一个辅助函数。
- 不修改原有的 `CallTool` 方法签名，保持向后兼容。
