# go-utils

## config.yaml 示例

```yaml
app:
  name: app-name
  mode: debug
  addr: :8080

log:
  encode_type: json
  stdout_encode: json  # json/console/none, 默认 none
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