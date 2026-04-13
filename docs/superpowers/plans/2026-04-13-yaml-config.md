# YAML 配置与环境变量支持实现计划

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 将配置系统从 INI 格式迁移到 YAML 格式，支持 .env 环境变量和 `${ENV_VAR:default}` 语法。

**Architecture:** 使用 godotenv 加载 .env 文件，yaml.v3 解析配置，os.Expand 替换环境变量。保持现有 API 兼容（Section/Key），新增路径式访问方法。

**Tech Stack:** Go 1.25, gopkg.in/yaml.v3, github.com/joho/godotenv v1.5.1

---

## 文件结构

| 文件 | 操作 | 说明 |
|------|------|------|
| `go.mod` | 修改 | 升级 yaml v2 → v3，新增 godotenv |
| `config/yaml.go` | 新建 | YAML 配置解析核心 |
| `config/yaml_test.go` | 新建 | 配置解析测试 |
| `config/testdata/config.yaml` | 新建 | 测试用配置文件 |
| `config/testdata/.env` | 新建 | 测试用环境变量文件 |
| `config/ini.go` | 删除 | 移除 INI 支持 |
| `database/database.go` | 修改 | 适配 YAML 嵌套结构 |
| `redis/redis.go` | 修改 | 适配 YAML 嵌套结构 |
| `logger/logger.go` | 修改 | 适配 YAML 嵌套结构 |
| `README.md` | 修改 | 更新配置示例 |

**设计说明：辅助函数重复定义**
- `getStringFromMap`、`getIntFromMapWithDefault`、`getBoolFromMapWithDefault` 在 database、redis、logger 模块中各自定义
- 这些是私有函数（不导出），每个模块只需要自己使用
- Go 不支持跨包私有函数共享，独立定义避免创建额外的 shared 包

---

### Task 1: 更新依赖

**Files:**
- Modify: `go.mod`

- [ ] **Step 1: 更新 go.mod 依赖**

将 yaml.v2 升级为 v3，添加 godotenv：

```go
require (
    github.com/coocood/freecache v1.2.5
    github.com/joho/godotenv v1.5.1
    github.com/redis/go-redis/v9 v9.18.0
    go.uber.org/zap v1.27.1
    gopkg.in/natefinch/lumberjack.v2 v2.2.1
    gopkg.in/yaml.v3 v3.0.1
    gorm.io/driver/mysql v1.6.0
    gorm.io/driver/postgres v1.6.0
    gorm.io/gorm v1.31.1
    gorm.io/plugin/dbresolver v1.6.2
)
```

删除 `gopkg.in/ini.v1` 和 `gopkg.in/yaml.v2` 行。

- [ ] **Step 2: 运行 go mod tidy**

```bash
go mod tidy
```

- [ ] **Step 3: 提交依赖变更**

```bash
git add go.mod go.sum
git commit -m "deps: upgrade yaml v3, add godotenv, remove ini.v1"
```

---

### Task 2: 创建 YAML 配置核心模块

**Files:**
- Create: `config/yaml.go`
- Create: `config/yaml_test.go`
- Create: `config/testdata/config.yaml`
- Create: `config/testdata/.env`

- [ ] **Step 1: 创建测试配置文件**

`config/testdata/config.yaml`:
```yaml
app:
  name: ${APP_NAME:test-app}
  mode: ${APP_MODE:release}
  addr: :8080

log:
  encode_type: json
  path: ./logs
  maxsize: 1024
  compress: true

gorm:
  trace_sql: false
  slow_threshold: 1s
  prepare_stmt: true

database:
  test:
    master:
      drive: mysql
      host: ${DB_HOST:127.0.0.1}
      port: ${DB_PORT:3306}
      username: ${DB_USER:root}
      password: ${DB_PASS}
      db: test
      charset: utf8
      max_idle: 10
      max_open: 20
    slaves:
      - drive: mysql
        host: 127.0.0.1
        port: 3306
        username: root
        password: ""
        db: test2
        charset: utf8

redis:
  default:
    host: ${REDIS_HOST:127.0.0.1}
    port: ${REDIS_PORT:6379}
    auth: ${REDIS_AUTH}
    db: 0
    max_idle: 5
    max_open: 10
```

`config/testdata/.env`:
```
APP_NAME=my-app
DB_HOST=192.168.1.100
REDIS_PORT=6380
```

- [ ] **Step 2: 编写失败测试**

`config/yaml_test.go`:
```go
package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestSetDefault(t *testing.T) {
	testdataDir := filepath.Join(".", "testdata")
	configFile := filepath.Join(testdataDir, "config.yaml")

	// 清理环境变量，确保测试独立
	os.Clearenv()

	SetDefault(configFile)

	if AppName != "my-app" {
		t.Errorf("AppName expected 'my-app', got '%s'", AppName)
	}

	if AppMode != "release" {
		t.Errorf("AppMode expected 'release', got '%s'", AppMode)
	}
}

func TestGetString(t *testing.T) {
	testdataDir := filepath.Join(".", "testdata")
	configFile := filepath.Join(testdataDir, "config.yaml")

	os.Clearenv()
	SetDefault(configFile)

	// 测试环境变量替换
	host := GetString("database.test.master.host")
	if host != "192.168.1.100" {
		t.Errorf("host expected '192.168.1.100', got '%s'", host)
	}

	// 测试默认值
	port := GetString("database.test.master.port")
	if port != "3306" {
		t.Errorf("port expected '3306', got '%s'", port)
	}
}

func TestGetStringWithDefault(t *testing.T) {
	testdataDir := filepath.Join(".", "testdata")
	configFile := filepath.Join(testdataDir, "config.yaml")

	os.Clearenv()
	SetDefault(configFile)

	val := GetStringWithDefault("nonexistent.key", "fallback")
	if val != "fallback" {
		t.Errorf("expected 'fallback', got '%s'", val)
	}
}

func TestGetInt(t *testing.T) {
	testdataDir := filepath.Join(".", "testdata")
	configFile := filepath.Join(testdataDir, "config.yaml")

	os.Clearenv()
	SetDefault(configFile)

	db := GetInt("redis.default.db")
	if db != 0 {
		t.Errorf("db expected 0, got %d", db)
	}

	maxsize := GetInt("log.maxsize")
	if maxsize != 1024 {
		t.Errorf("maxsize expected 1024, got %d", maxsize)
	}
}

func TestGetBool(t *testing.T) {
	testdataDir := filepath.Join(".", "testdata")
	configFile := filepath.Join(testdataDir, "config.yaml")

	os.Clearenv()
	SetDefault(configFile)

	compress := GetBool("log.compress")
	if !compress {
		t.Errorf("compress expected true, got %v", compress)
	}

	traceSQL := GetBool("gorm.trace_sql")
	if traceSQL {
		t.Errorf("trace_sql expected false, got %v", traceSQL)
	}
}

func TestGetStringMap(t *testing.T) {
	testdataDir := filepath.Join(".", "testdata")
	configFile := filepath.Join(testdataDir, "config.yaml")

	os.Clearenv()
	SetDefault(configFile)

	appMap := GetStringMap("app")
	if appMap["name"] != "my-app" {
		t.Errorf("app.name expected 'my-app', got '%s'", appMap["name"])
	}
}

func TestGetSlice(t *testing.T) {
	testdataDir := filepath.Join(".", "testdata")
	configFile := filepath.Join(testdataDir, "config.yaml")

	os.Clearenv()
	SetDefault(configFile)

	slaves := GetSlice("database.test.slaves")
	if len(slaves) != 1 {
		t.Errorf("slaves expected 1 element, got %d", len(slaves))
	}
}

func TestSection(t *testing.T) {
	testdataDir := filepath.Join(".", "testdata")
	configFile := filepath.Join(testdataDir, "config.yaml")

	os.Clearenv()
	SetDefault(configFile)

	appSection := Section("app")
	if appSection["name"] != "my-app" {
		t.Errorf("section app.name expected 'my-app', got '%v'", appSection["name"])
	}
}
```

- [ ] **Step 3: 运行测试验证失败**

```bash
cd config && go test -v
```

预期：测试失败，`SetDefault` 等函数未定义。

- [ ] **Step 4: 实现 yaml.go 核心功能**

`config/yaml.go`:
```go
package config

import (
	"fmt"
	"github.com/joho/godotenv"
	"gopkg.in/yaml.v3"
	"os"
	"path/filepath"
	"strings"
)

var (
	AppPath string
	AppName string
	AppMode string

	defaultConf map[string]interface{}
)

// SetDefault 初始化配置
func SetDefault(file string) {
	var err error
	if AppPath, err = filepath.Abs(filepath.Dir(os.Args[0])); err != nil {
		panic(err)
	}

	// 加载 .env 文件（忽略不存在错误）
	_ = godotenv.Load()

	// 加载配置文件
	defaultConf = NewConfig(file)
	AppName = GetStringWithDefault("app.name", "app")
	AppMode = GetStringWithDefault("app.mode", "release")
}

// NewConfig 加载 YAML 配置文件并替换环境变量
func NewConfig(configFile string) map[string]interface{} {
	content, err := os.ReadFile(configFile)
	if err != nil {
		panic(err)
	}

	// 替换环境变量
	expanded := expandEnv(string(content))

	// 解析 YAML
	var conf map[string]interface{}
	if err := yaml.Unmarshal([]byte(expanded), &conf); err != nil {
		panic(err)
	}

	return conf
}

// expandEnv 替换 ${ENV_VAR} 和 ${ENV_VAR:default} 格式的环境变量
func expandEnv(content string) string {
	return os.Expand(content, func(key string) string {
		// 解析 KEY:default 格式
		parts := strings.SplitN(key, ":", 2)
		envKey := parts[0]

		val := os.Getenv(envKey)
		if val != "" {
			return val
		}

		// 返回默认值（如果有）
		if len(parts) == 2 {
			return parts[1]
		}
		return ""
	})
}

// Section 获取指定节点的配置（兼容旧 API）
func Section(name string) map[string]interface{} {
	if defaultConf == nil {
		panic("default conf not found.")
	}

	val, ok := defaultConf[name]
	if !ok {
		return map[string]interface{}{}
	}

	section, ok := val.(map[string]interface{})
	if !ok {
		return map[string]interface{}{}
	}

	return section
}

// Key 获取 app section 下的 key（兼容旧 API）
func Key(name string) string {
	appSection := Section("app")
	val, ok := appSection[name]
	if !ok {
		return ""
	}
	return fmt.Sprintf("%v", val)
}

// getValueByPath 通过路径获取配置值
func getValueByPath(path string) (interface{}, bool) {
	parts := strings.Split(path, ".")
	var current interface{} = defaultConf

	for _, part := range parts {
		switch v := current.(type) {
		case map[string]interface{}:
			if next, ok := v[part]; ok {
				current = next
			} else {
				return nil, false
			}
		case []interface{}:
			// 处理数组索引
			idx := 0
			if _, err := fmt.Sscanf(part, "%d", &idx); err == nil && idx < len(v) {
				current = v[idx]
			} else {
				return nil, false
			}
		default:
			return nil, false
		}
	}

	return current, true
}

// GetString 通过路径获取字符串配置值
func GetString(path string) string {
	val, ok := getValueByPath(path)
	if !ok {
		return ""
	}
	return fmt.Sprintf("%v", val)
}

// GetStringWithDefault 通过路径获取字符串配置值，带默认值
func GetStringWithDefault(path string, defaultValue string) string {
	val, ok := getValueByPath(path)
	if !ok {
		return defaultValue
	}
	return fmt.Sprintf("%v", val)
}

// GetInt 通过路径获取整数配置值
func GetInt(path string) int {
	val, ok := getValueByPath(path)
	if !ok {
		return 0
	}

	switch v := val.(type) {
	case int:
		return v
	case int64:
		return int(v)
	case float64:
		return int(v)
	default:
		return 0
	}
}

// GetIntWithDefault 通过路径获取整数配置值，带默认值
func GetIntWithDefault(path string, defaultValue int) int {
	val, ok := getValueByPath(path)
	if !ok {
		return defaultValue
	}

	switch v := val.(type) {
	case int:
		return v
	case int64:
		return int(v)
	case float64:
		return int(v)
	default:
		return defaultValue
	}
}

// GetBool 通过路径获取布尔配置值
func GetBool(path string) bool {
	val, ok := getValueByPath(path)
	if !ok {
		return false
	}

	switch v := val.(type) {
	case bool:
		return v
	default:
		return false
	}
}

// GetStringMap 通过路径获取 map 类型配置
func GetStringMap(path string) map[string]interface{} {
	val, ok := getValueByPath(path)
	if !ok {
		return map[string]interface{}{}
	}

	if m, ok := val.(map[string]interface{}); ok {
		return m
	}
	return map[string]interface{}{}
}

// GetSlice 通过路径获取数组类型配置
func GetSlice(path string) []interface{} {
	val, ok := getValueByPath(path)
	if !ok {
		return []interface{}{}
	}

	if s, ok := val.([]interface{}); ok {
		return s
	}
	return []interface{}{}
}
```

- [ ] **Step 5: 运行测试验证通过**

```bash
cd config && go test -v
```

预期：所有测试通过。

- [ ] **Step 6: 提交配置模块**

```bash
git add config/yaml.go config/yaml_test.go config/testdata/
git commit -m "feat: implement YAML config with env var support"
```

---

### Task 3: 删除 INI 配置模块

**Files:**
- Delete: `config/ini.go`

- [ ] **Step 1: 删除 ini.go**

```bash
rm config/ini.go
```

- [ ] **Step 2: 验证编译**

```bash
go build ./...
```

预期：编译成功，无错误。

- [ ] **Step 3: 提交删除**

```bash
git add -A
git commit -m "remove: delete INI config support"
```

---

### Task 4: 适配 database 模块

**Files:**
- Modify: `database/database.go`

- [ ] **Step 1: 修改 database/database.go 导入**

将导入从：
```go
"github.com/qkzsky/gutils/config"
```
保持不变，但需要适配新的配置读取方式。

- [ ] **Step 2: 重写 InitDb 函数**

修改 `database/database.go` 第 59-100 行的 `InitDb` 函数：

```go
func InitDb() {
	mapConf := map[string]dbConfigList{}

	// 获取 database 配置
	dbConfigMap := config.GetStringMap("database")

	for dbName, dbConf := range dbConfigMap {
		confMap, ok := dbConf.(map[string]interface{})
		if !ok {
			continue
		}

		// 解析 master
		if masterConf, ok := confMap["master"].(map[string]interface{}); ok {
			c := parseDbConfig(masterConf, true)
			mapConf[dbName] = append(mapConf[dbName], c)
		}

		// 解析 slaves
		if slavesConf, ok := confMap["slaves"].([]interface{}); ok {
			for _, slave := range slavesConf {
				if slaveMap, ok := slave.(map[string]interface{}); ok {
					c := parseDbConfig(slaveMap, false)
					mapConf[dbName] = append(mapConf[dbName], c)
				}
			}
		}

		// 处理没有 master/slave 结构的简单配置
		if _, hasMaster := confMap["master"]; !hasMaster {
			c := parseDbConfig(confMap, true)
			mapConf[dbName] = append(mapConf[dbName], c)
		}
	}

	for dbName := range mapConf {
		sort.Sort(mapConf[dbName])
		db, err := makeDB(mapConf[dbName])
		if err != nil {
			panic(fmt.Sprintf("db init failed. name: %s, error: %s.", dbName, err.Error()))
		}

		dbMap[dbName] = db
	}
}

// parseDbConfig 从 map 解析 dbConfig
func parseDbConfig(conf map[string]interface{}, isMaster bool) *dbConfig {
	return &dbConfig{
		Drive:    getStringFromMap(conf, "drive"),
		Host:     getStringFromMap(conf, "host"),
		Port:     getStringFromMap(conf, "port"),
		Username: getStringFromMap(conf, "username"),
		Password: getStringFromMap(conf, "password"),
		DBName:   getStringFromMap(conf, "db"),
		SSlMode:  getStringFromMapWithDefault(conf, "ssl_mode", DefaultSSLMode),
		Charset:  getStringFromMapWithDefault(conf, "charset", DefaultCharset),
		MaxIdle:  getIntFromMapWithDefault(conf, "max_idle", defaultMaxIdle),
		MaxOpen:  getIntFromMapWithDefault(conf, "max_open", defaultMaxOpen),
		isMaster: isMaster,
	}
}

// 辅助函数：从 map 获取字符串
func getStringFromMap(m map[string]interface{}, key string) string {
	if val, ok := m[key]; ok {
		return fmt.Sprintf("%v", val)
	}
	return ""
}

// 辅助函数：从 map 获取字符串（带默认值）
func getStringFromMapWithDefault(m map[string]interface{}, key string, defaultVal string) string {
	if val, ok := m[key]; ok {
		return fmt.Sprintf("%v", val)
	}
	return defaultVal
}

// 辅助函数：从 map 获取整数（带默认值）
func getIntFromMapWithDefault(m map[string]interface{}, key string, defaultVal int) int {
	if val, ok := m[key]; ok {
		switch v := val.(type) {
		case int:
			return v
		case int64:
			return int(v)
		case float64:
			return int(v)
		}
	}
	return defaultVal
}
```

- [ ] **Step 3: 修改 makeDB 函数中的 gorm 配置读取**

修改 `database/database.go` 第 103-112 行：

```go
func makeDB(cs dbConfigList) (DB *gorm.DB, err error) {
	gormSC := config.GetStringMap("gorm")
	var gormConfig = &gorm.Config{
		SkipDefaultTransaction: true,
		PrepareStmt:            getBoolFromMapWithDefault(gormSC, "prepare_stmt", true),
		Logger: &gLogger{
			Logger:        logger.GetDefaultLogger().WithOptions(zap.AddCallerSkip(1)),
			TraceSQL:      getBoolFromMapWithDefault(gormSC, "trace_sql", false),
			SlowThreshold: getDurationFromMapWithDefault(gormSC, "slow_threshold", 1*time.Second),
		},
	}
	// ... 后续代码保持不变
}

// 辅助函数：从 map 获取布尔值（带默认值）
func getBoolFromMapWithDefault(m map[string]interface{}, key string, defaultVal bool) bool {
	if val, ok := m[key]; ok {
		if b, ok := val.(bool); ok {
			return b
		}
	}
	return defaultVal
}

// 辅助函数：从 map 获取 Duration（带默认值）
func getDurationFromMapWithDefault(m map[string]interface{}, key string, defaultVal time.Duration) time.Duration {
	if val, ok := m[key]; ok {
		if s, ok := val.(string); ok {
			d, err := time.ParseDuration(s)
			if err == nil {
				return d
			}
		}
		if f, ok := val.(float64); ok {
			return time.Duration(f) * time.Second
		}
	}
	return defaultVal
}
```

注意：这些辅助函数应放在文件末尾，与之前添加的辅助函数合并。

- [ ] **Step 4: 验证编译**

```bash
go build ./database/...
```

预期：编译成功。

- [ ] **Step 5: 提交 database 适配**

```bash
git add database/database.go
git commit -m "refactor: adapt database to YAML config structure"
```

---

### Task 5: 适配 redis 模块

**Files:**
- Modify: `redis/redis.go`

- [ ] **Step 1: 重写 InitRedis 函数**

修改 `redis/redis.go` 第 48-67 行：

```go
func InitRedis() {
	redisConfigMap := config.GetStringMap("redis")

	for name, redisConf := range redisConfigMap {
		confMap, ok := redisConf.(map[string]interface{})
		if !ok {
			continue
		}

		mu.Lock()
		redisMap[name], err = NewRedis(redisConfig{
			Host:    getStringFromMap(confMap, "host"),
			Port:    getStringFromMap(confMap, "port"),
			Auth:    getStringFromMap(confMap, "auth"),
			DB:      getIntFromMapWithDefault(confMap, "db", 0),
			MaxOpen: getIntFromMapWithDefault(confMap, "max_open", defaultPoolSize),
			MaxIdle: getIntFromMapWithDefault(confMap, "max_idle", defaultIdleSize),
		})
		mu.Unlock()

		if err != nil {
			panic(fmt.Sprintf("redis init failed. name: %s, error: %s.", name, err.Error()))
		}
	}
}

// 辅助函数：从 map 获取字符串
func getStringFromMap(m map[string]interface{}, key string) string {
	if val, ok := m[key]; ok {
		return fmt.Sprintf("%v", val)
	}
	return ""
}

// 辅助函数：从 map 获取整数（带默认值）
func getIntFromMapWithDefault(m map[string]interface{}, key string, defaultVal int) int {
	if val, ok := m[key]; ok {
		switch v := val.(type) {
		case int:
			return v
		case int64:
			return int(v)
		case float64:
			return int(v)
		}
	}
	return defaultVal
}
```

- [ ] **Step 2: 验证编译**

```bash
go build ./redis/...
```

预期：编译成功。

- [ ] **Step 3: 提交 redis 适配**

```bash
git add redis/redis.go
git commit -m "refactor: adapt redis to YAML config structure"
```

---

### Task 6: 适配 logger 模块

**Files:**
- Modify: `logger/logger.go`

- [ ] **Step 1: 修改 GetLevel 函数**

修改 `logger/logger.go` 第 26-32 行：

```go
func GetLevel() *zapcore.Level {
	l := new(zapcore.Level)
	mode := config.GetString("app.mode")
	if err := l.Set(mode); err != nil {
		_ = l.Set("info")
	}
	return l
}
```

- [ ] **Step 2: 修改 NewLogger 函数中的配置读取**

修改 `logger/logger.go` 第 79-85 行和第 94 行：

```go
func NewLogger(logName string, options ...zap.Option) *zap.Logger {
	// ... 前面代码保持不变

	fileName := fmt.Sprintf("%s/%s.log", logPath, logName)
	logSection := config.GetStringMap("log")
	fileWriters := []zapcore.WriteSyncer{zapcore.AddSync(&lumberjack.Logger{
		Filename:  fileName,
		MaxSize:   getIntFromMapWithDefault(logSection, "maxsize", defaultMaxSize),
		LocalTime: true,
		Compress:  getBoolFromMapWithDefault(logSection, "compress", true),
	})}

	// ... 中间代码保持不变

	// 文件日志格式
	switch getStringFromMap(logSection, "encode_type") {
	case "mis":
		cores = append(cores, zapcore.NewCore(NewMisEncoder(encoder), zap.CombineWriteSyncers(fileWriters...), logLevel))
	case "json":
		fallthrough
	default:
		cores = append(cores, zapcore.NewCore(zapcore.NewJSONEncoder(encoder), zap.CombineWriteSyncers(fileWriters...), logLevel))
	}

	// debug 日志输出至日志文件、标准输出
	if config.GetString("app.mode") == "debug" {
		// ... 后续代码保持不变
	}

	// ... 后续代码保持不变
}

// 辅助函数：从 map 获取字符串
func getStringFromMap(m map[string]interface{}, key string) string {
	if val, ok := m[key]; ok {
		return fmt.Sprintf("%v", val)
	}
	return ""
}

// 辅助函数：从 map 获取整数（带默认值）
func getIntFromMapWithDefault(m map[string]interface{}, key string, defaultVal int) int {
	if val, ok := m[key]; ok {
		switch v := val.(type) {
		case int:
			return v
		case int64:
			return int(v)
		case float64:
			return int(v)
		}
	}
	return defaultVal
}

// 辅助函数：从 map 获取布尔值（带默认值）
func getBoolFromMapWithDefault(m map[string]interface{}, key string, defaultVal bool) bool {
	if val, ok := m[key]; ok {
		if b, ok := val.(bool); ok {
			return b
		}
	}
	return defaultVal
}
```

- [ ] **Step 3: 验证编译**

```bash
go build ./logger/...
```

预期：编译成功。

- [ ] **Step 4: 提交 logger 适配**

```bash
git add logger/logger.go
git commit -m "refactor: adapt logger to YAML config structure"
```

---

### Task 7: 更新 README 文档

**Files:**
- Modify: `README.md`

- [ ] **Step 1: 更新 README.md 配置示例**

将 `README.md` 第 3-80 行的 INI 示例替换为 YAML 格式：

```markdown
# go-utils

## config.yaml 示例

```yaml
app:
  name: app-name
  mode: debug
  addr: :8080

log:
  encode_type: json
  path: ./logs
  maxsize: 1024
  compress: true

gorm:
  trace_sql: false
  slow_threshold: 1s
  prepare_stmt: true

database:
  test:
    master:
      drive: mysql
      host: ${DB_HOST:127.0.0.1}
      port: ${DB_PORT:3306}
      username: ${DB_USER:root}
      password: ${DB_PASS}
      db: test
      charset: utf8
      max_idle: 10
      max_open: 20
    slaves:
      - drive: mysql
        host: 127.0.0.1
        port: 3306
        username: root
        password: ""
        db: test2
        charset: utf8

redis:
  default:
    host: ${REDIS_HOST:127.0.0.1}
    port: ${REDIS_PORT:6379}
    auth: ${REDIS_AUTH}
    db: 0
    max_idle: 5
    max_open: 10
```

## .env 文件示例

```env
DB_HOST=192.168.1.100
DB_PORT=3307
REDIS_HOST=redis.internal
```

## 使用方式

```go
import "github.com/qkzsky/gutils/config"

func main() {
    // 初始化配置（自动加载 .env）
    config.SetDefault("config.yaml")

    // 获取配置值
    appName := config.GetString("app.name")
    dbHost := config.GetString("database.test.master.host")

    // 获取 section（兼容旧 API）
    appSection := config.Section("app")
}
```
```

- [ ] **Step 2: 提交文档更新**

```bash
git add README.md
git commit -m "docs: update README with YAML config examples"
```

---

### Task 8: 最终验证与提交

- [ ] **Step 1: 运行所有测试**

```bash
go test ./...
```

预期：所有测试通过。

- [ ] **Step 2: 验证完整编译**

```bash
go build ./...
```

预期：编译成功。

- [ ] **Step 3: 查看变更总结**

```bash
git log --oneline -8
git diff HEAD~7
```

预期：查看所有提交记录和变更内容。

---

## 自审清单

1. **Spec 覆盖检查**：
   - YAML 格式支持 ✓ Task 2
   - .env 文件加载 ✓ Task 2
   - ${ENV_VAR} 和 ${ENV_VAR:default} 语法 ✓ Task 2
   - config.yaml 默认文件名 ✓ Task 2
   - 移除 INI 支持 ✓ Task 3
   - database 适配 ✓ Task 4
   - redis 适配 ✓ Task 5
   - logger 适配 ✓ Task 6
   - README 更新 ✓ Task 7

2. **占位符检查**：无 TBD、TODO、模糊描述

3. **类型一致性检查**：
   - getStringFromMap、getIntFromMapWithDefault、getBoolFromMapWithDefault 在各模块中定义一致
   - config.GetString/GetInt/GetBool 等方法签名与 yaml.go 定义一致