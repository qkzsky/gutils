# Logger stdout JSON 输出设计

## 目标

支持在 stdout 输出 JSON 格式日志，通过配置项控制输出格式。

## 配置项

在 `log` 配置节下新增 `stdout_encode` 字段：

| 字段 | 类型 | 可选值 | 默认值 |
|------|------|--------|--------|
| `stdout_encode` | string | `json`, `console`, `none` | `none` |

- `json`：输出 JSON 格式到 stdout
- `console`：输出彩色文本格式到 stdout（保留当前 debug 模式的行为）
- `none`：不输出到 stdout

配置示例：

```yaml
log:
  encode_type: json
  stdout_encode: json
  maxsize: 1024
  compress: true
```

## 代码改动

修改 `logger/logger.go` 中的 `NewLogger` 函数：

1. 移除当前的 `app.mode == "debug"` 硬编码判断
2. 读取 `log.stdout_encode` 配置值
3. 根据配置值选择编码器：
   - `json`：使用 `zapcore.NewJSONEncoder(encoder)`
   - `console`：使用 `zapcore.NewConsoleEncoder(encoder)`，设置 `CapitalColorLevelEncoder`
   - `none` 或无效值：不添加 stdout core
4. stdout core 与文件输出 core 合并（使用 `zapcore.NewTee`）

## 边界情况

- 配置值无效时静默忽略，使用 `none`（不输出到 stdout）
- stdout 输出不影响文件日志输出
- stdout 输出的日志级别与文件输出保持一致（使用同一个 `logLevel`）