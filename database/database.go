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
	"strings"
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
	return l[i].isMaster == true
}
func (l dbConfigList) Swap(i, j int) {
	l[i], l[j] = l[j], l[i]
}

func InitDb() {
	mapConf := map[string]dbConfigList{}
	for _, section := range config.Section("database").ChildSections() {
		//var err error
		var dbName string
		isMaster := false
		secName := strings.TrimPrefix(section.Name(), "database.")
		if dotIndex := strings.Index(secName, "."); dotIndex != -1 {
			dbName = secName[:dotIndex]
			if secName[dotIndex+1:] == "master" {
				isMaster = true
			}
		} else {
			dbName = secName
		}

		// 判断是否支持keycenter
		var sid = section.Key("sid").String()
		if sid != "" {
			var oldPassword = section.Key("password").String()
			section.Key("password").SetValue(string(keycenter.DecryptSimple(sid, oldPassword)))
		}

		c := &dbConfig{
			Drive:    section.Key("drive").String(),
			Host:     section.Key("host").String(),
			Port:     section.Key("port").String(),
			Username: section.Key("username").String(),
			Password: section.Key("password").String(),
			DBName:   section.Key("db").String(),
			SSlMode:  section.Key("ssl_mode").MustString(DefaultSSLMode),
			Charset:  section.Key("charset").MustString(DefaultCharset),
			MaxIdle:  section.Key("max_idle").MustInt(defaultMaxIdle),
			MaxOpen:  section.Key("max_open").MustInt(defaultMaxOpen),
			isMaster: isMaster,
		}
		mapConf[dbName] = append(mapConf[dbName], c)
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

func makeDB(cs dbConfigList) (DB *gorm.DB, err error) {
	gormSC := config.Section("gorm")
	var gormConfig = &gorm.Config{
		SkipDefaultTransaction: true,
		PrepareStmt:            gormSC.Key("prepare_stmt").MustBool(true),
		Logger: &gLogger{
			Logger:        logger.GetDefaultLogger().WithOptions(zap.AddCallerSkip(1)),
			TraceSQL:      gormSC.Key("trace_sql").MustBool(false),
			SlowThreshold: gormSC.Key("slow_threshold").MustDuration(1 * time.Second),
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
