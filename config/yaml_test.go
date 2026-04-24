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

func TestGetStringNilValue(t *testing.T) {
	testdataDir := filepath.Join(".", "testdata")
	configFile := filepath.Join(testdataDir, "config.yaml")

	os.Clearenv()
	SetDefault(configFile)

	// 测试显式 null 值
	val := GetString("nil_test.value")
	if val != "" {
		t.Errorf("GetString for null value expected '', got '%s'", val)
	}

	// 测试空值（YAML 中也是 nil）
	val = GetString("nil_test.empty_key")
	if val != "" {
		t.Errorf("GetString for empty key expected '', got '%s'", val)
	}
}

func TestGetStringWithDefaultNilValue(t *testing.T) {
	testdataDir := filepath.Join(".", "testdata")
	configFile := filepath.Join(testdataDir, "config.yaml")

	os.Clearenv()
	SetDefault(configFile)

	// 测试显式 null 值应返回默认值
	val := GetStringWithDefault("nil_test.value", "fallback")
	if val != "fallback" {
		t.Errorf("GetStringWithDefault for null value expected 'fallback', got '%s'", val)
	}

	// 测试空值应返回默认值
	val = GetStringWithDefault("nil_test.empty_key", "fallback")
	if val != "fallback" {
		t.Errorf("GetStringWithDefault for empty key expected 'fallback', got '%s'", val)
	}
}

func TestGetValueByPathWithNil(t *testing.T) {
	testdataDir := filepath.Join(".", "testdata")
	configFile := filepath.Join(testdataDir, "config.yaml")

	os.Clearenv()
	SetDefault(configFile)

	// getValueByPath 应对 nil 值返回 true（key 存在）
	val, ok := getValueByPath("nil_test.value")
	if !ok {
		t.Errorf("getValueByPath for null value should return ok=true")
	}
	if val != nil {
		t.Errorf("getValueByPath for null value expected nil, got '%v'", val)
	}
}