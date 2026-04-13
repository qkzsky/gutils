# YAML 配置与环境变量支持设计

## 概述

将 gutils 的配置系统从 INI 格式迁移到 YAML 格式，并支持：
- `.env` 文件自动加载环境变量
- YAML 中使用 `${ENV_VAR}` 和 `${ENV_VAR:default}` 语法引用环境变量

## 需求

1. **配置格式**：仅支持 YAML，移除 INI 支持
2. **配置文件名**：`config.yaml` 作为默认配置文件
3. **环境变量语法**：`${ENV_VAR}` 或 `${ENV_VAR:default_value}`
4. **优先级**：YAML 值优先，环境变量仅在 YAML 中显式引用时生效
5. **.env 文件**：自动从工作目录加载，不存在则忽略
6. **依赖**：yaml.v3 + godotenv v1.5.1

## 配置文件结构

### YAML 格式示例

```yaml
app:
  name: app-name
  mode: debug
  addr: ":8080"

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
        host: ${DB_SLAVE1_HOST:127.0.0.1}
        port: 3306
        username: root
        password: ""
        db: test2
        charset: utf8
      - drive: mysql
        host: ${DB_SLAVE2_HOST:127.0.0.1}
        port: 3306
        username: root
        password: ""
        db: test3
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

### 环境变量替换逻辑

1. `SetDefault()` 调用时自动加载 `.env` 文件（工作目录）
2. 读取 YAML 文件原始内容
3. 使用 `os.Expand` 替换 `${ENV_VAR}` 和 `${ENV_VAR:default}` 格式
4. 将替换后的内容解析为 YAML 结构
5. 环境变量未设置且无默认值时，替换为空字符串

## API 设计

### 保持兼容的 API

```go
// 初始化配置（文件名改为 config.yaml）
config.SetDefault("config.yaml")

// 获取 section（返回 map[string]interface{}）
config.Section("database")  // 返回 database 节点下的所有配置

// 获取单个 key（从 app section 获取）
config.Key("name")  // 返回 app.name 的值
```

### 新增 API

```go
// 获取嵌套配置值（使用点号路径）
config.GetString("database.test.master.host")
config.GetInt("redis.default.db")
config.GetBool("app.pprof")

// 获取配置值（带默认值）
config.GetStringWithDefault("database.test.master.host", "localhost")
config.GetIntWithDefault("redis.default.db", 0)

// 获取 map 类型配置
config.GetStringMap("database")      // 返回 database 下所有配置
config.GetStringMap("redis.default") // 返回 redis.default 下所有配置

// 获取数组类型配置
config.GetSlice("database.test.slaves") // 返回 slaves 数组
```

### 配置结构类型

```go
var (
    AppPath string  // 应用路径
    AppName string  // 应用名称
    AppMode string  // 运行模式：debug/release

    defaultConf map[string]interface{}  // YAML 配置根节点
)
```

## 实现细节

### config/yaml.go 核心函数

```go
// SetDefault 初始化配置
func SetDefault(file string) {
    // 1. 加载 .env 文件
    godotenv.Load()
    
    // 2. 读取 YAML 文件
    content := readFile(file)
    
    // 3. 替换环境变量
    expanded := expandEnv(content)
    
    // 4. 解析 YAML
    yaml.Unmarshal([]byte(expanded), &defaultConf)
    
    // 5. 设置全局变量
    AppPath = filepath.Abs(...)
    AppName = GetString("app.name")
    AppMode = GetStringWithDefault("app.mode", "release")
}

// expandEnv 替换环境变量
func expandEnv(content string) string {
    return os.Expand(content, func(key string) string {
        // 解析 KEY:default 格式
        if parts := strings.SplitN(key, ":", 2); len(parts) == 2 {
            val := os.Getenv(parts[0])
            if val == "" {
                return parts[1]
            }
            return val
        }
        return os.Getenv(key)
    })
}
```

### database/database.go 适配

从平铺 section 改为嵌套 map 读取：

```go
func InitDb() {
    // 获取 database 配置节点
    dbConf := config.GetStringMap("database")
    
    for dbName, conf := range dbConf {
        // 解析 master 和 slaves
        master := parseDbConfig(conf["master"])
        slaves := parseSlaveConfigs(conf["slaves"])
        
        // 创建连接
        dbMap[dbName] = makeDB(master, slaves)
    }
}
```

### redis/redis.go 适配

```go
func InitRedis() {
    redisConf := config.GetStringMap("redis")
    
    for name, conf := range redisConf {
        redisMap[name] = NewRedis(redisConfig{
            Host:    config.GetString(fmt.Sprintf("redis.%s.host", name)),
            Port:    config.GetString(fmt.Sprintf("redis.%s.port", name)),
            Auth:    config.GetString(fmt.Sprintf("redis.%s.auth", name)),
            DB:      config.GetInt(fmt.Sprintf("redis.%s.db", name)),
            MaxOpen: config.GetIntWithDefault(fmt.Sprintf("redis.%s.max_open", name), defaultPoolSize),
            MaxIdle: config.GetIntWithDefault(fmt.Sprintf("redis.%s.max_idle", name), defaultIdleSize),
        })
    }
}
```

## 代码变更清单

| 文件 | 操作 | 说明 |
|------|------|------|
| `config/ini.go` | 删除 | 移除 INI 支持 |
| `config/yaml.go` | 新建 | YAML 配置解析 + 环境变量替换 |
| `go.mod` | 修改 | 升级 yaml v3，新增 godotenv |
| `database/database.go` | 修改 | 适配 YAML 嵌套结构 |
| `redis/redis.go` | 修改 | 适配 YAML 嵌套结构 |
| `logger/logger.go` | 修改 | 适配 YAML 嵌套结构 |
| `README.md` | 修改 | 更新配置示例为 YAML 格式 |

## 测试计划

1. 测试 YAML 解析基本功能
2. 测试环境变量替换（有默认值、无默认值、未设置）
3. 测试 `.env` 文件加载
4. 测试 database 主从配置解析
5. 测试 redis 多实例配置解析
6. 测试 logger 配置解析
7. 测试 API 兼容性