package database

import (
	"database/sql"
	"fmt"
	"github.com/qkzsky/gutils/config"
	"github.com/qkzsky/gutils/logger"
	"go.uber.org/zap"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/plugin/dbresolver"
	"runtime"
	"sort"
	"time"
)

const DefaultCharset = "utf8"
const DefaultSSLMode = "disable"

var defaultConnMaxIdleTime = time.Hour
var defaultConnMaxLifetime = 2 * time.Hour
var defaultMaxIdle = runtime.NumCPU() + 1
var defaultMaxOpen = runtime.NumCPU()*2 + 1

var (
	dbMap = map[string]*gorm.DB{}
)

type dbConfig struct {
	Drive    string
	Host     string
	File     string
	Port     string
	Username string
	Password string
	DBName   string
	SSlMode  string
	Charset  string
	MaxOpen  int
	MaxIdle  int

	isMaster bool
}

type dbConfigList []*dbConfig

func (l dbConfigList) Len() int {
	return len(l)
}
func (l dbConfigList) Less(i, j int) bool {
	return l[i].isMaster
}
func (l dbConfigList) Swap(i, j int) {
	l[i], l[j] = l[j], l[i]
}

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

	DB, err = gorm.Open(NewDialector(cs[0]), gormConfig)
	if err != nil {
		return
	}
	if len(cs) > 1 {
		err = DB.Use(
			dbresolver.Register(dbresolver.Config{
				Replicas: func(cs dbConfigList) (replicas []gorm.Dialector) {
					for _, c := range cs {
						replicas = append(replicas, NewDialector(c))
					}
					return
				}(cs[1:]),
				//Policy:   dbresolver.RandomPolicy{},
				Policy: dbresolver.RandomPolicy{},
			}).
				// 连接及连接池配置
				SetConnMaxIdleTime(defaultConnMaxIdleTime).
				SetConnMaxLifetime(defaultConnMaxLifetime).
				SetMaxIdleConns(cs[0].MaxIdle).
				SetMaxOpenConns(cs[0].MaxOpen),
		)
		if err != nil {
			return
		}
	} else {
		var db *sql.DB
		db, err = DB.DB()
		if err != nil {
			return
		}
		db.SetConnMaxIdleTime(defaultConnMaxIdleTime)
		db.SetConnMaxLifetime(defaultConnMaxLifetime)
		db.SetMaxIdleConns(cs[0].MaxIdle)
		db.SetMaxOpenConns(cs[0].MaxOpen)
	}

	return DB, err
}

func GetDB(name string) *gorm.DB {
	if client, ok := dbMap[name]; ok {
		return client
	}

	panic("db not found: " + name)
}

func NewDialector(c *dbConfig) gorm.Dialector {
	var dbDSN string
	switch c.Drive {
	case "mysql":
		dbDSN = fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=%s&parseTime=true&loc=Local&timeout=15s",
			c.Username, c.Password, c.Host, c.Port, c.DBName, c.Charset)
		return mysql.Open(dbDSN)
	case "postgres":
		dbDSN = fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
			c.Host, c.Port, c.Username, c.Password, c.DBName, c.SSlMode)
		return postgres.New(postgres.Config{
			DSN: dbDSN,
			//PreferSimpleProtocol: true, // disables implicit prepared statement usage
		})
	//case "sqlserver":
	//	dbDSN = fmt.Sprintf("sqlserver://%s:%s@%s:%s?database=%s",
	//		c.Username, c.Password, c.Host, c.Port, c.DBName)
	//	db, err = sqlserver.Open(dbDSN), Config
	//case "sqlite":
	//	db, err = sqlite.Open(filepath.Join(os.TempDir(), "gorm.db"))
	default:
		logger.Error(fmt.Sprintf("unknown database drive: %s", c.Drive))
	}

	return nil
}

func getStringFromMap(m map[string]interface{}, key string) string {
	if val, ok := m[key]; ok {
		return fmt.Sprintf("%v", val)
	}
	return ""
}

func getStringFromMapWithDefault(m map[string]interface{}, key string, defaultVal string) string {
	if val, ok := m[key]; ok {
		return fmt.Sprintf("%v", val)
	}
	return defaultVal
}

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

func getBoolFromMapWithDefault(m map[string]interface{}, key string, defaultVal bool) bool {
	if val, ok := m[key]; ok {
		if b, ok := val.(bool); ok {
			return b
		}
	}
	return defaultVal
}

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
