# go-utils

## app.ini 示例
```ini
[app]
name = app-name

mode = debug
;mode = release
addr = :8080

pprof = false
pprof.token = ""

session = false

[session]
addr = "127.0.0.1:6379"
auth = ""
; minute
expire = 30

[log]
; 文件日志格式：json|mis default: json
encode_type = json
path = ./logs
maxsize = 1024
; 压缩备份？
compress = true

[xorm]
; debug/info/warn/err/off
log_level = info
show_sql = true

[database.test.master]
drive = mysql
host = 127.0.0.1
port = 3306
username = root
password =
db = test
charset = utf8

[database.test.slave1]
drive = mysql
host = 127.0.0.1
port = 3306
username = root
password =
db = test2
charset = utf8

[database.test.slave2]
drive = mysql
host = 127.0.0.1
port = 3306
username = root
password =
db = test3
charset = utf8

[database.xxx]
drive = mysql
host = 127.0.0.1
port = 3306
username = root
password =
db = test
charset = utf8

[redis.default]
host = 127.0.0.1
port = 6379
auth =


```