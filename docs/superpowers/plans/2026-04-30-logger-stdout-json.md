# Logger stdout JSON 输出实现计划

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 支持通过配置项控制 stdout 输出格式（json/console/none）

**Architecture:** 修改 `NewLogger` 函数，将硬编码的 debug 模式判断替换为读取 `stdout_encode` 配置项，根据配置值选择编码器并添加 stdout core。

**Tech Stack:** Go, zap, zapcore

---

## 文件结构

| 文件 | 改动类型 | 职责 |
|------|----------|------|
| `logger/logger.go` | Modify | 修改 NewLogger 函数，添加 stdout_encode 配置读取逻辑 |

---

### Task 1: 修改 NewLogger 函数支持 stdout_encode 配置

**Files:**
- Modify: `logger/logger.go:105-117`

- [ ] **Step 1: 读取 stdout_encode 配置值**

在 `NewLogger` 函数中，文件日志 core 创建后、添加 stdout 判断前，添加读取 `stdout_encode` 配置的代码。

找到现有代码（约第 104-117 行）：

```go
// debug 日志输出至日志文件、标准输出
if config.GetString("app.mode") == "debug" {
    cores = append(cores, func() zapcore.Core {
        consoleWriter, closeOut, err := zap.Open("stdout")
        if err != nil {
            if closeOut != nil {
                closeOut()
            }
            panic(err)
        }
        encoder.EncodeLevel = zapcore.CapitalColorLevelEncoder
        return zapcore.NewCore(zapcore.NewConsoleEncoder(encoder), zap.CombineWriteSyncers(consoleWriter), logLevel)
    }())
}
```

替换为：

```go
// stdout 输出根据配置决定
stdoutEncode := getStringFromMap(logSection, "stdout_encode")
if stdoutEncode == "json" || stdoutEncode == "console" {
    stdoutWriter, closeOut, err := zap.Open("stdout")
    if err != nil {
        if closeOut != nil {
            closeOut()
        }
        panic(err)
    }

    stdoutEncoder := GetEncoder()
    if stdoutEncode == "console" {
        stdoutEncoder.EncodeLevel = zapcore.CapitalColorLevelEncoder
        cores = append(cores, zapcore.NewCore(zapcore.NewConsoleEncoder(stdoutEncoder), zap.CombineWriteSyncers(stdoutWriter), logLevel))
    } else {
        cores = append(cores, zapcore.NewCore(zapcore.NewJSONEncoder(stdoutEncoder), zap.CombineWriteSyncers(stdoutWriter), logLevel))
    }
}
```

注意：使用新的 `stdoutEncoder` 变量而非修改共享的 `encoder`，避免影响文件日志编码器。

- [ ] **Step 2: 运行现有代码确保无编译错误**

Run: `go build ./...`
Expected: 编译成功，无错误

- [ ] **Step 3: 提交改动**

```bash
git add logger/logger.go
git commit -m "$(cat <<'EOF'
feat: support stdout_encode config for logger stdout output

Add stdout_encode config option (json/console/none) to control
stdout log output format. Remove hardcoded debug mode check.

Co-Authored-By: Claude Opus 4.7 <noreply@anthropic.com>
EOF
)"
```

---

### Task 2: 更新 README 文档

**Files:**
- Modify: `README.md:11-16`

- [ ] **Step 1: 在 README 的 log 配置示例中添加 stdout_encode**

找到 README 第 11-16 行的 log 配置部分：

```yaml
log:
  encode_type: json
  path: ./logs
  maxsize: 1024
  compress: true
```

修改为：

```yaml
log:
  encode_type: json
  stdout_encode: json  # json/console/none, 默认 none
  path: ./logs
  maxsize: 1024
  compress: true
```

- [ ] **Step 2: 提交文档更新**

```bash
git add README.md
git commit -m "$(cat <<'EOF'
docs: add stdout_encode config to README

Co-Authored-By: Claude Opus 4.7 <noreply@anthropic.com>
EOF
)"
``

---

## 自检清单

- [x] Spec coverage: 配置项读取、json/console/none 三种格式、默认 none
- [x] Placeholder scan: 无 TBD/TODO，代码完整
- [x] Type consistency: encoder 变量命名一致