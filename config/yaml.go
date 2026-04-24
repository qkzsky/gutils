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
	// 首先尝试加载配置文件同目录下的 .env，然后尝试当前目录
	configDir := filepath.Dir(file)
	_ = godotenv.Load(filepath.Join(configDir, ".env"))
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
	if !ok || val == nil {
		return ""
	}
	return fmt.Sprintf("%v", val)
}

// GetStringWithDefault 通过路径获取字符串配置值，带默认值
func GetStringWithDefault(path string, defaultValue string) string {
	val, ok := getValueByPath(path)
	if !ok || val == nil {
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